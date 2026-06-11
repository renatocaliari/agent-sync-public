# Toast Pattern — Backend-Driven, Zero JavaScript

All toast logic lives on the backend via `data-init__delay` attributes + CSS transitions.
This is the **only** toast pattern — animated with entry/exit, icon, close button, and progress bar.

**Zero JavaScript files needed.** Datastar's `data-init` with `__delay` modifier handles the entire lifecycle declaratively.

---

## ToastAnimated: Animated Entry/Exit (Standard)

3-phase lifecycle per toast instance using `data-init__delay` + CSS transitions.
Each toast has its own signal (`$toast_vis_{uniqueID}`) so toasts stack independently.

### CSS (add to layout `<style>` or Tailwind)
```css
#toast-container {
    z-index: 50;
    position: fixed;
    right: 1rem;
    bottom: 1rem;
    display: flex;
    flex-direction: column-reverse;
    gap: 0.5rem;
}

.toast-msg {
    opacity: 0;
    margin-top: -2.5rem;
    transform: scale(0.9) translateY(50px);
    transition: margin-top 0.5s, transform 0.5s, opacity 0.5s;
}

.toast-msg.open {
    margin-top: 0.5rem;
    transform: scale(1) translateY(0);
    opacity: 1;
}
```

### Go Templ Component
```templ
package shared

import (
    "fmt"
    "time"
)

templ ToastAnimated(message string, toastType string, toastID string) {
    { signalName := fmt.Sprintf("toast_vis_%s", toastID) }
    <div id={ toastID }
        data-signals={ SafeJSON(map[string]interface{}{ signalName: false }) }
        data-class={ fmt.Sprintf(`{'open': $%s === true}`, signalName) }
        data-init__delay.10ms={ fmt.Sprintf(`$%s = true`, signalName) }
        data-init__delay.3s={ fmt.Sprintf(`$%s = false`, signalName) }
        data-init__delay.3500ms="el.remove()"
        class="toast-msg mb-2"
    >
        <div role="alert" class={ "alert shadow-lg",
            templ.KV("alert-success", toastType == "success"),
            templ.KV("alert-error", toastType == "error"),
            templ.KV("alert-warning", toastType == "warning"),
            templ.KV("alert-info", toastType == "info"),
        }>
            @ToastIcon(toastType)
            <span>{ message }</span>
            <button data-on:click="el.closest('.toast-msg').remove()"
                class="btn btn-xs btn-circle btn-ghost">✕</button>
        </div>
    </div>
}

func NewToastID() string {
    return fmt.Sprintf("toastmsg_%d", time.Now().UnixMilli())
}
```

### ToastIcon sub-component
```templ
templ ToastIcon(toastType string) {
    switch toastType {
    case "success":
        <iconify-icon icon="material-symbols:check-circle" class="text-lg"></iconify-icon>
    case "error":
        <iconify-icon icon="material-symbols:error" class="text-lg"></iconify-icon>
    case "warning":
        <iconify-icon icon="material-symbols:warning" class="text-lg"></iconify-icon>
    default:
        <iconify-icon icon="material-symbols:info" class="text-lg"></iconify-icon>
    }
}
```

### Backend Handler
```go
func (s *Service) showToast(sse *datastar.ServerSentEventGenerator, message, toastType string) {
    var buf strings.Builder
    shared.ToastAnimated(message, toastType, shared.NewToastID()).Render(context.Background(), &buf)
    sse.PatchElements(buf.String(),
        datastar.WithSelectorID("toast-container"),
        datastar.WithModePrepend(),  // newest on top
    )
}
```

### Rendered HTML
```html
<div id="toastmsg_1714512345678"
     data-signals='{"toast_vis_toastmsg_1714512345678":false}'
     data-class="{'open': $toast_vis_toastmsg_1714512345678 === true}"
     data-init__delay.10ms="$toast_vis_toastmsg_1714512345678 = true"
     data-init__delay.3000ms="$toast_vis_toastmsg_1714512345678 = false"
     data-init__delay.3500ms="el.remove()"
     class="toast-msg mb-2">
  <div role="alert" class="alert alert-success shadow-lg">
    <iconify-icon icon="material-symbols:check-circle" class="text-lg"></iconify-icon>
    <span>Configurações salvas</span>
    <button data-on:click="el.closest('.toast-msg').remove()" class="btn btn-xs btn-circle btn-ghost">✕</button>
  </div>
</div>
```

### Page Container
```html
<div id="toast-container"></div>
```

---

## Lifecycle Breakdown

| Phase | Delay | Attribute | CSS State |
|-------|-------|-----------|-----------|
| Mount | 0ms | Element appended, `$toast_vis_X` = `false` | `.toast-msg`: `opacity: 0, scale(0.9), translateY(50px), margin-top: -2.5rem` |
| Entry | 10ms | `$toast_vis_X = true` → `.open` class added | `.toast-msg.open`: `opacity: 1, scale(1), translateY(0), margin-top: 0.5rem` |
| Display | 0–3000ms | Toast visible | CSS transitions idle |
| Exit | 3000ms | `$toast_vis_X = false` → `.open` removed | Original CSS takes over, transition plays in reverse |
| Remove | 3500ms | `el.remove()` | Element gone from DOM |

The 500ms gap between exit trigger (3000ms) and remove (3500ms) ensures the CSS `transition: 0.5s` completes.

---

## Key Conventions (from Treinador project)

| Convention | Rule |
|------------|------|
| **SafeJSON** | Use `SafeJSON()` for `data-signals` — NOT `templ.JSONString()`. Escapes single quotes (`'` → `&#39;`) required for single-quoted HTML attributes. |
| **templ.KV** | Use `templ.KV("class-name", condition)` for conditional CSS classes. |
| **data-class** | Use backtick raw string: `data-class={ `{'key': $signal}` }`. For dynamic signal names use `fmt.Sprintf`. |
| **strings.Builder** | Always use a fresh `strings.Builder` per SSE fragment: `var buf strings.Builder; component.Render(ctx, &buf); sse.PatchElements(buf.String(), ...)`. |
| **data-init__delay** | Write as a single attribute: `data-init__delay.Xms="expr"` — the `__delay.Xms` modifier is part of the attribute name, NOT a separate attribute. |
| **No JS files** | Toast logic uses zero JS files. All behavior is `data-init__delay` + CSS transitions. |
| **iconify-icon** | Use `iconify-icon` with `material-symbols:*` icons consistently. |

---

## Go Templ Syntax Notes

### Attribute names with dots and underscores
Datastar modifier syntax (`data-init__delay.10ms`) is a single HTML attribute name. Go templ handles it natively:
```templ
data-init__delay.10ms="el.remove()"   // static value
data-init__delay.10ms={ fmt.Sprintf("$s = true", sig) }  // dynamic
```

### Dynamic signal names in data-class
Use `fmt.Sprintf` to interpolate the signal name into the Datastar expression:
```templ
data-class={ fmt.Sprintf(`{'open': $%s === true}`, signalName) }
```

### Dynamic signal names in data-signals
Use `SafeJSON` with a map built from the dynamic name:
```templ
data-signals={ SafeJSON(map[string]interface{}{ signalName: false }) }
```

### Variable declarations in templ
Use a script block for local variables:
```templ
{ signalName := fmt.Sprintf("toast_vis_%s", toastID) }
```
