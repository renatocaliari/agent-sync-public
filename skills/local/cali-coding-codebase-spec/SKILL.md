---
name: cali-coding-codebase-spec
description: >
  Use when user provides a codebase (file path, GitHub URL, uploaded files, pasted code, or "@file" reference)
  and asks to understand, document, reverse-engineer, or analyze what the product does.
  Triggers on: "spec", "specification", "document this", "what does this do", "reverse engineer",
  "analyze codebase", "extract product rules", "what can this do", "@/path/to/file", "@file".
  NOT for: reading logs, debugging specific errors (use agent-browser or ctx_execute instead).
version: 2.0.0
compatibility: Works with any codebase. For large repos (>200 files), uses sampling strategy.

metadata:
  frequency: monthly
  category: meta
  context-cost: medium
---

## Goal

Reverse-engineer a product from its source code into a comprehensive, tech-stack-agnostic specification document.
The output should be thorough enough that another team could rebuild the product in any language or framework.

## Input Detection

The skill accepts any of:

| Input Type | How to Detect | How to Handle |
|------------|---------------|---------------|
| Local path | `@/path` or `@path` or absolute path | `find` + `read` |
| GitHub URL | `github.com/owner/repo` | `ctx_fetch_and_index` + explore |
| Uploaded files | Session attachment | Extract and explore |
| Pasted code | User pasted snippet | Treat as module |
| Zip archive | `.zip` extension | `unzip -o <file> -d /tmp/cali-coding-codebase-spec-input` |

**Mixed input:** Handle each type appropriately and merge findings.

## Output Location

Write the spec to a **project-local path** in the current working directory:

```
{cwd}/docs/codebase/{project-name}-spec.md
```

If `{project-name}` cannot be determined from the input, use `codebase-spec.md`.

**Why project-local?** Living documents should live with the code. This enables:
- Version control (spec is tracked alongside code)
- CI/CD integration (auto-regenerate on PR)
- Team access (same repo, same spec)
- Drift detection (compare spec vs code on changes)

If the `docs/` directory does not exist, create it first.

## Saturation Signals

**STOP reading files when you can answer ALL of:**

1. What problem does this product solve?
2. Who are the users?
3. What are the 5 main things a user can DO?
4. Where does AI/LLM fit in (if at all)?
5. What are the 3 most important business rules?

If you cannot answer all 5, continue reading.

## Sampling Strategy

| Repo Size | Strategy |
|----------|----------|
| **Small** (<50 files) | Read everything |
| **Medium** (50-200 files) | Top 30 files by importance (see priority order below) |
| **Large** (200+ files) | Package roots → feature directories → key files |

**Large repo sampling:**
1. Identify root packages (look for `package.json`, `go.mod`, `pyproject.toml`, `Cargo.toml`)
2. Read each package's entry point
3. Sample 5-10 files per package based on importance
4. Stop when saturation signals met

**Monorepo handling:**
If multiple package manifests found:
- Treat each package as a sub-product
- Generate one master spec that links to sub-specs
- Cross-reference shared modules

## File Priority Order

Read files in this order (most important first):

### Tier 1 — Essential (always read)
1. `package.json` / `go.mod` / `pyproject.toml` / `Cargo.toml` — tech stack + dependencies
2. `src/main.ts` / `src/index.ts` / `cmd/` / `app.py` / `main.go` — entry point
3. `README.md` — human context (if exists)

### Tier 2 — Features (read most)
4. Route files (`routes/`, `controllers/`, `pages/`, `handlers/`, `endpoints/`)
5. Data models (`models/`, `schemas/`, `types/`, `entities/`)
6. Business logic (`services/`, `hooks/`, `middleware/`, `validation/`)

### Tier 3 — Configuration (read selectively)
7. API definitions (`openapi.yaml`, `graphql/`, `proto/`)
8. Auth config (`auth/`, `middleware/auth*`)
9. Database schema (`db/`, `migrations/`, `schema.*`)

### Tier 4 — Implementation Details (read if needed)
10. UI components (`components/`, `views/`)
11. Utilities (`utils/`, `helpers/`, `lib/`)
12. Tests (`test/`, `spec/`) — extract rules from test cases

## AI/LLM Detection (Critical)

**Always search for AI integrations.** This is a priority detection.

```bash
grep -r -i "openai\|anthropic\|claude\|gpt\|gemini\|mistral\|llama\|ollama\|langchain\|llamaindex\|huggingface\|transformers\|replicate\|groq\|cohere\|bedrock\|vertex" \
  --include="*.js" --include="*.ts" --include="*.py" --include="*.go" \
  --include="*.rb" --include="*.java" --include="*.cs" \
  --include="*.md" --include="*.json" \
  . 2>/dev/null | grep -v node_modules | head -50
```

For each AI integration found, extract:
- Provider and model name
- System prompts and user templates
- Inference parameters (temperature, max_tokens)
- Streaming configuration
- RAG setup (vector DB, embeddings, chunking)
- Tool/function calling definitions
- Error handling and fallbacks

## Phase 1: Reconnaissance

Execute these steps in order:

### 1.1 Map Structure
```bash
find . -type f \( -name "*.ts" -o -name "*.tsx" -o -name "*.js" -o -name "*.py" -o -name "*.go" \) \
  ! -path "./node_modules/*" ! -path "./.git/*" ! -path "./dist/*" ! -path "./build/*" \
  ! -path "./vendor/*" | head -200
```

### 1.2 Identify Tech Stack
Look for: `package.json`, `go.mod`, `pyproject.toml`, `Gemfile`, `pom.xml`, `Dockerfile`, `docker-compose.yml`.

### 1.3 Detect AI/LLM Usage
Run the grep command above. Read every file that matches.

### 1.4 Read Tier 1-2 Files
Start with essential files, then features. Stop when saturation signals met.

### 1.5 Read Data Models
Read models, schemas, Zod/Pydantic types, Prisma schema, GraphQL types.

### 1.6 Read UI Components
Forms, modals, pages — they reveal exact fields, validations, and UX flows.

### 1.7 Read Business Logic
Services, hooks, utilities, middleware — where the rules live.

### 1.8 Read API Definitions
REST routes, GraphQL resolvers, tRPC routers, gRPC protos.

## Phase 2: Synthesis

Before writing, ask yourself:
- What problem does this product solve?
- Who are the users?
- What can a user DO in this product?
- What rules govern those actions?
- Where does AI/LLM fit in?

**If AI/LLM found:** Document EVERY call in detail. This is often the most valuable section.

## Phase 3: Write Specification

Write the spec in the **same language the user wrote in** (Portuguese → Portuguese, English → English).

### Required Sections

```
1. Product Overview
2. Tech Stack Summary (context only)
3. User Roles & Permissions
4. Features & Product Rules
5. User Flows (ASCII art)
6. Screens & Components (ASCII art)
7. Data Models
8. API Surface
9. AI / LLM Integration (detailed — if any)
10. Business Rules Catalog
11. Open Questions / Inferred Behavior
```

### Section Templates

#### 1. Product Overview
One paragraph: what is this, who uses it, what core problem it solves.

#### 2. Tech Stack Summary

| Layer | Technology | Notes |
|-------|-----------|-------|
| Frontend | React + Next.js | App Router |
| Backend | Node / Express | REST API |
| Database | PostgreSQL | via Prisma ORM |
| AI | OpenAI SDK | GPT-4o, streaming |
| Auth | NextAuth.js | Google + email/password |
| Infra | Vercel + Railway | |

*Note: This section exists for context only. The rest of the spec is tech-agnostic.*

#### 3. User Roles & Permissions
List every role found in the code (admin, user, guest, etc.) and what each can do.

#### 4. Features & Product Rules

For each feature:
```
### Feature: [Name]
**Description:** What it does in plain English.

**Rules:**
- Rule 1 (extracted from validation / business logic)
- Rule 2
- ...

**Edge cases handled:**
- ...
```

#### 5. User Flows (ASCII art)

Draw flows using ASCII art. Use these symbols:

```
Arrows:     ──►   ◄──   ──►──►   ═══►
Vertical:    │     ▲     ▼
Decision:    ─┬─   ─┤    └──
Box styles:   ┌───────┐   ╭───────╮   ┏━━━━━━━┓
             │ text  │   │ text  │   ┃ text  ┃
             └───────┘   ╰───────╯   ┗━━━━━━━┛
```

#### 6. Screens & Components (ASCII art)

Draw wireframe-level ASCII for every distinct screen.

**If UI is not visible** (e.g., API-only product):
- Describe screens textually instead
- Focus on data flow and interactions
- Omit ASCII art if it would be pure inference

**ASCII wireframe template:**
```
┌─────────────────────────────────────────────────────┐
│ Header: [Page Title]                                │
├─────────────────────────────────────────────────────┤
│ Nav: [Nav1] [Nav2] [Nav3]            [User ▼]    │
├───────────┬─────────────────────────────────────┤
│ Sidebar   │ Content Area                          │
│ - Item 1  │ ┌─────────────────────────────┐     │
│ - Item 2  │ │ Widget / Component           │     │
│ - Item 3  │ └─────────────────────────────┘     │
└───────────┴─────────────────────────────────────┘
```

#### 7. Data Models

For each entity:
```
### Entity: [Name]
| Field         | Type      | Rules / Notes                    |
|---------------|-----------|----------------------------------|
| id            | UUID      | auto-generated                   |
| email         | string    | unique, required, validated      |
| role          | enum      | user | admin | guest             |
| created_at    | timestamp | auto                             |

Relations:
- [Entity] has many [OtherEntity]
- [Entity] belongs to [OtherEntity]
```

#### 8. API Surface

```
### POST /api/auth/login
**Purpose:** Authenticate user
**Auth:** Public
**Body:** { email: string, password: string }
**Returns:** { token: string, user: User }
**Rules:**
- Rate limited: 5 attempts / 15min
- Returns 401 if credentials invalid
- Logs failed attempts
```

#### 9. AI / LLM Integration (DETAILED)

**This is the most important section for AI-powered products.**

For each AI feature:
```
### AI Feature: [Name]

**Provider:** OpenAI / Anthropic / etc.
**Model:** gpt-4o / claude-3-5-sonnet-20241022 / etc.
**Endpoint:** /api/chat or similar
**Streaming:** yes/no

**System Prompt:**
\`\`\`
[exact system prompt extracted from code]
\`\`\`

**User Prompt Template:**
\`\`\`
[exact template, with {variable} placeholders shown]
\`\`\`

**Inference Parameters:**
| Param         | Value  |
|---------------|--------|
| temperature   | 0.7    |
| max_tokens    | 2000   |

**Input:** What is passed to the model
**Output:** How the response is used
**Post-processing:** Any parsing/validation/transform

**Tool/Function Calling (if any):**
\`\`\`json
[exact definitions from code]
\`\`\`

**Error handling:**
- What happens on API failure
- Fallback behavior
```

#### 10. Business Rules Catalog

A numbered, exhaustive list. **Include confidence level for each.**

```
BR-001: [Confidence: HIGH/MEDIUM/LOW]
Users cannot delete a project with active subscriptions.
**Evidence:** `src/services/project.ts:156`
```typescript
if (project.subscriptions.length > 0) {
  throw new Error("Cannot delete project with active subscriptions");
}
```

BR-002: [Confidence: HIGH]
Email must be verified before accessing paid features.
**Evidence:** `src/middleware/auth.ts:23`
```typescript
if (!user.emailVerified && plan === 'paid') {
  return res.status(403).json({ error: "Email not verified" });
}
```

BR-003: [Confidence: MEDIUM]
AI features are disabled for free tier users.
**Evidence:** `src/config/features.ts:15` + UI gate in `src/components/AIFeature.tsx`
**Note:** UI prevents access, but backend validation also exists.

BR-004: [Confidence: LOW]
Messages are truncated to 4000 tokens before sending to LLM.
**Evidence:** Comment in `src/services/chat.ts:89`
```typescript
// TODO: confirm token limit with API
```
**Warning:** This is inferred from comment, not verified behavior.
```

#### 11. Open Questions / Inferred Behavior

Mark each item with appropriate flag:

| Flag | Meaning |
|------|---------|
| ⚠️ INFERRED | Behavior assumed from code patterns, not explicit confirmation |
| ⚠️ INCOMPLETE | Handler exists but edge cases not fully traced |
| ❓ UNCLEAR | Cannot determine intent from available code |
| 🔍 VERIFIED | Confirmed against running system or tests |

```
⚠️  INFERRED: The `legacy_mode` flag appears in the user model but is never set 
    anywhere in the current codebase. Possibly dead code or migrated feature.

⚠️  INCOMPLETE: Payment webhook handler exists but only handles `payment_succeeded`. 
    No handler found for `payment_failed` or `subscription_cancelled`.

❓  UNCLEAR: The onboarding flow appears to be optional based on a `skip` button,
    but no skip logic was found in the backend.
```

## Confidence Levels

| Level | Criteria | Example |
|-------|----------|---------|
| **HIGH** | Exact code match, line referenced | `auth.ts:42` shows exact behavior |
| **MEDIUM** | Cross-reference inference | Called by A, used by B, behavior inferred |
| **LOW** | Comment, TODO, or guess | "// TODO: implement rate limiting" |

**Rule:** Never claim HIGH confidence for inferred behavior.

## Phase 4: Quality Gates

Before saving, verify:

- [ ] All 5 saturation questions answered
- [ ] Every route/page has corresponding entry (screen or description)
- [ ] Every AI call documented in Section 9 (if any AI found)
- [ ] Every validation rule captured in Section 10
- [ ] Confidence level on every business rule
- [ ] ASCII art renders correctly (consistent widths, aligned columns)
- [ ] No placeholder text ("TBD", "TODO", "FIXME" in output)
- [ ] Output path exists or was created
- [ ] Tech-stack specifics in Section 2 only (rest is agnostic)

## Phase 5: Post-Write Audit

After saving, add to the spec:

```
---

## Audit Trail
- Generated: [timestamp]
- Generator: cali-coding-codebase-spec v2.0.0
- Codebase: [git commit or version if available]
-饱和度: [met / partial] (saturation signals answered)

## How to Update This Spec

When code changes, re-run this skill to regenerate:
1. Check what changed: `git diff docs/codebase/{name}-spec.md`
2. If changes are significant, re-run the skill
3. Focus on: Features (4), Business Rules (10), AI Integration (9)

**Living document maintenance:**
- Run after major features
- Run when AI/LLM integration changes
- Run when business rules are added/modified
```

## Common Errors to Avoid

1. **Tech vs Product confusion**
   → If it requires code to verify, it's tech; if it's behavioral, it's product

2. **Inferring user roles that don't exist**
   → Only document roles found in code (auth checks, role enums, permission guards)

3. **Missing AI integration details**
   → Always run the grep command; AI features are often the most valuable output

4. **Incomplete business rules**
   → Extract from: validation, middleware, service logic, edge case handlers, test cases

5. **ASCII art as obligation**
   → If UI is not visible (API-only, backend), describe textually. Don't fake wireframes.

6. **All rules as HIGH confidence**
   → Include MEDIUM and LOW with appropriate evidence. False confidence is worse than uncertainty.

## Testing the Skill

After any changes to this skill, test with:

**Trigger tests:**
- "Document this codebase" + local path → should trigger
- "Analyze @src/" + pasted code → should trigger
- "What does this do?" + GitHub URL → should trigger
- "Debug the login bug" → should NOT trigger

**Output tests:**
- Spec contains all 11 sections
- AI/LLM section complete if AI found
- Every business rule has confidence level
- ASCII art renders without broken boxes

## References

- [SKILL.md Guide](https://lalitmadan.com/post/the-skill-md-guide) — Lalit Madan
- [Anthropic Skills Best Practices](https://github.com/anthropics/skills)
- [mgechev/skills-best-practices](https://github.com/mgechev/skills-best-practices)
- [MCP Build with Agent Skills](https://modelcontextprotocol.io/docs/develop/build-with-agent-skills)

---

*Skill v2.0.0 — Updated based on execution experience + 2026 skill design research*
