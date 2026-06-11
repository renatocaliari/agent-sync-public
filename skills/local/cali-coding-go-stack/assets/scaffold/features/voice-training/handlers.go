package voicetraining

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/delaneyj/toolbelt/id"
	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
	"northstar/features/voice-training/pages"
	"northstar/features/voice-training/services"
)

type Handlers struct {
	liveKitService *services.LiveKitService
	sessionService *services.SessionService
	geminiAPIKey   string
}

func NewHandlers(liveKit *services.LiveKitService, session *services.SessionService, geminiAPIKey string) *Handlers {
	return &Handlers{
		liveKitService: liveKit,
		sessionService: session,
		geminiAPIKey:   geminiAPIKey,
	}
}

type CreateRoomRequest struct {
	ClientName string `json:"clientName"`
}

type CreateRoomResponse struct {
	RoomName   string `json:"roomName"`
	Token      string `json:"token"`
	LiveKitURL string `json:"livekitUrl"`
}

func (h *Handlers) LandingPage(w http.ResponseWriter, r *http.Request) {
	if err := pages.LandingPage().Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handlers) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req CreateRoomRequest
	json.NewDecoder(r.Body).Decode(&req)

	var clientProfile pages.ClientProfile
	if req.ClientName != "" {
		clientProfile = pages.ClientProfile{Name: req.ClientName, VoiceTone: "female"}
	} else {
		clientProfile = pages.RandomClientProfile()
	}

	roomName := "training-" + id.NextEncodedID()

	ctx := r.Context()
	slog.Info("creating room", "room", roomName, "client", clientProfile.Name, "tone", clientProfile.VoiceTone)

	room, err := h.liveKitService.CreateRoom(ctx, roomName)
	if err != nil {
		slog.Error("failed to create room", "error", err)
		http.Error(w, "failed to create room", http.StatusInternalServerError)
		return
	}

	token, err := h.liveKitService.GenerateToken(roomName, "therapist-"+id.NextEncodedID())
	if err != nil {
		slog.Error("failed to generate token", "error", err)
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	session := &services.Session{
		ID:           roomName,
		RoomName:     roomName,
		ClientName:   clientProfile.Name,
		VoiceTone:    clientProfile.VoiceTone,
		SystemPrompt: services.DefaultSystemPrompt(clientProfile.Name, clientProfile.VoiceTone),
		CreatedAt:    time.Now().Format(time.RFC3339),
	}

	if err := h.sessionService.SaveSession(ctx, session); err != nil {
		slog.Error("failed to save session", "error", err)
	}

	slog.Info("room created", "room", roomName, "client", clientProfile.Name, "tone", clientProfile.VoiceTone)

	resp := CreateRoomResponse{
		RoomName:   room.Name,
		Token:      token,
		LiveKitURL: h.liveKitService.Host(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handlers) RoomPage(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "id")

	session, err := h.sessionService.GetSession(r.Context(), roomName)
	if err != nil {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	token, err := h.liveKitService.GenerateToken(roomName, "therapist-"+id.NextEncodedID())
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	if err := pages.RoomPage(session, token, h.liveKitService.Host()).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handlers) EndSession(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "id")

	if err := h.liveKitService.DeleteRoom(r.Context(), roomName); err != nil {
		slog.Error("failed to delete room", "error", err)
	}

	if err := h.sessionService.DeleteSession(r.Context(), roomName); err != nil {
		slog.Error("failed to delete session", "error", err)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ended"})
}

func (h *Handlers) JoinBot(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "id")
	slog.Info("[DEBUG] JoinBot requested", "room", roomName)

	session, err := h.sessionService.GetSession(r.Context(), roomName)
	if err != nil {
		slog.Error("[DEBUG] JoinBot: session not found", "room", roomName, "error", err)
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	slog.Info("[DEBUG] Session found for JoinBot", "room", roomName, "client", session.ClientName)

	geminiService := services.NewGeminiService(h.geminiAPIKey)
	botService := services.NewBotService(h.liveKitService, geminiService)

	go func() {
		slog.Info("[DEBUG] Starting background JoinRoom for bot", "room", roomName)
		if err := botService.JoinRoom(context.Background(), roomName, session.SystemPrompt, session.VoiceTone); err != nil {
			slog.Error("[DEBUG] Bot failed to join room", "room", roomName, "error", err)
		} else {
			slog.Info("[DEBUG] Bot JoinRoom process completed", "room", roomName)
		}
	}()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "bot_joining"})
}

func (h *Handlers) TranscriptsStream(w http.ResponseWriter, r *http.Request) {
	roomName := chi.URLParam(r, "id")
	slog.Info("[DEBUG] TranscriptsStream connection requested", "room", roomName)

	session, err := h.sessionService.GetSession(r.Context(), roomName)
	if err != nil {
		slog.Error("[DEBUG] TranscriptsStream: session not found", "room", roomName, "error", err)
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	sse := datastar.NewSSE(w, r)

	ch, sub, err := services.Broker().Subscribe(roomName)
	if err != nil {
		slog.Error("[DEBUG] Failed to subscribe to NATS broker", "error", err, "room", roomName)
		http.Error(w, "failed to subscribe to NATS", http.StatusInternalServerError)
		return
	}
	defer sub.Unsubscribe()

	slog.Info("[DEBUG] SSE connection established (Datastar)", "room", roomName)

	// Iniciar com uma mensagem de sistema
	err = sse.PatchElementTempl(
		pages.SystemMessage("Conectado! Fale com "+session.ClientName),
		datastar.WithModeAppend(),
		datastar.WithSelector("#chat-messages"),
	)
	if err != nil {
		slog.Error("[DEBUG] Failed to send system message via SSE", "error", err, "room", roomName)
	}

	// Keep track of active turns so we know when to append vs morph
	activeTurns := make(map[string]string)

	for {
		select {
		case <-r.Context().Done():
			slog.Info("[DEBUG] SSE connection closed by client", "room", roomName)
			return
		case event := <-ch:
			slog.Info("[DEBUG] Transcript event received from NATS", "room", roomName, "speaker", event.Speaker, "textLen", len(event.Text), "complete", event.IsTurnComplete)
			isUser := event.Speaker == "user"
			speakerName := event.Speaker
			if !isUser {
				speakerName = session.ClientName
			}

			// Generate an ID for the current turn
			turnID, exists := activeTurns[event.Speaker]
			isFirstChunk := !exists
			if !exists {
				turnID = fmt.Sprintf("msg-%s-%d", event.Speaker, time.Now().UnixNano())
				activeTurns[event.Speaker] = turnID
				slog.Debug("[DEBUG] New turn started", "speaker", event.Speaker, "turnID", turnID)
			}

			var err error
			if isFirstChunk {
				// First chunk: Append to the chat
				err = sse.PatchElementTempl(
					pages.TranscriptBubble(speakerName, event.Text, isUser, event.IsTurnComplete, turnID),
					datastar.WithModeAppend(),
					datastar.WithSelector("#chat-messages"),
				)
			} else {
				// Subsequent chunk: Morph the existing bubble
				err = sse.PatchElementTempl(
					pages.TranscriptBubble(speakerName, event.Text, isUser, event.IsTurnComplete, turnID),
				)
			}

			if err != nil {
				slog.Error("[DEBUG] Failed to send transcript bubble via SSE", "error", err, "room", roomName)
				return
			}

			if event.IsTurnComplete {
				slog.Debug("[DEBUG] Turn completed", "speaker", event.Speaker, "turnID", turnID)
				delete(activeTurns, event.Speaker)
			}
		}
	}
}
