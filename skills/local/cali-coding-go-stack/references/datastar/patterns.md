# Datastar v1.0.0 Patterns

## Installation

```bash
# Download latest Datastar JS
curl -O https://cdn.jsdelivr.net/npm/@starfederation/datastar/

# Or use CDN in your templ component:
// <script type="module" src="https://cdn.jsdelivr.net/npm/@starfederation/datastar/"></script>
```

## HTML Template Setup

```go
templ Layout(title string, content templ.Component) {
    <!DOCTYPE html>
    <html lang="en" data-theme="light">
        <head>
            <meta charset="UTF-8"/>
            <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
            <title>{ title }</title>
            <link href="/css/styles.css" rel="stylesheet"/>
            <script type="module" src="https://cdn.jsdelivr.net/npm/@starfederation/datastar/"></script>
        </head>
        <body>
            { content }
        </body>
    </html>
}
```

## Core Attributes

| Attribute | Purpose | Example |
|---|---|---|
| `data-signals` | Define reactive state | `data-signals="{ count: 0 }"` |
| `data-text` | Bind text content | `data-text="$count"` |
| `data-show` | Conditional visibility | `data-show="$isOpen"` |
| `data-class` | Dynamic CSS classes | `data-class="{ active: $isSelected }"` |
| `data-on:event` | Event handlers | `data-on:click="@get('/api/increment')"` |
| `data-bind` | Two-way binding | `data-bind="inputValue"` |
| `data-indicator` | Loading indicator | `data-indicator="isLoading"` |

## Signal Philosophy (IMPORTANT)

**Signals are global state.** They live in the `<body>` tag and are shared across all features.

```html
<!-- GOOD: Global signals in body -->
<body data-signals="{ isOpen: false, count: 0 }">
    <!-- All features can access these signals -->
</body>

<!-- BAD: Local signals in a feature -->
<div data-signals="{ localState: false }">
    <!-- Only accessible within this div -->
</div>
```

**Why?** Datastar's signal system is designed for global state. Local signals create confusion and debugging difficulties.

## Form Submissions

```html
<form data-on:submit="@post('/api/submit', { contentType: 'form' })">
    <input name="email" type="email" data-bind="email"/>
    <button type="submit">Submit</button>
</form>
```

**Key points:**
- Always include `name` attribute on inputs
- Use `{ contentType: 'form' } ` for form data
- Call `r.ParseForm()` in your Go handler

## What NOT to Use

| ❌ Don't | ✅ Do Instead |
|---|---|
| `data-on:click="count++"` | `data-on:click="@get('/api/increment')"` |
| `data-signals="{count: 0}"` in div | `data-signals="{count: 0}"` in body |
| `class="hidden" + data-show` | Only `data-show` |
| `data-on:input="el.value"` | `data-bind="fieldName"` |
| `fmt.Sprintf` with HTML | Templ components |

## Key Concepts

1. **Signals are global** - defined in `<body>`, shared across features
2. **Actions are server-driven** - `@get`, `@post` call your Go handlers
3. **SSE is the transport** - server sends events, client renders
4. **Delta sync** - send only what changed, not entire objects

## Example: Counter

```go
// handler.go
func HandleIncrement(w http.ResponseWriter, r *http.Request) {
    sse := datastar.NewSSE(w, r)
    // Read current count from signals
    count := datastar.GetSignals(r)["count"].(float64)
    // Send updated count
    sse.MarshalAndPatchSignals(map[string]any{"count": count + 1})
}
```

```templ
// counter.templ
templ Counter() {
    <div data-signals="{ count: 0 }">
        <button data-on:click="@get('/api/increment')">
            Count: <span data-text="$count">0</span>
        </button>
    </div>
}
```

## Example: TodoMVC (NATS KV)

A more complex example using NATS KV for persistence:

```go
// handler.go
func HandleAddTodo(w http.ResponseWriter, r *http.Request) {
    sse := datastar.NewSSE(w, r)
    input := datastar.GetSignals(r)["newTodo"].(string)
    
    // Save to NATS KV
    todo := Todo{ID: uuid.New(), Text: input, Done: false}
    repo.Save(todo)
    
    // Send updated list
    todos, _ := repo.ListAll()
    templ.Handler(todomvc.TodoList(todos)).ServeHTTP(w, r)
}
```

```templ
// todomvc.templ
templ TodoMVC() {
    <div data-signals="{ newTodo: '' }">
        <input data-bind="newTodo" placeholder="What needs to be done?"/>
        <button data-on:click="@post('/api/todos')">Add</button>
        <div id="todo-list"></div>
    </div>
}
```
