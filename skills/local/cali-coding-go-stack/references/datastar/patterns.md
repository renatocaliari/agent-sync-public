# Datastar Patterns

Advanced Datastar v1.0.2 patterns for reactive hypermedia.

> ⚠️ **IMPORTANT:** Use Datastar v1.0.2 from CDN or the included file.

> ⚠️ **Datastar has NO global JS API (`window.datastar` does not exist).**
> Unlike Alpine.js (`window.Alpine`) or React, Datastar v1.x is a **pure ES module** —
> it does NOT expose `window.datastar.setSignals()`, `window.datastar.getSignal()`,
> or any similar global. To read or write signals from a `<script>` block:
> - **Read initial values:** parse the `data-signals` attribute:
>   `JSON.parse(el.getAttribute('data-signals')).signalName`
> - **Set signals from JS:** dispatch a CustomEvent:
>   `document.dispatchEvent(new CustomEvent('datastar-signal-patch', {detail: {key: val}}))`
> - **For timers/countdowns/progress:** use direct DOM manipulation
>   (`document.getElementById`, `textContent`, `style.display`, `className`),
>   as shown in the `startRetryCountdown` pattern below. `setInterval` + DOM is
>   the correct approach for things "only JS can do."

## Signals

Signals are reactive state in the frontend. Define on the parent element:

```html
<div
    id="my-component"
    data-signals={ templ.JSONString(MySignals{Value: "initial"}) }
>
    <!-- Children inherit signals -->
</div>
```

### Go -> Frontend (struct)

```go
type MySignals struct {
    Count    int      `json:"count"`
    Name     string   `json:"name"`
    IsActive bool     `json:"isActive"`
    Items    []string `json:"items"`
}

templ MyComponent() {
    <div
        data-signals={ templ.JSONString(MySignals{
            Count: 0,
            Name:  "Hello",
        }) }
    >
        <span data-text="$count"></span>
    </div>
}
```

### Backend -> Frontend (patch)

```go
sse := datastar.NewSSE(w, r)

// Patch specific element
sse.PatchElementTempl(MyComponent(), datastar.WithModeAppend(), datastar.WithSelector("#target"))

// Update signals directly
store := gabs.New()
store.Set(42, "count")
store.Set("updated", "name")
sse.MarshalAndPatchSignals(store)
```

> **Prefer `datastar.ReadSignals(r, &store)` for reading signals from request** — it handles both query param (GET/DELETE) and body (POST/PUT/PATCH) methods correctly per the SDK spec. Manual `json.Unmarshal` works for simple cases but bypasses the spec.

## Event Handlers

### Click

```html
<button
    data-on:click={ datastar.GetSSE("/api/data") }
>
    Fetch
</button>

<!-- With parameters -->
<button
    data-on:click={ datastar.PostSSE("/api/item/%d/toggle", itemID) }
>
    Toggle
</button>
```

### Keydown

```html
<input
    class="input"
    data-bind:inputValue
    data-on:keydown={`
        if (evt.key !== 'Enter' || !$inputValue.trim().length) return;
        ${datastar.PutSSE("/api/search")}
        $inputValue = '';
    `}
/>
```

### Click Outside

```html
<!-- Close dropdown when clicking outside -->
<div
    id="dropdown"
    data-on:click__outside={ datastar.PutSSE("/ui/dropdown/close") }
>
    <button class="btn">Open</button>
</div>
```

## Loading Indicators (data-indicator)

Shows loading state while request is pending.

### Button with indicator

```html
<button
    class="btn btn-primary"
    data-on:click={ datastar.PostSSE("/api/action") }
    data-indicator="myIndicator"
>
    Submit
</button>

<!-- Visual indicator (spinner) -->
@common.SseIndicator("myIndicator")
```

### Input disabled during loading

```html
<input
    class="input"
    data-bind:inputValue
    data-attr-disabled="$isLoading"
    data-indicator="searchIndicator"
/>

<span data-text="$isLoading ? 'Searching...' : 'Ready'"></span>
```

### Indicator CSS

```css
[data-indicator] {
    position: relative;
}

[data-indicator-fetching] {
    opacity: 0.6;
    pointer-events: none;
}

[data-indicator-fetching]::after {
    content: '';
    position: absolute;
    /* spinner styles */
}
```

### Manual boolean signals (alternative)

Many projects use manual boolean signals instead of `data-indicator`:

```html
<button data-show="!$isSaving" data-on:click="$isSaving = true; @post('/api/action')">
  Save
</button>
<button data-show="$isSaving" disabled>
  <span class="loading loading-spinner"></span> Saving...
</button>
```

```go
// Backend: reset after operation
sse.MarshalAndPatchSignals(map[string]interface{}{"isSaving": false})
```

Both approaches work. Manual signals give more explicit control; `data-indicator` is more concise.

## Reactive Attributes (data-attr-*)

### Dynamic disabled

```html
<button
    class="btn"
    data-on:click={ datastar.PostSSE("/api/submit") }
    data-indicator="submitting"
    data-attr-disabled="$isSubmitting"
>
    Submit
</button>
```

### Dynamic class

```html
<div
    class={ "p-4", templ.KV("bg-primary", isActive) }
    data-on:click={ datastar.PutSSE("/api/toggle") }
>
    Click
</div>
```

### Dynamic style

```html
<div
    data-style:width={ fmt.Sprintf("%dpx", width) }
    data-style:background-color={ color }
>
    Content
</div>
```

## SSE (Server-Sent Events)

### Backend Go

```go
sse := datastar.NewSSE(w, r)

// PATCH - update HTML fragment
sse.PatchElementTempl(Component(), datastar.WithModeAppend(), datastar.WithSelector("#target"))

// SIGNALS - update JS state
update := gabs.New()
update.Set(42, "count")
sse.MarshalAndPatchSignals(update)

// EXECUTE - execute JS script
sse.ExecuteScript("alert('Done!')")
```

### Patching with Raw HTML Strings

For projects not using Templ, use `PatchElements` with raw HTML strings:

```go
sse := datastar.NewSSE(w, r)

// Patch raw HTML to element
sse.PatchElements(htmlString,
    datastar.WithSelectorID("chat-container"),
    datastar.WithModeAppend())

// Patch signals from map
sse.MarshalAndPatchSignals(map[string]interface{}{
    "contextBlocks": settings.ContextBlocks,
})
```

>`PatchElementTempl` uses Templ components; `PatchElements` uses raw HTML strings.

> **Prefer `sse.PatchElementTempl(component, opts...)` for Templ components** — it uses `bytebufferpool` (more efficient than `strings.Builder`) and inherits `sse.Context()` automatically. Manual rendering with `strings.Builder` + `sse.PatchElements` should be reserved for cases where you already have a string.

### Patch Modes

| Mode | Description |
|------|-------------|
| `Append` | Adds after element |
| `Prepend` | Adds before element |
| `Outer` | Replaces element |
| `Inner` | Replaces inner content |
| `Delete` | Removes element |
| `Morph` | Idiomorph (merge DOM) |

### Frontend: listening to SSE events

```javascript
// Connect to SSE endpoint
const sse = new EventSource('/api/stream');

sse.addEventListener('datastar', (event) => {
    const data = JSON.parse(event.data);
    // process patches, signals, etc
});

// Specific patches
sse.addEventListener('datastar-patch', (event) => {
    const patch = JSON.parse(event.data);
    applyPatch(patch);
});
```

## Forms with Datastar

### Input binding

```html
<form
    data-signals={ templ.JSONString(FormSignals{Email: ""}) }
    data-on:submit={ datastar.PostSSE("/api/register") }
>
    <input
        type="email"
        class="input input-bordered"
        placeholder="email@example.com"
        data-bind:email
    />
    <button type="submit" class="btn btn-primary">
        Register
    </button>
</form>
```

### Client-side validation

```go
type FormSignals struct {
    Email    string `json:"email"`
    Error    string `json:"error"`
    IsValid  bool   `json:"isValid"`
}

templ Form() {
    <form
        data-signals={ templ.JSONString(FormSignals{}) }
        data-on:submit={`
            if (!/$email/.test($email)) {
                $error = 'Invalid email';
                return;
            }
            ${datastar.PostSSE("/api/register")}
        `}
    >
        <input
            class="input input-bordered"
            class={ templ.KV("input-error", error !== "") }
            data-bind:email
            data-on:input={`$error = ''`}
        />
        if error != "" {
            <span class="text-error text-sm">{ error }</span>
        }
    </form>
}
```

## Best Practices

### Request cancellation for pages

```html
<div
    data-init={ datastar.GetSSE("/api/list", {requestCancellation: 'disabled'}) }
>
    <!-- does not cancel when navigating to another page -->
</div>
```

### try/catch on media operations

```javascript
async function enableMic() {
    try {
        await room.localParticipant.setMicrophoneEnabled(true);
    } catch (e) {
        console.warn('Mic access denied:', e);
    }
}
```

### View Transitions for smooth navigation

```css
@view-transition {
    navigation: auto;
}
```

---

## Complete Form Patterns (v1.0.2 Verified)

### Full Form with Validation and Loading State

```go
// Go handler
type SettingsSignals struct {
    Model         string `json:"model"`
    RefinePrompt  string `json:"refinePrompt"`
    CurrentTab    string `json:"currentTab"`
    IsSaving      bool   `json:"isSaving"`
    ShowSaveToast bool   `json:"showSaveToast"`
}

func (ns *NarrativeService) saveSettings(sse *datastar.ServerSentEventGenerator, signals map[string]interface{}) {
    // ... save to database ...

    // IMPORTANT: Reset loading signal!
    sse.MarshalAndPatchSignals(map[string]interface{}{
        "isSaving":      false,
        "showSaveToast": true,
    })
}
```

```html
<!-- HTML Template -->
<div data-signals={ templ.JSONString(SettingsSignals{
    Model:        "gpt-4",
    CurrentTab:   "refine",
    IsSaving:     false,
}) }>

    <!-- Tabs -->
    <div class="tabs">
        <button data-on:click="$currentTab = 'refine'">Refinar</button>
        <button data-on:click="$currentTab = 'generate'">Gerar</button>
    </div>

    <!-- Tab Content -->
    <div data-show="$currentTab === 'refine'">
        <input type="text" name="model" data-bind="model">
        <textarea name="refinePrompt" data-bind="refinePrompt"></textarea>
    </div>

    <div data-show="$currentTab === 'generate'">
        <!-- generate content -->
    </div>

    <!-- Save Button with Loading State -->
    <button
        data-show="!$isSaving"
        data-on:click="$isSaving = true; @post('/api/ui/action?action=save_settings')"
    >
        Salvar
    </button>

    <button data-show="$isSaving" disabled>
        <span class="loading loading-spinner"></span>
        Salvando...
    </button>
</div>
```

**Prevent FOUC (Flash of Unstyled Content):** For elements that start hidden, add `style="display: none"` as initial state:
```html
<div data-show="$currentTab === 'details'" style="display: none">
  <!-- Content hidden until Datastar initializes -->
</div>
```
This prevents FOUC by hiding the element before Datastar processes the DOM.

### Centralized Action Dispatch

A common pattern routes all UI actions through a single endpoint with a query parameter:

```html
<button data-on:click="@post('/api/ui/action?action=save_settings')">Save</button>
<button data-on:click="@post('/api/ui/action?action=delete_item&id=' + $itemId)">Delete</button>
```

```go
func handler(w http.ResponseWriter, r *http.Request) {
    sse := datastar.NewSSE(w, r)
    body, _ := io.ReadAll(r.Body)
    signals := map[string]interface{}{}
    json.Unmarshal(body, &signals)

    switch r.URL.Query().Get("action") {
    case "save_settings":
        // ...
    case "delete_item":
        // ...
    }
}
```

### Form Submit with Loading Signal

Set a loading signal before the `@post()` call — the JS executes left-to-right, so `$isSaving` is set before the fetch starts:

```html
<form data-on:submit="$isSaving = true; @post('/api/ui/action?action=send_message')">
  <button type="submit" data-show="!$isSaving">Send</button>
  <button type="submit" data-show="$isSaving" disabled>
    <span class="loading loading-spinner"></span> Sending...
  </button>
</form>
```

```go
// Backend: reset after completion
sse.MarshalAndPatchSignals(map[string]interface{}{"isSaving": false})
```

### JSON Escape Helper (REQUIRED for complex content)

```go
import "encoding/json"

// For content with newlines, quotes, special characters:
func escapeForJS(s string) string {
    b, _ := json.Marshal(s)
    return string(b)
}

// Usage in template:
// <div data-signals={ escapeForJS(complexContent) }>
```

**Never use strings.ReplaceAll for JSON escaping** - it breaks with nested quotes and special characters.

> **⚠️ Exception: Single-quoted HTML attributes.** When `data-signals` uses single-quoted JSON (`data-signals='{"key": %s}'`), single quotes inside values break the attribute boundary. In this case, replace `'` with `&#39;`:
> ```go
> func escJSON(b []byte) []byte {
>     s := string(b)
>     s = strings.ReplaceAll(s, "'", "&#39;")
>     return []byte(s)
> }
> ```
> This is the **only** safe use of `strings.ReplaceAll` for JSON escaping.

### Two-Way Data Binding

```html
<!-- Input binding - always include name attribute! -->
<input
    type="text"
    name="username"
    data-bind="username"
    class="input input-bordered"
/>

<!-- Textarea binding -->
<textarea
    name="content"
    data-bind="content"
    class="textarea textarea-bordered"
></textarea>
```

**Critical:** The `name` attribute is required for POST data to be sent correctly.

### Form Submission with contentType: 'form'

**Frontend - Use `{contentType: 'form'}` to send form-encoded data (NOT signals):**

```html
<form data-on:submit={`@post('/api/action', {contentType: 'form'})`}>
    <input name="username" data-bind="username">
    <input name="email" type="email" data-bind="email">
    <textarea name="message"></textarea>
    <button type="submit">Send</button>
</form>
```

**Backend - Parse form data:**
```go
func handleFormSubmit(sse *datastar.ServerSentEventGenerator, w http.ResponseWriter, r *http.Request) {
    // IMPORTANT: ParseForm() is required when using contentType: 'form'
    if err := r.ParseForm(); err != nil {
        sse.MarshalAndPatchSignals(map[string]interface{}{
            "error": "Failed to parse form",
        })
        return
    }

    username := r.FormValue("username")
    email := r.FormValue("email")
    message := r.FormValue("message")

    // Process...

    sse.MarshalAndPatchSignals(map[string]interface{}{
        "isSaving": false,
        "success":  true,
    })
}
```

**Key Points:**
- `{contentType: 'form'}` sends `application/x-www-form-urlencoded` (NOT JSON signals)
- `r.ParseForm()` must be called to read form values
- Use `r.FormValue("fieldname")` to get values
- **Large text content works naturally (no escaping issues!)**

### When to Use `{contentType: 'form'}` vs Signals

| Use Case | Recommended Approach | Reason |
|----------|---------------------|--------|
| Small forms (login, search) | JSON signals | Simple, reactive |
| **Large text (prompts, descriptions)** | **`{contentType: 'form'}`** | **Avoids JSON escaping issues** |
| File uploads | `{contentType: 'form'}` | Multipart encoding |
| Complex nested objects | JSON signals | Structured data |

**Rule of thumb:** If your text content has newlines, quotes, or special characters → use `{contentType: 'form'}`.

### Form Button Placement (CRITICAL)

When using `{contentType: 'form'}`, button placement and types are crucial:

**✓ CORRECT: Submit button inside form**
```html
<form data-on:submit="@post('/api/save', {contentType: 'form'})">
    <textarea name="content"></textarea>
    <button type="submit">Save</button>
</form>
```

**✓ CORRECT: Other buttons with `type="button"`**
Use `type="button"` for buttons inside the form that should NOT submit:
```html
<form data-on:submit="@post('/api/save', {contentType: 'form'})">
    <textarea name="content"></textarea>
    <button type="button" data-on:click="@post('/api/other')">
        Other Action
    </button>
    <button type="submit">Save</button>
</form>
```

**✗ WRONG: Submit button outside form**
```html
<form data-on:submit="@post('/api/save', {contentType: 'form'})">
    <textarea name="content"></textarea>
</form>
<button data-on:click="@post('/api/save', {contentType: 'form'})">
    Save  <!-- ERROR: FetchFormNotFound -->
</button>
```

**Common Error:** `FetchFormNotFound` occurs when:
- The button using `{contentType: 'form'}` is not inside a `<form>` element
- The button is inside the form but lacks `type="submit"`

**Solution:** Always place submit buttons INSIDE the form with `type="submit"` and use `data-on:submit` on the form element.

---

## Anti-Patterns to Avoid

### ❌ DON'T: Use Signals for Large Text Content

**Problem:** Using signals to store and update large text content (like prompts, descriptions, or multi-line text) causes issues:
- Complex escaping required for newlines, quotes, and special characters
- JavaScript parser errors with unescaped content
- Fragile code that breaks with edge cases

**Wrong approach (AVOID):**
```html
<!-- DON'T DO THIS - Sets large text via signal in data-on:click -->
<button data-on:click="$prompt = 'Very long text with \n newlines and \"quotes\"...'">
    Restore Default
</button>
<textarea data-bind="prompt"></textarea>
```

### ✅ DO: Use PatchElements for Large Text

**Solution:** Use `PatchElements` via SSE to update the textarea directly, then sync the signal.

**Right approach (RECOMMENDED):**

```html
<!-- Button triggers SSE request -->
<button data-on:click="@post('/api/ui/action?action=restore_prompt&type=refine')">
    Restore Default
</button>
<textarea id="refinePrompt" name="refinePrompt" data-bind="refinePrompt"></textarea>
```

```go
// Backend handler
func (ns *NarrativeService) restoreDefaultPrompt(sse *datastar.ServerSentEventGenerator, promptType string) {
    defaultPrompt := db.DefaultPrompts.RefineDescription
    
    // 1. Patch the textarea directly via SSE (no escaping needed!)
    textareaHTML := fmt.Sprintf(`<textarea id="refinePrompt" name="refinePrompt" data-bind="refinePrompt">%s</textarea>`, 
        defaultPrompt)
    sse.PatchElements(textareaHTML, datastar.WithSelectorID("refinePrompt"), datastar.WithModeOuter())
    
    // 2. Also update the signal so form submission works correctly
    sse.MarshalAndPatchSignals(map[string]string{
        "refinePrompt": defaultPrompt,
    })
}
```

**Benefits:**
- ✅ No complex escaping required
- ✅ Newlines and special characters work naturally
- ✅ More robust and maintainable
- ✅ Follows Datastar's hypermedia principles

**Rule of thumb:**
- Use **signals** for: small state values (booleans, IDs, counters, short strings)
- Use **PatchElements** for: large text content, HTML fragments, complex data

**Signal Philosophy:**
> ⚠️ **Overusing signals typically indicates trying to manage state on the frontend.**

Datastar works best when:
- Backend is the **source of truth**
- Signals are **ephemeral** (short-lived, user interaction focused)
- State is **fetched when needed**, not pre-loaded

Don't:
- Pre-load entire objects into signals
- Use signals as a client-side database
- Store fetched data in signals for later use

---

## Backend Patterns

### Go SSE Handler Template

```go
func handleAction(w http.ResponseWriter, r *http.Request) {
    sse := datastar.NewSSE(w, r)

    // Parse form if needed
    if err := r.ParseForm(); err != nil {
        sse.MarshalAndPatchSignals(map[string]interface{}{
            "error": "Failed to parse form",
        })
        return
    }

    model := r.FormValue("model")
    prompt := r.FormValue("refinePrompt")

    // Process...

    // IMPORTANT: Always reset loading states
    sse.MarshalAndPatchSignals(map[string]interface{}{
        "isSaving": false,
        "success":  true,
    })
}
```

### SSE Compression (v1.2.1+)

**Recommended for production with large SSE payloads** (chat streams, large HTML patches, real-time updates):

```go
// Default: ServerPriority strategy (Brotli > Zstd > Gzip > Deflate)
sse := datastar.NewSSE(w, r, datastar.WithCompression())

// Explicit strategy control:
sse := datastar.NewSSE(w, r, datastar.WithCompression(
    datastar.WithServerPriority(),
    datastar.WithBrotli(),
    datastar.WithZstd(),
))

// Force specific algorithm:
sse := datastar.NewSSE(w, r, datastar.WithCompression(
    datastar.WithForced(),
    datastar.WithGzip(),
))
```

> **Note:** Default strategy changed from `ClientPriority` to `ServerPriority` in v1.2.1. `ServerPriority` lets the server choose the most efficient encoding the client supports (usually Brotli or Zstd).

### Element-Scoped View Transitions (v1.2.2+)

**Requires datastar.js v1.0.2+.** Scopes view transitions to a specific element instead of the whole document:

```go
// Scopes the view transition to #main-content
sse.PatchElements(html,
    datastar.WithUseViewTransitions(),
    datastar.WithViewTransitionSelector("#main-content"),
)

// Or with PatchElementTempl:
sse.PatchElementTempl(MyComponent(),
    datastar.WithUseViewTransitions(),
    datastar.WithViewTransitionSelector("#target"),
)
```

> The `viewTransitionSelector` CSS selector is sent in the SSE `datastar-patch-elements` event. The frontend applies `element.startViewTransition()` on that element. Only works in Chromium-based browsers with [element-scoped view transitions](https://developer.chrome.com/blog/element-scoped-view-transitions) support.

### Store vs PatchSignals

```go
// For simple signal updates, use Map:
sse.MarshalAndPatchSignals(map[string]bool{"isSaving": false})

// For nested data, use store:
// store := gabs.New()
// store.Set(true, "user", "isActive")
// sse.MarshalAndPatchSignals(store)
```

---

## SafeJSON Global Pattern (data-signals)

> ⚠️ **CRITICAL:** `templ.JSONString()` does NOT escape single quotes (`'`). When used inside `data-signals='{...}'` (single-quoted HTML attribute), a single quote in the JSON breaks the attribute boundary.

### Problem

```templ
<!-- BROKEN: value contains ' which breaks the attribute -->
<div data-signals={ templ.JSONString(map[string]string{"name": "O'Brian"}) }>
```

This renders as:
```html
<div data-signals='{"name": "O'Brian"}'>   <!-- WRONG: quote breaks here -->
```

### Solution: SafeJSON

```go
// components/shared/attrs.go
package shared

import (
    "encoding/json"
    "strings"
)

// SafeJSON marshals v to JSON and escapes single quotes for use in single-quoted HTML attributes.
// Always use this for data-signals over templ.JSONString().
func SafeJSON(v any) string {
    b, _ := json.Marshal(v)
    return strings.ReplaceAll(string(b), "'", "&#39;")
}
```

```templ
<!-- CORRECT: SafeJSON escapes single quotes -->
<div data-signals={ shared.SafeJSON(map[string]any{"count": 0, "name": "O'Brian"}) }>
```

This renders as:
```html
<!-- SafeJSON replaces ' with &#39; -->
<div data-signals='{"count":0,"name":"O&#39;Brian"}'>
```

### Rule of Thumb

| Context | Use |
|---------|-----|
| `data-signals='{...}'` (single-quoted attribute) | **SafeJSON** — always |
| `data-signals={ `\`{...}\`` }` (backtick attribute) | `templ.JSONString()` is safe |
| Any value with user input | **SafeJSON** — defensive |

**Exception:** `SafeJSON` is the **only** safe use of `strings.ReplaceAll` for JSON escaping. All other manual escaping (`ReplaceAll("\\n")`, etc.) should be replaced with `json.Marshal`.


## DaisyUI Modal + data-class Pattern

> ⚠️ **CRITICAL:** DaisyUI modals use `<dialog class="modal">` with the CSS class `modal-open` for visibility. `data-show` does NOT work with `<dialog>` elements — use `data-class` instead.

### Correct Pattern

```templ
<dialog class="modal" data-class="{'modal-open': $showModal}">
    <div class="modal-box max-w-3xl max-h-[80vh] overflow-y-auto">
        <!-- Modal header -->
        <h3 class="text-lg font-bold">Título</h3>
        <!-- Modal content -->
        <p>Conteúdo...</p>
        <!-- Modal actions -->
        <div class="modal-action">
            <button class="btn" data-on:click="$showModal = false">Fechar</button>
        </div>
    </div>
    <!-- Backdrop — click to close -->
    <div class="modal-backdrop" data-on:click="$showModal = false"></div>
</dialog>

<!-- Trigger button -->
<button class="btn" data-on:click="$showModal = true">Abrir Modal</button>
```

### Multiple Modals

Use unique signal names per modal:

```templ
<dialog class="modal" data-class="{'modal-open': $showSettingsModal}">...</dialog>
<dialog class="modal" data-class="{'modal-open': $showConfirmModal}">...</dialog>
```

### Close All Modals on Escape

```templ
data-on:keydown__window="if (evt.key === 'Escape') { $showModal = false; $showOther = false }"
```

### Why Not data-show

```templ
<!-- ❌ BROKEN: data-show on dialog does nothing -->
<dialog class="modal" data-show="$showModal">
```

DaisyUI's `<dialog class="modal">` uses CSS class `modal-open` to toggle `display: flex`/`display: none`. Since Datastar's `data-show` sets `display: none` via inline style, and DaisyUI controls display via class, they conflict. Always use `data-class` to toggle `modal-open`.


## Keyboard Shortcuts with `data-on:keydown__window`

Datastar's `data-on:*` supports the `__window` modifier to listen for events globally (on `window`), without `addEventListener`.

### Basic Pattern

```templ
<div data-on:keydown__window={`
    if (evt.key === 'Escape') {
        $showModal = false;
        $showDebug = false;
        document.activeElement?.blur();
    }
`}>
```

### Common Shortcut Combinations

```templ
<div data-on:keydown__window={`
    if (evt.key === 'Escape') {
        // Close all modals/panels
        $showModal = false;
        $showPanel = false;
        document.activeElement?.blur();
    } else if (evt.key === 'Enter' && evt.altKey) {
        // Alt+Enter: trigger voice/send action
        window.startAction();
        evt.preventDefault();
    } else if (evt.key === '/' && (evt.ctrlKey || evt.metaKey) && !evt.shiftKey) {
        // Ctrl+/ or Cmd+/: open search/shortcuts modal
        evt.preventDefault();
        $showShortcuts = true;
    }
`}>
```

### Modifier Reference

| Modifier | Description |
|----------|-------------|
| `__window` | Binds to `window` (global listener) |
| `__outside` | Fires when click occurs outside the element |
| `.10ms` | Debounce delay (e.g., `data-on:input__debounce.200ms`) |
| `__delay.300ms` | Delayed execution (used in toast lifecycle) |

### When to Use vs JavaScript

| Scenario | Recommendation |
|----------|---------------|
| Escape to close | `data-on:keydown__window` — prefer this |
| Complex key combos (Ctrl+Shift+?) | `data-on:keydown__window` — prefer this |
| Per-element key handling (Enter in input) | `data-on:keydown` on the element |
| Global shortcuts that need preventDefault logic | `data-on:keydown__window` |
| Shortcuts that depend on focus state | JavaScript `addEventListener` (rare) |


## SSE Error Dialog Pattern

Instead of inline error toasts or `alert()`, use a Datastar-driven modal controlled by backend SSE signals.

### Frontend

```templ
<!-- Error dialog — hidden by default, shown by backend -->
<dialog class="modal" data-class="{'modal-open': $showErrorDialog}">
    <div class="modal-box">
        <div class="flex items-center gap-3">
            <iconify-icon icon="material-symbols:error-outline" class="text-3xl text-error"></iconify-icon>
            <h3 class="font-bold text-lg">Erro</h3>
        </div>
        <p class="py-4 text-base-content/80" data-text="$errorMessage"></p>
        <div class="modal-action">
            <button class="btn btn-primary" data-on:click="$showErrorDialog = false">OK</button>
        </div>
    </div>
    <div class="modal-backdrop" data-on:click="$showErrorDialog = false"></div>
</dialog>
```

### Backend Helper

```go
// Handler helper — centralized error signal dispatch
func (s *Service) showError(sse *datastar.ServerSentEventGenerator, msg string) {
    sse.MarshalAndPatchSignals(map[string]interface{}{
        "showErrorDialog": true,
        "errorMessage":    msg,
    })
}

// Usage in any handler
if err != nil {
    s.showError(sse, "Falha ao salvar: " + err.Error())
    return
}
```

### Why This Pattern

| Approach | Problem |
|----------|---------|
| `alert()` | Blocking, ugly, no styling |
| Toast | Easy to miss, auto-dismisses too fast |
| Inline error in form | Only works near the form, not for global errors |
| **Error dialog** | **Modal, styled, user must acknowledge — never missed** |

### Alternative: Toast for Non-Blocking Errors

For non-critical messages ("Saved successfully"), use the toast pattern:

```go
s.showToast(sse, "Configurações salvas", "success")
```

See [toast.md](toast.md) for full toast implementation.


## Streaming Bubble Pattern

When streaming content from a backend (e.g., LLM response), show a loading bubble immediately and replace it when streaming completes.

### Frontend: Streaming Bubble Template

```templ
<!-- streaming_bubble.templ -->
templ StreamingBubble(msgID string) {
    <div id={ "streaming-" + msgID } class="chat chat-end">
        <div class="chat-bubble bg-primary text-primary-content max-w-full">
            <span class="loading loading-dots loading-sm"></span>
            <span class="text-sm opacity-80">Gerando resposta...</span>
        </div>
    </div>
}
```

### Backend: SSE Sequence

```go
// 1. Show loading bubble immediately
ns.renderAndPatch(sse, chat.StreamingBubble(msg.ID),
    datastar.WithSelectorID("chat-container"),
    datastar.WithModeAppend(),
)
sse.MarshalAndPatchSignals(map[string]interface{}{"isStreaming": true})

// 2. Start streaming (async goroutine or retry loop)
go func() {
    result, err := s.callLLM(ctx, settings, prompt)

    // 3. Replace loading bubble with final content
    finalBubble := chat.ChatBubble(chat.ChatMessageData{
        ID:      msg.ID,
        Role:    "assistant",
        Content: result,
    })
    s.renderAndPatch(sse, finalBubble,
        datastar.WithSelectorID("streaming-"+msg.ID),
        datastar.WithModeOuter(),  // Replaces the loading bubble entirely
    )
    sse.MarshalAndPatchSignals(map[string]interface{}{"isStreaming": false})
    sse.ExecuteScript("document.getElementById('chat-scroll').scrollTop = document.getElementById('chat-scroll').scrollHeight")
}()
```

### Key Points

| Step | Action | Mode |
|------|--------|------|
| 1 | Append streaming bubble | `Append` |
| 2 | Backend processes (LLM, DB, etc.) | — |
| 3 | Replace streaming bubble with final content | `Outer` |
| 4 | Reset `isStreaming` signal, scroll | `MarshalAndPatchSignals` + `ExecuteScript` |

- Use `WithModeOuter()` to **replace** the loading element completely (not `Inner` which keeps the container)
- The streaming element has a **stable ID** (`streaming-{msgID}`) so the backend can target it
- Reset signals (`isStreaming: false`) after the replacement so loading states update


## Progress Bar with `data-style:width`

Use Datastar's `data-style:*` to bind CSS properties to signals for animated progress bars.

### Frontend

```templ
<!-- Progress bar container -->
<div class="w-full bg-base-200 rounded-full h-2 overflow-hidden">
    <div id="progress-bar"
         class="h-full bg-primary rounded-full transition-all duration-300 ease-out"
         data-style:width={ fmt.Sprintf("%d%%", progress) }>
    </div>
</div>
<span class="text-xs text-base-content/60" data-text={ fmt.Sprintf("%d%%", progress) }></span>
```

### Backend

```go
// Update progress via SSE signals
sse.MarshalAndPatchSignals(map[string]interface{}{
    "progress": 65,
})

// Or use ExecuteScript for real-time updates during long operations
sse.ExecuteScript(fmt.Sprintf(
    "document.getElementById('progress-bar').style.width = '%d%%'",
    percent,
))
```

### Retry Countdown with Progress Bar

A common pattern for LLM retry feedback:

```templ
<!-- Retry countdown with progress bar -->
<div id={ "cd-container-" + id } class="rounded-lg px-3 py-2 text-sm border bg-base-200/70 border-l-4 border-l-warning">
    <div class="flex items-center gap-2 mb-1">
        <span class="loading loading-spinner loading-xs"></span>
        <span class="text-sm font-medium">{ title }</span>
        if attempt > 0 {
            <span class="badge badge-warning badge-xs">{ fmt.Sprint(attempt) + "/" + fmt.Sprint(totalRetries) }</span>
        }
    </div>
    if waitMs > 0 {
        <div class="text-xs text-base-content/60">
            Próxima tentativa em <span id={ "cd-" + id }>{ (waitMs + 999) / 1000 }</span>s
        </div>
        <div class="h-1.5 bg-base-300 rounded-full overflow-hidden mt-1.5">
            <div id={ "pg-" + id } class="h-full bg-warning/70 rounded-full" style="width: 0%"></div>
        </div>
    }
</div>
```

```javascript
// Client-side countdown (called via sse.ExecuteScript)
function startRetryCountdown(totalSeconds, countdownId, progressId) {
    var counter = document.getElementById(countdownId);
    var progress = document.getElementById(progressId);
    if (!counter || !progress) return;

    var remaining = totalSeconds;
    counter.textContent = remaining;
    progress.style.width = '0%';

    var interval = setInterval(function() {
        remaining--;
        counter.textContent = Math.max(0, remaining);
        progress.style.width = ((totalSeconds - remaining) / totalSeconds * 100) + '%';

        if (remaining <= 0) {
            clearInterval(interval);
            progress.style.width = '100%';
        }
    }, 1000);
}
```

### Backend: Trigger Countdown

```go
func retryCountdownJS(waitMs int, cdID, pgID string) string {
    sec := (waitMs + 999) / 1000
    return fmt.Sprintf("startRetryCountdown(%d, '%s', '%s')", sec, cdID, pgID)
}

// Usage in retry callback
func onRetry(attempt, totalRetries int, waitMs int) {
    s.renderAndPatch(sse, retryLoadingComponent(attempt, totalRetries, waitMs, "retry-1"),
        datastar.WithSelectorID("retry-container"),
        datastar.WithModeInner(),
    )
    if waitMs > 0 {
        sse.ExecuteScript(retryCountdownJS(waitMs, "cd-retry-1", "pg-retry-1"))
    }
}
```


## Toast Patterns

See **[toast.md](toast.md)** for the toast pattern — animated with entry/exit, icon, close button, and progress bar. Uses per-instance Datastar signals `$toast_vis_{uniqueID}` so toasts stack independently.
