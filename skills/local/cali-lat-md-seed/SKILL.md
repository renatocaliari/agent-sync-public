---
name: cali-lat-md-seed
description: >
  Auto-initialize structured documentation for any project using lat.md
  (knowledge graph of markdown files with [[wiki links]], // @lat: code
  refs, and semantic search). Detects cali-product-workflow artifacts
  (spec-product.md, spec-tech.md, critiques) and uses them as seed
  material. Falls back to extracting business rules, architecture, and
  design decisions directly from the codebase. Use when a project lacks
  structured documentation or when lat.md/ is missing. After seeding,
  lat.md extension hooks keep documentation alive automatically.
metadata:
  frequency: monthly
  category: product
  context-cost: high
---

# cali-lat-md-seed

> **Purpose:** Transform any project — with or without planning artifacts — into one with
> living documentation that agents automatically maintain.
>
> **Tools:** See `references/cli-tools/lat-md.md` for lat.md command reference.

## Overview

This skill installs and seeds `lat.md` — a markdown knowledge graph that agents
use as shared context across sessions. It handles three scenarios:

| Scenario | Behavior |
|---|---|
| **Has `lat.md/` already** | Verify freshness, skip seed |
| **Has cali-pw artifacts** | Use `spec-product.md`, `spec-tech.md`, critiques as seed |
| **Neither exists** | Extract knowledge directly from the codebase |

After seeding, the `lat.md` extension hooks (installed by `lat init`) automatically
enforce documentation freshness on every agent session.

---

## Activation

### Standalone (triggered via AGENTS.md)
```
on any project without lat.md/
```

### Via cali-product-workflow (execution stage)
The workflow's execution phase can seed `lat.md/` during implementation.

### Manual
```
/skill:cali-lat-md-seed
```

---

## 🔀 Input Detection

```
Working directory has no lat.md/?:
  ├── Does .cali-product-workflow/ exist?
  │   ├── Yes → Mode: Seed-from-Plans
  │   └── No  → Mode: Seed-from-Codebase
  └── After seed → Mode: Verify

lat.md/ already exists:
  └── Mode: Verify (check freshness)
```

---

## How to Execute

### Step 1 — Install / Update lat.md

First check if lat.md is installed and up to date:

```bash
# Install if missing
which lat || npm install -g lat.md

# Check for updates and apply silently
npm outdated -g lat.md && npm update -g lat.md && lat init --update
# If no update needed, this does nothing
```

Then initialize or re-sync:

```bash
cd /project/root && lat init
```

> **Note:** If this is a Verify pass (lat.md/ already exists), skip the npm
> check unless the user explicitly asks for an upgrade.

> **Auto-update note:** This skill performs one-shot version checks on seed.
> For ongoing updates during regular development, the agent should check
> `npm outdated -g lat.md` roughly once per month and apply if safe.

This:
- Creates `lat.md/` directory structure
- Installs 6 native Pi tools (`lat_search`, `lat_section`, `lat_check`, etc.)
- Registers `before_agent_start` and `agent_end` lifecycle hooks
- Configures MCP server and skills for other platforms (Cursor, Claude Code, etc.)

> **Note:** If `lat search` fails due to missing API key, explain to the user:
> Set `LAT_LLM_KEY`, `LAT_LLM_KEY_FILE`, or `LAT_LLM_KEY_HELPER` with an OpenAI
> (`sk-...`) or Vercel (`vck_...`) key. If they decline, `lat locate` works
> without embeddings for direct lookups.

### Step 2 — Detect planning artifacts (cali-product-workflow)

Search for artifacts in order of value:

```bash
# Business rules and product spec
find . -path '*/.cali-product-workflow/*/plans/specs/spec-product*.md' -type f 2>/dev/null
# Technical architecture
find . -path '*/.cali-product-workflow/*/plans/spec-tech*.md' -type f 2>/dev/null
# Tech planning with scopes
find . -path '*/.cali-product-workflow/*/plans/tech-planning*.md' -type f 2>/dev/null
# Critique reports (gaps, risks, assumptions)
find . -path '*/.cali-product-workflow/*/critiques/critique-report*.md' -type f 2>/dev/null
# Interface decisions
find . -path '*/.cali-product-workflow/*/interfaces/interfaces*.md' -type f 2>/dev/null
# Verification and audit
find . -path '*/.cali-product-workflow/*/plans/verification*.md' -type f 2>/dev/null
# Execution critique
find . -path '*/.cali-product-workflow/*/plans/execution-critique*.md' -type f 2>/dev/null
# Scope definitions
find . -path '*/.cali-product-workflow/*/plans/scopes/*.md' -type f 2>/dev/null
```

Also check for alternate locations or formats used in project-specific setups.

#### If found → Mode: Seed-from-Plans

Read each artifact found and map to `lat.md/` files:

| Artifact | Seeds | Strategy |
|---|---|---|
| `spec-product*.md` | `lat.md/business-rules.md`, `lat.md/glossary.md` | Extract IN/OUT scope → business rules. Extract appetite, rabbit holes → constraints. |
| `spec-tech*.md` | `lat.md/architecture.md`, `lat.md/decisions.md` | Extract architecture decisions, stack choices, data flow. |
| `tech-planning*.md` | `lat.md/architecture.md`, `lat.md/guides/setup.md` | Extract sequence diagrams, dependency graph, typed scopes. |
| `critique-report*.md` | `lat.md/decisions.md`, `lat.md/risks.md` | Extract gaps, assumptions, questions as decision context. |
| `interfaces*.md` | `lat.md/interfaces.md` | Extract UI decisions, component hierarchy, navigation. |
| `scopes/*.md` | Populate from execution results | Map each implemented scope to its section. |

For each artifact, read it in full and create corresponding `lat.md/` entries.
Use `[[wiki links]]` to cross-reference concepts between files.

After seeding from plans, add `// @lat:` code refs by searching the codebase
for functions/types/classes mentioned in the artifacts:

```bash
# Example: if spec mentions "OrderService", find it in code
grep -rn "OrderService" --include='*.ts' --include='*.go' --include='*.py' src/ 2>/dev/null
# Add // @lat: [[business-rules#feature-name]] above the definition
```

### Step 3 — Extract from codebase (fallback) [Mode: Seed-from-Codebase]

If no `cali-product-workflow` artifacts exist, extract knowledge from the codebase
directly. This is a thorough but efficient process — read at most 15-20 key files.

#### 3a. Map structure

Run the following (use whatever is fastest in the project's stack):

```bash
# General structure
find . -maxdepth 1 -type f -name '*.md' | head -5   # README, AGENTS.md etc.
find . -maxdepth 1 -type d -not -path './.*' | sort # src, tests, cmd, etc.

# Project config files
ls package.json go.mod Cargo.toml pyproject.toml 2>/dev/null

# Entry points
find . -name 'main.go' -o -name 'main.ts' -o -name 'app.ts' -o -name 'index.ts' \
  -o -name 'main.py' -o -name 'main.rs' -o -name 'cmd/' 2>/dev/null | head -5

# Routes / handlers (varies by stack)
grep -rn "HandleFunc\|\.Get(\|\.Post(\|app\.\(get\|post\|put\|del\)" --include='*.go' 2>/dev/null | head -15
grep -rn "router\.\|@app\.\|Route\|app\.router" --include='*.ts' 2>/dev/null | head -10
grep -rn "def \|class " --include='*.py' src/ 2>/dev/null | head -20

# Data models / types
grep -rn "type .* struct\|interface\|enum " --include='*.go' src/ 2>/dev/null | head -15
grep -rn "interface\|type " --include='*.ts' src/ --include='*.tsx' src/ 2>/dev/null | head -15

# Tests structure
find . -name '*test*' -o -name '*spec*' 2>/dev/null | head -10
```

#### 3b. Read the critical files

Read 5-10 key files to understand:
- **README.md** or **AGENTS.md** — project overview, setup, conventions
- **Entry point** — how the app boots, dependencies
- **Main handler/controller** — primary flows
- **Core domain type/model** — entities and relationships
- **Test file** — how tests are structured

#### 3c. Create lat.md files

Create the following initial files based on your findings:

**`lat.md/business-rules.md`** — the most important file
```markdown
# Business Rules

Core domain logic, invariants, and constraints derived from codebase analysis.

## [Domain Entity/Rule Name]

[Description of what it does and the business rule it enforces]

### [Rule detail]

[Specific invariant or constraint]
```

**`lat.md/architecture.md`**
```markdown
# Architecture

System architecture, component relationships, and design patterns.

## [Component Name]

[Description of the component and its responsibilities]
```

**`lat.md/data-model.md`**
```markdown
# Data Model

Entities, relationships, and data flow.

## [Entity]

[Description of the entity, key fields, relationships]
```

**`lat.md/decisions.md`**
```markdown
# Design Decisions

Architecture Decision Records and rationale for key choices.

## ADR-1: [Decision Title]

[Context, decision, consequences]
```

**`lat.md/glossary.md`**
```markdown
# Glossary

Domain-specific terms and their meanings.

## [Term]

[Definition and context]
```

**`lat.md/tests.md`**
```markdown
---
lat:
  require-code-mention: true
---
# Tests

Test structure and specifications.

## [Test Area]

[Description of what is tested and how]
```

**`lat.md/guides/setup.md`**
```markdown
# Setup

How to set up and run this project.
```

For every section, follow the lat.md section rules:
- **Leading paragraph required** — at least one sentence immediately after heading, ≤250 chars (excluding `[[wiki link]]` content)
- **Cross-reference with `[[wiki links]]`** — connect related concepts
- **Add `// @lat:` comments** in source code pointing to the relevant section

### Step 4 — Add `// @lat:` code refs

For every section in `lat.md/` that corresponds to a specific piece of code,
add a `// @lat:` comment next to that code:

```typescript
// src/orders/service.ts:42
// @lat: [[business-rules#Order Processing#Payment Required Before Shipping]]
async function processOrder(orderId: string): Promise<OrderResult> {
```

Language syntax:
- TypeScript/JS/Go/Rust/C: `// @lat: [[section-id]]`
- Python: `# @lat: [[section-id]]`
- HTML/Templ: `<!-- @lat: [[section-id]] -->`

Run this after creating documentation to validate:

```bash
lat check
```

Fix any broken links or missing code refs until all checks pass.

### Step 5 — Verify

```bash
lat check
```

**All checks must pass.** If errors exist, fix them:
- Broken `[[wiki links]]` → correct the target section name
- Missing `// @lat:` for `require-code-mention` sections → add refs
- Missing leading paragraphs → add them (≤250 chars)

### Periodic Update Check (when lat.md/ exists)

If this is a Verify pass (not a seed), and it's been ~30 days since
last check, the agent may run:

```bash
npm outdated -g lat.md 2>/dev/null | grep -q lat && echo "lat.md has update available"
```

If an update exists, ask the user: "lat.md vX.Y is available. Update now?
(Just `npm update -g lat.md`, takes 5 seconds, risk-free.)"

Only check this when the user is in a natural break (not mid-task).
This keeps the tool current without annoying the user.

---

## cali-product-workflow Integration

When planning artifacts ARE detected, the skill not only seeds `lat.md/`
but also establishes a contract between planning and documentation:

### Before Implementation
```
Read spec-product.md      → populate lat.md/business-rules.md
Read spec-tech.md         → populate lat.md/architecture.md, lat.md/decisions.md
Read critique-report.md   → populate lat.md/risks.md (assumptions, open questions)
```

### During Implementation
Every change to the codebase should update `lat.md/` via the lat.md extension hooks
(`agent_end` enforcement). If the agent is about to implement from a plan:

- Keep `lat.md/` synchronized with what is **actually** built
- If implementation diverges from plan, document the divergence in `lat.md/decisions.md`
  with rationale

### After Implementation
The plan artifact remains as historical record. The `lat.md/` is now the
**living source of truth** — superseding the planning artifacts.

> **Key insight:** `cali-product-workflow` documents what you *planned* to build.
> `lat.md` documents what you *actually* built. Both coexist but `lat.md` wins
> for future decision-making.

---

## Quality Gates

| Check | Command | Required |
|---|---|---|
| Structure | `ls lat.md/` has at least 3 files | ✅ |
| Links | `lat check` exits 0 | ✅ |
| Code refs | Key functions have `// @lat:` comments | ✅ |
| Coverage | Business rules cover main domain concepts | ✅ |
| Freshness | `lat.md/` reflects current code (spot-check 2-3 claims) | ✅ |

---

## Undo / Revert

If the seed produced poor results or the user wants to start over:

```bash
rm -rf lat.md/
lat init  # fresh structure
# Re-run this skill with different guidance
```
