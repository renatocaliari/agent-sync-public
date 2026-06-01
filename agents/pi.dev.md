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
2. `cymbal_search <symbol>` or `ffgrep <text>` or `ffind <name>` → narrow
3. `cymbal_show <symbol>` or `read <path>` → read
4. `edit` / `write` → make the change

| Task | Command | When |
|------|---------|------|
| Find a file by name | `ffind` | Always (fuzzy, frecency, git-aware) |
| Search arbitrary text | `ffgrep` | Content has no symbol structure |
| Find a symbol definition | `cymbal_search` | You know the exact symbol name |
| Understand references/impact | `cymbal_refs` / `cymbal_impact` | Before any refactor |
| Multi-pattern OR search | `fff-multi-grep` | Matching any of several literals |
| Non-Git directories | `bash find` / `bash ls` | Cymbal requires a Git repo |
| Process large output | use output truncation or targeted reads | — |

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
| Go standards | `/skill:cali-go-standards` |
| Go stack | `/skill:cali-go-stack` |
| Package audit | `/skill:cali-package-audit` |
| Releases | `/skill:cali-github-releases` |
| Testing | `/skill:cali-product-testing-execution` |
| Deploy | `/skill:cali-deploy-github-tailscale` |
| Security | `/skill:cali-server-security` |
| Create/update AGENTS.md | `/skill:cali-agents-md-generator` |
| Validate AGENTS.md | `/skill:cali-agents-md-validator` |
| Parallel work | `/skill:pi-subagents` |
| Post-execution check | `/skill:cali-post-execution-check` |
| Auto-documentation | `/skill:cali-lat-md-seed` |

### Auto-Documentation

If the current project lacks a `lat.md/` directory AND the project is non-trivial
(multiple source files with domain logic), load `/skill:cali-lat-md-seed` to
auto-initialize structured documentation. This skill:
- Installs lat.md via `npm install -g lat.md && lat init`
- Detects `cali-product-workflow` artifacts and uses them as seed material
- Falls back to extracting business rules from the codebase
Once seeded, the lat.md Pi extension maintains documentation freshness automatically.

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
**Always activate** `/skill:cali-post-execution-check` — verifies plan compliance, edge cases, and needed doc/test updates.

### AGENTS.md for Projects

- **Create:** `/skill:cali-agents-md-generator`
- **Validate:** `/skill:cali-agents-md-validator` (after setup changes)
- **Evolve:** update when stack, skills, or conventions change
- **Staleness:** sem warns on commit if code changed but AGENTS.md didn't
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
