# Datastar Patterns & Pitfalls

## ⚠️ v1.0.0 Critical Setup

**This section is REQUIRED for Datastar v1.0.0 to work!**

### Version Notice
Use **v1.0.0** from CDN or self-hosted. Other versions (preview, beta) may have different behavior.

### The Problem
Datastar v1.0.0 requires the STANDARD datastar.js (not aliased version)

### Correct Setup

```html
<!-- In your HTML <head> -->
<script defer type="module" src="/static/datastar/datastar.js"></script>
```

### About data-init
`data-init` is **optional** - it executes an expression when an element is initialized. Use it only when you need to:
- Fetch initial data on page/component load (e.g., lazy loading)
- Execute any action exactly once when the element first renders

**Do NOT use `data-init` "by default"** - only add it when there's a concrete reason.

---

## Critical Pitfalls Discovered Through Debugging

### 1. JSON Escaping with Complex Content

**Symptom:** `SyntaxError: Invalid or unexpected token` in browser console

**Wrong approach:**
```go
func escapeForJS(s string) string {
    s = strings.ReplaceAll(s, `\`, `\\`)
    s = strings.ReplaceAll(s, `'`, `\'`)
    s = strings.ReplaceAll(s, "\n", `\n`)
    return s
}
```

**Correct approach:**
```go
import "encoding/json"

func escapeForJS(s string) string {
    b, _ := json.Marshal(s)
    return string(b)
}
```

**Why:** Manual escaping breaks with nested quotes, unicode, and complex content. `json.Marshal` handles all cases.

**Note:** `json.Marshal` returns the string with quotes included, so use `%s` directly in templates, not `'%s'`.

### 2. Form Data Not Sending

**Symptom:** POST body arrives empty or without field values

**Wrong:**
```html
<input type="text" id="model" data-bind="model">
```

**Correct:**
```html
<input type="text" id="model" name="model" data-bind="model">
```

**Why:** Datastar v1 requires the `name` attribute to include the field in POST body.

### 3. Textareas Not Syncing

**Symptom:** Changing textarea doesn't update the signal, or value not sent in POST

**Wrong:**
```html
<textarea data-on:input="$refinePrompt = el.value">
```

**Correct:**
```html
<textarea name="refinePrompt" data-bind="refinePrompt"></textarea>
```

**Why:** `data-bind` provides two-way synchronization. Without it, changes don't propagate properly.

### 4. Tabs Not Showing Content

**Symptom:** Clicking tab button doesn't reveal content

**Wrong:**
```html
<div data-show="$currentTab === 'refine'" class="hidden" data-class-hidden="$currentTab !== 'refine'">
```

**Correct:**
```html
<div data-show="$currentTab === 'refine'">
```

**Why:** Mixing `class="hidden"` with `data-show` causes conflicts. Datastar v1 manages visibility purely through `data-show`.

### 5. Loading Button Never Resets

**Symptom:** Clicking save shows spinner forever

**Wrong (Go handler):**
```go
func saveSettings(sse *datastar.ServerSentEventGenerator) {
    repo.Save(data)
    // isSaving signal is never reset!
}
```

**Correct (Go handler):**
```go
func saveSettings(sse *datastar.ServerSentEventGenerator) {
    repo.Save(data)

    // IMPORTANT: Reset the loading signal!
    sse.MarshalAndPatchSignals(map[string]bool{"isSaving": false})
}
```

### 6. Static Files 404

**Symptom:** `GET /static/datastar/datastar.js net::ERR_ABORTED 404`

**Cause:** Path resolution fails when running binary from different directory

**Solution (dev mode):**
```go
func findStaticDir() string {
    candidates := []string{
        "./web/resources/static",
        "../web/resources/static",
        "../../web/resources/static",
    }

    if ex, err := os.Executable(); err == nil {
        exDir := filepath.Dir(ex)
        candidates = append(candidates,
            filepath.Join(exDir, "web", "resources", "static"),
            filepath.Join(exDir, "..", "..", "web", "resources", "static"),
        )
    }

    for _, dir := range candidates {
        if info, err := os.Stat(dir); err == nil && info.IsDir() {
            abs, _ := filepath.Abs(dir)
            return abs
        }
    }
    return candidates[0]
}
```

**Note:** The boilerplate uses build tags (`//go:build dev`) which works in most cases. Only use `findStaticDir()` if you need to run the binary from arbitrary directories.

### 7. Go Variables in @post/@get Expressions

**Symptom:** `ReferenceError: model is not defined` in browser console when clicking.

**Wrong approach (Go variable in Datastar JS expression):**
```templ
<!-- model.ID is a Go variable — Datastar evaluates as JavaScript, which fails -->
<button data-on:click="@post('/api/action?id=' + model.ID)">Click</button>
```

**Correct approach (Templ interpolates Go variable server-side):**
```templ
<!-- { } tells Templ to render model.ID as a string at build time -->
<button data-on:click={ "@post('/api/action?id=" + model.ID + "')" }>Click</button>
```

**Why:** Datastar evaluates `@post()` expressions as JavaScript on the client. Go variables like `model.ID`, `session.ID` don't exist in the browser — they produce `ReferenceError`. Using `{ }` Templ expressions renders the Go value into the HTML string at build time.

**Exception — client-side signals don't need interpolation:**
```templ
<!-- CORRECT: $itemId is a client-side Datastar signal, not a Go variable -->
<button data-on:click="@post('/api/action?id=' + $itemId)">Click</button>
```

---

## Troubleshooting

| Symptom | Cause | Solution |
|---------|-------|----------|
| `SyntaxError: Invalid or unexpected token` | Malformed JSON in data-signals | Use `json.Marshal` for escaping |
| POST body empty | Missing `name` attribute | Add `name="fieldName"` to inputs |
| Signal not updating | Missing `data-bind` | Add `data-bind="varName"` |
| Tabs not showing | `class="hidden"` + `data-show` conflict | Remove `class="hidden"`, use only `data-show` |
| Loading never resets | Signal not reset in handler | Call `sse.MarshalAndPatchSignals({"isLoading": false})` |
| 404 for datastar.js | Path resolution issue | Verify static file serving setup |
| Clicks not triggering | Check `data-on:click` syntax | Verify action uses `@post('/api/action')` format |

### Debugging Checklist

1. **Check browser console** for JavaScript errors
2. **View page source** to verify data-signals JSON is valid
3. **Network tab** - verify SSE connection established
4. **Server logs** - check for Datastar-Request headers
5. **curl test** - `curl -v http://localhost:8080/static/datastar/datastar.js`

---

## When to Use Datastar

Datastar excels at **reactive UI state** — small, ephemeral state that changes frequently:
- Toggle states (`isOpen`, `isLoading`)
- Tab selection (`activeTab`)
- Form input binding (small forms)
- Conditional rendering of small UI fragments

Datastar is **NOT** designed for:
- Large text content (prompts, articles, long descriptions)
- Complex nested data structures
- Initial page content that doesn't change reactively

## The Go SDK vs JavaScript API

### Go SDK (Server-side)
```go
import "github.com/starfederation/datastar-go/datastar"

// In .templ files - these work:
sse := datastar.NewSSE(w, r)
sse.PatchElementTempl(Component())
sse.MarshalAndPatchSignals(map[string]any{"count": 5})
```

### JavaScript Actions (Client-side)
```html
<!-- CORRECT: Using @post() action -->
<button data-on:click="@post('/api/action')">Submit</button>

<!-- CORRECT: Raw fetch with expression -->
<button data-on:click="fetch('/api/action', {method: 'POST'})">Submit</button>

<!-- WRONG: datastar.PostSSE() does NOT exist in JavaScript -->
<button data-on:click="datastar.postSSE('/api/action')">Broken!</button>
```

### Go SDK Functions in Templ

These **DO** exist in the Go SDK for use in templ expressions:
```templ
data-on:click={ datastar.PostSSE("/api/action") }
data-on:click={ datastar.PutSSE("/api/action/%d", id) }
data-on:click={ datastar.DeleteSSE("/api/action/%d", id) }
```

But they generate `@post()`, `@put()`, `@delete()` in HTML — they don't become `datastar.postSSE()` in JavaScript.

## Data Size Guidelines

### Signals (Small Data Only)
```html
<!-- GOOD: Small booleans, strings, numbers -->
<div data-signals='{"isOpen": false, "count": 0, "name": "test"}'>

<!-- BAD: Long text content will break parsing -->
<div data-signals='{"prompt": "This is a very long prompt...\nwith newlines...\nand more text..."}'>
```

### Server-Side Rendering (Large Data)
```html
<!-- CORRECT: Render large content directly in the template -->
<textarea name="prompt">{ user.PromptContent }</textarea>

<!-- OR: Use data-bind for small inputs, render initial value server-side -->
<input type="text" data-bind="title" value={ pageTitle } />
```

## Common Pitfalls

### 1. JSON in data-signals with Newlines
```go
// BROKEN: Newlines in JSON break Datastar parsing
`{"text": "Hello\nWorld"}`

// FIXED: Use HTML entities or avoid newlines in signals
`{"text": "Hello&#10;World"}`
// OR render the content directly in HTML
<textarea>{ largeText }</textarea>
```

### 2. Trying to Pass Functions via Signals
```html
<!-- BROKEN: Can't pass functions through signals -->
<div data-signals='{"onClick": "() => doSomething()"}'>

<!-- FIXED: Define functions in <script> tags or use data-on:click -->
<button data-on:click="doSomething()">Click</button>
```

### 3. Large Data in data-show Conditions
```html
<!-- WORKS but inefficient: -->
<div data-show="longSignalName === 'some long value'">

<!-- BETTER: Use small boolean flags -->
<div data-show="$isVisible">
```

### 4. Mixing SSR and Datastar for Same Data
```html
<!-- CONFUSING: Value both in signals and as initial input -->
<input data-bind="name" value="John">

<!-- FIXED: Choose one approach -->
<input data-bind="name">  <!-- Datastar controls value -->
<!-- OR -->
<input value={ user.Name }>  <!-- Server renders, no binding -->
```

## Fetch Pattern (Recommended for Forms)

For forms with validation, loading states, and error handling, use vanilla JavaScript:

```html
<form id="my-form">
    <input name="email" type="email" required>
    <button type="submit">Submit</button>
</form>

<script>
document.getElementById('my-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const form = e.target;
    const data = { email: form.email.value };
    
    try {
        const res = await fetch('/api/subscribe', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(data)
        });
        if (res.ok) {
            // Success
        }
    } catch (err) {
        // Error
    }
});
</script>
```

## Signal Reactivity Pattern

For complex UI with multiple states:

```html
<div data-signals='{"tab": "home", "loading": false}'>
    <button data-on:click="$tab = 'home'">Home</button>
    <button data-on:click="$tab = 'settings'">Settings</button>
    
    <div data-show="$tab === 'home'">Home content</div>
    <div data-show="$tab === 'settings'">Settings content</div>
    
    <div data-show="$loading" class="spinner"></div>
</div>
```

Keep signals flat and simple. Use meaningful names with `$` prefix convention.

## When to Use Datastar vs Vanilla JS

| Use Case | Recommendation |
|----------|---------------|
| Toggle visibility | Datastar `data-show` |
| Tab switching | Datastar signals + `data-show` |
| Form input binding | Datastar `data-bind` (small forms) |
| Button click → API call | Datastar `@post()` or vanilla fetch |
| Loading states | Datastar signals |
| Large text content | SSR (render in template) |
| Complex forms | Vanilla JS with fetch |
| Page initialization data | SSR (not signals) |

**Key Principle:**
> ⚠️ **Overusing signals indicates managing state on the frontend. Backend should be the source of truth.**

Signals are for:
- User interactions (toggles, tabs, visibility)
- Sending form data to backend

Signals are NOT for:
- Storing fetched/loaded data
- Client-side state management
- Persisting data between page loads
