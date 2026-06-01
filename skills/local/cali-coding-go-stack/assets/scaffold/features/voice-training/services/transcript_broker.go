package services

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
)

type TranscriptEvent struct {
	Speaker        string `json:"speaker"`
	Text           string `json:"text"`
	IsTurnComplete bool   `json:"isTurnComplete"`
}

type NatsTranscriptBroker struct {
	nc *nats.Conn
}

var transcriptBroker *NatsTranscriptBroker

func InitBroker(nc *nats.Conn) {
	transcriptBroker = &NatsTranscriptBroker{nc: nc}
}

func Broker() *NatsTranscriptBroker {
	return transcriptBroker
}

func (b *NatsTranscriptBroker) Subject(roomName string) string {
	return fmt.Sprintf("voice.training.transcripts.%s", roomName)
}

func (b *NatsTranscriptBroker) Subscribe(roomName string) (chan *TranscriptEvent, *nats.Subscription, error) {
	ch := make(chan *TranscriptEvent, 64)
	sub, err := b.nc.Subscribe(b.Subject(roomName), func(msg *nats.Msg) {
		var event TranscriptEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			slog.Error("failed to unmarshal transcript event", "error", err)
			return
		}
		ch <- &event
	})
	if err != nil {
		return nil, nil, err
	}
	return ch, sub, nil
}

func (b *NatsTranscriptBroker) Publish(roomName, speaker, text string, isTurnComplete bool) {
	event := TranscriptEvent{
		Speaker:        speaker,
		Text:           text,
		IsTurnComplete: isTurnComplete,
	}

	jsonBytes, err := json.Marshal(event)
	if err != nil {
		slog.Error("failed to marshal transcript event", "error", err)
		return
	}

	if err := b.nc.Publish(b.Subject(roomName), jsonBytes); err != nil {
		slog.Error("failed to publish transcript to NATS", "error", err)
	}
}
