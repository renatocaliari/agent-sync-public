package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type TranscriptMessage struct {
	Speaker        string
	Text           string
	IsTurnComplete bool
}

type GeminiService struct {
	apiKey         string
	conn           *websocket.Conn
	mu             sync.Mutex
	onAudioOut     func([]byte)
	onTranscript   func(TranscriptMessage)
	onInterruption func()
}

const geminiWSURL = "wss://generativelanguage.googleapis.com/ws/google.ai.generativelanguage.v1alpha.GenerativeService.BidiGenerateContent"

func NewGeminiService(apiKey string) *GeminiService {
	return &GeminiService{apiKey: apiKey}
}

func (s *GeminiService) SetOnInterruption(f func()) {
	s.onInterruption = f
}

func (s *GeminiService) Connect(ctx context.Context, systemPrompt, voiceTone string) error {
	url := geminiWSURL + "?key=" + s.apiKey

	slog.Info("connecting to gemini", "url", geminiWSURL, "voiceTone", voiceTone)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		slog.Error("failed to connect to gemini", "error", err)
		return err
	}
	s.conn = conn
	slog.Info("connected to gemini websocket")

	voiceName := "Charon" // Male voice
	if voiceTone == "female" {
		voiceName = "Aoede" // Female voice
	}

	// Add strong instructions to prevent hallucination of Chinese and enforce Portuguese
	sttInstruction := "\n\nCRITICAL INSTRUCTION: The user will speak exclusively in Portuguese (pt-BR). You must transcribe their speech in Portuguese. Ignore background noise and silence. Never transcribe as Chinese or any other language."
	fullSystemPrompt := systemPrompt + sttInstruction

	setupMsg := map[string]interface{}{
		"setup": map[string]interface{}{
			"model": "models/gemini-2.5-flash-native-audio-preview-12-2025",
			"generationConfig": map[string]interface{}{
				"responseModalities": []string{"AUDIO"},
				"speechConfig": map[string]interface{}{
					"voiceConfig": map[string]interface{}{
						"prebuiltVoiceConfig": map[string]string{
							"voiceName": voiceName,
						},
					},
				},
			},
			"inputAudioTranscription": map[string]interface{}{
				"enabled": true,
			},
			"outputAudioTranscription": map[string]interface{}{
				"enabled": true,
			},
			"systemInstruction": map[string]interface{}{
				"parts": []map[string]string{
					{"text": fullSystemPrompt},
				},
			},
		},
	}

	slog.Info("sending setup message to gemini")
	if err := s.conn.WriteJSON(setupMsg); err != nil {
		slog.Error("failed to send setup", "error", err)
		return err
	}

	go s.readResponses(ctx)

	return nil
}

var audioChunkCount int

func (s *GeminiService) SendAudio(audioData []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn == nil {
		return nil
	}

	audioChunkCount++
	if audioChunkCount%100 == 0 {
		slog.Info("[DEBUG] bot: audio chunks sent to gemini", "total", audioChunkCount)
	}

	msg := map[string]interface{}{
		"realtimeInput": map[string]interface{}{
			"mediaChunks": []map[string]interface{}{
				{
					"mimeType": "audio/pcm;rate=16000",
					"data":     base64.StdEncoding.EncodeToString(audioData),
				},
			},
		},
	}

	if err := s.conn.WriteJSON(msg); err != nil {
		slog.Error("WriteJSON to gemini failed", "error", err)
		return err
	}

	return nil
}

func (s *GeminiService) SetOnAudioOut(callback func([]byte)) {
	s.onAudioOut = callback
}

func (s *GeminiService) SetOnTranscript(callback func(TranscriptMessage)) {
	s.onTranscript = callback
}

func (s *GeminiService) readResponses(ctx context.Context) {
	slog.Info("starting to read gemini responses")
	var clientTranscript strings.Builder
	var userTranscript strings.Builder
	var currentSpeaker string // "user", "client", or ""

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if s.conn == nil {
				return
			}

			_, message, err := s.conn.ReadMessage()
			if err != nil {
				slog.Error("gemini read error", "error", err)
				return
			}

			var response map[string]interface{}
			if err := json.Unmarshal(message, &response); err != nil {
				continue
			}

			// Global turnComplete flag
			isModelTurnComplete := false
			if _, ok := response["turnComplete"]; ok {
				isModelTurnComplete = true
			}

			if serverContent, ok := response["serverContent"].(map[string]interface{}); ok {
				if _, ok := serverContent["turnComplete"]; ok {
					isModelTurnComplete = true
				}

				// 1. Interruption
				if interrupted, ok := serverContent["interrupted"].(bool); ok && interrupted {
					slog.Info("[DEBUG] interruption detected")
					s.finalizeTranscript(&clientTranscript, "client")
					currentSpeaker = ""
				}

				// 2. User Speech (Input)
				if inputTranscription, ok := serverContent["inputTranscription"].(map[string]interface{}); ok {
					if text, ok := inputTranscription["text"].(string); ok && text != "" {
						slog.Info("[DEBUG] gemini input transcript", "text", text)

						// If model was speaking, it means a new user turn started
						if currentSpeaker == "client" {
							s.finalizeTranscript(&clientTranscript, "client")
						}
						currentSpeaker = "user"

						userTranscript.WriteString(text)
						if s.onTranscript != nil {
							s.onTranscript(TranscriptMessage{
								Speaker:        "user",
								Text:           userTranscript.String(),
								IsTurnComplete: false,
							})
						}
					}
				}

				// 3. Model Speech (Output)
				if outputTranscription, ok := serverContent["outputTranscription"].(map[string]interface{}); ok {
					if text, ok := outputTranscription["text"].(string); ok && text != "" {
						slog.Info("[DEBUG] gemini output transcript", "text", text)

						// Transition from user to client
						if currentSpeaker == "user" {
							s.finalizeTranscript(&userTranscript, "user")
						}
						currentSpeaker = "client"

						clientTranscript.WriteString(text)
						if s.onTranscript != nil {
							s.onTranscript(TranscriptMessage{
								Speaker:        "client",
								Text:           clientTranscript.String(),
								IsTurnComplete: false,
							})
						}
					}
				}

				// 4. Model Audio
				if modelTurn, ok := serverContent["modelTurn"].(map[string]interface{}); ok {
					if parts, ok := modelTurn["parts"].([]interface{}); ok {
						for _, part := range parts {
							if p, ok := part.(map[string]interface{}); ok {
								if inlineData, ok := p["inlineData"].(map[string]interface{}); ok {
									if data, ok := inlineData["data"].(string); ok {
										if currentSpeaker == "user" {
											s.finalizeTranscript(&userTranscript, "user")
										}
										currentSpeaker = "client"

										if s.onAudioOut != nil {
											audio, _ := base64.StdEncoding.DecodeString(data)
											s.onAudioOut(audio)
										}
									}
								}
							}
						}
					}
				}
			}

			// 5. Finalize Model turn
			if isModelTurnComplete {
				slog.Info("[DEBUG] model turn complete signal")
				s.finalizeTranscript(&clientTranscript, "client")
				if currentSpeaker == "client" {
					currentSpeaker = ""
				}
			}
		}
	}
}

func (s *GeminiService) finalizeTranscript(sb *strings.Builder, speaker string) {
	if sb.Len() > 0 && s.onTranscript != nil {
		text := sb.String()
		slog.Info("finalizing turn", "speaker", speaker, "text", text)
		s.onTranscript(TranscriptMessage{
			Speaker:        speaker,
			Text:           text,
			IsTurnComplete: true,
		})
		sb.Reset()
	}
}

func (s *GeminiService) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
