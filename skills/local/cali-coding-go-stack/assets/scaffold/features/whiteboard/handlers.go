package whiteboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"northstar/features/whiteboard/pages"
	"northstar/features/whiteboard/services"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/starfederation/datastar-go/datastar"
)

type FabricHandlers struct {
	fabricService *services.FabricService
	cursorService *services.CursorService
	sessionStore  sessions.Store
}

func NewFabricHandlers(
	fabricService *services.FabricService,
	cursorService *services.CursorService,
	sessionStore sessions.Store,
) *FabricHandlers {
	return &FabricHandlers{
		fabricService: fabricService,
		cursorService: cursorService,
		sessionStore:  sessionStore,
	}
}

func (h *FabricHandlers) getUserID(r *http.Request, w http.ResponseWriter) string {
	if tabId := r.URL.Query().Get("tabId"); tabId != "" {
		return tabId
	}
	return "anon_" + fmt.Sprintf("%d", time.Now().UnixNano())
}

func (h *FabricHandlers) FabricWhiteboardPage(w http.ResponseWriter, r *http.Request) {
	if err := pages.FabricWhiteboardPage().Render(r.Context(), w); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (h *FabricHandlers) StreamFabricCanvas(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := h.getUserID(r, w)
	sse := datastar.NewSSE(w, r)

	state, err := h.fabricService.GetCanvasState(ctx)
	if err != nil {
		state = &services.FabricCanvasState{Version: "5.3.0", Objects: []*services.FabricObject{}}
	}
	data, _ := json.Marshal(state)
	sse.ExecuteScript(fmt.Sprintf(
		`window.dispatchEvent(new CustomEvent('wb:canvas_init',{detail:%s}))`,
		string(data),
	))

	ch, err := h.fabricService.SubscribeDeltas(ctx)
	if err != nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case delta := <-ch:
			if delta == nil || delta.UserID == userID {
				continue
			}
			data, _ := json.Marshal(delta)
			sse.ExecuteScript(fmt.Sprintf(
				`window.dispatchEvent(new CustomEvent('wb:canvas_delta',{detail:%s}))`,
				string(data),
			))
		}
	}
}

func (h *FabricHandlers) SaveFabricObject(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r, w)

	var req struct {
		ID         string                 `json:"id"`
		Tool       string                 `json:"tool"`
		Color      string                 `json:"color"`
		FabricData *services.FabricObject `json:"fabricData"`
		Delta      *services.FabricObject `json:"delta"`
		Live       bool                   `json:"_live"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var obj *services.FabricObject
	isDelta := req.Delta != nil

	if isDelta {
		obj = req.Delta
	} else {
		obj = req.FabricData
	}

	if obj == nil {
		http.Error(w, "no fabricData or delta provided", http.StatusBadRequest)
		return
	}

	if obj.ID == "" {
		obj.ID = req.ID
	}

	if obj.Type == "" {
		obj.Type = req.Tool
	}

	delta := &services.FabricDelta{
		Type:       "object",
		ObjectID:   obj.ID,
		FabricData: obj,
		UserID:     userID,
		Timestamp:  time.Now().UnixMilli(),
	}

	// Always save to KV for persistence (not for live drawing)
	if !req.Live {
		if err := h.fabricService.SaveObject(r.Context(), obj); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	h.fabricService.BroadcastDelta(r.Context(), delta)
}

func (h *FabricHandlers) ClearFabricCanvas(w http.ResponseWriter, r *http.Request) {
	if err := h.fabricService.ClearCanvas(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *FabricHandlers) DeleteFabricObject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := h.getUserID(r, w)
	objectID := chi.URLParam(r, "objectId")

	if objectID == "" {
		http.Error(w, "object ID required", http.StatusBadRequest)
		return
	}

	// Delete from KV
	if err := h.fabricService.DeleteObject(r.Context(), objectID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast delete event
	delta := &services.FabricDelta{
		Type:      "delete",
		ObjectID:  objectID,
		UserID:    userID,
		Timestamp: time.Now().UnixMilli(),
	}

	h.fabricService.BroadcastDelta(r.Context(), delta)
}

// Cursor handlers
func (h *FabricHandlers) StreamCursors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := h.getUserID(r, w)
	sse := datastar.NewSSE(w, r)

	cursor := h.cursorService.GetOrCreateCursor(userID)

	allCursors := h.cursorService.GetAllCursors()
	otherCursors := make([]*services.Cursor, 0)
	for _, c := range allCursors {
		if c.UserID != userID {
			otherCursors = append(otherCursors, c)
		}
	}

	data, _ := json.Marshal(map[string]interface{}{
		"myCursor": cursor,
		"cursors":  otherCursors,
	})
	sse.ExecuteScript(fmt.Sprintf(
		`window.dispatchEvent(new CustomEvent('wb:cursor_init',{detail:%s}))`,
		string(data),
	))

	ch, err := h.cursorService.Subscribe(ctx)
	if err != nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case c := <-ch:
			data, _ := json.Marshal(c)
			sse.ExecuteScript(fmt.Sprintf(
				`window.dispatchEvent(new CustomEvent('wb:cursor_update',{detail:%s}))`,
				string(data),
			))
		}
	}
}

func (h *FabricHandlers) UpdateCursor(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r, w)

	var req struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.cursorService.UpdateCursor(userID, req.X, req.Y)
}
