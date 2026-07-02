---
name: cali-code-navigation
description: >
  [Cali] Tool priority rules for code navigation: Cymbal-first, FFF fallback,
  batch guidance for token efficiency, and semantic diff workflow. Activates
  on any file-task within a Git repository to guide agent tool selection.
metadata:
  frequency: daily
  category: meta
  context-cost: low
---

# Code Navigation Rules

## Tool Priority

For every file-task, ask: "Can Cymbal do this?"
1. `cymbal_search` beats `fffind` (symbol-aware, git-aware, ranked)
2. `fffind` beats raw `find`/`ls` (fuzzy, frecency, git-aware)
3. `cymbal_show` beats `read` (structured output, symbol-aware, refs-aware)
4. `cymbal_search --text` beats `ffgrep` (git-aware, indexed)
5. `read` = exact file path only (never directory). `grep`/`rg` only for reading a single file's content, not for cross-file search.

| Task | Command | Priority | When |
|------|---------|----------|------|
| Find a symbol/name | `cymbal_search` | 1st | Know the symbol name or concept |
| Find a file by name | `fffind` | 2nd (fallback) | After `cymbal_search` returned nothing |
| Search arbitrary text (Git repo) | `cymbal_search --text` | 1st | Text content inside a Git repo |
| Search text outside Git | `ffgrep` | 2nd (fallback) | Outside Git or Cymbal has no index. `read`/`grep` only for single-file content |
| Read a file (Git repo) | `cymbal_show` | 1st | File inside a Git repo (Cymbal-indexed) |
| Read a file (outside Git) | `read` | 2nd (fallback) | Exact file path only. Never directory path — `read` can't list dirs |
| Understand references/impact | `cymbal_refs` / `cymbal_impact` | 1st | Before any refactor |
| Diff-scoped impact | `cymbal_changed` | 1st | **NEW.** Unstaged changes; `--staged` for staged |
| Cross-language resolution | `cymbal_impact`/`trace --resolve-scope` | use flag | `same`/`family`/`all` (default `family`) |
| Multi-pattern OR search | `ffgrep` with regex alternation (e.g. `auth|login`) | 3rd | fff-multi-grep disabled — use regex alternation instead |
| Non-Git directories | `bash find` / `bash ls` | only option | Cymbal requires a Git repo |
| Process large output | truncation or targeted reads | — | — |

### Cymbal: Direct Usage
- No `init` needed: Cymbal auto-indexes.
- Unknown repo: start with `cymbal_structure`.
- Search/read: `cymbal_investigate` (adaptive) > `cymbal_search` → `cymbal_show` → `cymbal_refs`/`cymbal_impact`/`cymbal_trace`.
- **New in v0.14:** `cymbal_changed` for one-call diff-to-impact. `--no-tests` to exclude test callers. `--resolve-scope` for polyglot repos.
- `cymbal_index` only if repo seems stale or too new; use `--force` once.
- If repo not detected, fallback: `bash find` / `bash grep`.
- All commands support `--json`. Piped batching via `--stdin` on search/show/refs/impact/trace/impls/investigate.

## Batch for Token Efficiency

Several tools accept multi-target in one call. Prefer batch when reading/searching multiple symbols, URLs, or page elements. Each batch saves ~400 tokens and one round-trip.

| Batch this | Instead of |
|---|---|
| `cymbal_show A B C` (`targets: [A,B,C]`) | 3 serial calls |
| `cymbal_search X Y` (`queries: [X,Y]`) | 2 calls |
| `cymbal_refs A B` (`symbols: [A,B]`) | 2 calls |
| `cymbal_impact A B` (`symbols: [A,B]`) | 2 calls |
| `cymbal_impls A B` (`symbols: [A,B]`) | 2 calls |
| `cymbal_investigate A B` (`symbols: [A,B]`) | 2 calls |
| `cymbal_changed` (`--staged` or unstaged) | replaces manual diff + impact loop |
| `cymbal_outline a.ts b.ts` (`files: [a.ts,b.ts]`) | 2 calls |
| `cymbal_trace A B` (`symbols: [A,B]`) | 2 calls |
| `fetch_content urls: [A,B]` | 2 calls |
| `web_search queries: [Q1,Q2]` | 2 calls |
| `agent_browser job: {steps: [...]}` or `batch` stdin | N serial calls |
| `subagent tasks: [...]` or `chain: [...]` | N calls |

Don't batch: `edit`, `write`, `read` (each different context). `ffgrep`/`fffind` are already single-pattern fast.

## Cymbal v0.14+ — New & Notable

| Feature | How | Why |
|---------|-----|-----|
| `changed` | `cymbal changed` (unstaged) / `--staged` / `--base <ref>` | One-call diff-to-impact. Replaces manual diff + per-symbol impact loop. |
| `--resolve-scope` | `--resolve-scope family` (default) / `same` / `all` | Constrain cross-language name resolution. `same` = same language only, `family` = interop families (jvm, js, c), `all` = any language. |
| `--no-tests` | `cymbal impact X --no-tests` | Exclude test-file callers from impact. Useful for production blast radius. |
| Caller classification | `impact` reports `P production, T test, U unknown` | Split production vs test callers at a glance. |
| Reference metrics | `impact` shows exact `reference_rows` / `referencing_files` | Get true reference breadth even when callers are truncation-capped. |
| `--stdin` on investigate | `cymbal outline svc.go -s --names | cymbal investigate --stdin` | Pipe file outlines directly into adaptive investigation. |
| Git worktree federation | Auto-includes indexed sibling worktrees in lookup commands | Work across branches without re-indexing. Pass `--no-federate` to opt out. |

## sem — entity-level diff (not navigation)

sem (`sem diff`/`blame`/`log`) is for **entity-level version control**. Not code navigation.

| Command | Purpose | Unique vs cymbal? |
|---------|---------|-------------------|
| `sem diff` | Overview: what entities changed in working tree | ✅ cymbal only does per-symbol scoped diff |
| `sem blame <file>` | Who last modified each function/method | ✅ cymbal has no blame |
| `sem log <entity>` | Evolution of an entity through git history | ✅ cymbal has no log |
| `sem setup` | Replace `git diff` with entity-level diff | ✅ cymbal can't do this |
| `sem verify` | Check function call arity | ✅ unique |
| ~~`sem impact`~~ | **Use `cymbal impact` instead** — same job, better index | ❌ overlap |
| ~~`sem context`~~ | **Use `cymbal context` or `cymbal investigate` instead** | ❌ overlap |

**Rule:** `sem` is for diff overview and history. For code navigation — symbols, refs, callers, impact, context — always use `cymbal` first.
