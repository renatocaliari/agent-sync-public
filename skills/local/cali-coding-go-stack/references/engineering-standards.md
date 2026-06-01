# Go + Datastar Engineering Standards

These rules are CRITICAL — violations block merge.

## 1. Maximum 500 Lines per File

No `.go`, `.templ`, or `.js` file shall exceed 500 lines.

- Files approaching 400 lines MUST be split before adding new code
- Use single-responsibility: one handler group per file, one repository domain per file

**Split patterns:**
- `handlers/chat.go` (god function) → `chat_send.go` + `chat_load.go` + `chat_types.go`
- `db/repository.go` (all repos) → `session_repo.go` + `message_repo.go` + `settings_repo.go`

## 2. DRY: Zero-Tolerance for SSE Boilerplate

Never write this pattern manually:
```go
var buf strings.Builder
component.Render(context.Background(), &buf)
sse.PatchElements(buf.String(), datastar.WithSelectorID("target"), datastar.WithModeInner())
```

Instead, use the shared helper:
```go
ns.renderAndPatch(sse, component, datastar.WithSelectorID("target"), datastar.WithModeInner())
```

Extract shared helpers BEFORE adding new code — never repeat patterns.

## 3. KISS: God Functions Are Forbidden

No function shall exceed 100 lines.

Functions >100 lines MUST be refactored into smaller focused functions before adding new logic.

Common god functions: `sendChatMessage`, `HandleChat`, large handler switch statements.

## 4. LoC (Locality of Behavior) for Datastar

Prefer `data-show`, `data-class`, `data-on:*` over JavaScript DOM manipulation.

Only write JavaScript for things ONLY JS can do:
- Web Speech API (microphone recording)
- Drag-to-resize (60fps pointer tracking)
- Scroll position tracking (no `data-on:scroll` in Datastar Free)
- Clipboard (Pro: `@clipboard()`, Free: JS)

For visual feedback (class toggling, show/hide): ALWAYS use Datastar signals. Never `classList.add/remove`.

Small JS helpers should be colocated in Templ files via `<script>` blocks, not separate `.js` files.

## 5. `fmt.Sprintf` for HTML Is Forbidden

All HTML rendering MUST use Templ components.

If existing `fmt.Sprintf` HTML is found, convert to Templ before adding new code.

**Why:** Go's `html/template` handles XSS escaping automatically. `fmt.Sprintf` bypasses this safety, creating XSS vulnerabilities.

**CI enforcement:** `grep -r 'fmt\.Sprintf.*<' .` must return empty.

## 6. Before Adding Any Code

Checklist:
- [ ] Is the target file close to 500 lines (>400)?
- [ ] Does the same pattern already exist as a helper?
- [ ] Does the function you're editing exceed 100 lines?
- [ ] Can `fmt.Sprintf` with HTML tags be replaced with a Templ component?

## Universal Principles

These Go-specific rules extend the universal principles from `cali-coding-standards`:

| Rule | Universal | Go Override |
|---|---|---|
| Lines per function | 50 | 100 |
| Lines per file | 400 | 500 |
| HTML in code | Allowed | `fmt.Sprintf` with HTML: BLOCKED |

For universal principles (KISS, DRY, LoB, SoC, etc.), see `/skill:cali-coding-standards`.
