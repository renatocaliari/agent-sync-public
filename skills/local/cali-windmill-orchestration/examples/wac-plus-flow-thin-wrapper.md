# WAC + Flow Thin Wrapper Pattern

Use when: tools are LLM-callable scripts, you want single source of truth
across WAC and Flow as Chat entry points, and you need provider
flexibility (custom base_url, custom body params).

## Runtime: Bun (always)

**Use Bun como runtime único. Nunca misture Deno + Bun no mesmo projeto.**

Razões:
- Lockfile Deno tem bug recorrente (`//bun.lock` header trava worker)
- Cross-runtime taskScript trava (Bun→Deno sem `tag: "deno"` trava)
- Padronização > soma das vantagens individuais

```typescript
// ✅ CORRETO: Bun, bare import
import { getResource, getVariable, setVariable, workflow, taskScript } from "windmill-client";
const config = process.env.MY_VAR || Bun.env?.MY_VAR;

// ❌ ERRADO: Deno, npm: prefix
// import { getResource } from "npm:windmill-client@1.722.0";
// const config = Deno.env.get("MY_VAR");
```

```yaml
# .script.yaml — SEMPRE "bun"
language: bun
```

Migrar Deno → Bun:
1. `npm:windmill-client@1.722.0` → `windmill-client`
2. `Deno.env.get(x)` → `process.env.x || Bun.env?.x`
3. `language: deno` → `language: bun`
4. `wmill generate-metadata` + `wmill script push`

## WAC v1 vs v2 vs Direct Import (recommended)

| | WAC v1 (legacy) | WAC v2 (server >= 1.735) | **Direct Import** 🏆 |
|---|---|---|---|
| **Wrapper** | `workflow<T>(...)` | `workflow(...)` | `workflow(...)` ou sem wrapper |
| **Tool call** | `taskScript(path)(args)` | `taskScript(path, { timeout, tag })(args)` | `import { main as fn }; fn(...)` |
| **Timeout** | Manual `Promise.race` | Native (segundos) | Try/catch (síncrono) |
| **Worker isolation** | Child job | Child job | **Mesmo processo** |
| **Deadlock risk** | Alto | Baixo (checkpoint) | **Zero** |
| **Cross-runtime** | ✅ Bun→Deno | ✅ Com tag | ❌ Mesmo runtime |
| **auto_kind detection** | ✅ wac | Server>=1.735 | ✅ script (não precisa) |

**Verifique antes de escolher:**
```bash
wmill script get f/path --json | python3 -c "import json,sys; print(json.loads(sys.stdin.read())['auto_kind'])"
wmill version  # servidor
```

**Escolha o pattern baseado no runtime das tools:**

| Tools em... | Pattern recomendado |
|---|---|
| **Mesmo runtime** (ex: Bun) | **Direct Import** 🏆 — zero child jobs, sem deadlock |
| **Runtime diferente** (Python, Go, Deno sem migração) | WAC v2 com `{ timeout, tag }` |
| **Server < 1.735** | WAC v1 + `Promise.race` manual |

**Critical pattern mistakes (WAC v2):**
1. ❌ `workflow<T>(async ...)` (typed) — quebra auto_kind detection
2. ❌ Wrapper around export — deve ser `export const main = workflow(...)`

**Critical pattern mistakes (Direct Import):**
1. ❌ Import de `npm:windmill-client` em tool → trocar pra `windmill-client`
2. ❌ `language: deno` no script.yaml → trocar pra `language: bun`
3. ⚠️ Tools devem aceitar params individuais (orchestrator faz dispatch via switch)

## The Pattern

```
f/your-feature/
├── tool_a.ts            ← script: single source, used by both
├── tool_b.ts            ← script
├── orchestrator.ts      ← WAC: system prompt + LLM loop + tool dispatch
└── your-flow__flow/
    └── flow.yaml        ← 1 module: script step calling orchestrator
```

**Tool scripts:** standalone, with `input_transforms` for static values
and exposed args for LLM-supplied params. Mixed runtimes (Bun + Deno) OK
when orchestrator specifies `tag` per task.

**Orchestrator (WAC):**
- Owns the system prompt (1 source)
- Owns the LLM provider config (reads from Windmill resource)
- Owns the tool definitions (passed to LLM as JSON Schema)
- Owns the memory logic (`getVariable`/`setVariable`)
- Calls tools via `taskScript("f/path/tool")({...args})`
- Returns `{ windmill_chat_answer: response, ...debug }` for chat UI

**Flow (thin wrapper):**
- Schema declares `user_message` + `memory_id` (auto-injected by chat runtime)
- Single script step that calls the orchestrator
- Static `input_transforms` for: `llm_resource_path`, `gcal_resource_path`,
  any config that's the same for every conversation
- Zero business logic in the flow

## File: orchestrator.ts

```typescript
import {
  getResource,
  getVariable,
  setVariable,
  taskScript,
  workflow,
} from "windmill-client";

const SYSTEM_PROMPT = `You are a helpful assistant.

## Tools
- tool_a: does X. Args: foo (string, required), bar (number, optional)
- tool_b: does Y. Args: baz (string, required)
`;

type Message = {
  role: "system" | "user" | "assistant" | "tool";
  content: string | null;
  tool_calls?: Array<{ id: string; type: "function"; function: { name: string; arguments: string } }>;
  tool_call_id?: string;
  name?: string;
};

export const main: (
  text: string,
  memory_id: string,
  llm_resource_path: string,
) => Promise<any> = workflow<any>(async (text, memory_id, llm_resource_path) => {

  // 1. Load history
  const key = `u/admin/conversation/${memory_id}`;
  let messages: Message[];
  try {
    const raw = await getVariable(key);
    messages = raw ? JSON.parse(raw) : [{ role: "system", content: SYSTEM_PROMPT }];
  } catch {
    messages = [{ role: "system", content: SYSTEM_PROMPT }];
  }
  messages.push({ role: "user", content: text });

  // 2. Get LLM config
  const llmRes = await getResource<{ api_key: string; base_url?: string }>(llm_resource_path);
  const apiKey = llmRes.api_key;
  const baseUrl = llmRes.base_url || "https://api.openai.com/v1";
  if (!apiKey) throw new Error(`Resource ${llm_resource_path} missing api_key`);

  // 3. LLM loop
  const toolsCalled: string[] = [];
  let finalContent = "";
  for (let iter = 0; iter < 5; iter++) {
    const resp = await callLLM(apiKey, baseUrl, messages);
    const msg = resp.choices?.[0]?.message;
    if (!msg) break;

    // Strip think blocks (reasoning models)
    if (typeof msg.content === "string") {
      msg.content = msg.content.replace(
        new RegExp("<think>[\\s\\S]*?</think>\\s*", "g"), ""
      ).trim();
    }
    messages.push(msg);

    if (!msg.tool_calls?.length) {
      finalContent = msg.content || "";
      break;
    }

    for (const tc of msg.tool_calls) {
      const args = JSON.parse(tc.function.arguments);
      toolsCalled.push(tc.function.name);
      try {
        // WAC v2 nativo: timeout + tag + checkpoint automatico
        // (requer server >= 1.735 + auto_kind="wac" detectado)
        const result = await taskScript(
          `f/your-feature/${tc.function.name}`,
          { timeout: 30, tag: "deno" }
        )(args);

        messages.push({
          role: "tool",
          content: typeof result === "string" ? result : JSON.stringify(result),
          tool_call_id: tc.id,
          name: tc.function.name,
        });
      } catch (err: any) {
        messages.push({
          role: "tool",
          content: `Error: ${err.message}`,
          tool_call_id: tc.id,
          name: tc.function.name,
        });
      }
    }
  }

  // 4. Save history (capped)
  await setVariable(key, JSON.stringify(messages.slice(-50)));

  // 5. Return shape optimized for chat UI
  return {
    response: finalContent,
    tools_called: toolsCalled,
    windmill_chat_answer: finalContent,  // chat UI shows this, ignores other fields
  };
});

async function callLLM(apiKey: string, baseUrl: string, messages: Message[]): Promise<any> {
    tool_choice: "auto",
    max_tokens: 1024,
    temperature: 0.3,
    // Provider-specific: reasoning_split, thinking, etc
    reasoning_split: true,
  };

  const resp = await fetch(`${baseUrl.replace(/\/+$/, "")}/chat/completions`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${apiKey}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });

  if (!resp.ok) throw new Error(`LLM API ${resp.status}: ${await resp.text()}`);
  return resp.json();
}

function buildToolDefinitions() {
  return [
    {
      type: "function",
      function: {
        name: "tool_a",
        description: "Does X",
        parameters: {
          type: "object",
          properties: {
            foo: { type: "string", description: "required" },
            bar: { type: "number" },
          },
          required: ["foo"],
        },
      },
    },
    // ... more tools
  ];
}
```

## File: flow.yaml

```yaml
summary: Your feature (thin wrapper over WAC orchestrator)
description: |
  Flow as Chat wrapper. All logic (prompt, tools, memory) lives in
  f/your-feature/orchestrator. This file only routes inputs.
value:
  modules:
    - id: a
      summary: Call orchestrator WAC
      value:
        type: script
        path: f/your-feature/orchestrator
        input_transforms:
          text:
            type: javascript
            expr: "flow_input.user_message || ''"
          memory_id:
            type: javascript
            expr: "flow_input.memory_id || ('flow-' + Date.now())"
          llm_resource_path:
            type: static
            value: f/your-feature/llm-resource
  schema:
    $schema: https://json-schema.org/draft/2020-12/schema
    type: object
    properties:
      user_message:
        type: string
        description: Auto-injected in chat mode
      memory_id:
        type: string
        description: Auto-injected in chat mode
  chat_input_enabled: true
```

## Resource: LLM provider

```yaml
# f/your-feature/llm-resource.resource.yaml
summary: Your LLM provider
description: OpenAI-compatible LLM credentials
value:
  api_key: $var:f/your-feature/api_key
  base_url: https://api.provider.com/v1
  headers:
    X-Session-Affinity: your-app
```

```bash
wmill resource-type new openai_compatible
# Edit .resource-type.yaml with schema {api_key, base_url, headers}
wmill resource-type push openai_compatible.resource-type.yaml openai_compatible
wmill resource push llm-resource.resource.yaml f/your-feature/llm-resource
```

## When NOT to use this pattern

- **No provider quirks, standard OpenAI/Anthropic:** use Flow AI Agent
  directly with `memory: { kind: auto }`. Less code, native memory.
- **No tools, just chat:** Flow AI Agent with no tools is simpler.
- **Complex custom retry/error logic:** WAC without Flow (call WAC via
  HTTP route or schedule, skip the flow wrapper).
- **Multiple unrelated entry points:** if 3+ different ways to trigger
  the same logic, each gets its own flow/handler, all call the WAC.

## Testing the pattern

```bash
# Smoke test WAC directly (no flow):
wmill script run f/your-feature/orchestrator -d '{
  "text": "hello",
  "memory_id": "smoke-1",
  "llm_resource_path": "f/your-feature/llm-resource"
}'

# Then test in Flow as Chat UI:
# https://<host>/flows/edit/f/your-feature/your-flow
```

## Migration from AI Agent → WAC

When AI Agent step's limitations (no extra_body, no reasoning_split, no
custom base URL) block you:

1. Create `orchestrator.ts` (WAC) — move system prompt + tool defs
2. Update flow.yaml: replace AI Agent module with script step
3. Reuse existing `llm-resource` (no change)
4. Reuse existing tool scripts (no change)
5. Test smoke → chat UI → verify `windmill_chat_answer` displays

Don't delete the old AI Agent step until WAC version validated.
