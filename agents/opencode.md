# AGENTS.md

## Cymbal

At session start, run `cymbal hook remind --update=if-stale` and treat its output as persistent navigation guidance. Prefer cymbal over Read/Grep/Glob/Bash for code navigation (symbol search, refs, impact, trace, etc.). Falls back to fff tools (fffind/ffgrep) for pure file or content search only — never raw `find`/`grep` from system.

See **`cali-code-navigation`** skill for full workflow.

## Code Navigation

Tool priority rules, Cymbal workflow, batch guidance for token efficiency, semantic diff commands — see **`cali-code-navigation`** skill.

> **Prefer `sem diff` over `git diff`** for entity-level change detection (functions, types, methods). `git diff` shows raw lines; `sem diff` shows meaningful symbols. If `sem` is unavailable, fall back to `git diff` — never block on its absence.

> **Active loading:** `cali-code-navigation` is referenced here but NOT auto-loaded. Load it explicitly when navigating code: use the `skill` tool with `name: cali-code-navigation`.

### Cymbal Commands Quick Reference

| Goal | Command |
|------|---------|
| Adaptive exploration | `cymbal investigate <symbol>` |
| Find symbol | `cymbal search <name>` |
| Read source | `cymbal show <symbol>` |
| File outline | `cymbal outline <file>` |
| References | `cymbal refs <symbol>` |
| Upstream impact | `cymbal impact <symbol>` |
| Downstream trace | `cymbal trace <symbol>` |
| Implementations | `cymbal impls <iface>` |
| Diff-to-impact | `cymbal changed` (unstaged) / `--staged` |
| Repo map | `cymbal structure` |
| Import fan-in | `cymbal importers <pkg>` |
| Graph topology (add `--graph` to trace/impact/impls/importers) |

---

## ast-grep (Structural Code Search)

`ast-grep` (`sg`) is a structural (AST-based) search/lint/rewrite tool. Installed
via Homebrew. **Different modality from fff and cymbal:**

| Tool | Layer | Answers |
|------|-------|---------|
| `fff` (pi-fff) | **Lexical** | "Where does this **text** appear?" — fuzzy file find, SIMD grep, frecency |
| `ast-grep` | **Structural** | "Where does this **code pattern** appear?" — AST pattern match, language-aware |
| `cymbal` | **Graph/Semantic** | "Who calls this **symbol**?" — references, impact, call graph |

**No overlap or conflict.** Use priority:
1. **Text/filename** → `fffind` / `ffgrep` (fff)
2. **Code shape/pattern** → `ast-grep` (structural)
3. **Symbol refs/impact** → `cymbal` (graph)

### When to use ast-grep

Use **ast-grep over fff/grep** when you need structural awareness:
- "Find all unchecked `err` returns" → `sg -p 'f(), err := $F()'` (finds ignored errors WITHOUT `_ =`)
- "Find all `context.Background()` in HTTP handlers" → `sg -p 'context.Background()'`
- "Find all `sync.Mutex` copies" → `sg -p 'func $F($$$) { $$$ }'` + filter types
- Pattern rewrite: `sg -p '$A && $A()' -r '$A?.()'` (replace with optional chaining)

Use **fff/grep over ast-grep** when:
- You need a filename or path
- The pattern is trivial text (a string literal, a config key)
- You need frecency ranking (most-accessed files first)

### Basic usage

```bash
ast-grep -p 'fmt.Sprintf("$MSG", $ARGS)'   # structural search
ast-grep -p 'func $NAME($$$) error'          # find all error-returning funcs
ast-grep -l go -p 'defer $FILE.Close()'      # find all deferred Close() calls
ast-grep --stdin                              # batch queries via stdin
```

### Refactoring (rewrite with AST safety)

Use `-r` (rewrite) for cross-file refactoring. Prefer over grep+sed — ast-grep
preserves AST structure and never matches inside strings, comments, or tokens
that happen to look like code:

```bash
# Add ctx param to all error-returning funcs that lack it
ast-grep -p 'func ($R *$T) $F($$$) error' \
  -r 'func ($R *$T) $F(ctx context.Context, $$$) error' \
  --rewrite

# Rename method across all files
ast-grep -p 'func ($R *$T) oldName($$$) $RES' \
  -r 'func ($R *$T) newName($$$) $RES' \
  --rewrite

# Replace pattern structurally (not textually)
ast-grep -p '$A && $A()' -r '$A?.()' --rewrite
```

**Regra:** refatoração cross-file que mexe em assinatura/estrutura → ast-grep.
Alteração localizada (1-2 arquivos, corpo de função) → edit/write normal.

Install check: `which ast-grep || brew install ast-grep`

### Linting with ast-grep

`ast-grep` can also run rule files (`sg scan`) for project-specific lint rules.
Not a replacement for `golangci-lint` — use for custom structural rules the Go
type system cannot express. See `ast-grep new` for scaffolding.

---

---

## Skills Reference

| When | Skill |
|------|-------|
| Code navigation (cymbal), file search (fff), structural search (ast-grep), entity diff (sem) | `cali-code-navigation` |
| Cross-cutting debug (framework JS API, binary freshness) | `cali-cross-cutting-debug` |
| New Feature or Product | `/skill:stelow-product-orchestrator` |
| Coding standards (universal: KISS, DRY, LoB, SoC, Fail Fast, YAGNI) | `/skill:cali-product-coding-standards` |
| Go standards (idiomatic Go, concurrency, linting) | `/skill:cali-coding-go-standards` |
| Go stack (Datastar, Templ, NATS, Air) | `/skill:cali-coding-go-stack` |
| Package audit (socket.dev, Trivy, OSV-Scanner, dotenvx) | `/skill:cali-ops-package-audit` |
| Releases | `/skill:cali-ops-github-releases` |
| Testing execution (unit → subagent → UI → browser) | `/skill:cali-product-testing-execution` |
| Deploy | `/skill:cali-ops-deploy-github-tailscale` |
| Security | `/skill:cali-ops-server-security` |
| AGENTS.md create/update/validate | `/skill:cali-agents-md-generator` |
| AGENTS.md validate | `/skill:cali-agents-md-validator` |
| Post-execution critique | `/skill:cali-product-execution-critique` |
| Parallel work | `/skill:pi-subagents` |

### Skill Prefix Convention
- `cali-coding-*` — práticas/padrões de escrita de código (standards, go-stack, starhtml)
- `cali-code-*` — ferramentas de análise de código existente (navigation)
- `cali-cross-cutting-*` — regras transversais (debug)
- `stelow-*` — skills de planejamento de produto (fonte: `stelow-product-orchestrator`)
- `cali-ops-*` — operações (deploy, security, package-audit)

### Project-Level AGENTS.md

Every project **must** have a lean `AGENTS.md` (20-30 lines max, no placeholders). See `cali-agents-md-generator` for create/update/validate workflow, pre-commit hook setup, and CI guard.

Update AGENTS.md when:
- Entry point / cmd moved (e.g. `cmd/web/` → `cmd/server/`)
- New top-level package or `features/` module added/removed/renamed
- Build/test/lint command changed
- Framework swap (e.g. Templ → html/template, Datastar → HTMX, NATS → in-process)
- Database / schema migration system changed
- New convention enforced (e.g. `errors.New` → `fmt.Errorf`)

**Human must approve all AGENTS.md changes** — never auto-update.

Cosmetic changes (CSS, copy, refactors inside functions, dependency bumps, test cases) do NOT require an update.

---

## Conventions

### Language
All code, URLs, routes, handler IDs: **English**. Exception: user-facing UI text, LLM prompts, database content.

### File Naming
All project files: `lowercase-kebab-case`.

### Estimation Bias Correction
LLMs are trained on human data that **systematically overestimates implementation time** (Hofstadter's Law). This creates a bias toward choosing "simpler" but worse solutions to avoid perceived complexity.

**Rule:** Never let estimated implementation cost outweigh solution quality. A correct, maintainable, well-structured solution that takes longer is preferable to a quick hack. Push back against your own instinct to choose "cheap" approaches. Quality first — if a pattern is architecturally correct (SSE over polling, proper schema over flat storage, real auth over stub), implement it properly regardless of estimated effort.

### Post-Execution Verification
After ANY multi-step task: trigger **`/skill:cali-product-execution-critique`** — verifies plan compliance, edge cases, NFRs, and needed doc/test updates.

---

## Don'ts

- Never add dependencies without asking
- Never put secrets in AGENTS.md
- Never modify vendor/ without asking
- Never use Axios — use native `fetch`
