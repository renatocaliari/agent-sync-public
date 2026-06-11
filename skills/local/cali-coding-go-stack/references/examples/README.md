# Examples Reference

This directory contains standalone examples adapted for the boilerplate structure.

## Available Examples

| File | Pattern | Description |
|------|---------|-------------|
| `click_to_edit.templ` | Inline Editing | Click to edit a record inline |
| `active_search.templ` | Debounced Search | Search as you type with 200ms debounce |
| `counter.templ` | Real-time Counter | Shared counter with SSE updates |
| `infinite_scroll.templ` | Infinite Scroll | Load more on scroll with `data-on-intersect` |
| `lazy_load.templ` | Lazy Loading | Load content on demand |
| `file_upload.templ` | File Upload | Upload files via base64 encoding |
| `todo_mvc.templ` | Full CRUD | Complete todo app with session persistence |
| `whiteboard.templ` | Collaborative | Real-time collaborative whiteboard with NATS |

## Usage

Each example is a self-contained `.templ` file with:
1. Route setup function (e.g., `SetupClickToEdit`)
2. Handler functions
3. Templ templates

### Integrating an Example

1. Copy the `.templ` file to `features/<name>/pages/`
2. Add the setup call in `router/router.go`:

```go
import "yourproject/features/<name>/pages"

func SetupRoutes(...) error {
    pages.SetupClickToEdit(router)
    // ...
}
```

3. Run `go tool templ generate` to compile templates

## Common Patterns Demonstrated

### Click to Edit
- `@get()` to fetch edit form
- `@put()` to save changes
- `MarshalAndPatchSignals()` to sync state

### Active Search
- `data-on:input__debounce.200ms` for debouncing
- `ReadSignals()` to get search query
- Filter and return results

### Counter
- `atomic.Int32` for shared state
- `PostSSE()` for actions
- Real-time updates via SSE

### Infinite Scroll
- `data-on-intersect` to trigger load
- `WithModeAppend()` to add elements
- Offset-based pagination

### File Upload
- `data-bind:files` for file input
- Base64 encoding in signals
- Max size validation

### Todo MVC
- Session-based persistence
- CRUD operations
- Route parameters with `chi.URLParam`

### Whiteboard
- NATS KV for persistence
- Real-time updates via watchers
- Multi-user collaboration