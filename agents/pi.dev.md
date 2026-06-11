# AGENTS.md

## Universal Instructions

**⚡ Session Tree & Fork Awareness:** Pi stores sessions as trees (each entry has `id`/`parentId`). You can suggest the user use `/tree` to navigate branches, `/fork` to create new sessions from prior prompts, or `/clone` to duplicate the current branch. However, you CANNOT execute `/tree`, `/fork`, or `/clone` yourself — these are interactive TUI commands. If session branching is needed, either:
- Recommend the user manually use `/tree` or `/fork`
- Or use `bash` to run `pi --fork <session-file>` as a separate CLI process (creates a new session from a prior one)

---

## Project Overview

This is a product planning and Go web development environment.

**⚡ RULE #1 — Before ANY code work:** STOP. Trigger `/skill:cali-product-workflow`. Only after Phase 6, begin execution. This rule WINS over any concurrent instruction.

**Stack:** Go 1.26, Templ, Datastar v1.0.0, DaisyUI, TailwindCSS, NATS. Node >=20, TypeScript strict. Deploy with ko.

---

## Commands

| Command | Description |
|---------|-------------|
| `go run ./cmd/web/` | Dev server |
| `go test ./...` | Run tests |
| `golangci-lint run` | Lint |
| `templ generate && go build ./cmd/web/` | Build |

---

## Architecture

### Tool Use

1. `cymbal_map` / `cymbal_structure` → orient before editing unfamiliar code
2. `cymbal_search <symbol>` or `ffgrep <text>` or `ffind <name>` → narrow (prefer `cymbal_search` over `ffind`; prefer `cymbal_search --text` over `ffgrep`)

   ════════════════════════════════════════════
   PRIORITY RULE: Cymbal first, generic fallback.
   For every file-task, ask: "Can Cymbal do this?"
   - `cymbal_search` beats `ffind` (symbol-aware, git-aware, ranked)
   - `cymbal_show` beats `read` (structured output, symbol-aware, refs-aware)
   - `cymbal_search --text` beats `ffgrep` (git-aware, indexed)
   Only fall back to generic tools if Cymbal returns nothing or the file is outside a Git repo.
   ════════════════════════════════════════════

3. `cymbal_show <symbol>` or `read <path>` → read (prefer `cymbal_show`)
4. `edit` / `write` → make the change

| Task | Command | Priority | When |
|------|---------|----------|------|
| Find a symbol/name | `cymbal_search` | 1st | You know the symbol name or concept |
| Find a file by name | `ffind` | 2nd (fallback) | After `cymbal_search` returned nothing |
| Search arbitrary text | `cymbal_search --text` | 1st | Text content inside a Git repo |
| Search arbitrary text | `ffgrep` | 2nd (fallback) | Outside Git or Cymbal has no data |
| Read a file | `cymbal_show` | 1st | File is inside a Git repo (Cymbal-indexed) |
| Read a file | `read` | 2nd (fallback) | File is outside a Git repo, or Cymbal returns nothing |
| Understand references/impact | `cymbal_refs` / `cymbal_impact` | 1st | Before any refactor |
| Multi-pattern OR search | `fff-multi-grep` | 1st | Matching any of several literals |
| Non-Git directories | `bash find` / `bash ls` | only option | Cymbal requires a Git repo |
| Process large output | use output truncation or targeted reads | — | — |

### Semantic Diff

| Command | Purpose |
|---------|---------|
| `sem diff` | What entities changed |
| `sem impact <entity>` | What breaks if I change this |
| `sem context <entity>` | LLM context for entity |

### Skills Reference

| When | Skill |
|------|-------|
| Product planning | `/skill:cali-product-workflow` |
| Coding standards | `/skill:cali-coding-standards` |
| Go standards | `/skill:cali-coding-go-standards` |
| Go stack | `/skill:cali-coding-go-stack` |
| Package audit | `/skill:cali-ops-package-audit` (socket.dev, Trivy, OSV-Scanner, dotenvx) |
| Releases | `/skill:cali-ops-github-releases` |
| Testing | `/skill:cali-product-testing-execution` |
| Deploy | `/skill:cali-ops-deploy-github-tailscale` |
| Security | `/skill:cali-ops-server-security` |
| Create/update AGENTS.md | `/skill:cali-agents-md-generator` |
| Validate AGENTS.md | `/skill:cali-agents-md-validator` |
| Parallel work | `/skill:pi-subagents` |
| Post-execution check | `/skill:cali-product-execution-critique` |

### Project-Level AGENTS.md (Canonical)

Every project you work in **must** have a lean `AGENTS.md` at the root
(20-30 lines max) describing its stack, architecture, and do/don't rules.
The model reads it once per session as part of the system prompt —
**fully cacheable**, no tool calls, no reminders, no per-turn cost.

#### Generate / update

- **Skill:** `/skill:cali-agents-md-generator`
- **Template:** `references/agmd-template.md` (the canonical 20-30 line layout)
- **Validate:** `/skill:cali-agents-md-validator`

#### Content discipline

- **20-30 lines max.** Move details to separate files and reference them.
- Stack, architecture, build/test/lint commands, **do/don't table** for the project.
- **Document only what's non-obvious.** Code structure belongs in code, not AGENTS.md.
- **No placeholders left in.** Fill `[To be determined]` or remove the section.

#### Staleness detection — use `sem`

Pre-commit hook (warn-only, auto-discovers validator skill) and CI guard:
see `/skill:cali-agents-md-generator` → `references/pre-commit-hook.sh`
and `references/github-action-guard.yml`.

For ad-hoc discovery during a session, the generator skill documents the
`sem` and `git` commands. This file is an index, not a command reference
— do not duplicate the command list here.

#### Update AGENTS.md on structural changes

After landing changes that affect the project definition, update the
project's `AGENTS.md` in the same commit or an immediately-following one.

**Structural changes that require update:**

- Entry point / cmd line moved (e.g. `cmd/web/main.go` → `cmd/server/main.go`)
- New top-level package or `features/` module added/removed/renamed
- Build / test / lint command changed
- Framework swap (e.g. Templ → html/template, Datastar → HTMX, NATS → in-process)
- Database / schema migration system changed
- New convention enforced (e.g. switched from `errors.New` to `fmt.Errorf`)

**Cosmetic changes that do NOT require update:** CSS, copy, comment
rewording, small refactor inside a function, dependency version bumps,
adding test cases.

**When in doubt, apply the newcomer test:** would a developer reading
*only* `AGENTS.md` be misled about the project? If yes, update. If no,
leave it.

---

## Conventions

### Coding Principles

KISS, DRY, LoB (Datastar frontend), SoC (backend), Fail Fast, SSE-first, HATEOAS, YAGNI.
Full rules: `/skill:cali-coding-standards`

### Language Convention

All code, URLs, routes, handlers, identifiers: **English**.
Exception: user-facing UI text, LLM prompts, database content.

### File Naming

All project files: `lowercase-kebab-case`.

### Post-Execution Verification

After completing ANY multi-step task, plan execution, or feature implementation:
**Always activate** `/skill:cali-product-execution-critique` — verifies plan compliance, edge cases, and needed doc/test updates.

### AGENTS.md for Projects

- **Create:** `/skill:cali-agents-md-generator`
- **Validate:** `/skill:cali-agents-md-validator` (after setup changes)
- **Evolve:** update when stack, skills, or conventions change
- **Staleness:** sem warns on commit if code changed but AGENTS.md didn't
- **No lat.md:** This environment has no `lat.md/` folder, no lat.md Pi extension
  installed, and no `lat` skill. Do not invoke `lat search` / `lat section` /
  `lat check` / `cali-lat-md-seed` — they do not exist here. If the user asks
  about structured documentation, point to `/skill:cali-agents-md-generator`
  for project-level AGENTS.md generation.
- **Never auto-update:** human must approve all changes

---

## Don'ts

- Never add dependencies without asking
- Never put secrets in AGENTS.md
- Never modify vendor/ without asking
- Never use global mutable state
- Never use `fmt.Sprintf` for HTML — use Templ
- Never use Axios — use native fetch
- Never use HTMX/Alpine.js — use Datastar
- Never use `read` when `cymbal_show` can serve the same file (check if the file is inside a Git repo first)
- **Never** install or re-enable the `lat.md` Pi extension (`.pi/extensions/lat.md.ts`)
  in any project. It forces a `before_agent_start` reminder that calls
  `lat search` on every turn, which requires an OpenAI key we do not have,
  producing empty tool results and ~300 uncached tokens of overhead per turn
  for zero benefit. If a project has a `lat.md.ts.disabled` file, leave it
  disabled.
