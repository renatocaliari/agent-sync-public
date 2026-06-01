package services

import (
	"context"
	"log/slog"

	"github.com/livekit/media-sdk"
	"github.com/livekit/protocol/logger"
	lksdk "github.com/livekit/server-sdk-go/v2"
	lkmedia "github.com/livekit/server-sdk-go/v2/pkg/media"
	"github.com/pion/webrtc/v4"
)

type BotService struct {
	liveKitService *LiveKitService
	geminiService  *GeminiService
	room           *lksdk.Room
	audioTrack     *lkmedia.PCMLocalTrack
}

func NewBotService(liveKit *LiveKitService, gemini *GeminiService) *BotService {
	return &BotService{
		liveKitService: liveKit,
		geminiService:  gemini,
	}
}

func (b *BotService) JoinRoom(ctx context.Context, roomName, systemPrompt, voiceTone string) error {
	slog.Info("joining room", "room", roomName, "voiceTone", voiceTone)

	b.liveKitService.currentRoomName = roomName

	err := b.geminiService.Connect(ctx, systemPrompt, voiceTone)
	if err != nil {
		slog.Error("failed to connect to gemini", "error", err)
		return err
	}
	slog.Info("gemini connected")

	b.geminiService.SetOnAudioOut(func(audio []byte) {
		b.publishAudio(audio)
	})

	b.geminiService.SetOnTranscript(func(msg TranscriptMessage) {
		slog.Info("received transcript", "speaker", msg.Speaker, "text", msg.Text, "isTurnComplete", msg.IsTurnComplete)
		transcriptBroker.Publish(b.liveKitService.currentRoomName, msg.Speaker, msg.Text, msg.IsTurnComplete)
	})

	b.geminiService.SetOnInterruption(func() {
		slog.Info("interruption detected - user started speaking. Flushing audio queue.")
		if b.audioTrack != nil {
			// Clear the audio track queue to stop model speech immediately
			b.audioTrack.ClearQueue()
		}
	})

	room, err := lksdk.ConnectToRoom(b.liveKitService.Host(), lksdk.ConnectInfo{
		APIKey:              b.liveKitService.apiKey,
		APISecret:           b.liveKitService.apiSecret,
		RoomName:            roomName,
		ParticipantIdentity: "client-bot",
	}, &lksdk.RoomCallback{
		ParticipantCallback: lksdk.ParticipantCallback{
			OnTrackSubscribed: b.onTrackSubscribed,
		},
	})
	if err != nil {
		slog.Error("failed to connect to room", "error", err)
		return err
	}

	b.room = room

	// Create audio track for bot responses
	// IMPORTANT: Use 24kHz to match Gemini output - LiveKit SDK will resample to 48kHz automatically
	b.audioTrack, err = lkmedia.NewPCMLocalTrack(24000, 1, logger.GetLogger())
	if err != nil {
		slog.Error("failed to create PCM local track", "error", err)
		return err
	}

	_, err = b.room.LocalParticipant.PublishTrack(b.audioTrack, &lksdk.TrackPublicationOptions{
		Name: "client-audio",
	})
	if err != nil {
		slog.Error("failed to publish audio track", "error", err)
		return err
	}

	slog.Info("bot joined room and audio track published", "room", roomName)
	return nil
}

func (b *BotService) onTrackSubscribed(track *webrtc.TrackRemote, publication *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {
	slog.Info("[DEBUG] bot: track subscribed", "kind", track.Kind(), "participant", rp.Identity(), "trackID", track.ID())

	if track.Kind() == webrtc.RTPCodecTypeAudio {
		slog.Info("[DEBUG] bot: processing audio track", "participant", rp.Identity())
		go b.processAudioTrack(track, rp.Identity())
	}
}

func (b *BotService) processAudioTrack(track *webrtc.TrackRemote, participantIdentity string) {
	slog.Info("[DEBUG] bot: creating PCM remote track", "participant", participantIdentity)

	writer := &pcmWriter{
		geminiService: b.geminiService,
		participant:   participantIdentity,
	}

	pcmTrack, err := lkmedia.NewPCMRemoteTrack(track, writer, lkmedia.WithTargetSampleRate(16000), lkmedia.WithTargetChannels(1))
	if err != nil {
		slog.Error("[DEBUG] bot: failed to create PCM remote track", "error", err, "participant", participantIdentity)
		return
	}
	defer pcmTrack.Close()

	slog.Info("[DEBUG] bot: PCM remote track created, receiving audio...", "participant", participantIdentity)

	select {}
}

type pcmWriter struct {
	geminiService *GeminiService
	participant   string
	sampleCount   int
}

func (w *pcmWriter) WriteSample(sample media.PCM16Sample) error {
	w.sampleCount++
	if w.sampleCount <= 10 || w.sampleCount%100 == 0 {
		slog.Info("received PCM samples from therapist", "count", w.sampleCount, "len", len(sample), "participant", w.participant)
	}

	buf := make([]byte, len(sample)*2)
	for i, s := range sample {
		buf[i*2] = byte(s & 0xff)
		buf[i*2+1] = byte(s >> 8)
	}

	if w.sampleCount <= 10 || w.sampleCount%100 == 0 {
		slog.Info("sending audio to gemini", "count", w.sampleCount, "bufSize", len(buf))
	}

	if err := w.geminiService.SendAudio(buf); err != nil {
		if w.sampleCount <= 10 || w.sampleCount%100 == 0 {
			slog.Error("failed to send audio to gemini", "error", err, "count", w.sampleCount)
		}
		return err
	}

	if w.sampleCount <= 10 || w.sampleCount%100 == 0 {
		slog.Info("audio sent to gemini successfully", "count", w.sampleCount)
	}

	return nil
}

func (w *pcmWriter) Close() error {
	slog.Info("PCM writer closed", "totalSamples", w.sampleCount)
	return nil
}

func (b *BotService) publishAudio(audioData []byte) {
	if b.room == nil || b.audioTrack == nil {
		slog.Warn("room or audioTrack is nil, cannot publish audio")
		return
	}

	slog.Info("publishing audio from gemini", "size", len(audioData), "trackRate", "24kHz")

	samples := make(media.PCM16Sample, len(audioData)/2)
	for i := range samples {
		samples[i] = int16(audioData[i*2]) | int16(audioData[i*2+1])<<8
	}

	slog.Info("writing samples to LiveKit track", "samples", len(samples), "expectedSampleRate", "24kHz -> 48kHz")

	if err := b.audioTrack.WriteSample(samples); err != nil {
		slog.Error("failed to write sample", "error", err)
	} else {
		slog.Info("audio published successfully", "samples", len(samples))
	}
}

func (b *BotService) Leave() {
	if b.audioTrack != nil {
		b.audioTrack.Close()
	}
	if b.room != nil {
		b.room.Disconnect()
	}
	b.geminiService.Close()
}
