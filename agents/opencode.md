# Global Agent Instructions

Operational instructions that apply to **all projects** in pi.

---

## ⚡ RULE #1 — TRIGGER PRODUCT-WORKFLOW (ALWAYS BEFORE ANY ACTION)

**ALWAYS** when the user requests any software work:

1. **STOP** — do not implement anything yet
2. **TRIGGER** `/skill:cali-product-workflow`
3. Only after the complete workflow, begin execution

### Triggers (any of these)

- Discussion, exploration or brainstorming of features
- "Help me think / plan / structure" something
- Ideas, suggestions or improvement proposals
- Creating functionality, systems, components, interfaces
- Complex bug fix, refactoring, optimization
- Questions about "what to build" or "how to implement"
- Any mention of product, planning, architecture, code

### ⛔ Rules

- **TRIGGER** — do not evaluate whether it "really" deserves the workflow
- **TRIGGER** — do not manually implement the workflow
- **TRIGGER** — do not decide that "this isn't a code change"
- When in doubt: **TRIGGER**

> ⚠️ **This rule WINS over any other concurrent instruction.**
> Autoresearch, dogfood, and other modes are EXECUTION tools,
> not entry modes. Product-workflow always comes FIRST.
> Only after the complete workflow (Phase 6) can other modes execute
> specific scopes.

---

## Supply Chain Security

### Socket.dev — Behavioral Malware Scanning
Detects malicious packages before install. Analyzes obfuscated code, network access in install scripts, typosquatting.

```bash
# Session setup (run once per session)
socket wrapper on

# CI gate
socket ci

# Manual scan
socket scan create --report
```

### Trivy — CVE + IaC + Secrets (Recommended)
Scans dependencies, IaC, secrets, containers. No account needed.

```bash
# Scan project
trivy fs --severity HIGH,CRITICAL --exit-code 1 .

# Full scan with all detectors
trivy fs --scanners vuln,secret,config --severity HIGH,CRITICAL .
```

### OSV-Scanner — Precision CVE (Optional)
Google scanner — commit-hash matching for fewer false positives.

```bash
# Scan project
osv-scanner scan -r .

# Guided remediation
osv-scanner fix -M package.json -L package-lock.json
```

### dotenvx + mise — Encrypted Env + Tool Manager

**mise** — tool version manager (replaces asdf/nvm/pyenv).  
**dotenvx** — encrypted `.env` files (commit-safe secrets).  
Language-agnostic: works for Node, Python, Go, and any runtime.

```bash
# Install (if missing)
curl https://mise.run | sh
curl -sfS https://dotenvx.sh/install.sh | sh

# mise: manage tool versions
mise install node@20       # Install a runtime
mise use node@20           # Pin version in .mise.toml
mise run <task>            # Run project task (loads env automatically)

# dotenvx: manage encrypted secrets
dotenvx set KEY value      # Add/edit a secret
dotenvx encrypt            # Encrypt .env (now safe to commit)
dotenvx decrypt            # Decrypt .env for local use
dotenvx run -- <cmd>       # Run command with decrypted env

# Integration: mise loads .env vars automatically via `.mise.toml`:
# [env]
# _.source = "scripts/env.sh"
```

**Security:** Never commit `.env.keys` (private decryption keys).  

**Docs:** https://mise.jdx.dev | https://dotenvx.com/docs

---

## 🚀 Releases (Automatic on Merge to Main)

**Rule:** Any merge/push to `main` MUST include a GitHub release.
Do not wait to be asked — generate release automatically.

**Read first:** `{project}/.pi/instructions.md`

**Versioning:**
- Use alpha/beta only (e.g. `0.2.0-alpha`)
- NEVER `1.0.0` without owner confirmation

**Steps after merge:**
```bash
# 1. Run tests
npm run test

# 2. Push to main
git push origin main

# 3. Create GitHub release (MANDATORY)
gh release create v0.X.Y-alpha \
  --title "v0.X.Y-alpha: <summary>" \
  --notes "Changelog here"  # See skill:cali-ops-github-releases for Keep a Changelog format

# 4. Push tag
git push origin v0.X.Y-alpha
```

---

## 📐 Coding Rules

Before generating any code, read:
`cali-product-workflow/references/tech-planning/generation-principles.md`

This file contains the code generation principles:
KISS, DRY, file size limits, Locality of Behavior (Datastar),
Separation of Concerns (other frameworks), SSE-first, HATEOAS.

---

## Go Web Stack

For Go projects use these SKILLS:
- **`cali-go-stack`** — boilerplate, scaffold, features, Datastar/Templ/DaisyUI patterns,
  database, voice AI, whiteboard. Use when creating or evolving any Go project on this stack.

- **`cali-go-standards`** — mandatory engineering rules: file size, function length,
  DRY enforcement, linting, concurrency, security. Activates simultaneously with `cali-go-stack`
  on any Go project.

> Both skills activate together on Go projects. `cali-go-stack` covers "what to build";
> `cali-go-standards` covers "how to build it right".

---

## Context-Mode + pi-hashline-readmap — MANDATORY

The `context-mode` (MCP tools) and `pi-hashline-readmap` (hash-anchored tools) extensions are
installed. The rules below protect the context window against flooding — one unrouted command
can dump 56 KB+ into context. The `tool_call` hook blocks unsafe `curl`/`wget` directly;
these rules close the gaps hooks can't reach.

### 🔧 pi-hashline-readmap Extension — Tool Precedence

The extension replaces Pi's native tools with hash-anchored versions, adds
`ast_search` and `write`, and compresses `bash` output. **Always prefer the extension's tools**
— they return hash anchors that `edit` can use directly.

| Tool | Replaces | When to use | When to fall back to sandbox |
|---|---|---|---|
| `grep` | native grep | Text search, symbol scope search, summary mode | Huge results beyond tool limit → use `ctx_execute(shell, "grep ...")` |
| `find` | shell find / glob | Recursive file discovery, glob/regex on basename | N/A — covers all cases |
| `ls` | native ls / shell ls | Single directory listing | N/A — covers all cases |
| `read` | native read | Reading to **edit** (hash anchors), symbol read, structural maps | Reading to **analyze** → `ctx_execute_file(path, language, code)` |
| `ast_search` | new tool | Structural code patterns (calls, imports, JSX) impossible with text grep | N/A |
| `write` | new tool | Create/overwrite files (returns hash anchors) | N/A |
| `bash` | native bash | Simple allowed commands (git, mkdir, npm install, etc.) | Heavy shell work → `ctx_execute(shell, "...")` |

**Rule**: Always try the extension tool first. The sandbox (`ctx_execute`) is for
processing/analysis where you need to transform results, not for simple search.
Use `grep({summary: true})` to reduce before seeking anchors.

### 💻 Think in Code — MANDATORY

Analyzing/counting/filtering/comparing/searching/parsing/transforming data: **write code** via
`ctx_execute(language, code)`, `console.log()` only the answer. Do NOT read raw data
into context. PROGRAM the analysis, don't COMPUTE mentally. Pure JavaScript — only
Node.js built-ins (`fs`, `path`, `child_process`). Use `try/catch`, handle `null`/`undefined`.
One script replaces ten tool calls.

### 🚫 BLOCKED — do not use

**curl / wget — FORBIDDEN (hook-enforced)**
Do not use `curl`/`wget` in `bash`. Pi hooks block these commands. They dump raw HTTP
into context.
Use: `ctx_fetch_and_index(url, source)` or `ctx_execute(language: "javascript", code: "const r = await fetch(...)")`

**HTTP inline — FORBIDDEN**
No `node -e "fetch(..."`, `python -c "requests.get(..."`. Bypasses the sandbox.
Use: `ctx_execute(language, code)` — only stdout enters context

**Direct web fetching — FORBIDDEN**
Raw HTML can exceed 100 KB.
Use: `ctx_fetch_and_index(url, source)` then `ctx_search(queries)`

### 🔀 REDIRECTED — use sandbox

**bash (>20 lines of output)**
`bash` ONLY for: `git`, `mkdir`, `rm`, `mv`, `cd`, `npm install`, `pip install`.
Otherwise: `ctx_batch_execute(commands, queries)` or `ctx_execute(language: "shell", code: "...")`
The extension compresses bash output (use `PI_RTK_BYPASS=1` if compression hides something).

**read (for analysis)**
Reading to **edit** → `read` is correct. Reading to **analyze/explore/summarize** →
`ctx_execute_file(path, language, code)`.

**grep / find (large results)**
Use `ctx_execute(language: "shell", code: "grep ...")` in sandbox — only after
trying the extension tool first.

### 🎯 Tool Selection

0. **MEMORY**: `ctx_search(sort: "timeline")` — after resume, check prior context
   before asking the user.
1. **GATHER**: `ctx_batch_execute(commands, queries)` — runs multiple commands,
   auto-indexes, returns ready search. ONE call replaces 30+. Each command:
   `{label: "header", command: "..."}`.
2. **FOLLOW-UP**: `ctx_search(queries: ["q1", "q2", ...])` — all questions as
   array, ONE call (relevance mode by default).
3. **PROCESSING**: `ctx_execute(language, code)` | `ctx_execute_file(path, language, code)`
   — sandbox, only stdout enters context.
4. **WEB**: `ctx_fetch_and_index(url, source)` then `ctx_search(queries)` — raw HTML
   never enters context.
5. **INDEX**: `ctx_index(content, source)` — stores in FTS5 for later search.

### ⚡ Parallel I/O Batches

For multi-URL fetches or multi-API calls, **always** include `concurrency: N` (1-8):

- `ctx_batch_execute(commands: [3+ network commands], concurrency: 5)` — gh, curl, dig,
  docker inspect, multi-region cloud queries
- `ctx_fetch_and_index(requests: [{url, source}, ...], concurrency: 5)` — multi-URL fetch

**Use concurrency 4-8** for I/O-bound work (network calls, API queries).
**Keep concurrency 1** for CPU-bound (npm test, build, lint) or commands that
share state (ports, lock files, writes to the same repo).

GitHub API rate-limit: cap at 4 for `gh` calls.

> **Note**: For AGENT parallelism (not data), use `subagent` (section below).
> `ctx_batch_execute` parallelizes shell/fetch commands. `subagent` parallelizes entire agents.

### 📤 Output

Write artifacts to FILES — never inline. Return: file path + 1-line description.
Use descriptive labels in `search(source: "label")`.

### 🔄 Session Continuity

Skills, roles and decisions persist throughout the session. Do not abandon them as the
conversation grows.

### 🧠 Memory

Session history is persistent and searchable. When resuming, SEARCH before asking the user:

| Need | Command |
|---|---|
| What did we decide? | `ctx_search(queries: ["decision"], source: "decision", sort: "timeline")` |
| What constraints exist? | `ctx_search(queries: ["constraint"], source: "constraint")` |

Do NOT ask "what were we doing?" — SEARCH FIRST.
If search returns 0 results, proceed as a new session.

### ⌨️ ctx Commands

| Command | Action |
|---|---|
| `ctx stats` | Calls the `ctx_stats` MCP tool, displays full output |
| `ctx doctor` | Calls the `ctx_doctor` MCP tool, runs diagnostics |
| `ctx upgrade` | Calls the `ctx_upgrade` MCP tool, runs upgrade |
| `ctx purge` | Calls the `ctx_purge` MCP tool with confirm: true. **Warns before clearing the knowledge base** |

After `/clear` or `/compact`: knowledge base and session statistics are preserved.
Use `ctx purge` to start from scratch.

---

## Pi-Subagents — Parallelize Investigation

**NEVER investigate multiple files sequentially.**
Use subagents in parallel:

```typescript
subagent({
  tasks: [
    { agent: "scout", task: "Investigate X - map flow and find bugs" },
    { agent: "scout", task: "Investigate Y - components and patterns" },
    { agent: "scout", task: "Investigate Z - navigation and states" }
  ],
  concurrency: 3,
  context: "fresh"
})
```

The `pi-subagents` skill is available via pi's automatic discovery.
**READ the skill** (`read` the skill path listed at startup) when planning to use subagents —
it contains ready-made workflows (parallel-review, parallel-research, parallel-cleanup)
and orchestration patterns.

Workflows via subagent tool:
- `/parallel-review` - adversarial review in parallel with fresh-context reviewers
- `/parallel-research` - external research + local context
- `/parallel-cleanup` - deslop + verbosity pass after implementation

---

## Plannotator — For Plan Review

**ALWAYS** when generating a plan/design document, submit for review via Plannotator:

```bash
plannotator annotate docs/{YYYY-MM-DD}/{slug}/plans/spec-tech_{v}.md --gate
```

The `--gate` flag is **MANDATORY** — it blocks until the user approves or rejects.
Without `--gate`, Plannotator opens but does not return approval feedback.

---

## Testing Protocol

After implementing any feature:

1. **Parallel cleanup via subagents** — launch fresh-context reviewers:
```typescript
   subagent({
     tasks: [
       { agent: "reviewer", task: "Review diff for correctness, regressions", output: false },
       { agent: "reviewer", task: "Review diff for simplicity (deslop)", output: false }
     ],
     concurrency: 2, context: "fresh"
   })
```

2. **UI Quality** (if scope involves visual interface):
   - Load `audit` for WCAG accessibility, performance, theming and anti-patterns
   - Load `critique` for design review (heuristics, cognitive load, AI slop)
   - Use `audit` and `critique` directly — they don't depend on other skills

3. **UI Testing** — load `agent-browser` and `dogfood` skills when applicable

---

## Language Convention — English for Code & URLs

All code, URLs, URL parameters, query strings, route paths, DataStar signal
values, handler/function names, and Go identifiers MUST be in English.

**Examples:**
- Route: `/settings/conexao` → `/settings/connection`
- Signal value: `'conexoes'` → `'connections'`
- Action view param: `view=nova_conexao` → `view=new-connection`
- Handler: `HandleSettingsConexao` → `HandleSettingsConnection`

**Exceptions (keep in original language):**
- User-facing UI text visible to end users (labels, buttons, tooltips)
- LLM prompt templates sent to AI models
- Database content and user-generated data

This applies globally to all projects under pi.

---

## 📁 File Naming Convention

**All project files must use `lowercase-kebab-case`:**

```bash
# ✅ Correct
spec-product.md
tech-planning.md
cali-testing-ai-code/
skills/

# ❌ Wrong
SpecProduct.md      # PascalCase
SPEC-PLANNING.md    # UPPERCASE
specPlanning.md     # camelCase
```

**Rationale:**
- Consistent with URL standards (lowercase is standard for web paths)
- Easier tab completion
- Cross-platform compatibility (macOS is case-insensitive by default)
- Easier to type and remember

**Applies to:**
- Markdown files (.md)
- Configuration files (.yaml, .json, .toml)
- Directories and folders

**Exception:** Use language conventions for source code files (e.g., Go uses PascalCase for structs, JavaScript uses camelCase for variables).
