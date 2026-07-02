package services

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/nats-io/nats.go"
)

type Session struct {
	ID           string `json:"id"`
	RoomName     string `json:"room_name"`
	TherapistID  string `json:"therapist_id"`
	ClientName   string `json:"client_name"`
	VoiceTone    string `json:"voice_tone"`
	SystemPrompt string `json:"system_prompt"`
	CreatedAt    string `json:"created_at"`
}

type SessionService struct {
	nc *nats.Conn
	kv nats.KeyValue
}

func NewSessionService(nc *nats.Conn) (*SessionService, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}

	kv, err := js.CreateKeyValue(&nats.KeyValueConfig{
		Bucket: "voice_training_sessions",
	})
	if err != nil {
		slog.Warn("kv bucket might already exist", "error", err)
		kv, err = js.KeyValue("voice_training_sessions")
		if err != nil {
			return nil, err
		}
	}

	return &SessionService{nc: nc, kv: kv}, nil
}

func (s *SessionService) SaveSession(ctx context.Context, session *Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	_, err = s.kv.Put(session.ID, data)
	return err
}

func (s *SessionService) GetSession(ctx context.Context, id string) (*Session, error) {
	entry, err := s.kv.Get(id)
	if err != nil {
		return nil, err
	}

	var session Session
	if err := json.Unmarshal(entry.Value(), &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *SessionService) DeleteSession(ctx context.Context, id string) error {
	return s.kv.Delete(id)
}

func DefaultSystemPrompt(clientName, voiceTone string) string {
	return `Você é ` + clientName + `, um cliente simulado para treinamento de terapeutas.

Você está em uma sessão de terapia. Responda de forma natural, demonstrando emoções apropriadas ao contexto da conversa.

Características:
- Mostre ansiedade leve e hesitação ao falar sobre problemas
- Seja receptivo às intervenções do terapeuta
- Expresse emoções genuínas (tristeza, esperança, frustração)
- Responda de forma concisa, como em uma conversa real
- Fale sempre em português brasileiro

Mantenha consistência na sua história e emoções ao longo da sessão.`
}
