---
name: cali-coding-go-stack
description: >
  [Cali] Go web applications using Datastar, Templ, DaisyUI, TailwindCSS, NATS,
  and optional features (Fabric.js whiteboard, LiveKit+Gemini voice AI, PocketBase/SQLite).
  
  Use when: starting a new Go web project OR evolving/extending an existing one that uses
  this stack. Triggers for scaffolding ("new go project", "scaffold go app", "go web boilerplate"),
  real-time work ("datastar go", "templ project", "go sse server", "go realtime app", "hypermedia go"),
  and feature evolution ("add feature to go app", "refactor go handler", "add NATS", "migrate to Templ",
  "extend boilerplate", "add auth", "add database", "add whiteboard", "add voice AI").
  Also triggers for AI/LLM with voice ("voice AI", "voice bot", "livekit gemini", "real-time voice").
  For text-only AI use the goai module in features/ai/ (included in scaffold).
---

### Datastar Go SDK v2 Watch

Monitor at start of every session:
- Issue #8 (OPEN): GetSSE/PostSSE/etc. breaking change to support action options (contentType, openWhenHidden, retry). Maintainers agreed to a v2 SDK. NO PR yet.
- PR #18 (OPEN): Refactor PatchElement* sugars to be library-agnostic. Backward-compatible wrappers preserved. Pending @delaneyj review.

When v2 is released, alert the user. Today (2026-06): v1.2.2 is the latest stable.


# Go Stack Web Application

Go boilerplate inspired by [Northstar](https://github.com/delaneyj/toolbelt), featuring:

### Engineering Standards
> **Standards:** This skill covers stack-specific patterns only.
> Concurrency, linting, security, and supply chain rules are in `cali-coding-go-standards` — both skills apply simultaneously on this stack.
> Local dev tooling (Air, `.air.toml`, Makefile dev target) is also defined in `cali-coding-go-standards`; do not duplicate that policy here.
These are enforced via `cali-coding-go-standards`, which activates simultaneously with this skill.

> ⚠️ **Datastar v1.0.2 Required** - This boilerplate uses Datastar v1.0.2 (not RC.8). See [Installation](#datastar-v102-installation) below.
- **[Datastar](https://data-star.dev)** - Reactive hypermedia via SSE
- **[Templ](https://templ.guide)** - Go components that generate HTML
- **[DaisyUI + TailwindCSS](https://daisyui.com)** - UI components and styling
- **[NATS](https://nats.io)** - Real-time messaging (optional: JetStream for persistence)
- **[Fabric.js](http://fabricjs.com)** - Collaborative whiteboard (optional)
- **[LiveKit + Gemini](https://livekit.io)** - Voice AI (optional)
- **[PocketBase](https://pocketbase.io)** - Advanced database (optional)
- **[SQLite](https://sqlite.org)** - Simple database (optional)

## When to Activate This Skill

Activate this skill when the user wants:

| Intent | Example Prompt |
|--------|----------------|
| Create Go web project from scratch | "create a new go web app" |
| Create new feature in a Go web project | "create a feature" |
| Add real-time/SSE | "add real-time updates to my Go app" |
| UI with ready-made components | "add a dashboard with Tailwind components" |
| Hypermedia-driven app | "build like Datastar/HTMX style app in Go" |
| Voice AI with LiveKit | "add a voice assistant to my app" |
| Collaborative whiteboard | "add a collaborative whiteboard" |
| Database | "add persistence with SQLite/PocketBase" |

---

## 🚨 MANDATORY: templ for ALL HTML

**Read this first:** [MANDATORY_TEMPL_USAGE.md](./references/MANDATORY_TEMPL_USAGE.md)

This project has a ZERO-TOLERANCE policy for HTML in Go source files.
- ALWAYS create `.templ` files for HTML
- NEVER use `fmt.Sprintf` with HTML tags
- NEVER use indexed format specifiers (`%[N]s`)
- Blocked by CI: `grep -r 'fmt\.Sprintf.*<'` must return empty

## Quick Start

### 1. Ask the User: Which Features Are Needed?

Not every project needs everything. Use this decision tree:

```
Go Web Project
├── Need UI?
│   ├── YES → DaisyUI (always included by default)
│   └── NO → Skip to "Need data?"
│
├── Need real-time?
│   ├── Simple Pub/Sub (NATS Core) → fire-and-forget messaging
│   ├── With persistence/history (JetStream) → streams + replay
│   └── Reactive frontend only (Datastar SSE) → no NATS needed
│
├── Need database?
│   ├── Simple (1 instance, local) → SQLite
│   ├── Advanced (multi-instance, auth, REST, realtime) → PocketBase
│   └── None → in-memory data or NATS KV
│
├── Need voice AI?
│   ├── YES → LiveKit + Gemini Live API
│   └── NO → Skip
│
├── Need whiteboard?
│   ├── YES → Fabric.js
│   └── NO → Skip
```

### 2. Decision Checklist

Before starting, confirm with the user:

```markdown
## Project Configuration

- [ ] **UI**: DaisyUI + TailwindCSS (default, always recommended)
- [ ] **Real-time messaging**: NATS Core / JetStream / None
- [ ] **Database**: SQLite / PocketBase / None
- [ ] **Voice AI**: LiveKit + Gemini / None
- [ ] **Whiteboard**: Fabric.js / None
- [ ] **Module name**: `github.com/user/projectname`
- [ ] **Deploy target**: your-server.com / other / none
```

---

## 3. Deploy & Versioning (OPTIONAL)

> ⚠️ **Only generate this section if the user confirms a deploy target.**

Ask the user: *"Este projeto será deployado em produção? Onde?"*

### Target Options

| Target | Action |
|--------|--------|
| `your-server.com` | Generate full CI/CD pipeline with ghcr.io + cron |
| `other` | Generate pipeline with placeholders (`{{SERVER_HOST}}`, `{{IMAGE_NAME}}`) |
| `none` | Skip this section entirely |

### If Deploy Target = `your-server.com`

Generate these files with **concrete values** (no placeholders):

> **Branch note**: The examples use `main` (GitHub default). Use `master` if the project's default branch is `master`.

**`.github/workflows/deploy.yml`**:
```yaml
name: Build and Publish Docker Image
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  workflow_dispatch:
permissions:
  packages: write
  contents: read
jobs:
  build-and-push:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: docker/setup-buildx-action@v3
      - run: |
          echo "${{ secrets.GITHUB_TOKEN }}" | docker login ghcr.io -u ${{ github.actor }} --password-stdin
      - id: version
        run: |
          VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
          echo "version=$VERSION" >> $GITHUB_OUTPUT
      - uses: docker/metadata-action@v5
        with:
          images: ghcr.io/{{GITHUB_REPO}}
          tags: |
            type=ref,event=branch
            type=sha
            type=raw,value=latest
      - uses: docker/build-push-action@v6
        with:
          context: .
          push: true
          platforms: linux/amd64,linux/arm64
          tags: ${{ steps.meta.outputs.tags }}
          build-args: |
            CGO_ENABLED=0
            VERSION=${{ steps.version.outputs.version }}
```

**`.github/workflows/release.yml`**:
```yaml
name: Release
on:
  push:
    branches: [main]
concurrency:
  group: release-please
  cancel-in-progress: false
permissions:
  contents: write
  pull-requests: write
jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: googleapis/release-please-action@v4
        id: release
        with:
          release-type: go
          config-file: release-please-config.json
          manifest-file: .release-please-manifest.json
          target-branch: ${{ github.ref_name }}
          token: ${{ secrets.GITHUB_TOKEN }}

      - name: Auto-merge Release PR
        if: ${{ steps.release.outputs.release_created != 'true' }}
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          PR_NUMBER="${{ fromJSON(steps.release.outputs.pr).number }}"
          if [ -n "$PR_NUMBER" ]; then
            gh pr merge --merge "$PR_NUMBER"
          fi
```

Also create `release-please-config.json`:
```json
{
  "packages": {
    ".": {
      "release-type": "go"
    }
  }
}
```

And `.release-please-manifest.json`:
```json
{
  ".": "0.0.0"
}
```

**Important**: release-please requires **conventional commits** format. Agents must use:
- `fix: descricao` for patch bumps
- `feat: descricao` for minor bumps
- `chore:`, `docs:`, etc. for no bump
- Avoid emoji prefixes like `:bug: fix:` — release-please ignores these

**⚠️ Critical `prs_created` trap**: The release-please-action outputs `prs_created` as a boolean (`true`/`false`), not a PR number. Using `${{ steps.release.outputs.prs_created }}` in shell will expand to `<<< "true"` and break. Always use `steps.release.outputs.pr` (actual PR number) instead.

**`update.sh`** (on server at `/opt/{{APP_NAME}}/update.sh`):
```bash
#!/bin/bash
set -e
IMAGE="ghcr.io/{{GITHUB_REPO}}"
CONTAINER_NAME="{{APP_NAME}}"
TOKEN_FILE="/opt/{{APP_NAME}}/.gh_token"

log() { echo "$(date): $1" | tee -a /opt/{{APP_NAME}}/update.log; }

log "Checking for updates..."
echo "$(cat $TOKEN_FILE)" | docker login ghcr.io -u ${{ GITHUB_USERNAME }} --password-stdin > /dev/null 2>&1

docker pull "$IMAGE:latest" > /dev/null 2>&1

# Compare image IDs (not manifest digests — multi-arch bug: RepoDigests != pull output)
REMOTE_ID=$(docker inspect "$IMAGE:latest" --format="{{.Id}}" 2>/dev/null || echo "")
LOCAL_ID=$(docker inspect "$CONTAINER_NAME" --format="{{.Image}}" 2>/dev/null || echo "")

if [ -z "$LOCAL_ID" ]; then
    log "Container not running, will start..."
fi

if [ "$REMOTE_ID" != "$LOCAL_ID" ]; then
    log "New version detected ($REMOTE_ID), deploying..."
    docker stop "$CONTAINER_NAME" 2>/dev/null || true
    docker rm "$CONTAINER_NAME" 2>/dev/null || true
    SESSION_SECRET=$(openssl rand -base64 32 2>/dev/null || head -c32 /dev/urandom | base64)
    docker run -d \
        --name "$CONTAINER_NAME" \
        --restart unless-stopped \
        -p 127.0.0.1:8080:8080 \
        -e SESSION_SECRET="$SESSION_SECRET" \
        -v /opt/{{APP_NAME}}/data:/app/data \
        "$IMAGE:latest"
    log "SUCCESS: Deployed new version (image: $REMOTE_ID)"
else
    log "No changes detected, skipping restart"
fi
```

**Dockerfile** (multi-stage):
- Builder: `golang:{{GOVSN}}-alpine` + `templ generate` + `go build -ldflags="-X main.Version=${VERSION}"`
- Runtime: `alpine:3.21` (no root, healthcheck via TCP port check)
- Expose: `8080`

### If Deploy Target = `other`

Same structure but use **placeholders** that the agent/user must fill:
- `{{IMAGE_NAME}}` = full image path
- `{{SERVER_HOST}}` = deploy server SSH host
- `{{APP_DIR}}` = path on server
- `{{DEPLOY_TOKEN_PATH}}` = path to ghcr.io token

### Go Configuration Variables

Go projects use environment variables directly (no `.env` by default). For production, pass via `docker run -e VAR=value`. For local dev, optionally add `github.com/joho/godotenv` if user requests `.env` support.

### Version Display in UI

To show version in the app UI, inject via ldflags at build time:
```bash
go build -ldflags="-X main.Version=${VERSION}" -o app ./cmd/web/
```

In Go code:
```go
var Version = "dev"  // set via ldflags in Dockerfile

func main() {
    cfg := config.Load()
    cfg.Version = Version
}
```

In `config.go`:
```go
type Config struct {
    // ... other fields
    Version string
}
```

Pass version to templates via page data structs:
```go
type HomePageData struct {
    Sessions []SessionCardData
    Version  string  // added to struct
}
```

---

### 4. Generated Structure

```
project/
├── cmd/web/main.go           # Entry point
├── config/
│   ├── config.go             # Configuration
│   ├── config_dev.go         # Dev config
│   └── config_prod.go        # Prod config
├── router/router.go          # Main router
├── nats/nats.go              # NATS setup
├── features/                 # Self-contained features
│   ├── common/
│   │   ├── layouts/base.templ
│   │   └── components/
│   ├── index/                # Home page
│   ├── todos/                # CRUD example (NATS KV)
│   ├── counter/              # Global vs user state
│   ├── monitor/              # System info
│   ├── sortable/             # Lit + SortableJS
│   ├── reverse/              # Streaming demo
│   ├── whiteboard/           # [OPTIONAL] Fabric.js
│   └── voice-training/       # [OPTIONAL] LiveKit + Gemini
├── web/
│   └── resources/
│       └── styles/
│           └── styles.css    # DaisyUI + Tailwind
├── go.mod
└── Taskfile.yml
```

---

## Feature Modules (Self-Contained)

Each feature in `features/<name>/` is 100% self-contained:
- Its own routes, handlers, services
- Its own static assets (via `go:embed`)
- Its own templates and components

### How to Add/Remove Features

```bash
# To add: copy the feature directory to the project
cp -r boilerplate-go/assets/scaffold/features/whiteboard myproject/features/

# To remove: delete the directory
rm -rf myproject/features/whiteboard
```

### Pattern: Embedded Assets (go:embed)

```go
// features/whiteboard/static.go
package whiteboard

import (
    "embed"
    "io/fs"
    "net/http"
)

//go:embed static/*
var staticEmbed embed.FS

func StaticFS() http.FileSystem {
    fsys, _ := fs.Sub(staticEmbed, "static")
    return http.FS(fsys)
}
```

---

## Detailed Architectural Decisions

### UI: DaisyUI (Always Recommended)

DaisyUI is ready-made TailwindCSS components. **Always recommended** for Go web projects because:
- 0 custom JavaScript for basic UI
- Consistent themes (automatic light/dark)
- Accessible components by default
- Customizable via Tailwind config

**When NOT to use DaisyUI:**
- Project needs 100% custom design (UI as differentiator)
- Team already has their own design system

### Real-time: NATS Core vs JetStream

| Need | Solution |
|------|----------|
| Simple broadcast (1→N) | NATS Core |
| Messages with history | JetStream |
| Work queues | JetStream Consumer |
| Simple Key-Value | JetStream KV |
| High performance, low latency | NATS Core |

### Database: SQLite vs PocketBase

| Criteria | SQLite | PocketBase |
|----------|--------|------------|
| Multiple instances | ❌ | ✅ |
| Built-in auth | ❌ | ✅ |
| Automatic REST API | ❌ | ✅ |
| Realtime subscriptions | ❌ | ✅ |
| Migrations | Manual | Automatic |
| Simplicity | ✅ | ✅ |
| Local data only | ✅ | ❌ |

### Voice AI: When to Activate

**Activate LiveKit + Gemini when:**
- Voice assistant/chatbot that speaks
- Real-time transcription - meetings, calls
- Visual assistants - screen sharing + voice
- Automated IVR/NPS - phone support
- Remote education - tutoring with voice

**Do NOT activate (use `features/ai/` goai module instead):**
- Text/chat generation
- Embeddings
- Image analysis
- Function calling without voice

---

## Datastar v1.0.2 Patterns

> ⚠️ **CRITICAL: This boilerplate uses Datastar v1.0.2. Read this section carefully!**

### Datastar v1.0.2 Installation

#### Option 1: CDN (Recommended for quick setup)
```html
<script type="module" src="https://cdn.jsdelivr.net/gh/starfederation/datastar@v1.0.2/bundles/datastar.js"></script>
```

#### Option 2: Self-hosted (Included in boilerplate)
The boilerplate includes datastar.js at `web/resources/static/datastar/datastar.js`.
```html
<script defer type="module" src="/static/datastar/datastar.js"></script>
```

#### Option 3: npm/deno/bun
```javascript
import 'https://cdn.jsdelivr.net/gh/starfederation/datastar@v1.0.2/bundles/datastar.js'
```

### HTML Template Setup
Your HTML template must include the Datastar script:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <script defer type="module" src="/static/datastar/datastar.js"></script>
</head>
<body>
    { children... }
</body>
</html>
```

### Core Attributes

| Attribute | Example | Purpose |
|-----------|---------|---------|
| `data-on:click` | `data-on:click="@post('/api/action')"` | Click handler |
| `data-init` | `data-init="$count = 1"` | Optional: executes expression on element init |
| `data-signals` | `data-signals="{count: 0}"` | Reactive state |
| `data-bind` | `data-bind="model"` | Two-way binding |
| `data-text` | `data-text="$count"` | Text content |
| `data-class` | `data-class="{'text-primary': $active}"` | Conditional classes |
| `data-show` | `data-show="$visible"` | Conditional visibility |

### Signal Philosophy (IMPORTANT)

> ⚠️ **Overusing signals typically indicates trying to manage state on the frontend.**

**Favor fetching current state from the backend rather than pre-loading and assuming frontend state is current.**

**Good rule of thumb:**
- ✅ Use signals for **user interactions** (e.g., toggling element visibility, accordion state)
- ✅ Use signals for **sending new state to backend** via form bindings
- ❌ Don't use signals to **store fetched data** (fetch it when needed instead)
- ❌ Don't **pre-load entire objects** into signals and manage frontend state

**Why this matters:**
- Backend is the source of truth, not the frontend
- Signals are ephemeral - they don't persist across page loads
- Pre-loaded state can become stale when data changes elsewhere

### Form Submissions

**Standard forms (small data):**
```html
<form data-on:submit={`@post('/api/action', {contentType: 'form'})`}>
    <input name="name" data-bind="name">
    <button type="submit">Submit</button>
</form>
```

**Backend handler:**
```go
func handler(w http.ResponseWriter, r *http.Request) {
    sse := datastar.NewSSE(w, r)
    // For production with large SSE payloads, enable compression:
    // sse := datastar.NewSSE(w, r, datastar.WithCompression())
    r.ParseForm() // Required for form contentType
    name := r.FormValue("name")
    // ...
}
```

> ⚠️ **Important:** Always use `{contentType: 'form'}` for form submissions to send form-encoded data, not JSON signals.

### What NOT to Use
- ❌ HTMX
- ❌ Alpine.js
- ❌ Vanilla JavaScript to replace Datastar functionality
- ❌ Other reactive frameworks

---

> ⚠️ **Read [references/datastar/patterns.md](references/datastar/patterns.md)** for complete patterns.

### Key Concepts

1. **Signals** - Reactive state in the frontend (JavaScript)
2. **SSE** - Server-Sent Events for push updates
3. **Patches** - HTML fragments updated by the server
4. **Indicators** - Visual loading states

### Example: Counter

```go
// Backend: signals and patches
type CounterSignals struct {
    Global uint32 `json:"global"`
    User   uint32 `json:"user"`
}

templ Counter(signals CounterSignals) {
    <div data-signals={ templ.JSONString(signals) }}>
        <button data-on:click={ datastar.PostSSE("/counter/increment/global") }>
            + Global
        </button>
        <span data-text="$global"></span>
    </div>
}
```

### Example: TodoMVC (NATS KV)

```go
// Backend: CRUD with NATS KV
js.Set("todos", "key", []byte(jsonData))

// Frontend: reactive signals
<div data-signals={ templ.JSONString(TodoSignals{
    Todos: []*Todo{},
    Mode:  "all",
}) }>
    <ul>
        for _, todo := range mvc.Todos {
            @TodoRow(todo)
        }
    </ul>
</div>
```

---

> 💡 **Skill complementar:** Para planejamento estratégico completo (shaping, análise de riscos,
> planejamento técnico, gates de qualidade), use a skill `cali-product-workflow` via `/skill:cali-product-workflow`.
> Ela gerencia todo o workflow de produto — este boilerplate cobre a parte de implementação.

---


## References (Progressive Disclosure)

| Reference | When to Read | What It Contains |
|-----------|--------------|------------------|
| **[references/README.md](references/README.md)** | First | Index with reading path |
| **[references/templ/rules.md](references/templ/rules.md)** | Before generating ANY HTML | Zero-tolerance rules, anti-patterns, CI enforcement for Templ |
| **[references/datastar/patterns.md](references/datastar/patterns.md)** | When using Datastar | Signals, SSE, events, indicators |
| **[references/datastar/pitfall.md](references/datastar/pitfall.md)** | When Datastar behaves unexpectedly | Known pitfalls and fixes |
| **[references/datastar/toast.md](references/datastar/toast.md)** | When adding notifications | Backend-driven toasts (zero JS, animated) |
| **[references/datastar/versus_javascript.md](references/datastar/versus_javascript.md)** | When deciding JS vs Datastar | Decision matrix per browser feature |
| **[references/daisyui/datastar-integration.md](references/daisyui/datastar-integration.md)** | When combining DaisyUI + Datastar | Integration rules and pitfalls (modal, show, signals) |
| **[DaisyUI llms.txt](https://daisyui.com/llms.txt)** | When needing any DaisyUI component | All components, class names, syntax — fetch via curl for current info |
| **[references/nats/when-to-use-jetstream.md](references/nats/when-to-use-jetstream.md)** | When configuring real-time | NATS Core vs JetStream vs KV |
| **[references/voice-ai/when-to-use.md](references/voice-ai/when-to-use.md)** | When adding voice AI | LiveKit + Gemini |
| **[references/whiteboard/fabric_patterns.md](references/whiteboard/fabric_patterns.md)** | When using whiteboard | Fabric.js + synchronization |
| **[references/database/README.md](references/database/README.md)** | When choosing a database | SQLite vs PocketBase decision guide |
| **[references/database/database.go](references/database/database.go)** + **[user_crud_example.templ](references/database/user_crud_example.templ)** | When implementing DB layer | Concrete setup and CRUD example |
| **[references/examples/](references/examples/)** | When implementing a specific UI pattern | Active search, click-to-edit, counter, file upload, infinite scroll, lazy load, todo MVC, whiteboard |
| **cali-product-workflow/references/tech-planning/generation-principles.md** | **Sempre** | Princípios de geração de código (KISS, DRY, LoB, SoC) |


---

## Working with Existing Projects

When updating an existing Go project to use Datastar v1, follow this checklist:

### Migration Checklist

1. **Datastar Script Setup**
   - Ensure datastar.js v1.0.2 is in `web/resources/static/datastar/`
   - Add to HTML: `<script defer type="module" src="/static/datastar/datastar.js"></script>`
   - `data-init` is optional (not required) - only use when you need to execute an action on page/component load

2. **Static Files Configuration**
   - Use build tags: `//go:build dev` for dev, `//go:build !dev` for prod
   - In dev: `http.FileServerFS(os.DirFS(StaticDirectoryPath))`
   - In prod: `embed.FS` + `hashfs.FileServer`
   - Test: `curl http://localhost:8080/static/datastar/datastar.js` should return 200

3. **Form Migration**
   - Add `name` attribute to all inputs/textareas
   - Add `data-bind` for two-way sync
   - Replace `data-on:input="$var = el.value"` with `data-bind="varName"`

4. **JSON Escaping**
   - Replace manual escape with `json.Marshal` approach
   - Test with special characters: newlines, quotes, unicode

5. **Visibility Patterns**
   - Replace `class="hidden"` + `data-show` with `data-show` only
   - Test tab switching

### Common Migration Issues

See [Common Pitfalls](#common-pitfalls) section below.

---

## Common Pitfalls

### 1. data-signals JSON Escaping

**❌ Don't:** Manual escaping with strings.ReplaceAll
```go
s = strings.ReplaceAll(s, "\n", "\\n")
s = strings.ReplaceAll(s, "'", "\\'")
```

**✅ Do:** Use json.Marshal
```go
func escapeForJS(s string) string {
    b, _ := json.Marshal(s)
    return string(b)
}
```

**Note:** `json.Marshal` returns the string with quotes included, so use `%s` directly in templates, not `'%s'`.

### 2. Form Data Not Sending

**❌ Don't:** Inputs without name attribute
```html
<input data-bind="model">
```

**✅ Do:** Always include name
```html
<input name="model" data-bind="model">
```

### 3. Textareas Not Syncing

**❌ Don't:** Only data-on:input
```html
<textarea data-on:input="$prompt = el.value">
```

**✅ Do:** Use data-bind
```html
<textarea name="prompt" data-bind="prompt">
```

### 4. Tabs Not Working

**❌ Don't:** Mix class="hidden" with data-show
```html
<div data-show="$tab === 'a'" class="hidden">
```

**✅ Do:** Use only data-show
```html
<div data-show="$tab === 'a'">
```

### 5. Loading State Stuck

**❌ Don't:** Forget to reset signal
```go
// Handler only saves, doesn't reset
repo.Save(data)
```

**✅ Do:** Reset signal after operation
```go
repo.Save(data)
sse.MarshalAndPatchSignals(map[string]bool{"isLoading": false})
```

---

## Troubleshooting

### "Invalid or unexpected token" in console

**Cause:** Malformed data-signals JSON
**Fix:**
1. Check json.Marshal escaping
2. View page source and validate JSON
3. Check for unescaped newlines/quotes

### "POST body empty"

**Cause:** Inputs missing name attribute
**Fix:** Add `name="fieldName"` to all form inputs

### "datastar not defined"

**Cause:** Script not loaded
**Fix:**
1. Check 404 for datastar.js
2. Verify static file serving
3. Check script path

### "Signals not updating"

**Cause:** Typically a data-bind or signal definition issue
**Fix:**
1. Ensure `data-bind` attribute is present on form elements
2. Check signal name matches between definition and usage
3. Verify server returns proper SSE response

### Handler never called

**Cause:** Wrong data-on:click syntax
**Fix:** Use `@post('/api/action')` format with quotes and leading slash

### Form data not received at backend

**Cause:** Missing `{contentType: 'form'}` in action
**Fix:** Use `@post('/api/action', {contentType: 'form'})` and call `r.ParseForm()` in handler

---

## Best Practices

1. **Self-contained features** - Each feature in its own directory
2. **Delta sync** - Send only what changed (not entire objects)
3. **KISS (SSE)** - Prefer SSE over WebSockets for one-way updates
4. **Indicators** - Always show loading states with `data-indicator`
5. **Error boundaries** - try/catch on media and storage operations
6. **View transitions** - Use `@view-transition { navigation: auto; }` for smooth navigation

### Engineering Standards (CRITICAL — Always Enforce)

This project follows strict engineering principles. Violations block merge.

#### 1. Maximum 500 lines per file
No `.go`, `.templ`, or `.js` file shall exceed 500 lines.
- Files approaching 400 lines MUST be split before adding new code.
- Use single-responsibility: one handler group per file, one repository domain per file.
- Split patterns:
  - `handlers/chat.go` (god function) → `chat_send.go` + `chat_load.go` + `chat_types.go`
  - `db/repository.go` (all repos) → `session_repo.go` + `message_repo.go` + `settings_repo.go` + `feedback_repo.go`

#### 2. DRY: Zero-tolerance for SSE boilerplate
Never write this pattern manually:
```go
var buf strings.Builder
component.Render(context.Background(), &buf)
sse.PatchElements(buf.String(), datastar.WithSelectorID("target"), datastar.WithModeInner())
```
Instead, use the shared helper on NarrativeService:
```go
ns.renderAndPatch(sse, component, datastar.WithSelectorID("target"), datastar.WithModeInner())
```
Extract shared helpers BEFORE adding new code — never repeat patterns.

#### 3. KISS: God functions are forbidden
No function shall exceed 100 lines.
Functions >100 lines MUST be refactored into smaller focused functions before adding new logic.
Common god functions: `sendChatMessage`, `HandleChat`, large handler switch statements.

#### 4. LoC (Locality of Behavior) for Datastar
Prefer `data-show`, `data-class`, `data-on:*` over JavaScript DOM manipulation.
Only write JavaScript for things ONLY JS can do:
- Web Speech API (microphone recording)
- Drag-to-resize (60fps pointer tracking)
- Scroll position tracking (no `data-on:scroll` in Datastar Free)
- Clipboard (Pro: `@clipboard()`, Free: JS)

For visual feedback (class toggling, show/hide): ALWAYS use Datastar signals. Never `classList.add/remove`.
Small JS helpers should be colocated in Templ files via `<script>` blocks, not separate `.js` files.

#### 5. `fmt.Sprintf` for HTML is forbidden
All HTML rendering MUST use Templ components. If existing `fmt.Sprintf` HTML is found, convert to Templ before adding new code.

#### 6. Before adding any code
- Check if the target file is close to 500 lines (>400)
- Check if the same pattern already exists as a helper
- Check if the function you're editing exceeds 100 lines
- Check if `fmt.Sprintf` with HTML tags can be replaced with a Templ component

### Testing Protocol (MANDATORY — Frontend Changes)

After any browser-facing change:

1. **Load the `agent-browser` skill** — navigate pages, click buttons, verify no JS errors in console.
2. **Load the `dogfood` skill** — systematically explore the feature, test edge cases, find bugs.
3. **Only then consider the feature complete.** Do NOT skip browser testing.

Use `skill("agent-browser")` and `skill("dogfood")` tools to load them.

---

## Inspiration

This boilerplate is inspired by [Northstar](https://github.com/delaneyj/toolbelt) by Delaney Johnson.
