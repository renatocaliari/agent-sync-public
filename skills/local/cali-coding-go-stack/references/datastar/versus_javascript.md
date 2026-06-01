# JavaScript vs Datastar: Decision Guide for Boilerplate-Go

When implementing features, prefer Datastar declarative attributes. JavaScript is only justified for browser APIs, performance-critical continuous operations, or features without a Datastar primitive.

---

## ✅ Always Use Datastar (Replace JS)

### Typing Indicator / Streaming State
```html
<!-- Good: signal-based show/hide -->
<div id="typing-indicator" data-show="$isStreaming" style="display: none">
  <span class="loading loading-dots"></span>
</div>
```
```go
// Backend controls via SSE signals, NOT ExecuteScript
sse.MarshalAndPatchSignals(map[string]bool{"isStreaming": true})
// ...later...
sse.MarshalAndPatchSignals(map[string]bool{"isStreaming": false})
```

### Password Visibility Toggle
```html
<!-- Good: signal + data-attr:type + conditional icons -->
<input data-attr:type="$showPassword ? 'text' : 'password'" data-bind="password"/>
<button data-on:click="$showPassword = !$showPassword">
  <iconify-icon data-show="!$showPassword" icon="material-symbols:visibility"></iconify-icon>
  <iconify-icon data-show="$showPassword" icon="material-symbols:visibility-off"></iconify-icon>
</button>
```

### Setting/Focusing an Input After Action
```html
<!-- Good: data-ref creates a signal reference to the DOM element -->
<input id="chat-input" data-bind="message" data-ref="chatInputRef"/>
<!-- In any data-on:click expression elsewhere: -->
<button data-on:click='$message = "hello"; $chatInputRef.focus()'>Use</button>
```
`data-ref` creates a global signal (`$chatInputRef`) pointing to the element. Works across components, templates, and even Go-generated HTML fragments.

### Toggle State Without JS
```html
<!-- Good: signal toggle -->
<button data-on:click="$isOpen = !$isOpen">
  <iconify-icon data-show="$isOpen" icon="..."></iconify-icon>
  <iconify-icon data-show="!$isOpen" icon="..."></iconify-icon>
</button>
```

---

## ✅ Use `sse.ExecuteScript()` (Official Datastar Pattern)

These are documented as **first-class Datastar patterns**, not workarounds. The `ExecuteScript` sends a `<script>` tag via SSE's `PatchElements`.

### Scroll Container to Bottom
```go
// Acceptable — no declarative scroll primitive in Datastar Free
sse.ExecuteScript("document.getElementById('chat-container').scrollTop = document.getElementById('chat-container').scrollHeight")
```

### Scroll Container to Top
```go
sse.ExecuteScript("document.getElementById('feedback-content').scrollTop = 0")
```

### Delayed Scroll (after render)
```go
sse.ExecuteScript("setTimeout(function(){var c=document.getElementById('chat-container');if(c)c.scrollTop=c.scrollHeight},50)")
```

### Focus Input from Backend
```go
// Only when the target element is in a different SSE fragment
sse.ExecuteScript("document.getElementById('message-input').focus()")
```
When the input is in the same template/page, prefer `data-ref`.

---

## ✅ Use `sse.Redirect()` (Official Datastar SDK)

The SDK provides `sse.Redirect(url)` and `sse.Redirectf(format, args...)` — these wrap `setTimeout(() => window.location.href = url)` to handle Firefox's script-triggered redirect issue.

```go
// Good — official Datastar redirect pattern
sse.Redirectf("/chat?session=%s", session.ID)
sse.Redirect("/api/export/" + sessionID)
```
Do NOT use raw `ExecuteScript("window.location.href = ...")` — use the SDK method.

---

## ❌ Justified JavaScript (No Datastar Equivalent)

### Speech Recognition (Microphone)
```js
// Web Speech API — inherently a browser JS API
const recognition = new (window.SpeechRecognition || window.webkitSpeechRecognition)();
recognition.continuous = true;
recognition.lang = 'pt-BR';
recognition.onresult = (event) => { /* update input */ };
```
Datastar has no wrapper for this. **Keep JS.**

**Visual feedback refinement**: The feedback UI (`btn-error`, `animate-pulse` classes) called from `startVisualFeedback()`/`stopVisualFeedback()` CAN use Datastar signals. Replace DOM class manipulation with `window.Datastar.patchSignals({isRecording: true/false})` and use `data-class`/`data-show` in templates.

### Panel Drag-to-Resize
```js
// Continuous pointer tracking — Datastar's reactive system is too slow for 60fps drag
handle.addEventListener('mousedown', startDrag);
document.addEventListener('mousemove', onDrag);
document.addEventListener('mouseup', stopDrag);
```
**Keep JS.** Datastar events fire per interaction, not per animation frame. Direct DOM manipulation is correct here.

### Scroll Pills (Position Indicator)
```js
// No data-on:scroll in Datastar Free
container.addEventListener('scroll', updatePillPosition);
```
**Keep JS.** Datastar Free has no scroll event listener. The `data-on-raf` and `data-scroll-into-view` attributes are Pro-only.

### Insert Text at Cursor Position (Textarea)
```html
<!-- Small helper function serving multiple textareas — cleaner as JS -->
<script>
function insertVar(textareaId, text) {
  var ta = document.getElementById(textareaId);
  if (!ta) return;
  var start = ta.selectionStart;
  var end = ta.selectionEnd;
  ta.value = ta.value.substring(0, start) + text + ta.value.substring(end);
  ta.selectionStart = ta.selectionEnd = start + text.length;
  ta.focus();
  ta.dispatchEvent(new Event('input', {bubbles: true}));
}
</script>
```
**Keep JS.** Doing this inline in `data-on:click` would create unreadably long expressions, especially since it serves multiple textareas.

### Clipboard Copy
```js
// @clipboard() is Pro-only. Free alternative:
navigator.clipboard.writeText(document.getElementById('debug-panel-0').textContent)
```
**Keep JS** or purchase Datastar Pro for `@clipboard()` action.

### Page Reload
```js
// No Datastar equivalent
location.reload()
```
**Keep JS** (`sse.ExecuteScript("location.reload()")`).

### Toast Auto-Remove via Timer
```html
<!-- Simple: single data-init__delay for auto-remove -->
<div class="alert alert-success" data-init__delay.3s="el.remove()">
  <span>Mensagem salva</span>
</div>
```
For animated toasts with entry/exit transitions, see [toast.md](./datastar/toast.md).
**PREFER this Datastar approach** over `setTimeout` + `getElementById`. Only one line, self-contained, no JS file needed.

---

## Decision Matrix

| Feature | Use | Why |
|---------|-----|-----|
| Show/hide elements | `data-show` / `data-class` | Declarative, reactive |
| DaisyUI modals | `data-class="{'modal-open': $x}"` | `data-show` doesn't work on `<dialog>` |
| Keyboard shortcuts | `data-on:keydown__window` | Global listener, no addEventListener |
| Error dialogs | Signals + `data-class` modal | Backend-driven, user must acknowledge |
| Loading → result swap | `PatchElements` + `WithModeOuter()` | Append loading, replace when done |
| Progress bars | `data-style:width` + signals | Animated CSS transitions |
| Global JSON escaping | `SafeJSON()` helper | Only safe way for `data-signals='{...}'` |
| Dynamic attributes | `data-attr:*` | Type, disabled, aria, etc. |
| Two-way form binding | `data-bind` | Auto sync signal ↔ element |
| Toggle state | `data-on:click="$x = !$x"` | Pure signals |
| Focus an input | `data-ref` + `$ref.focus()` | Zero JS |
| Redirect page | `sse.Redirect()` / `sse.Redirectf()` | Official SDK method |
| Scroll container | `sse.ExecuteScript("...scrollTop = ...")` | No declarative primitive |
| Speech/Mic | JS file | Web Speech API |
| Drag resize | JS file | Performance-critical |
| Scroll position tracking | JS file | No `data-on:scroll` |
| Clipboard | JS (or Pro) | `@clipboard()` is Pro |
| Textarea cursor insert | JS helper | Multiple textareas, complex |
| Page reload | `sse.ExecuteScript("location.reload()")` | No equivalent |
| Timer-based remove | `data-init__delay.Xs="el.remove()"` | Declarative, no JS |

## General Rule

If you need to **reactively show/hide, toggle, bind, or style** an element → **Datastar attribute**.

If you need to **interact with a browser API** (Speech, Clipboard) or **track continuous events** (scroll, drag, pointer move) → **JavaScript**, preferably wrapped in small, well-named functions.

`sse.ExecuteScript()` is NOT an anti-pattern — but prefer signals + `data-*` attributes for UI state (class toggles, visibility, text content) and reserve `ExecuteScript` for browser API calls and scroll manipulation.
