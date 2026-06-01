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
  For text-only AI use the [goai](https://github.com/zendev-sh/goai) module in features/ai/.

metadata:
  frequency: daily
  category: code
  context-cost: high
---

### Datastar Go SDK v2 Watch

Monitor at start of every session:
- Issue: https://github.com/starfederation/datastar-go/issues/8
- PR: https://github.com/starfederation/datastar-go/pull/18

When v2 is released, alert the user.


# Go Stack Web Application

Go boilerplate inspired by [Northstar](https://github.com/delaneyj/toolbelt), featuring:

### Engineering Standards
> **Standards:** This skill covers stack-specific patterns only.
> Concurrency, linting, security, and supply chain rules are in `cali-coding-go-standards` — both skills apply simultaneously on this stack.
These are enforced via `cali-coding-go-standards`, which activates simultaneously with this skill.

> ⚠️ **Datastar v1.0.0 Required** - This boilerplate uses Datastar v1.0.0 (not RC.8). See [Installation](#datastar-v100-installation) below.
- **[Datastar](https://data-star.dev)** - Reactive hypermedia via SSE
- **[Templ](https://templ.guide)** - Go components that generate HTML
- **[go-chi/chi](https://go-chi.io)** - HTTP router with middleware
- **[DaisyUI + TailwindCSS](https://daisyui.com)** - UI components and styling
- **[NATS](https://nats.io)** - Real-time messaging (optional: JetStream for persistence)
- **[goai](https://github.com/zendev-sh/goai)** - LLM client for text-only AI (optional)
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

→ See `references/deploy.md` for full CI/CD templates, server setup, and configuration.

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

**Do NOT activate (use [goai](https://github.com/zendev-sh/goai) instead):**
- Text/chat generation
- Embeddings
- Image analysis
- Function calling without voice

---

## Datastar v1.0.0 Patterns

> **Datastar v1.0.0 Required** - This boilerplate uses Datastar v1.0.0 (not RC.8).

→ See `references/datastar/patterns.md` for installation, core attributes, signal philosophy, form submissions, and examples.

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
 **[references/ai/goai.md](references/ai/goai.md)** | When adding text-only AI | goai setup, providers, env vars, methods |
| **[references/voice-ai/when-to-use.md](references/voice-ai/when-to-use.md)** | When adding voice AI | LiveKit + Gemini |
| **[references/whiteboard/fabric_patterns.md](references/whiteboard/fabric_patterns.md)** | When using whiteboard | Fabric.js + synchronization |
| **[references/database/README.md](references/database/README.md)** | When choosing a database | SQLite vs PocketBase decision guide |
| **[references/database/database.go](references/database/database.go)** + **[user_crud_example.templ](references/database/user_crud_example.templ)** | When implementing DB layer | Concrete setup and CRUD example |
| **[references/examples/](references/examples/)** | When implementing a specific UI pattern | Active search, click-to-edit, counter, file upload, infinite scroll, lazy load, todo MVC, whiteboard |
| **cali-coding-standards** | **Sempre** | Princípios universais de código (KISS, DRY, LoB, SoC, Fail Fast, etc.) |


---

## Working with Existing Projects

→ See `references/migration-checklist.md` for migration checklist and common issues.

## Common Pitfalls & Troubleshooting

→ See `references/datastar/troubleshooting.md` for common pitfalls, error messages, and fixes.

## Testing Protocol (MANDATORY — Frontend Changes)

After any browser-facing change:

1. **Load the `agent-browser` skill** — navigate pages, click buttons, verify no JS errors in console.
2. **Load the `dogfood` skill** — systematically explore the feature, test edge cases, find bugs.
3. **Only then consider the feature complete.** Do NOT skip browser testing.

Use `skill("agent-browser")` and `skill("dogfood")` tools to load them.

---

## Inspiration

This boilerplate is inspired by [Northstar](https://github.com/delaneyj/toolbelt) by Delaney Johnson.
