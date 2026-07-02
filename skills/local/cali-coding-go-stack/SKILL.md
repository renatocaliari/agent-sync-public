---
name: cali-coding-go-stack
description: >
  [Cali] Go web stack using Datastar, Templ, DaisyUI, TailwindCSS, NATS, optional:
  Fabric.js whiteboard, LiveKit+Gemini voice AI, PocketBase/SQLite, GoAI LLM SDK,
  Zenflow declarative multi-agent workflows, goqite lightweight queue, Hatchet durable
  execution.
  Use when: starting a new Go project OR evolving an existing one.
  Triggers for scaffolding ("new go project", "scaffold go app", "go web boilerplate"),
  real-time ("datastar go", "templ project", "go sse server", "go realtime app", "hypermedia go"),
  feature evolution ("add feature", "refactor handler", "add NATS", "migrate to Templ",
  "extend boilerplate", "add auth", "add database", "add whiteboard", "add voice AI"),
  AI/LLM with voice ("voice AI", "voice bot", "livekit gemini", "real-time voice"),
  and agent/AI ("goai LLM", "zenflow agent", "hatchet workflow", "multi-agent",
  "LLM tools", "durable execution").
  For text-only AI use GoAI; for multi-agent use Zenflow; for code-based workflows use Hatchet (n8n replacement); for simple background queues use goqite.
---

### Datastar Go SDK v2 Watch

Monitor: Issue #8 (GET/POST SSE action options) and PR #18 (PatchElement* refactor) pending. Alert user if merged. Today: v1.2.2 stable.

# Go Stack Web Application

Go boilerplate inspired by [Northstar](https://github.com/delaneyj/toolbelt), featuring:

### Engineering Standards

Stack-specific patterns only. Concurrency, linting, security in `cali-coding-go-standards` (activated simultaneously).
- Local dev (Air, `.air.toml`, Makefile) also defined in `cali-coding-go-standards` — do not duplicate.
> ⚠️ **Datastar v1.0.2 Required** - This boilerplate uses Datastar v1.0.2 (not RC.8). See the Datastar section below for setup.
- **[Datastar](https://data-star.dev)** - Reactive hypermedia via SSE
- **[Templ](https://templ.guide)** - Go components that generate HTML
- **[DaisyUI + TailwindCSS](https://daisyui.com)** - UI components and styling
- **[NATS](https://nats.io)** - Real-time messaging (optional: JetStream for persistence)
- **[Fabric.js](http://fabricjs.com)** - Collaborative whiteboard (optional)
- **[LiveKit + Gemini](https://livekit.io)** - Voice AI (optional)
- **[GoAI](https://github.com/zendev-sh/goai)** - LLM SDK (tools, structured output, streaming, MCP, providers) (optional)
- **[Zenflow](https://github.com/zendev-sh/zenflow)** - Declarative multi-agent harness with YAML workflows, LLM coordinator, MCP tools, race-safe delivery (optional)
- **[Hatchet](https://hatchet.run)** - Code-based workflow engine: DAGs, retries, scheduling, monitoring (n8n replacement for code-first teams). Runs as separate server instance — connect via Go SDK. Use Hatchet Lite for local dev (optional)
- **[goqite](https://github.com/maragudk/goqite)** - Lightweight persistent queue on SQLite/PG, zero deps, SQS-like (optional, for simple background jobs; uses its own SQLite file, not PocketBase's internal DB)
- **[PocketBase](https://pocketbase.io)** - Advanced database, auth, REST API, realtime, file storage (default)
  - **[sqlc](https://github.com/sqlc-dev/sqlc)** (recommended): typed Go from checked SQL — escape hatch when PB record API/dbx too narrow (stats aggregations, multi-level joins, bulk ops, schema migrations). Install upfront: need unpredictable, zero cost when unused.
  - **SQLite driver**: `ncruces/go-sqlite3` (pure Go, wasm2go, no CGO). Registers as `"sqlite3"` — use via PocketBase `Config.DBConnect`. This is the recommended driver for this stack. The default modernc driver that PocketBase ships with is usable but ncruces provides better extension support (FTS5, spellfix1, unicode).
  - **Extensions via sqlite3.AutoExtension()** (only when using ncruces driver):
    - ext/unicode → unaccent(), collation pt_BR
    - ext/fts5 → FTS5 (required for PB EnsureFTS5Index)
    - ext/spellfix1 → fuzzy matching
    - ⚠️ ext/vec1 does NOT work with the default WASM binary
  - **Vector search**: vector math in pure Go (not SQL) — see Hybrid search strategy section below.
- **[SQLite](https://sqlite.org)** - Simple database (optional)

## When to Use This Skill

Activate this skill when the user wants:

| Intent | Example Prompt |
|--------|----------------|
| Create Go web project from scratch | "create a new go web app" |
| Create new feature in a Go web project | "create a feature" |
| Add real-time/SSE | "add real-time updates to my Go app" |
| UI with ready-made components | "add a dashboard with Tailwind components" |
| Hypermedia-driven app | "build like Datastar/HTMX style app in Go" |
| Voice AI with LiveKit | "add a voice assistant to my app" |
| LLM integration | "add LLM calls to my Go app" |
| Multi-agent orchestration | "coordinate multiple AI agents in Go" |
| Lightweight background queue | "add a queue to my SQLite Go app" |
| Code-based workflow engine | "replace n8n with Go code" |
| Collaborative whiteboard | "add a collaborative whiteboard" |
| Database | "add persistence with SQLite/PocketBase" |

---

## 🚨 MANDATORY: templ for ALL HTML

**Read this first:** [references/templ/rules.md](./references/templ/rules.md)
This project has a ZERO-TOLERANCE policy for HTML in Go source files.
- ALWAYS create `.templ` files for HTML
- NEVER use `fmt.Sprintf` with HTML tags
- NEVER use indexed format specifiers (`%[N]s`)
- Blocked by CI: `grep -r 'fmt\.Sprintf.*<'` must return empty

## Quick Start

### Examples

**Example 1: Scaffold a new Go web project** — answer the decision tree questions in the [Feature Checklist](#1-ask-the-user-which-features-are-needed) below, then follow the generated structure.
**Example 2: Add GoAI + Zenflow to an existing project** — add GoAI for LLM calls, add Zenflow for declarative YAML multi-agent workflows, add Hatchet for code-based workflow engine (replaces n8n), or add goqite for simple background queues.
**Example 3: Run a visual workflow** — define a Hatchet DAG in Go, generate a Mermaid diagram from the DAG, render it in the browser for user review, then execute with full durability.

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
│   ├── Simple (1 instance, local) → SQLite (uses ncruces/go-sqlite3)
│   ├── Advanced (multi-instance, auth, REST API, realtime, file storage) → PocketBase
│   └── None → in-memory data or NATS KV
│
├── Need hybrid search (text + vector)?
│   ├── YES → **Bleve** in-memory (`NewMemOnly`, native RRF fusion) — see references/database/bleve.md
│   └── NO → FTS5 only (sqlite unicode61 tokenizer)
│
├── Need voice AI?
│   ├── YES → LiveKit + Gemini Live API
│   └── NO → Skip
│
├── Need AI/agents?
│   ├── Simple LLM calls (chat, structured output, tools) → GoAI
│   ├── Multi-agent YAML workflows with LLM coordinator → GoAI + Zenflow
│   ├── Durable workflows (retries, queue, monitoring) → Hatchet
│   └── None → Skip
│
├── Need code-based workflows (replace n8n)?
│   ├── Complex DAGs, retries, monitoring → Hatchet
│   ├── Simple background jobs, same DB → goqite (uses its own SQLite file)
│   └── None → Skip
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
- [ ] **Hybrid search**: Bleve (native text+vector RRF) / None
- [ ] **Voice AI**: LiveKit + Gemini / None
- [ ] **AI/Agent**: GoAI / GoAI+Zenflow / GoAI+Zenflow+Hatchet / None
- [ ] **Code-based workflows**: Hatchet (n8n replacement) / goqite (simple queue) / None
- [ ] **Whiteboard**: Fabric.js / None
- [ ] **Module name**: `github.com/user/projectname`
- [ ] **Deploy target**: your-server.com / other / none

```

---

## 3. Deploy & Versioning (OPTIONAL)

> ⚠️ **Only load `references/deploy.md` if the user confirms a deploy target.**
Ask the user: *"Will this project be deployed to production? Where?"*

| Target | Action |
|--------|--------|
| `your-server.com` | Generate full CI/CD pipeline with ghcr.io + cron |
| `other` | Generate pipeline with placeholders |
| `none` | Skip entirely |

See **`references/deploy.md`** for workflow YAML, update.sh, Dockerfile, and version display patterns.

### 4. Generated Structure

See **`references/examples/`** for the complete project scaffold structure (cmd/web, config, router, nats, features/, web/).

---

## Feature Modules (Self-Contained)

Each feature in `features/<name>/` is 100% self-contained: routes, handlers, services, static assets (`go:embed`), templates.

```bash
# Add: cp -r boilerplate-go/assets/scaffold/features/<name> myproject/features/
# Remove: rm -rf myproject/features/<name>

```

See `references/examples/` for ready-made feature patterns.

---

## Detailed Architectural Decisions

### UI: DaisyUI (Always Recommended)

Ready-made TailwindCSS components. Zero JS for basic UI, consistent themes, accessible defaults. Avoid only when project needs 100% custom design or already has a design system.

### Real-time: NATS Core vs JetStream

| Need | Solution |
|------|----------|
| Simple broadcast (1→N) | NATS Core |
| Messages with history | JetStream |
| Work queues | JetStream Consumer |
| Simple Key-Value | JetStream KV |
| High performance, low latency | NATS Core |

### Database: SQLite vs PocketBase

#### Decision Guide

```
Need database?
├── Simple, local, embedded → SQLite (uses ncruces/go-sqlite3)
├── Multi-instance, auth, REST API, realtime, file storage → PocketBase
└── None → in-memory or NATS KV
```

#### Comparison

| Criteria | SQLite | PocketBase |
|----------|--------|------------|
| Engine | Embedded SQLite | SQLite |
| Multi-instance | ❌ | ✅ |
| Built-in auth | ❌ | ✅ |
| Automatic REST API | ❌ | ✅ |
| Realtime subscriptions | ❌ | ✅ |
| Migrations | Manual | Automatic |
| SQL control | Full | Limited (PB abstraction) — escape via **[sqlc](https://github.com/sqlc-dev/sqlc)** against `pocketbase.db` |
| Simplicity | ✅ | ✅ |

Two strategies via `MessageSearcher`:
1. **FTS5** (tokenizer `unicode61 remove_diacritics 2 prefix=3 4`) — accent-flexible term search. Default.
   - **Trigger gotcha (standalone FTS5 only):** use `DELETE FROM messages_fts WHERE rowid=old.rowid` — plain DELETE is the correct pattern for standalone tables (no `content=`). The FTS5 special `INSERT ... VALUES('delete', ...)` command is only needed for *contentless* tables (`content=''`) and can silently fail with `sqlite3: SQL logic error` on certain segment-tree states, rolling back the entire source-table UPDATE.
2. **Semantic embedding** (Phase 2): `multilingual-e5-small` ONNX int8, cosine similarity in pure Go.

**Embeddings (local ONNX):** Hugot v0.7.4+ via `ONNXEmbedder` in `features/<name>/embeddings/`. Uses `FeatureExtractionPipeline` pure-Go backend (gomlx, no CGO). Model: `intfloat/multilingual-e5-small` (384d). Lazy init + auto-download from HF. EmbedBatch serialized by mutex. Cosine similarity in pure Go. Dev: `FakeEmbedder`. **For long-form documents (>50KB / ~50 pages)**, `EmbedBatch` caps the sliding-window count at 16 and samples evenly via stride — keeps a 158KB article under 30s on the pure-Go backend. See `references/embeddings/README.md` for the implementation pattern.

**Pipeline for texts >512 tokens (SBD + sliding window):** `SplitSentencesPT` (pt-BR SBD, abbreviation-aware) → `ChunkSentences` (sliding window, maxByte+overlap, stride-sampled for long texts) → Hugot embed windows → `meanPoolVectors` → single 384-d vector. Callers see one embedding regardless of length. SBD is pt-BR specific — for other languages adapt the abbreviation list. Requires Hugot for inference backend.

**Why not sqlite-vec:** ext/vec1 needs custom WASM; sqlite-vec-go-bindings API removed in ncruces v0.35.1+. Solution: vector math in pure Go.

**Bleve alternative:** `github.com/blevesearch/bleve/v2` — native hybrid search (k-NN vector + full-text + RRF fusion), pure Go, zero CGO. Use in-memory mode (`NewMemOnly`) to avoid disk state sync (no split-state in backups). **Consider Bleve before** building manual FTS5+embedding+RRF — see `references/database/bleve.md`.

**Hybrid RRF (FTS5 + embedding scores merged):** Reciprocal Rank Fusion k=60 over FTS5 + embedding search ranks. Default for message search.

**Relevance strategy stack (token overlap → bi-encoder → cross-encoder rerank):** When you have a corpus of user-authored texts that should influence LLM prompts, compose three strategies in priority order with circuit-breaker fallbacks at each tier: **(1)** TokenOverlap (Title + 200 bytes Content, zero-cost lexical, always-safe) → **(2)** SemanticEmbed (bi-encoder cosine over pre-computed block embeddings, ~50–100ms) → **(3)** CrossEncoderRerank (BERT ms-marco-MiniLM-L-6-v2 reranks top-K bi-encoder candidates, ~300ms for K=10). Saves 200-400ms/prompt vs naive re-embedding; prompts never fail because of embedding. Opt-out: `ENABLE_SEMANTIC_BLOCK_RELEVANCE=false`. Why cross-encoder over LLM judge: same ONNX infra, ~30ms/pair vs 200-2000ms, no hallucination, no API cost. Full implementation + interfaces + failsafe design in `references/embeddings/README.md` under "Relevance Strategy Stack" section.

### ⚠️ Embedding Gotchas

1. **`truncateText` (MaxSeqLen):** XLMRoberta tokenizer averages ~2 chars/token for dense Portuguese (bullet lists, punctuation). `MaxSeqLen*4` overestimates and produces chunks >512 tokens → ONNX panic. **Safe cap: `MaxSeqLen*2` (1024 bytes).** `shortThreshold` = `MaxSeqLen*3` (1536 bytes) for sentence splitting.

2. **ChunkSentences overflow gap:** `ChunkSentences` only checks `nBytes + len(s) + 1 > maxBytes && nBytes > 0`. If the FIRST sentence exceeds `maxBytes`, it passes through anyway. Apply `truncateText` on each chunk after `ChunkSentences` to catch this.

3. **Embedding storage is in `user_settings`, not `model_context_blocks`:** The `ContextBlock.Embedding` field lives in the `user_settings` JSON column, not in the `model_context_blocks` ORM collection. Code that reads blocks from `model_context_blocks` and shows an `Embedded` status indicator must cross-reference `user_settings` for the truth.

4. **`filterNonEmpty` on structured output:** Free models returning `ClientTips` may produce `["","",""]` for `Paths`/`Questions` instead of `[]`. Always filter empty strings before checking `HasTips()`.

### ⚠️ Datastar Gotchas

1. **`data-signals` uses literal JSON, not K=V.** From v1.0.2, `data-signals="theme: 'light'"` (K=V syntax) fails with `GenerateExpression`. Format: `data-signals='{"theme":"light"}'` (JSON only).

2. **Initial signal values on server-rendered pages:** Use `data-signals` attribute (literal JSON) for initial values. For dynamic signals that depend on server-side data (e.g. `hasModels`), seed via the `datastar-signal-patch` CustomEvent in an inline `<script>` that waits for `datastar-ready`:
```html
<script>function __seedSigs(){if(!document.documentElement.classList.contains('ds-ready'))
{requestAnimationFrame(__seedSigs);return}document.dispatchEvent(new CustomEvent('datastar-signal-patch',
{detail:{key:val}}))}requestAnimationFrame(__seedSigs)</script>
```

3. **`renderMessageContent` for `client` messages:** Content in the DB may be raw markdown OR pre-rendered HTML. Check `looksLikeHTML()` (starts with `<p>`, `<div>`, `<br>`, `<span>`) before calling `RenderMarkdown` — goldmark omits raw HTML tags.

### AI & Agent Stack: GoAI + Zenflow + Hatchet/goqite

The stack has distinct layers for LLM, orchestration, workflow engine, and simple queues:

| Layer | Library | Purpose |
|-------|---------|---------|
| **LLM SDK** | [GoAI](https://github.com/zendev-sh/goai) | Calls, tools, structured output, streaming, MCP, multi-provider |
| **Orchestration** | [Zenflow](https://github.com/zendev-sh/zenflow) | Declarative YAML multi-agent harness, LLM coordinator, hub-and-spoke mailboxes, native MCP tools, race-safe delivery |
| **Workflow Engine** | [Hatchet](https://hatchet.run) | Code-based DAG workflows, retries, scheduling, monitoring (n8n replacement for code). **Separate server instance** — projects connect via Go SDK. |
| **Simple Queue** | [goqite](https://github.com/maragudk/goqite) | Persistent queue on SQLite/PG, SQS-like, zero non-test deps (for background jobs only). See [references/queue/goqite-patterns.md](./references/queue/goqite-patterns.md) for `ctx` rules and SSE Hub patterns. |

**GoAI + Zenflow + Hatchet/goqite** is the pragmatic choice for:
- Multi-step agents with coordination (Zenflow)
- Agentic RAG + tools (GoAI)
- Code-defined automation workflows with retries and monitoring (Hatchet — replaces n8n)
- Simple background jobs where you already use SQLite/PB (goqite)

#### Hatchet Architecture — Separate Instance

Hatchet runs as its own control plane, not embedded in your app:

- **Production:** Deploy a shared Hatchet instance on the server. All projects connect to it via the Go SDK (`hatchet.NewClient()` pointed at the server URL). The app's own data lives in PocketBase or SQLite — Hatchet manages its own Postgres for workflow state independently.
- **Local dev:** Use **Hatchet Lite** — a single Docker image bundling API, engine, and dashboard. Start with `hatchet server start` (requires Docker). For headless local mode (no dashboard, embedded Postgres), use `hatchet server start --local`.
- **App database unchanged:** Your project still uses PocketBase (or SQLite) as its data store. Hatchet is only for workflow orchestration, not app data.

#### Queue Decision Guide

```

Need code-based automation workflows (replace n8n)?
├── Complex DAGs, retries, scheduling, monitoring → Hatchet
├── Simple background jobs, same SQLite/PocketBase DB → goqite (uses its own SQLite file, not PB's internal DB — PB manages its own connection)
└── None → Skip

```

#### Zenflow Modes

| Mode | What it does | Use when |
|------|-------------|----------|
| `zenflow flow workflow.yaml` | Runs a declared YAML DAG to completion | Plan is fixed up-front; deterministic execution |
| `zenflow goal "build a thing"` | Coordinator plans + runs workflow on the fly | Plan must adapt to user input or interim results |
| `zenflow agent "<prompt>"` | Single-agent chat with optional tool loop | One-shot agent calls reusing zenflow's lifecycle |

#### LLM Decision Guide

```

Need LLM?
├── Direct calls only (chat, tool calling) → GoAI
├── Multiple coordinated agents → GoAI + Zenflow
├── Long workflows with retry/queue/monitor → GoAI + Zenflow + Hatchet
└── No LLM needed yet, but may in the future → Hatchet (start with code-based
    workflows, add GoAI+Zenflow later when LLM features are needed)

```
#### Workflow Visualization (n8n Replacement Pattern)

Hatchet is the code-based equivalent of n8n — define workflows in Go, not a visual editor. For simple background jobs (single queue, no DAGs), use goqite instead.

1. **Define workflows as Go code** — the LLM generates the Hatchet workflow DAG in Go.
   All steps, conditions, retries, and error handling are code (version-controlled, testable).

2. **Generate a visual diagram** — alongside each workflow file, generate a Mermaid
   flowchart that renders the full DAG: steps, branches, parallel forks, error paths,
   and retry policies. The user sees exactly how the workflow connects before execution.

   ```go
   // Pattern: workflow + diagram generator
   // workflow.go — Hatchet workflow definition
   // workflow_diagram.go — generates Mermaid from the same DAG
   func WorkflowMermaid(wf interface{}) string { ... }
   ```

3. **Display in the UI** — render the Mermaid diagram in the browser using
   Datastar signals and a client-side Mermaid renderer.
   Users can inspect workflow structure, step dependencies, and execution paths
   without running anything.
4. **LLM-coded, human-reviewed** — the LLM writes the workflow code and the diagram
   generator. The user reviews the visual diagram, approves the structure, then
   Hatchet executes it with full durability (retries, queues, monitoring).

5. **Start without LLM** — pure automation workflows run on Hatchet. When
   LLM steps are needed later (classification, summarization, routing), add GoAI
   to specific workflow steps without rewriting the workflow structure.

6. **Need only a simple queue?** — if your app already uses SQLite/PocketBase and
   you just need background job processing (no DAG, no monitoring), use goqite
   instead. Zero infrastructure, same DB. Upgrade to Hatchet when complexity grows.
#### Voice AI: When to Activate LiveKit + Gemini

**Activate LiveKit + Gemini when:**
- Voice assistant/chatbot that speaks
- Real-time transcription - meetings, calls
- Visual assistants - screen sharing + voice
- Automated IVR/NPS - phone support
- Remote education - tutoring with voice
**Do NOT activate (use GoAI instead):**
- Text/chat generation
- Embeddings
- Image analysis
- Function calling without voice

---

## Datastar v1.0.2 Patterns

> ⚠️ **CRITICAL: This boilerplate uses Datastar v1.0.2. See `references/datastar/patterns.md` for complete patterns.**
| Attribute | Example | Purpose |
|-----------|---------|---------|
| `data-on:click` | `data-on:click="@post('/api/action')"` | Click handler |
| `data-signals` | ``data-signals={`{"count":0}`}`` 🔴 JSON only | Reactive state |
| `data-bind` | `data-bind="model"` | Two-way binding |
| `data-text` | `data-text="$count"` | Text content |
| `data-class` | `data-class="{'text-primary': $active}"` | Conditional classes |
| `data-show` | `data-show="$visible"` | Conditional visibility |
**Key rules:** Prefer backend as source of truth. Forms need `name` + `{contentType: 'form'}`. No HTMX, Alpine.js, or other reactive frameworks.
See **`references/datastar/patterns.md`** for counters, TodoMVC, SSE, indicators.
See **`references/datastar/pitfalls.md`** for common pitfalls.
## References (Progressive Disclosure)

| Reference | When to Read | What It Contains |
|-----------|--------------|------------------|
| **[references/templ/rules.md](references/templ/rules.md)** | Before generating ANY HTML | Zero-tolerance rules, anti-patterns, CI enforcement for Templ |
| **[references/datastar/patterns.md](references/datastar/patterns.md)** | When using Datastar | Signals, SSE, events, indicators |
| **[references/datastar/pitfalls.md](references/datastar/pitfalls.md)** | When Datastar behaves unexpectedly | Known pitfalls and fixes |
| **[references/datastar/toast.md](references/datastar/toast.md)** | When adding notifications | Backend-driven toasts (zero JS, animated) |
| **[references/datastar/versus_javascript.md](references/datastar/versus_javascript.md)** | When deciding JS vs Datastar | Decision matrix per browser feature |
| **[references/daisyui/datastar-integration.md](references/daisyui/datastar-integration.md)** | When combining DaisyUI + Datastar | Integration rules and pitfalls (modal, show, signals) |
| **[DaisyUI llms.txt](https://daisyui.com/llms.txt)** | When needing any DaisyUI component | All components, class names, syntax — fetch via curl for current info |
| **[references/nats/when-to-use-jetstream.md](references/nats/when-to-use-jetstream.md)** | When configuring real-time | NATS Core vs JetStream vs KV |
| **[references/voice-ai/when-to-use.md](references/voice-ai/when-to-use.md)** | When adding voice AI | LiveKit + Gemini |
| **[references/whiteboard/fabric_patterns.md](references/whiteboard/fabric_patterns.md)** | When using whiteboard | Fabric.js + synchronization |
| **[references/pii-masking/cloakpipe.md](references/pii-masking/cloakpipe.md)** | When sending PII to LLM | CloakPipe sidecar, BR PII patterns (CPF/CNPJ/CEP/phone) |
| **[references/context-management/strategy.md](references/context-management/strategy.md)** | When context exceeds window limits | Sliding window + FTS5/BM25 hybrid strategy |
| **[references/embeddings/README.md](references/embeddings/README.md)** | When adding semantic search, relevance ranking, or long-form documents | ONNX infra + SBD + sliding window + strategy stack (token/bi-encoder/cross-encoder) + circuit breakers + long-form sampling |
| **[references/queue/goqite-patterns.md](references/queue/goqite-patterns.md)** | When using goqite (background jobs) | `ctx` rules (never nil) + SSE Hub race condition + when NOT to use |
| **[references/database/README.md](references/database/README.md)** | When choosing a database | SQLite vs PocketBase decision guide |
| **[references/database/database.go](references/database/database.go)** + **[user_crud_example.templ](references/database/user_crud_example.templ)** | When implementing DB layer | Concrete setup and CRUD example |
| **[references/examples/](references/examples/)** | When implementing a UI pattern | Active search, click-to-edit, counter, file upload, infinite scroll, lazy load, todo MVC, whiteboard |
| **[references/datastar-lint/main.go](references/datastar-lint/main.go)** | After `templ generate` | Validates Datastar attrs: typos, JSON, modifiers, Alpine/Vue detection |
| **[references/ci/docker-cache.md](references/ci/docker-cache.md)** | Setting up CI / Dockerfile | Fast Docker builds (cache mounts + GHA backend) |
| **[references/deploy.md](references/deploy.md)** | When user confirms a deploy target | CI/CD workflows, update.sh, Dockerfile, version display |
| **[references/llm-streaming.md](references/llm-streaming.md)** | When using LLM + SSE | Streaming architecture, visible/hidden modes, retry, conventions |
| **[references/troubleshooting.md](references/troubleshooting.md)** | When Datastar behaves unexpectedly | Error diagnostics and fixes |
| **cali-product-workflow/references/tech-planning/generation-principles.md** | **Always** | Code generation principles (KISS, DRY, LoB, SoC) |

---

## Working with Existing Projects

When updating an existing Go project to use Datastar v1, follow the migration checklist in **`references/datastar/patterns.md`** (Datastar Script Setup, Static Files Configuration, Form Migration, JSON Escaping, Visibility Patterns).

---

## Common Pitfalls

See **`references/datastar/pitfalls.md`** for detailed fixes. Key patterns:
| Pitfall | Fix |
|---------|-----|
| data-signals JSON escaping | Use `json.Marshal`, not `strings.ReplaceAll` |
| Form data not sending | Always add `name` attribute to inputs |
| Textareas not syncing | Use `data-bind`, not just `data-on:input` |
| Tabs not working | Use `data-show` only — no `class="hidden"` |
| Loading state stuck | Reset signal with `MarshalAndPatchSignals` |

---

## Troubleshooting

See **`references/troubleshooting.md`** for full diagnostics.
**Quick checks:**
- "Invalid token" → malformed data-signals JSON
- "POST body empty" → inputs missing `name` attribute
- "datastar not defined" → script not loaded / 404
- "Signals not updating" → missing `data-bind` or signal name mismatch
- Handler not called → use `@post('/api/action')` with quotes
- Form data not received → add `{contentType: 'form'}` and `r.ParseForm()`

---

## Best Practices

1. **Self-contained features** - Each feature in its own directory
2. **Delta sync** - Send only what changed (not entire objects)
3. **KISS (SSE)** - Prefer SSE over WebSockets for one-way updates
4. **Indicators** - Always show loading states with `data-indicator`
5. **Error boundaries** - try/catch on media and storage operations
6. **View transitions** - Use `@view-transition { navigation: auto; }` for smooth navigation
### Engineering Standards (CRITICAL — Always Enforce)

See `cali-coding-go-standards` for file (500 lines) and function (100 lines) limits.
**1. DRY: Zero-tolerance for SSE boilerplate.** Extract a shared `renderAndPatch` helper before adding new code — never repeat the `strings.Builder` + `Render` + `PatchElements` pattern.
**2. LoC (Locality of Behavior).** Prefer Datastar signals (`data-show`, `data-class`, `data-on:*`) over JS DOM manipulation. Only write JS for Web Speech API, drag-to-resize, scroll tracking, clipboard. Colocate small JS helpers in Templ `<script>` blocks.
**3. `fmt.Sprintf` for HTML is forbidden.** All HTML rendering MUST use Templ components.
**4. Before adding code**, check for existing helpers, replace `fmt.Sprintf` HTML with Templ, check file/function limits.
### Datastar + LLM Streaming Pattern (Critical for SSE Reliability)

> ⚠️ **Principle:** EVERY LLM call with SSE access MUST use streaming (`goai.StreamText`) to keep SSE alive. Without it, proxies close the connection and duplicate LLM calls loop.
See **`references/llm-streaming.md`** for full architecture and usage.
### Datastar HTML Validation (Recommended)

Run `references/datastar-lint/main.go` after `templ generate`.

```bash
cd references/datastar-lint && go install . && datastar-lint -r ./web/

```

Add to CI or Air `post_cmd`: `templ generate && datastar-lint -r ./web/`.

---

### Testing Protocol (MANDATORY — Frontend Changes)

After any browser-facing change:
1. **Load the `agent-browser` skill** — navigate pages, click buttons, verify no JS errors in console.
2. **Load the `dogfood` skill** — systematically explore the feature, test edge cases, find bugs.
3. **Only then consider the feature complete.** Do NOT skip browser testing.
Use `skill("agent-browser")` and `skill("dogfood")` tools to load them.

---

## Test Cases

### Should activate

- "Create a new Go web app with Datastar and Templ"
- "Add real-time updates to my Go project using SSE"
- "Add LLM calls with GoAI and multi-agent workflows with Zenflow"
- "Add a background queue with goqite to my SQLite Go app"
- "Set up a Hatchet workflow for durable job execution (replace n8n)"
- "Add voice AI with LiveKit to my Go app"
- "Create a new feature module in my Go web project"
- "Migrate an existing Go project to use Datastar"
- "Set up Postgres with sqlc and golang-migrate"
### Should NOT activate

- "Write a Go CLI tool" (use `cali-coding-go-standards` instead)
- "Review my Go code for bugs" (use `cali-coding-go-standards`)
- "Debug a goroutine leak" (use `cali-coding-go-standards`)
- "Set up golangci-lint configuration" (use `cali-coding-go-standards`)

---

## Inspiration

This boilerplate is inspired by [Northstar](https://github.com/delaneyj/toolbelt) by Delaney Johnson.
