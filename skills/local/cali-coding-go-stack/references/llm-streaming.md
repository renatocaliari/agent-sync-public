# Datastar + LLM Streaming Pattern

**Principle:** EVERY LLM call with SSE access MUST use streaming (`goai.StreamText`),
regardless of output size. The PRIMARY goal is keeping the SSE connection alive
(data flows every 80ms via throttle), NOT displaying progress.

**Bug scenario:** Without streaming, the SSE handler blocks for 30-60s while
the LLM generates a response. With no data flowing, proxies/browsers close the
connection. Datastar's automatic `@get` retry fires → duplicate LLM calls loop.

## Architecture

```
streamTextChunks(ctx, settings, promptID, prompt)
  → mask PII → goai.StreamText → returns *goai.TextStream
  ↑
  |__ StreamCompletion (patches DOM via SSE) — VISIBLE mode
  |      Supervision, client simulation, discussion
  |      Target: #bubble-{id} (inside chat bubble with bg-primary)
  |
  |__ streamTextChunks directly + MarshalAndPatchSignals — HIDDEN mode
  |      Case summary, block summary
  |      Target: #*-buffer (div display:none)
  |      Only reveals final result when stream completes
```

## Usage

**Visible (default):**
```go
ns.renderAndPatch(sse, chat.StreamingBubble(id, timestamp), ...)
response, err := StreamCompletion(ctx, sse, settings, promptID, prompt,
    "bubble-"+id,  // target: inner of the styled bubble
    true,            // markdown render
)
```

**Hidden (buffer):**
```go
response, err := StreamCompletion(ctx, sse, settings, promptID, prompt,
    "my-buffer",  // target: hidden div
    false,          // markdown: only plain text needed
)
```

## Retry Button on Error

`StreamingError` accepts `retryAction` parameter (Datastar expression):
```templ
StreamingError(id, message, timestamp, retryAction)
```
If non-empty, renders a "Try again" button firing the action.

## Error Toast Duration

Errors stay 8s (vs 3s default): implemented via `toastDuration(toastType)`.

## Conventions

| Context | Target | Markdown | Visible | Example |
|---------|--------|----------|---------|---------|
| Chat bubble | `bubble-{id}` | true | ✅ | Supervision, client, discussion |
| Panel/editor | `{target}-buffer` | false | ❌ | Case summary, block summary |
| Signal direct | `streamTextChunks`+signals | N/A | ❌ | Block summary (legacy) |

## DRY Helpers

- `streamTextChunks(ctx, settings, promptID, prompt) (*goai.TextStream, error)` —
  core: mask PII, start `goai.StreamText`, return stream object.
- `StreamCompletion(ctx, sse, settings, promptID, prompt, targetID, markdown) (string, error)` —
  wraps `streamTextChunks` + DOM patching (throttle 80ms). Returns unmasked accumulated text.
