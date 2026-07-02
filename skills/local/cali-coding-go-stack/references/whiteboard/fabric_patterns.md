# Fabric.js Whiteboard: Synchronization Patterns

## Architecture

```
┌─────────────┐     SSE      ┌─────────┐     NATS     ┌─────────────┐
│   Browser   │ ◄──────────► │  Go     │ ◄──────────► │  NATS JS    │
│  (Fabric.js)│              │ Server  │              │  (state)    │
└─────────────┘              └─────────┘              └─────────────┘
```

## Principles

1. **Isolation** - CSS/JS self-contained in feature
2. **Delta sync** - send only changed properties
3. **KISS** - HTTP POST for mutations, SSE for stream
4. **NATS KV** - shared state between instances

## Delta Synchronization

### What to send (minimize payload):

```javascript
// ❌ WRONG - send entire object
canvas.on('object:modified', (e) => {
    fetch('/whiteboard/update', {
        method: 'POST',
        body: JSON.stringify({ object: e.target })
    });
});

// ✅ CORRECT - only relevant properties
canvas.on('object:modified', (e) => {
    const obj = e.target;
    fetch('/whiteboard/update', {
        method: 'POST',
        body: JSON.stringify({
            id: obj.id,
            left: obj.left,
            top: obj.top,
            angle: obj.angle,
            scaleX: obj.scaleX,
            scaleY: obj.scaleY,
            // don't send: object, canvas, etc
        })
    });
});
```

## go:embed for Assets

Each feature has its own assets:

```go
// features/whiteboard/static.go
package whiteboard

import (
    "embed"
    "io/fs"
    "net/http"
)

//go:embed static/*
var staticEmbed embed.FS

func StaticFS() http.FileSystem {
    fsys, _ := fs.Sub(staticEmbed, "static")
    return http.FS(fsys)
}

// features/whiteboard/routes.go
func SetupRoutes(router chi.Router, ...) {
    router.Handle("/whiteboard/static/*", 
        http.StripPrefix("/whiteboard/static/", 
            http.FileServer(StaticFS())))
}
```

## Pattern: SSE for Broadcast

```go
// Server: SSE endpoint for whiteboard updates
func (h *Handlers) WhiteboardStream(w http.ResponseWriter, r *http.Request) {
    sse := datastar.NewSSE(w, r)
    
    // Subscribe to NATS channel
    sub, _ := h.js.Subscribe("whiteboard.updates", func(msg *nats.Msg) {
        // Broadcast to all connected clients
        sse.MarshalAndSend(string(msg.Data))
    })
    defer sub.Unsubscribe()
    
    // Keep connection open
    <-r.Context().Done()
}
```

## Storage Polyfill (LocalStorage)

For restricted environments:

```javascript
// static/whiteboard.js
if (!window.localStorage) {
    const store = new Map();
    window.localStorage = {
        getItem: (k) => store.get(k) ?? null,
        setItem: (k, v) => store.set(k, v),
        removeItem: (k) => store.delete(k),
        clear: () => store.clear()
    };
}
```

## Available Commands in Scaffold

| Feature | Path | Description |
|---------|------|-------------|
| Whiteboard | `/whiteboard` | Collaborative canvas |

## Whiteboard Routes

```go
// features/whiteboard/routes.go
func SetupRoutes(router chi.Router, sessionStore sessions.Store, ns *embeddednats.Server) error {
    whiteboard := handlers.NewHandlers(ns)
    
    // Main page
    router.Get("/whiteboard", whiteboard.Page)
    
    // Update API
    router.Post("/whiteboard/update", whiteboard.Update)
    router.Post("/whiteboard/clear", whiteboard.Clear)
    
    // SSE stream for updates
    router.Get("/whiteboard/stream", whiteboard.Stream)
    
    // Serve static assets
    router.Handle("/whiteboard/static/*", 
        http.StripPrefix("/whiteboard/static/", 
            http.FileServer(StaticFS())))
    
    return nil
}
```

## Turn Tracking for SSE

For incremental updates (not replacing entire DOM):

```javascript
// Use stable IDs for fragments
<div id="whiteboard-canvas-{unique-id}">
    <!-- canvas content -->
</div>

// Partial updates via SSE
sse.on('whiteboard:update', (data) => {
    const element = document.getElementById(`whiteboard-canvas-${data.id}`);
    if (element) {
        // Idiomorph or partial patch
        morph(element, data.fragment);
    }
});
```
