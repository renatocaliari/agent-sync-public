package whiteboard

import (
	"net/http"

	"northstar/features/whiteboard/services"

	"github.com/delaneyj/toolbelt/embeddednats"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
)

func SetupRoutes(
	router chi.Router,
	sessionStore sessions.Store,
	ns *embeddednats.Server,
) error {
	cursorService, err := services.NewCursorService(ns)
	if err != nil {
		return err
	}

	fabricService, err := services.NewFabricService(ns)
	if err != nil {
		return err
	}

	fabricHandlers := NewFabricHandlers(fabricService, cursorService, sessionStore)

	// Serve static files for the whiteboard module
	router.Handle("/whiteboard/static/*", http.StripPrefix("/whiteboard/static/", http.FileServer(StaticFS())))

	// Fabric.js whiteboard (main whiteboard route)
	router.Get("/whiteboard", fabricHandlers.FabricWhiteboardPage)
	router.Get("/whiteboard/stream", fabricHandlers.StreamFabricCanvas)
	router.Post("/whiteboard/drawing", fabricHandlers.SaveFabricObject)
	router.Delete("/whiteboard/drawings", fabricHandlers.ClearFabricCanvas)
	router.Delete("/whiteboard/object/{objectId}", fabricHandlers.DeleteFabricObject)

	// Cursor sync
	router.Get("/whiteboard/cursors", fabricHandlers.StreamCursors)
	router.Post("/whiteboard/cursor", fabricHandlers.UpdateCursor)

	return nil
}
