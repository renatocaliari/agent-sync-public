---
name: cali-coding-go-standards
description: "Use this skill for any Go (Golang) backend task — writing services, APIs, CLI tools, concurrent code, or reviewing/refactoring existing Go code. Triggers on: any .go file, mention of goroutines, channels, mutexes, Go modules, or requests to build backend systems in Go. Also triggers when setting up linters, running tests, managing Go dependencies, or running the app locally during development."
---

# Go Backend Development Guidelines

## When to Use This Skill

Activate when the user works with Go backend code — writing services, APIs, CLI tools, concurrent code, or reviewing/refactoring existing Go. Also triggers on linting, testing, dependency management, or local dev tooling (Air).

## Examples

**Example 1: Review Go code for idiomatic correctness** — check error handling, context propagation, interface design, naming, and avoid `any` in business logic.

**Example 2: Set up a new Go project** — configure Air for live reload, add golangci-lint with recommended rules, install pre-commit hooks, set up CI check workflow, and enforce file/function size limits with automated enforcement.

**Example 3: Debug concurrent code** — identify the correct concurrency primitive (channel, mutex, atomic, errgroup) based on the data flow pattern, and verify no race conditions.

**Example 4: Refactor oversized code** — a file at 480 lines needs splitting before new logic; a god function at 120 lines needs extraction into focused helpers. Propose the split structure explicitly.

## Table of Contents

1. [Core Principles](#core-principles)
2. [Idiomatic Go Rules](#idiomatic-go-rules)
   - [Errors are values](#errors-are-values--treat-them-as-such)
   - [Interfaces at the consumer](#interfaces-at-the-consumer-not-the-producer)
   - [context.Context](#contextcontext-is-always-the-first-parameter)
   - [Structured logging with slog](#structured-logging-with-slog)
   - [Dependency injection](#dependency-injection-via-constructor-functions)
   - [Idiomatic naming](#idiomatic-naming)
   - [Avoid any / interface{}](#avoid-any--interface)
   - [No init() side effects](#no-init-side-effects)
   - [Table-driven tests](#table-driven-tests)
3. [Concurrency Safety](#concurrency-safety)
4. [Development Workflow](#development-workflow)
   - [Local development with Air](#local-development-with-air)
   - [Staging and production](#staging-and-production)
5. [Automated Enforcement (Hooks)](#automated-enforcement-hooks)
6. [Explicit Prohibitions](#explicit-prohibitions)
7. [Mandatory Commands](#mandatory-commands)
8. [Static Analysis & Linting](#static-analysis--linting)
9. [Supply Chain Rules](#supply-chain-rules)

---

## Core Principles

1. **Always use the latest stable Go version** — never pin to an older release unless the user explicitly requests it.
2. **Idiomatic over clever** — prefer readable, standard Go over clever tricks.
3. **Standard library first** — exhaust stdlib options before reaching for third-party packages.
4. **Every error must be handled** — never discard errors with `_` in security-sensitive or production paths.
5. **Stack context** — For projects using Datastar, Templ, NATS, or DaisyUI, use alongside `cali-coding-go-stack`.

---

## File and Function Size Limits

These limits apply to ALL Go projects, regardless of framework or stack. They are enforced via pre-commit hooks (see Automated Enforcement below).

### Maximum 500 lines per file
No `.go` file shall exceed 500 lines.
- Files approaching 400 lines MUST be split before adding new code.
- Files at 300-400 lines are yellow flags — consider splitting if the new code adds a distinct responsibility.
- Use single-responsibility: one handler group per file, one repository domain per file.
- Split patterns:
  - `handlers/chat.go` → `chat_send.go` + `chat_load.go` + `chat_types.go`
  - `db/repository.go` → `session_repo.go` + `message_repo.go` + `settings_repo.go`

### Maximum 100 lines per function
No function shall exceed 100 lines.
- Functions >80 lines SHOULD be refactored into smaller focused functions.
- Functions >100 lines MUST be refactored before adding new logic.
- Common god functions: `sendChatMessage`, `HandleChat`, large handler switch statements, `init`-like setup functions.
- When refactoring a function >80 lines, extract helpers with descriptive names — one clear job per function.

### Before adding any code — ALWAYS check and refactor first

Before writing or modifying any Go code, the agent MUST:

1. **Check file size** — if the target `.go` file is > 400 lines, PROPOSE a split before adding new code. Suggest the split structure explicitly (e.g., "this file has 430 lines, I suggest splitting into `handler_a.go`, `handler_b.go`"). Do NOT silently add lines to an oversized file.

2. **Check function size** — if the function being edited or any function in the hot path exceeds 80 lines, PROPOSE extraction of helper functions before modifying. Functions at 60-80 lines are yellow flags — note them.

3. **Check for refactoring debt** — when editing existing code, scan the file for: god functions (>80 lines), duplicate patterns, `fmt.Sprintf` with HTML tags, missing error handling. If found, propose a refactoring pass before adding new logic.

4. **Never silently grow tech debt** — if a file is 450 lines and you need to add 50 more lines, the correct action is: (a) flag the issue, (b) propose the split, (c) only after user approval, proceed with the split + new code. Adding 50 lines to a 450-line file without comment violates this standard.

## Automated Enforcement (Hooks)

These checks MUST be automated via hooks so violations are caught before commit or build, not after.

**Agent MUST auto-create these hooks.** When scaffolding or modifying a Go project, run `bash references/setup-hooks.sh` from the project root. This script handles all hook setup automatically.

### What `setup-hooks.sh` does

| Layer | Scope | Type | What it checks |
|-------|-------|------|----------------|
| pi.dev hooks | Global (user machine) | Post-build (info) | File size, lint, deadcode |
| pi.dev hooks | Global (user machine) | Pre-commit (blocking) | File size, lint, tidy |
| pi.dev hooks | Global (user machine) | Pre-push (blocking) | govulncheck |
| Git hooks | Per-project | Pre-commit (blocking) | File size, lint, tidy |
| Air `post_cmd` | Per-project | On-save (info) | Lint (fast) |
| GitHub Actions | Per-project | CI (blocking) | Tidy, size, vet, lint, tests |
| Makefile | Per-project | Manual | File size, lint |

### pi.dev Hook Integration

If pi.dev is installed (`~/.pi/agent/hook/` exists), `setup-hooks.sh` registers global hooks that run on every Go project:

- **Post-build** (`go build`): file size check, golangci-lint, deadcode — non-blocking info
- **Pre-commit** (`git commit`): file size check, golangci-lint, `go mod tidy` — **blocking** (exit code 2)
- **Pre-push** (`git push`): govulncheck — **blocking**

### Git Hook Fallback (works everywhere)

If pi.dev is not detected, standard `.githooks/pre-commit` hooks are created instead. These work on any machine regardless of pi.dev.

### Air Post-Build Hook

`setup-hooks.sh` injects `post_cmd` into `.air.toml` if it exists — runs fast lint on every save (non-blocking):

```toml
[build]
  post_cmd = ["golangci-lint run ./... --timeout 3m --fast || true"]
```

### File Sizes Created by `setup-hooks.sh`

| File | Purpose |
|------|---------|
| `.githooks/pre-commit` | Blocking git hook (file size, lint, tidy) |
| `.air.toml` (updated) | Non-blocking lint on every save |
| `.github/workflows/check.yml` | CI (tidy, size, vet, lint, tests) |
| `Makefile` (updated) | `make lint` target with file size check |

---

## Idiomatic Go Rules

### Errors are values — treat them as such

Handle errors at the call site. Never use `_` to discard errors in production code.
Always wrap with context so stack traces are readable:

```go
// ✅ Do
if err := store.Save(ctx, user); err != nil {
    return fmt.Errorf("saving user %d: %w", user.ID, err)
}

// ❌ Don't
store.Save(ctx, user)
```

**Sentinel errors** for expected conditions; `fmt.Errorf("%w")` for propagation:
```go
var ErrNotFound = errors.New("not found")

if errors.Is(err, ErrNotFound) { ... }
```

### Interfaces at the consumer, not the producer

Define interfaces where they are *used*, not where the type is defined.
Keep interfaces small — 1 to 3 methods. The standard library is the model.

```go
// ✅ Do — consumer defines what it needs
type UserLoader interface {
    LoadUser(ctx context.Context, id int) (*User, error)
}

// ❌ Don't — giant interface at the producer
type UserRepository interface {
    Load(...)
    Save(...)
    Delete(...)
    List(...)
    Search(...)
}
```

**Rule:** Prefer `io.Reader`, `io.Writer`, `fmt.Stringer` over custom interfaces when stdlib suffices.

### `context.Context` is always the first parameter

All functions that perform I/O, call external services, or may be cancelled must accept
`ctx context.Context` as the first argument. No exceptions.

```go
// ✅ Do
func (s *UserService) Load(ctx context.Context, id int) (*User, error)

// ❌ Don't
func (s *UserService) Load(id int) (*User, error)
```

Never store a context in a struct — pass it through the call chain.

### Structured logging with `slog`

Use `log/slog` (stdlib since Go 1.21). Always include structured key-value pairs.
Never use `fmt.Println` or bare `log.Printf` for application logs.

```go
// ✅ Do
slog.InfoContext(ctx, "user loaded", "user_id", id, "duration_ms", elapsed.Milliseconds())
slog.ErrorContext(ctx, "save failed", "user_id", id, "error", err)

// ❌ Don't
fmt.Printf("loaded user %d\n", id)
log.Printf("error: %v", err)
```

### Dependency injection via constructor functions

No `init()` for dependencies. No package-level vars for services.
Constructors make dependencies explicit, testable, and traceable from `main`.

```go
// ✅ Do
type UserService struct {
    store  UserStore
    logger *slog.Logger
}

func NewUserService(store UserStore, logger *slog.Logger) *UserService {
    return &UserService{store: store, logger: logger}
}

// ❌ Don't
var globalUserService = &UserService{store: defaultStore}
```

### Idiomatic naming

- **Receivers**: short, consistent, 1–2 letters — `s *Server`, `h *Handler`, not `this` or `self`
- **Acronyms**: `userID`, `httpClient`, `urlPath` — not `userId`, `HttpClient`, `UrlPath`
- **No stuttering**: `user.UserID` → `user.ID`, `auth.AuthToken` → `auth.Token`
- **Exported names**: document every exported identifier with a comment starting with the name

### Avoid `any` / `interface{}`

Use generics (Go 1.18+) or concrete types. `any` hides intent from the compiler,
static analysis, and LLMs. Use it only at true boundaries (JSON decode, plugin systems).

```go
// ✅ Do — generic constraint
func Map[T, U any](slice []T, fn func(T) U) []U { ... }

// ❌ Don't — unconstrained any in business logic
func Process(data any) any { ... }
```

### No `init()` side effects

`init()` is reserved for registering drivers and codecs only.
Never put business logic, config loading, or network calls in `init()`.
They make initialization order non-obvious and untestable.

### Table-driven tests

All tests for logic with multiple cases must use table-driven format:

```go
func TestValidate(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    bool
        wantErr bool
    }{
        {"empty string", "", false, true},
        {"valid email", "a@b.com", true, false},
        {"missing @", "invalid", false, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Validate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Fatalf("wantErr=%v, got %v", tt.wantErr, err)
            }
            if got != tt.want {
                t.Errorf("want %v, got %v", tt.want, got)
            }
        })
    }
}
```

---

## Concurrency Safety

Use this priority order — do not skip levels without justification:

| Scenario | Preferred solution |
|---|---|
| Transferring data between goroutines | `channel` |
| Shared struct fields | `sync.Mutex` or `sync.RWMutex` |
| Simple counter or boolean flag | `sync/atomic` generic types |
| Fan-out with error propagation | `golang.org/x/sync/errgroup` |

### Channel-first pattern (data ownership transfer)
```go
// coordinator owns the map; workers send updates via channel
type update struct{ key string; delta int }

func coordinator(updates <-chan update) map[string]int {
    counts := make(map[string]int)
    for u := range updates {
        counts[u.key] += u.delta
    }
    return counts
}
```

### Mutex pattern (shared struct state)
```go
type SafeCache struct {
    mu    sync.RWMutex
    items map[string]string
}

func (c *SafeCache) Set(key, val string) {
    c.mu.Lock()
    defer c.mu.Unlock() // always defer immediately after locking
    c.items[key] = val
}

func (c *SafeCache) Get(key string) (string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    v, ok := c.items[key]
    return v, ok
}
```

### Atomic pattern (counters and flags)
```go
var requestCount atomic.Uint64  // generic, no casting needed

func handler() {
    requestCount.Add(1)
}
```

### errgroup pattern (concurrent tasks with error propagation)
```go
import "golang.org/x/sync/errgroup"

g, ctx := errgroup.WithContext(context.Background())

g.Go(func() error { return fetchUser(ctx, id) })
g.Go(func() error { return fetchOrders(ctx, id) })

if err := g.Wait(); err != nil {
    return fmt.Errorf("parallel fetch failed: %w", err)
}
```

---

## Development Workflow

### Local development with Air

**Air** is the standard live reload tool. See **`references/air-guide.md`** for:
- Installation, `.air.toml` config, `.gitignore`
- Project policy (single source of truth for Air)
- Agent rules (never ask user to restart manually)
- Frontend hot reload with Templ
- Staging/production deploy commands

---

## Explicit Prohibitions

- **No concurrent map writes** without a mutex or channel coordinator — this panics at runtime.
- **No naked returns** in functions longer than ~5 lines.
- **No unclosed resources** — always `defer f.Close()` or `defer resp.Body.Close()` immediately after opening.
- **No third-party packages** for things stdlib already covers: HTTP, JSON, crypto, templating, testing.
- **No micro-packages** (string utilities, math helpers, etc.) — implement them natively.
- **Never bypass `go.sum`** — always commit it to version control; never use `GONOSUMCHECK` in production.
- **No `init()` side effects** — see Idiomatic Go Rules above.
- **No `any` in business logic** — see Idiomatic Go Rules above.
- **No context stored in structs** — always pass through the call chain.
- **No manual `go build` during a dev session** — use Air; see Development Workflow above.

---

## Mandatory Commands

### Testing — always include the race detector
```bash
go test -race ./...
```

### Staging/CI build — race instrumentation enabled
```bash
go build -race -o app ./cmd/app
```

### Vulnerability scan — run before every release
```bash
govulncheck ./...
# Install if missing:
go install golang.org/x/vuln/cmd/govulncheck@latest
```

### Dead code scan — run regularly to prevent LLM accumulation
```bash
go install golang.org/x/tools/cmd/deadcode@latest
deadcode -test ./...
```

---

## Static Analysis & Linting

See **`references/lint-config.md`** for:
- golangci-lint installation and full `.golangci.yml` config
- File size check Makefile target (500-line max per `.go` file)
- Ruleguard setup (`references/gorules.go`) for Datastar + Go LLM patterns
- Lint command: `golangci-lint run ./...`

---

## Supply Chain Rules

- Audit new dependencies with `go mod why <package>` before adding them.
- Prefer packages from `golang.org/x/...` (official extended stdlib) over unknown third parties.
- After any `go get`, run `govulncheck ./...` to catch newly introduced CVEs.
- Keep `go.mod` and `go.sum` in sync — CI must fail if they are dirty (`go mod tidy` check).

```bash
# CI check: fails if go.mod/go.sum are not tidy
go mod tidy
git diff --exit-code go.mod go.sum
```
