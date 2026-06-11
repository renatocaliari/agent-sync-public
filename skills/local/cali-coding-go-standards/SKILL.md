---
name: cali-coding-go-standards
description: "Use this skill for any Go (Golang) backend task — writing services, APIs, CLI tools, concurrent code, or reviewing/refactoring existing Go code. Triggers on: any .go file, mention of goroutines, channels, mutexes, Go modules, or requests to build backend systems in Go. Also triggers when setting up linters, running tests, managing Go dependencies, or running the app locally during development."
---

# Go Backend Development Guidelines

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
5. [Explicit Prohibitions](#explicit-prohibitions)
6. [Mandatory Commands](#mandatory-commands)
7. [Static Analysis & Linting](#static-analysis--linting)
8. [Supply Chain Rules](#supply-chain-rules)

---

## Core Principles

1. **Always use the latest stable Go version** — never pin to an older release unless the user explicitly requests it.
2. **Idiomatic over clever** — prefer readable, standard Go over clever tricks.
3. **Standard library first** — exhaust stdlib options before reaching for third-party packages.
4. **Every error must be handled** — never discard errors with `_` in security-sensitive or production paths.
5. **Stack context** — For projects using Datastar, Templ, NATS, or DaisyUI, use alongside `cali-coding-go-stack`.

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

**Air** is the standard live reload tool for local development. It watches source
files, runs pre-build hooks, recompiles, and restarts the server on every save —
no manual rebuild steps needed.

#### Installation

```bash
go install github.com/air-verse/air@latest
```

#### Project policy

Use this policy for every new Go project unless the user explicitly chooses a
different local dev workflow:

- `cali-coding-go-standards` is the single source of truth for Air policy.
- `cali-coding-go-stack` must not duplicate the Air policy; it may only reference
  `cali-coding-go-standards`.
- Project `AGENTS.md` must list `air` as the default local dev command.
- `Makefile dev` must start Air.
- Commit `.air.toml` when it is the project's standard dev config.
- Ignore only `.air.toml.local` for user-specific overrides.
- Keep `.air.toml.example` only when `.air.toml` is intentionally local and not
  committed.
- Do not duplicate Air processes in `solo.yml` (`air` and `make dev` are the same
  workflow).

#### `.air.toml` — standard config (projects with Templ)

Place this file at the project root:

```toml
[build]
  pre_cmd        = ["templ generate"]
  cmd            = "go build -ldflags='-w -X main.Version=dev' -o ./tmp/app ./cmd/web"
  bin            = "./tmp/app"
  include_ext    = ["go", "templ", "yaml", "css", "js"]
  exclude_dir    = ["tmp", "vendor", "node_modules", "testdata"]
  exclude_regex  = ["_templ\\.go$"]  # prevents rebuild loop from generated files
  kill_delay     = "500ms"
  send_interrupt = true
  stop_on_error  = true

[log]
  time = true

[misc]
  clean_on_exit = true
```

For projects **without Templ**, remove the `pre_cmd` line entirely. For projects
using Go tool dependencies, `go tool templ generate` is also acceptable.

#### `.gitignore`

Always ignore Air's compiled binary output and local overrides:

```gitignore
tmp/
.air.toml.local
```

Do **not** ignore `.air.toml` when it is the committed project config.

#### Running Air

```bash
air   # run from the project root; keep the terminal open while developing
```

Air runs for the duration of a dev session. Stop it with `Ctrl+C` when done.

#### Frontend hot reload

- `.templ` changes are handled by `templ generate` in Air `pre_cmd`.
- Go changes rebuild and restart the server.
- DaisyUI/Tailwind via CDN need no local watcher.
- Local CSS/JS/static assets need an explicit watcher or include rule if edits
  must reload without manual restart.
- Do not switch to Templier for this stack unless Air cannot cover the required
  hot reload behavior.

#### Rules for the agent

- **Never** instruct the user to run `go build` + `./app` manually during a dev session.
- **Never** ask the user to restart the server manually when Air is active.
- Before suggesting a code change that affects the running server, check whether Air is running:

```bash
pgrep -x air
```

- If Air is **running**: tell the user to save the file — Air will rebuild automatically.
- If Air is **not running**: tell the user to start it with `air` or `make dev` before testing the change.

### Staging and production

Never use Air outside of local development. For deploying or restarting the app in
staging or production, follow this priority order:

1. **If the project has a deploy mechanism** — a shell script, a `Makefile` target,
   a `Dockerfile`, or a process manager config — use it. Do not improvise.

2. **If no deploy mechanism exists**, perform these steps in order:

```bash
# 1. Run templ generate if the project uses Templ
templ generate

# 2. Kill any process currently holding the port
lsof -ti:<PORT> | xargs kill -9 2>/dev/null || true
sleep 1

# 3. Build with version metadata
go build \
  -ldflags="-w \
    -X main.Version=$(git describe --tags --abbrev=0 2>/dev/null || echo dev) \
    -X main.CommitHash=$(git rev-parse --short HEAD 2>/dev/null || echo unknown) \
    -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o ./app ./cmd/web

# 4. Start detached
nohup ./app > server.log 2>&1 &
echo "PID: $!"
```

If the project does not expose `main.Version`, `main.CommitHash`, or `main.BuildTime`,
drop the corresponding `-X` flags.

| Tool | Context | Builds with ldflags | Watches files | Long-running process |
|---|---|---|---|---|
| `air` | Local dev only | ❌ (uses `dev`) | ✅ | No — terminal session |
| deploy mechanism / steps above | Staging / production | ✅ | ❌ | Yes — nohup |

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

Use `golangci-lint` with the recommended linter set below. If `golangci-lint` is not installed, install it first:

```bash
# Install (latest version, no pinning)
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

### Recommended `.golangci.yml`
```yaml
linters:
  enable:
    - errcheck        # unchecked errors
    - govet           # suspicious constructs (shadow, printf misuse, etc.)
    - staticcheck     # comprehensive static analysis (SA*, S*, QF*)
    - gosec           # security anti-patterns (hardcoded creds, weak crypto, etc.)
    - revive          # idiomatic style (golint successor)
    - gocritic        # opinionated but high-signal checks
    - ineffassign     # assignments whose value is never used
    - unused          # unexported identifiers never referenced
    - goimports       # import ordering and formatting
    - noctx           # http requests without context.Context

linters-settings:
  gosec:
    severity: medium
  revive:
    rules:
      - name: exported
      - name: var-naming
      - name: error-return
      - name: context-as-argument   # ctx must be first param
      - name: context-keys-type     # typed context keys only
```

### Run linting
```bash
golangci-lint run ./...
```

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
