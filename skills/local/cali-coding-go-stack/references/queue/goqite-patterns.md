# goqite — Specific Patterns

[goqite](https://github.com/maragudk/goqite) is the lightweight persistent queue used for background jobs in this stack. This reference covers ONLY rules not already in other skills (lint rules in `cali-coding-go-standards/gorules.go`, SSE/Datastar patterns in `references/datastar/*`).

## Rule #1: `ctx context.Context` is never `nil`

goqite forwards `ctx` to `database/sql.DB.BeginTx` via `internal/sql.InTx`. That helper dereferences `ctx` internally. Passing `nil` panics at runtime:

```
runtime error: invalid memory address or nil pointer dereference
  maragu.dev/goqite/internal/sql.InTx
  maragu.dev/goqite.(*Queue).SendAndGetID
  maragu.dev/goqite/jobs.Create
  ...your handler...
```

Use `context.Background()` when no cancellation is needed. Convention over configuration.

### UI symptoms when ctx=nil

1. Initial SSE patches (loading bubbles) reach the browser
2. Panic in `Enqueue` before the worker picks up the job
3. `panicRecover` middleware (PocketBase) absorbs it silently
4. UI stuck on loading; refresh shows only the previous message
5. `queue.db` stays empty — job was never enqueued
6. **server stderr contains the stack trace** — always inspect logs when a worker is "stuck"

## Rule #2: SSE Hub race

goqite runs the job in a goroutine separate from the HTTP handler. To stream results back:

```go
// Handler: Register BEFORE Enqueue
hub.Register(sessionID)
defer hub.Unregister(sessionID, handlerID)
// ... render loading bubbles ...
jobs.Enqueue(ctx, payload)  // safe to call now

// Worker (separate goroutine): Send events
hub.Send(sessionID, SSEEvent{Type: "chunk", ...})
hub.Send(sessionID, SSEEvent{Type: "complete", ...})
hub.Send(sessionID, SSEEvent{Type: "error", ...})
```

⚠️ **Race**: Calling `Register` AFTER `Enqueue` lets the worker send events before the handler is listening — events get dropped.

## When NOT to use goqite

- **Work queues with multiple consumers** → use NATS JetStream (see `references/nats/when-to-use-jetstream.md`)
- **Durable workflows with retries/DAGs** → use Hatchet (see main SKILL.md)
- **Fire-and-forget trivial work** with no persistence requirement → plain goroutine

## Related References

- SSE/Datastar patterns: `references/datastar/patterns.md`
- Stack decision tree: main SKILL.md (section "AI & Agent Stack")
- goqite docs upstream: https://github.com/maragudk/goqite