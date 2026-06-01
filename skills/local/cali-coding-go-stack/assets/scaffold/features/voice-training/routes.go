package voicetraining

import (
	"fmt"

	"northstar/config"
	"northstar/features/voice-training/services"

	"github.com/delaneyj/toolbelt/embeddednats"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(router chi.Router, ns *embeddednats.Server) error {
	nc, err := ns.Client()
	if err != nil {
		return fmt.Errorf("error creating nats client: %w", err)
	}

	// Initialize the NATS broker for transcripts
	services.InitBroker(nc)

	liveKitService := services.NewLiveKitService(
		config.Global.LiveKitURL,
		config.Global.LiveKitAPIKey,
		config.Global.LiveKitAPISecret,
	)

	sessionService, err := services.NewSessionService(nc)
	if err != nil {
		return fmt.Errorf("error creating session service: %w", err)
	}

	handlers := NewHandlers(liveKitService, sessionService, config.Global.GeminiAPIKey)

	router.Mount("/voice-training", Routes(handlers))

	return nil
}

func Routes(h *Handlers) chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.LandingPage)
	r.Post("/create", h.CreateRoom)
	r.Get("/{id}", h.RoomPage)
	r.Post("/{id}/end", h.EndSession)
	r.Post("/{id}/join-bot", h.JoinBot)
	r.Get("/{id}/transcripts", h.TranscriptsStream)

	return r
}
