---
name: cali-windmill-orchestration
description: >
  [Cali] Windmill workflow orchestration strategies, gotchas, and decision
  frameworks. Use this skill when building, debugging, or maintaining
  Windmill scripts, flows, and Workflows as Code (WAC) orchestrators.

  Triggers: "build AI agent with Windmill", "integrate [API] with Windmill",
  "tool calling", "chat mode", "WAC vs Flow", "Windmill orchestration",
  "Windmill script hangs", "freeBusy error 403", "AI Agent memory",
  "orchestrator design", "script reuse across flows".

  This is a STRATEGY skill, not a tutorial. The official Windmill
  CLI skills (write-flow, write-workflow-as-code, write-script-*, etc)
  cover HOW to write code. This skill covers WHEN and WHY to choose
  each approach, plus gotchas discovered in real deployments.

  Does NOT cover: project-specific business decisions (identifiers,
  business rules, UI copy, etc). General patterns only.
---

# Windmill Orchestration Strategy

Decision frameworks and gotchas for Windmill self-hosted deployments.
Distilled from real production scenarios.

## When to Activate

Activate when:
- Choosing between WAC, Flow declarativo, and Flow visual for a new project
- Designing a script that will be consumed by multiple orchestration modes
- Building an AI agent with tool calling
- Debugging Windmill flow errors that aren't in the error message
- Setting up OAuth for a third-party API with custom scopes
- User asks "should I use WAC or Flow?"

## Decision: WAC vs Flow Declarativo vs Flow Visual

### Three orchestration modes

| Mode | What it is | When to use |
|------|-----------|-------------|
| **WAC** (Workflows as Code) | TypeScript/Python/Bun with `taskScript()` for custom logic | Custom orchestration: loops, retries, dynamic steps, error handling |
| **Flow declarativo** (YAML) | JSON/YAML structure of steps (scripts, AI agent, branches) | AI Agent + tools, chat mode, branching, parallel |
| **Flow visual** (UI drag-and-drop) | Same as Flow declarativo but edited in UI | Non-developer builders, visual debugging only |

### Choice framework

| Your situation | Recommended mode |
|---|---|
| AI Agent + tools + memory + multi-turn chat | **Flow declarativo** (native AI Agent step) |
| 5-10 sequential deterministic steps | **Flow declarativo** |
| Complex conditional/parallel/branching logic | **Flow declarativo** |
| Custom retry/loop/error handling logic | **WAC** |
| Dynamic step generation (e.g., loop over 1000 items) | **WAC** (via `taskScript()` per item) |
| Webhook handler with simple logic | **Script** (Deno/Bun), not a flow |
| Trigger from external API | **HTTP route** with script |

### LLM maintainer vs human maintainer

| Maintainer | Recommended mode | Why |
|---|---|---|
| **LLM will maintain code** (no human in loop) | **WAC** | TypeScript natural for LLMs; code review tools work; no YAML gotchas |
| **Human will maintain code** | **Flow declarativo** | YAML readable, version control simple, visual diff |
| **AI Agent + tools + flows (hybrid)** | **WAC orchestrator + Flow declarative tools** | WAC does LLM loop, Flow declarativo defines tool scripts |

Rule: if you (the LLM) are the only maintainer, prefer WAC. If a human will review the code, Flow declarativo is friendlier.

## Script Reuse: Single Source of Truth

A single tool script can be consumed by multiple orchestration patterns:

- Flow as Chat (AI Agent step tool description)
- WAC orchestrator via `taskScript(path, { timeout, tag })` (WAC v2)
- HTTP routes (direct invocation)
- Apps (form-based invocation)
- CLI testing (`wmill script run`)

**Example: same `check_availability.ts` script used by:**

1. **Flow as Chat** — `tools: [{ id: "check_availability", value: { type: script, path: "f/path/check_availability" } }]`
2. **WAC orchestrator** — `const result = await taskScript("f/path/check_availability")(args);`
3. **CLI** — `wmill script run f/path/check_availability -d '{"param":"value"}'`
4. **HTTP route** — script gets auto-webhook URL

Benefits:
- One place to fix bugs
- One place to add validation
- One place to test
- Tool schema (parameter names + descriptions) is the same for LLM and CLI

When designing a script for reuse:
- Use individual function params, never an `input` object
- Take `resource_path` strings as params, not hardcoded paths
- Take timezone, language, etc as params with sensible defaults
- Test standalone with `wmill script run` before wiring into Flow/WAC

## AI Agent Step (Flow Declarativo) — Gotchas

When using the AI Agent step in Flow as Chat:

1. **Tool `summary` must match `^[a-zA-Z0-9_]+$`** — Windmill uses it as
   the tool name sent to the LLM. Spaces/accents → "Invalid tool name" error.

2. **Tool `input_transforms` with `type: javascript` REMOVES that param
   from the LLM's schema.** Use `type: static` for fixed values, let
   the LLM pass dynamic params directly via its tool call args.

3. **`memory: { kind: auto }` is REQUIRED for chat mode context.** Without
   it, the agent forgets past messages each turn. Chat mode provides
   `memory_id` via `flow_status`, but the AI Agent only loads history if
   `memory` is configured as `auto`.

4. **`omit_output_from_conversation: true` causes chat UI to show NO
   responses.** It suppresses ALL messages including the final reply.
   Default is false; remove this field or set to false.

5. **System prompt must include explicit date + day of week.** LLMs
   cannot compute dates correctly. Inject `new Date()` server-side or
   pre-compute before pushing.

6. **`chat_input_enabled` lives in `value.chat_input_enabled`** for
   API update calls. Top-level persistence is unreliable.

7. **API endpoint uses `p/` prefix for scripts**:
   `POST /api/w/{ws}/jobs/run_wait_result/p/f/{path}`
   Without `p/` → "not found: flow not found at name..." (routing tries flow first)

## WAC (Workflows as Code) Pattern

When you need custom orchestration, use WAC with Bun runtime.

### Orchestrator template structure

```typescript
import {
  getResource,
  getVariable,
  setVariable,
  taskScript,
  workflow,
} from "windmill-client";

export const main: (
  text: string,
  memory_id: string,
  /* ... params ... */
) => Promise<Output> = workflow<Output>(async (
  text, memory_id, /* ... */
) => {
  // 1. Load conversation history
  const messages = await loadConversation(memory_id);
  messages.push({ role: "user", content: text });

  // 2. Get LLM credentials from resource
  const llmRes = await getResource<{ api_key: string }>(llmResourcePath);

  // 3. Tool-calling loop (ALWAYS set max iterations)
  const toolsCalled: string[] = [];
  let finalContent = "";
  for (let iter = 0; iter < MAX_ITERATIONS; iter++) {
    const llmResp = await callLLM(llmRes.api_key, model, messages);
    const msg = llmResp.choices?.[0]?.message;
    if (!msg) break;
    messages.push(msg);

    if (!msg.tool_calls?.length) {
      finalContent = msg.content || "";
      break;
    }

    for (const tc of msg.tool_calls) {
      const args = JSON.parse(tc.function.arguments);
      args.llmResourcePath = llmResourcePath;  // inject common params

      const result = await taskScript(`f/path/${tc.function.name}`)(args);
      toolsCalled.push(tc.function.name);
      messages.push({
        role: "tool",
        content: JSON.stringify(result),
        tool_call_id: tc.id,
        name: tc.function.name,
      });
    }
  }

  // 4. Save history
  await saveConversation(memory_id, messages);
  return { response: finalContent, /* ... */ };
});
```

### Critical WAC gotchas (v2 server >= 1.735)

1. **MUST wrap in `workflow(...)` (untyped!)** for `taskScript` to work.
   Without it, `taskScript` is not available. **DO NOT use type param
   `<T>`** — `workflow<T>(async ...)` breaks `auto_kind = "wac"` detection.

2. **`taskScript(path, opts)(args)` returns a FUNCTION** with native options:
   ```typescript
   const fn = taskScript("f/path/to/script", {
     timeout: 30,   // seconds, native (no Promise.race needed)
     tag: "deno",   // forces Deno worker (cross-runtime from Bun orchestrator)
   });
   const result = await fn({ param: "value" });
   ```

3. **WAC Bun scripts need `language: bun` in their `.script.yaml`.**
   Deno scripts use `language: deno` (or omit, defaults to deno).

4. **Cache bug**: Windmill caches Bun bundles by content hash. If a push
   doesn't take effect, edit the file (add a noop comment) to force a
   fresh bundle.

5. **WAC via CLI can hang**: `wmill script run` on a Bun WAC script
   that calls `taskScript` may hang indefinitely if the child job
   doesn't complete. Use UI Runs page to cancel. Worker slots consumed
   by hung jobs block other jobs.

6. **Push order matters**: `defaultTs: deno` for tool scripts, then
   switch to `bun` in wmill.yaml, push the WAC orchestrator, switch back.
   Or set `language:` per-script yaml.

7. **Server-side `language` can drift from YAML**: even with
   `language: deno` in the `.script.yaml`, the server may register the
   script as `bun` if it was pushed while `defaultTs: bun` was active.
   Symptom at runtime: `Invalid requirements, expected to find
   //bun.lock split pattern in reqs`. Verify with
   `wmill script get f/.../<script> --json | jq .language` for each
   deno tool. Fix: `wmill script push f/.../<script>.ts` again with
   `defaultTs: deno` in `wmill.yaml`.

8. **Variable paths use `/` not `_`**: `u/admin/conv/abc` works;
   `u_admin_conv_abc` fails with "proper_id" constraint violation.

## Custom OAuth for Third-Party APIs

Many providers (Google, Microsoft, etc) have Windmill built-in
resource types, but they often request **limited OAuth scopes** that
don't include what you need (e.g., Google Calendar built-in only
requests `calendar.events`, missing `calendar.readonly` for `freeBusy`).

### Pattern: custom OAuth resource

1. **Instance Settings → Resources → Add OAuth** (or "Connect")
2. Set:
   - Name: `gcal_full` (or whatever you want)
   - Auth URL: `<provider>/oauth/authorize`
   - Token URL: `<provider>/oauth/token`
   - Scopes: `<provider>/auth/your.needed.scopes` (space-separated)
   - Client ID/Secret: from your app's console
3. Save
4. Create resource of new type in workspace
5. Login with the provider — consent screen asks for all the scopes
6. Token is stored; use `getResource("f/path/to/your_resource")` in scripts

Now the token has the scopes you specified, including ones the
built-in resource type doesn't request.

## LLM Provider Setup (OpenRouter, OpenAI, etc)

Windmill supports many providers natively:
- `openai`, `azure_openai`, `anthropic`, `mistral`, `deepseek`,
- `googleai`, `groq`, `openrouter`, `togetherai`, `aws_bedrock`

For the API:
- **Base URL:** `https://<provider>/v1/chat/completions`
- **Auth:** `Authorization: Bearer $api_key`
- **Headers:** `HTTP-Referer` and `X-Title` recommended (OpenRouter)
- **Token limits:** vary by model/key tier. Check 402 errors for credit issues.

OpenRouter is a good default — one API, many models, easy to switch.

## Multi-Turn Conversation Memory

For chatbot orchestrators (WAC or AI Agent), you need to preserve
context across turns. Two approaches:

### WAC: manual via variables

```typescript
async function loadConversation(memoryId: string): Promise<Message[]> {
  const key = `u/admin/conversation/${memoryId}`;  // / not _
  try {
    const raw = await getVariable(key);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [{ role: "system", content: SYSTEM_PROMPT }];
  }
}

async function saveConversation(memoryId: string, messages: Message[]) {
  const key = `u/admin/conversation/${memoryId}`;
  await setVariable(key, JSON.stringify(messages.slice(-50)));  // cap at 50
}
```

`memory_id` is generated externally (UUID, phone hash, etc).

### Flow as Chat: automatic via `memory: { kind: auto }`

AI Agent step with this config automatically loads history from the
chat conversation storage (`flow_conversations` table). No manual
save/load needed.

But: `memory` is REQUIRED. Without it, each turn starts fresh.

## Test Patterns

```bash
# Test script standalone (fastest, no LLM)
wmill script run f/path/to/script -d '{"param":"value"}'

# Test WAC orchestrator with same params
wmill script run f/path/to/orchestrator -d '{
  "text":"test message",
  "memory_id":"test-1",
  "llmResourcePath":"u/admin/openrouter"
}'

# Test Flow with chat mode (requires UI - CLI doesn't simulate chat runtime)
# Go to: /flows/edit/{path} and use the chat interface
```

## Common Pitfalls

1. **`wmill script run` on WAC can hang** — if child jobs block, workers
   fill up. Cancel via UI Runs page.

2. **WAC variable names with `_`** — fail with "proper_id" constraint.
   Use `/`.

3. **Forgetting `value.chat_input_enabled: true`** in API update —
   silently ignored. Set both top-level AND `value.chat_input_enabled`.

4. **`memory: { kind: auto }` in AI Agent** — without it, agent
   forgets each turn in chat mode.

5. **Tool `summary` with spaces** — fails "Invalid tool name" error.
   Must be `^[a-zA-Z0-9_]+$`.

6. **`omit_output_from_conversation: true`** — chat UI shows nothing.

7. **API endpoint without `p/` prefix** — routing tries flow first,
   fails with "not found: flow not found".

## Deep Dives

For operational gotchas that take hours to debug (token scopes, lockfile
mismatch, customai provider quirks, reasoning model handling), see:

- **[references/gotchas.md](references/gotchas.md)** — token scopes,
  defaultTs/lockfile format mismatch, Flow as Chat context variables,
  customai provider body limitations, reasoning/think block handling,
  runtime mixing (Deno+Bun), variable size limits, WAC v2 detection.
  wmill 1.722 known bugs, cache invalidation, output shape for chat.

- **[examples/wac-plus-flow-thin-wrapper.md](examples/wac-plus-flow-thin-wrapper.md)**
  — DRY pattern: tools + WAC orchestrator + thin Flow wrapper.
  Single source of truth for system prompt, tools, memory. Use when AI
  Agent step's body param limitations (no extra_body, no reasoning_split)
  block you, or when you want provider flexibility.

## Official Windmill Skills Directory

These skills are created by `wmill init` in the project's `.agents/skills/`
or `.claude/skills/`. They cover HOW to write code in each aspect.
This skill (the one you're reading) covers WHEN and WHY — it does
NOT replace these sub-skills.

The official skills can be referenced by path: project's `.agents/skills/`.

### How to use this directory

When you need to write a specific type of Windmill asset, first read
this skill for strategy, then read the appropriate sub-skill for syntax.

### Script writing (per language)
| Skill | What to write | Activate when |
|---|---|---|
| `write-script-deno` | Deno TypeScript scripts (individual function params) | Writing or editing a Deno script |
| `write-script-bun` | Bun TypeScript scripts (workflow, taskScript) | Writing or editing a Bun script |
| `write-script-python3` | Python scripts | Writing or editing a Python script |
| `write-script-go` | Go scripts | Writing or editing a Go script |
| `write-script-bash` | Bash scripts | Writing or editing a Bash script |
| `write-script-rust` | Rust scripts | Niche, performance-critical |
| `write-script-php` | PHP scripts | Webhook endpoints |
| `write-script-csharp` | C# scripts | Enterprise integrations |
| `write-script-java` | Java scripts | Enterprise integrations |
| `write-script-rlang` | R scripts | Data analysis |
| `write-script-snowflake` | Snowflake SQL | Data warehouse queries |
| `write-script-postgresql` | PostgreSQL SQL | Database queries |
| `write-script-mysql` | MySQL SQL | Database queries |
| `write-script-mssql` | MSSQL SQL | Database queries |
| `write-script-duckdb` | DuckDB SQL | Embedded analytics |
| `write-script-bigquery` | BigQuery SQL | Cloud data warehouse |
| `write-script-bunnative` | Native Bun (non-WAC) | High-performance Deno alternative |
| `write-script-graphql` | GraphQL queries | API endpoints |
| `write-script-powershell` | PowerShell scripts | Windows automation |

### Flow orchestration
| Skill | What to write | Activate when |
|---|---|---|
| `write-flow` | Flow declarativo YAML (AI Agent branch, loop, parallel) | Writing or editing a Flow YAML |
| `write-workflow-as-code` | WAC Bun (workflow, task, taskScript, step) | Writing or editing a WAC orchestrator |

### Infrastructure
| Skill | What to configure | Activate when |
|---|---|---|
| `triggers` | Webhooks, HTTP routes, email triggers, Kafka, NATS, etc | Setting up a trigger or HTTP route |
| `schedules` | Cron schedules for scripts/flows | Setting up recurring jobs |
| `resources` | OAuth, API keys, custom resource types | Creating or managing resources |
| `cli-commands` | wmill CLI reference (commands, flags, env vars) | Using the wmill CLI |
| `preview` | Local testing without deploying (wmill script preview) | Testing scripts locally |
| `raw-app` | Windmill Apps (low-code) | Building internal tools with App editor |

### Decision matrix: which sub-skill to activate

| You are doing... | Read this sub-skill first |
|---|---|
| Choosing between WAC and Flow | `wmill-orchestration` (this skill) |
| Writing a Deno script | `write-script-deno` |
| Writing a Bun WAC orchestrator | `write-workflow-as-code` |
| Debugging AI Agent step | `wmill-orchestration` (AI Agent Gotchas section) |
| Setting up Google Calendar OAuth | `wmill-orchestration` (Custom OAuth section) |
| Creating a webhook | `triggers` |
| Setting up a cron | `schedules` |
| Creating OAuth resource | `resources` |
| Testing locally | `preview` |
| Using wmill command | `cli-commands` |

## Related Resources

- Windmill docs: https://www.windmill.dev/docs
- AI Agent docs: https://www.windmill.dev/docs/core_concepts/ai_agents
- Flow as Chat: https://www.windmill.dev/docs/flows/flow_as_chat
- OpenFlow spec: https://github.com/windmill-labs/windmill/blob/main/openflow.openapi.yaml
