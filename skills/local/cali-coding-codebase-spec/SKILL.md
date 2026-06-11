---
name: cali-coding-codebase-spec
description: >
  [Cali] Use this skill whenever the user wants to analyze a codebase and generate a detailed product spec document.
  Triggers for: "analyze this codebase", "generate a spec", "document this project", "reverse engineer this app",
  "what does this system do", "create product documentation", "map out the features", "extract product rules from code",
  "what does this codebase do", "generate product spec from code", "document this app", "spec out this repo",
  "read this code and tell me what it does", "create a product spec", "reverse engineer the product",
  "what features does this app have", "document the flows", "extract the prompts from this code",
  or any request to extract product knowledge from source code. Supports any input method: file uploads, pasted code,
  filesystem paths, or zip archives. Always use this skill when the user shares code AND wants to understand
  what the product does — even if they don't explicitly say "spec" or "documentation".
---

# Codebase → Product Spec

You are a senior product analyst reverse-engineering a product from its source code.
Your job: produce a **living product specification** that is completely tech-stack-agnostic —
so thorough that another team could rebuild the product in any language or framework.

---

## Phase 0 — Intake

Accept codebase via any of:
- **Uploaded files / zip** → extract and explore with bash tools
- **Pasted code snippets** → treat each as a module
- **Filesystem path** → `find` + `cat` to read
- **Mixed** → handle each source appropriately

If the input is a zip file:
```bash
unzip -o <file> -d /tmp/cali-codebase-spec-input
```

---

## Phase 1 — Reconnaissance (read silently, don't output yet)

### 1.1 Map the repository structure
```bash
find /tmp/cali-codebase-spec-input -type f | sort
# or for uploads: ls -R /mnt/user-data/uploads/
```

### 1.2 Identify what kind of system this is
Look for: `package.json`, `pyproject.toml`, `go.mod`, `Gemfile`, `pom.xml`, `*.csproj`,
`Dockerfile`, `docker-compose.yml`, `*.env.example`, CI configs, README files.

### 1.3 Identify AI/LLM usage (HIGH PRIORITY)
Search for:
```bash
grep -r "openai\|anthropic\|claude\|gpt\|gemini\|mistral\|llama\|ollama\|langchain\|llamaindex\|huggingface\|transformers\|replicate\|groq\|cohere\|bedrock\|vertex" \
  --include="*.js" --include="*.ts" --include="*.py" --include="*.go" \
  --include="*.rb" --include="*.java" --include="*.cs" -l -i /tmp/cali-codebase-spec-input 2>/dev/null
```

Then read every file that matches to extract:
- Which AI provider / SDK is used
- Model names (e.g. `gpt-4o`, `claude-3-5-sonnet`, `gemini-1.5-pro`)
- Every prompt string (system prompts, user prompt templates, few-shot examples)
- Temperature, max_tokens, and other inference params
- Whether streaming is used
- RAG setup (vector DBs, embeddings, chunking strategy)
- Agent loops, tool use, function calling definitions

### 1.4 Read all route/controller/page files
These reveal every feature the product has.

### 1.5 Read data models / schemas
DB models, Zod/Pydantic schemas, GraphQL types, Prisma schema, etc.

### 1.6 Read UI components
Especially forms, modals, pages — they reveal exact fields, validations, and UX flows.

### 1.7 Read business logic
Services, hooks, utilities, middleware — where the rules live.

### 1.8 Read API definitions
REST routes, GraphQL resolvers, tRPC routers, gRPC protos.

---

## Phase 2 — Synthesis

After reading everything, synthesize the full picture before writing.
Ask yourself:
- What problem does this product solve?
- Who are the users?
- What are all the things a user can DO in this product?
- What are all the rules that govern those actions?
- Where does AI/LLM fit in the product experience?

---

## Phase 3 — Write the Spec

Output a single Markdown document with the sections below, **in the same language the user wrote in**
(if user wrote in Portuguese, spec is in Portuguese; if English, spec is in English).

Save as `/mnt/user-data/outputs/<project-name>-spec.md`

---

### SPEC STRUCTURE

```
# [Product Name] — Product Spec
> Reverse-engineered from source code. Tech-stack agnostic.
> Generated: [date]

## 1. Product Overview
## 2. Tech Stack Summary (context only)
## 3. User Roles & Permissions
## 4. Features & Product Rules
## 5. User Flows (with ASCII art)
## 6. Screens & Components (with ASCII art)
## 7. Data Models
## 8. API Surface
## 9. AI / LLM Integration (detailed)
## 10. Business Rules Catalog
## 11. Open Questions / Inferred Behavior
```

---

### Section Templates

#### 1. Product Overview
```
One paragraph: what is this, who uses it, what core problem it solves.
```

#### 2. Tech Stack Summary
```
| Layer        | Technology       | Notes                    |
|--------------|-----------------|--------------------------|
| Frontend     | React + Next.js  | App Router               |
| Backend      | Node / Express   | REST API                 |
| Database     | PostgreSQL       | via Prisma ORM           |
| AI           | OpenAI SDK       | GPT-4o, streaming        |
| Auth         | NextAuth.js      | Google + email/password  |
| Infra        | Vercel + Railway |                          |
```
Note: This section exists for context only. The rest of the spec is tech-agnostic.

#### 3. User Roles & Permissions
List every role found in the code (admin, user, guest, etc.) and what each can do.

#### 4. Features & Product Rules
For each major feature:
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

Use ASCII art to draw every major flow. Examples:

**Navigation flow:**
```
[Login] ──► [Dashboard]
               │
     ┌─────────┼──────────┐
     ▼         ▼          ▼
 [Projects] [Settings] [Profile]
     │
     ▼
 [Project Detail] ──► [Edit Project]
                            │
                    [Save] / [Cancel]
```

**Decision flow:**
```
User submits form
       │
       ▼
  [Validate input]
       │
   ┌───┴───┐
  PASS    FAIL
   │        │
   ▼        ▼
[Save]  [Show errors]
   │
   ▼
[Send confirmation email]
   │
   ▼
[Redirect to dashboard]
```

**Sequence / interaction flow:**
```
User          Frontend       Backend        AI (LLM)
 │                │              │              │
 │──send msg─────►│              │              │
 │                │──POST /chat─►│              │
 │                │              │──prompt─────►│
 │                │              │◄─stream──────│
 │                │◄─SSE stream──│              │
 │◄─render tokens─│              │              │
```

#### 6. Screens & Components (ASCII art)

Draw wireframe-level ASCII for every distinct screen. Be thorough — include labels, fields, buttons, states.

**Example — Dashboard screen:**
```
┌─────────────────────────────────────────────────────┐
│  🏠 MyApp          [Notifications 🔔]  [User ▼]     │
├──────────────┬──────────────────────────────────────┤
│              │  Welcome back, João!                  │
│  📊 Dashboard│                                       │
│  📁 Projects │  ┌──────────┐ ┌──────────┐           │
│  ⚙️ Settings │  │ 12       │ │ 3        │           │
│  👤 Profile  │  │ Projects │ │ Active   │           │
│              │  └──────────┘ └──────────┘           │
│              │                                       │
│              │  Recent Activity                      │
│              │  ┌─────────────────────────────────┐  │
│              │  │ • Project X updated   2min ago  │  │
│              │  │ • New comment on Y    1hr ago   │  │
│              │  └─────────────────────────────────┘  │
│              │                                       │
│              │              [+ New Project]          │
└──────────────┴──────────────────────────────────────┘
```

For each screen, also list:
- **Purpose:** what the user accomplishes here
- **Components:** list of reusable components used
- **States:** loading / empty / error / populated
- **Actions:** what buttons/forms do

#### 7. Data Models
For each entity:
```
### Entity: User
| Field         | Type      | Rules / Notes                    |
|---------------|-----------|----------------------------------|
| id            | UUID      | auto-generated                   |
| email         | string    | unique, required, validated      |
| role          | enum      | user | admin | guest             |
| created_at    | timestamp | auto                             |

Relations:
- User has many Projects
- User has many Comments
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

This is the most important section for AI-powered products.

```
### AI Feature: [Name] (e.g. "Chat Assistant", "Document Summarizer")

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

**Few-shot Examples (if any):**
\`\`\`
[extracted examples]
\`\`\`

**Inference Parameters:**
| Param         | Value  |
|---------------|--------|
| temperature   | 0.7    |
| max_tokens    | 2000   |
| top_p         | 1.0    |

**Input:** What is passed to the model (user message, context, retrieved docs, etc.)
**Output:** What the model returns and how it's used
**Post-processing:** Any parsing, validation, or transformation applied to the response

**RAG / Context injection (if any):**
- Vector DB: Pinecone / Chroma / pgvector / etc.
- Embedding model: text-embedding-3-small / etc.
- Chunking: [strategy extracted from code]
- Top-k: [number]
- What documents are indexed

**Tool / Function calling (if any):**
\`\`\`json
[exact tool definitions extracted from code]
\`\`\`

**Error handling:**
- What happens on API failure
- Fallback behavior
- Retry logic
```

#### 10. Business Rules Catalog

A numbered, exhaustive list of every rule found in the codebase:
```
BR-001: Users cannot delete a project with active subscriptions.
BR-002: Email must be verified before accessing paid features.
BR-003: AI features are disabled for free tier users.
BR-004: Messages are truncated to 4000 tokens before sending to LLM.
BR-005: Admin can impersonate any user account.
...
```

#### 11. Open Questions / Inferred Behavior

Be honest about what was inferred vs. explicitly coded:
```
⚠️  INFERRED: The onboarding flow appears to be optional based on a `skip` button, 
    but no skip logic was found in the backend — unclear if this is tracked.

⚠️  INCOMPLETE: Payment webhook handler exists but only handles `payment_succeeded`. 
    No handler found for `payment_failed` or `subscription_cancelled`.

❓  UNCLEAR: The `legacy_mode` flag appears in the user model but is never set anywhere 
    in the current codebase. Possibly dead code or migrated feature.
```

---

## Phase 4 — Quality Check

Before saving, verify:
- [ ] Every route/page has a corresponding screen drawing
- [ ] Every AI call has a full entry in Section 9
- [ ] Every validation rule is captured in Section 10
- [ ] ASCII art renders correctly (consistent box widths, aligned columns)
- [ ] The spec could be handed to a designer/developer with zero access to the code and they'd understand the full product

---

## ASCII Art Reference

### Box styles
```
Simple:     +-------+    Rounded:   ╭───────╮    Heavy:  ┏━━━━━━━┓
            | text  |              │ text  │            ┃ text  ┃
            +-------+              ╰───────╯            ┗━━━━━━━┛

Double:     ╔═══════╗    Dotted:   ┌╌╌╌╌╌╌╌┐
            ║ text  ║              ╎ text  ╎
            ╚═══════╝              └╌╌╌╌╌╌╌┘
```

### Arrow styles
```
Flow:       ──►   ◄──   ──►──►   ═══►
Vertical:    │     ▲     ▼
Conditional: ─┬─   ─┤    └──
```

### Decision diamond
```
        ┌─────────┐
        │ Action  │
        └────┬────┘
             │
        ┌────▼────┐
        │Decision?│
        └────┬────┘
         ┌───┴───┐
        YES      NO
         │        │
```

### Sequence diagram template
```
Actor A       Actor B       Actor C
   │              │              │
   │──action─────►│              │
   │              │──action─────►│
   │              │◄─response────│
   │◄─response────│              │
```

---

## Output File

Always save the final spec to:
```
/mnt/user-data/outputs/<project-name>-spec.md
```

Then use `present_files` to share it with the user.

Finish with a brief summary: number of features documented, screens drawn, AI integrations found, and business rules catalogued.