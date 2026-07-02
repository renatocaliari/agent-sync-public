# Windmill Operational Gotchas

Deep dives on issues that cost hours of debugging. Load this when
something is silently broken (push fails, run errors, response is wrong).

## Token scopes (silent push failure)

`wmill` CLI does NOT show token scopes. Symptoms of insufficient scopes:

| Scope missing | Symptom |
|---|---|
| `flows:write` | `Failed to create flow: Permission denied: Access denied. Required scope: flows:write` |
| `flows:read` | `Forbidden: Permission denied: Access denied. Required scope: flows:read` |
| `jobs:run` | `Failed to queue dependencies job: 403 Forbidden, Required scope: jobs:run` |
| `jobs:run:scripts` | Script execution blocked, but push works |
| `users:write` | Can't self-promote scopes via API |

**Check current token:** `wmill token list` shows prefix + label but
NOT scopes. Use API:
```bash
curl -s https://<host>/api/tokens/list/scopes \
  -H "Authorization: Bearer <token>" | jq '.[].name'
```

**Fix:** edit `~/Library/Preferences/windmill/remotes.ndjson` (macOS) or
`$XDG_CONFIG_HOME/windmill/remotes.ndjson` (linux) — swap `token` field
for one with full scopes. Get a full-access token via Windmill UI →
Settings → Tokens → Create with all scopes.

**Auto-promote via API does NOT work** for limited tokens:
`POST /users/tokens/update_scopes/{prefix}` requires `users:write`
on the calling token (chicken-and-egg).

## defaultTs vs lockfile format

`wmill.yaml` `defaultTs` controls what runtime workers USE for the script.
Mismatch with `language` in `.script.yaml` causes silent failures.

| `defaultTs` | `language` in .yaml | Result |
|---|---|---|
| `deno` | `bun` | Worker runs as deno, fails on `//bun.lock` header |
| `bun` | `deno` | Worker runs as bun, fails on deno-style imports |
| `deno` | `deno` | ✅ |
| `bun` | `bun` | ✅ |

**Lockfile detection:**
- **JSON puro** (`{"lockfileVersion":1, "configVersion":1, ...}`) → bun format, OK
- **`//bun.lock` header + JSON** → bun format antigo, FAIL em deno worker
- **Plain JSON without header** → both runners OK

**Symptom of mismatch:** `error: Failed reading lockfile at 'lock.json':
Caused by: Failed parsing. Lockfile may be corrupt`

**Fix:** align yaml or regenerate metadata:
```bash
# 1. Edit wmill.yaml defaultTs
# 2. Regenerate lock for all scripts
wmill generate-metadata f/your-folder/ -r
# 3. Push
wmill script push f/your-folder/<script>.ts
```

## Flow as Chat context (NOT `flow_status`)

In `chat_input_enabled: true` flows, runtime injects:
- `flow_input.user_message` — the user's latest message
- `flow_input.memory_id` — UUID of the conversation

`flow_status` is **NOT available** in input_transforms expressions.
Common mistake: copying an AI Agent step expression that uses
`flow_status.conversation_id` — fails with:
```
Error during isolated evaluation of expression flow_status.X:
QuickJS evaluation error: Error: flow_status is not defined
```

**Correct pattern:**
```yaml
modules:
  - id: a
    value:
      type: script
      path: f/path/to/script
      input_transforms:
        text:
          type: javascript
          expr: "flow_input.user_message || ''"
        memory_id:
          type: javascript
          expr: "flow_input.memory_id || ('flow-' + Date.now())"
schema:
  type: object
  properties:
    user_message:
      type: string
    memory_id:
      type: string
  required: [user_message]
chat_input_enabled: true
```

**Critical:** `memory_id` MUST be in `schema.properties` for it to be
available as `flow_input.memory_id`. Without it, runtime doesn't inject.

## customai provider (OpenAI-compatible with quirks)

`kind: customai` in AI Agent step is the escape hatch for non-builtin
providers (MiniMax, OpenRouter via custom base, local LLMs).

**Resource schema (built-in):**
```json
{
  "api_key": "<key>",
  "base_url": "https://api.provider.com/v1",
  "headers": { "X-Custom-Header": "value" }
}
```

**Authorization header is auto-injected** as `Authorization: Bearer ${api_key}`.
Do NOT duplicate in `headers`.

**ProviderConfig limits (AI Agent step):**
- ❌ No `extra_body` pass-through
- ❌ No `reasoning_split` / `reasoning_effort` controls
- ❌ No custom query params

**Workaround for body quirks:** use WAC orchestrator (script step) that
makes raw `fetch()` to `${baseUrl}/chat/completions`. Flow becomes thin
wrapper calling WAC.

## Reasoning / think blocks

Models that return thinking (MiniMax M3, DeepSeek R1, Claude extended thinking)
leak `<think>...</think>` into `content` by default. Chat UI shows noise.

**API-level fix (preferred):** request provider to split thinking from content.

| Provider | Fix |
|---|---|
| MiniMax M3 | `reasoning_split: true` in body → `reasoning_details` field separate, `content` clean |
| DeepSeek R1 | `reasoning_content` always separate field, no opt-in needed |
| Claude | `thinking: { type: "enabled", budget_tokens: N }` → `thinking` block separate |
| OpenAI o1/o3 | `reasoning_tokens` in usage, content clean by default |

**Client-level fallback (any provider):** strip before returning to user.
```typescript
function stripThinkBlock(content: string): string {
  // Use string-regex to avoid TS parser issues with /<think>/ in literals
  return content.replace(
    new RegExp("<think>[\\s\\S]*?</think>\\s*", "g"),
    ""
  ).trim();
}
```

Apply in WAC orchestrator AFTER receiving LLM response, BEFORE returning.

## wmill 1.722 known bugs

| Bug | Workaround | Fixed in |
|---|---|---|
| `wmill flow push <file>` adds `/flow.yaml` to path → ENOTDIR | Pass path ending in `/` (folder) | 1.735 |
| No `wmill resource delete` subcommand | Use Windmill UI | Later |
| No `wmill variable delete` subcommand | Use Windmill UI | Later |
| `wmill generate-metadata` requires `jobs:run` scope | Use full-access token | N/A |
| `wmill token update-scopes` subcommand missing (API exists) | Use UI | N/A |

## Cache invalidation

Windmill caches Bun bundles by content hash. If a script push "doesn't
take effect" but the response is `Updated`:

```bash
# Force fresh bundle: add a noop comment
echo '// cache bust' >> f/path/script.ts
wmill script push f/path/script.ts
```

WAC scripts pushed with stale cache cause confusing "old behavior" bugs
even after logic changes. This is a known gotcha — check if behavior
changed before debugging logic.

## WAC + taskScript hang (orchestrator trava em "Processing...")

**Sintoma:** Flow as Chat mostra "Processing..." indefinido. Job aparece
como `running` por minutos. `wmill job logs <id>` mostra última linha
tipo `--- WAC: check_availability ---` sem completion.

**Causa:** Orchestrator (Bun) chama `taskScript("f/path/tool")`. Windmill
cria job filho (pode ser deno). Job filho trava por:
- Lockfile deno corrompido (worker hung)
- Cross-runtime bun→deno (incompatibilidade)
- Worker zombie de runs anteriores
- Tool chama API externa (Google Calendar) com retry infinito

**Diagnóstico:**
```bash
wmill job list --limit 5    # procura "running" há muito tempo
wmill job get <id>          # vê step atual
wmill job logs <id>         # última linha identifica onde travou
```

**Fix imediato:**
```bash
wmill job cancel <id>       # mata o job travado
```

**Fix preventivo (no orchestrator):** adicionar timeout no `executeTool`:
```typescript
async function executeTool(fnName: string, args: any): Promise<any> {
  const script = taskScript(`f/agendamento/${fnName}`);
  const TIMEOUT_MS = 30_000;
  return await Promise.race([
    script(args),
    new Promise<never>((_, reject) =>
      setTimeout(() => reject(new Error(`${fnName} timeout`)), TIMEOUT_MS)
    ),
  ]);
}
```

Catches exception e injeta no `messages.push({ role: "tool", content: "Tool X timeout..." })`.
LLM responde com fallback gracioso ("tive problema técnico, tenta de novo") em vez de
deixar o usuário pendurado.

**Fix definitivo (verified):** **direct import** de tools no orchestrator
(mesmo runtime Bun). Elimina `taskScript` e child jobs completamente.

 1. **Tools como Bun** — import `from "windmill-client"` (bare), `language: bun`
 2. **Orchestrator importa funções** — `import { main as checkAvail } from "./check_availability.ts"`
 3. **Dispatcher switch/case** — mapeia args object da LLM pra params individuais
 4. **Zero child job, zero checkpoint, zero deadlock**

Recomendação: **sempre usar direct import** quando tools e orchestrator podem
compartilhar o mesmo runtime (Bun). Tools continuam standalone pra CLI
(`wmill script run`) — nada se perde.

**Quando NÃO usar direct import:** se tools são Python/Go/Deno e converter é
inviável. Nesse caso, use `taskScript(path, { timeout, tag })` com WAC v2
checkpoint (server >= 1.735) + limite de 30s.

## WAC v2 detection quebrada em server < 1.735 (issue #8951)

**Sintoma:** Você escreve `export const main = workflow(async (...) => {...})`
com `taskScript(path, { timeout, tag })` (syntax oficial WAC v2 2026). Mas:
- Job nao usa checkpoint (parent espera blocking)
- `taskScript(path, opts)` ignora silenciosamente o `{ timeout, tag }`
- `wmill script get f/path --json` mostra `auto_kind: null` (não `wac`)

**Causa:** Windmill server < 1.735 tem bug na detecção automática de WAC.
O heuristic só roda se NÃO tem função `main`, então o template padrão cai pra
`auto_kind = NULL`. Issue #8951 (merged em 1.735+) fix isso.

**Diagnóstico:**
```bash
wmill script get f/path --json | python3 -c "import json,sys; print(json.loads(sys.stdin.read())['auto_kind'])"
# "wac" = WAC v2 active, native timeout/tag/checkpoint working
# "script"/None = legacy path, opts ignoradas
```

**Workaround enquanto server < 1.735:**
- Use legacy `taskScript(path)` + manual timeout via `Promise.race` (seção acima)
- Adicione `tag` no script yaml: `tag: "deno"` força runtime do worker
- Se precisar cross-runtime routing, use Flow AI Agent ao invés de WAC

**Fix permanente (server upgrade):**
```bash
# Self-hosted: SSH + docker
ssh deploy@server
cd /opt/windmill
docker exec windmill-db-1 pg_dump -U postgres windmill | gzip > /tmp/backup-$(date +%Y%m%d).sql.gz
docker compose pull
docker compose up -d --no-deps windmill_server windmill_worker windmill_worker_native windmill_extra
# Wait 30s for warm-up
curl -s https://<host>/api/version  # confirm 1.735+
# Re-push orchestrator
wmill script push f/path/to/orchestrator.ts
wmill script get f/path --json | python3 -c "import json,sys; print(json.loads(sys.stdin.read())['auto_kind'])"
# Should print "wac"
```

Depois do upgrade, troque `taskScript(path)` por `taskScript(path, { timeout, tag })`
e remova o `Promise.race` manual. O checkpoint/replay passa a funcionar
automaticamente (parent suspende enquanto child roda, worker slot liberado).

**Critical: NÃO use type param no wrapper:**
```typescript
// ❌ BUG: typed wrapper breaks auto_kind detection
export const main = workflow<{ text: string }>(async (text) => { ... });

// ✅ CORRETO: untyped wrapper
export const main = workflow(async (text: string) => { ... });
```

## Runtime mixing (Deno + Bun) — sempre evitar

**Decisão de arquitetura: 1 runtime por projeto. Use Bun para tudo.**

Misturar Deno e Bun no mesmo projeto causa:
- Cross-runtime taskScript trava (Bun→Deno sem `tag: "deno"` trava)
- Lockfile corrompido (Deno tem bug `//bun.lock` que trava worker)
- Mental overhead (2 runtimes = 2 mental models)
- Debug complexo (logs diferentes, env var APIs diferentes)

### Como migrar Deno → Bun

```diff
- // Deno
- import { getResource } from "npm:windmill-client@1.722.0";
- const token = Deno.env.get("API_TOKEN");
+ // Bun
+ import { getResource } from "windmill-client";
+ const token = process.env.API_TOKEN || Bun.env?.API_TOKEN;
```

```diff
# .script.yaml
- language: deno
+ language: bun
```

Depois: `wmill generate-metadata <script>.ts` + `wmill script push <script>.ts`

### Quando Deno faz sentido (raro)

- Precisa de versioning EXPLICITO de npm packages (`npm:windmill-client@1.722.0`)
- Precisa do permissions model do Deno (`--allow-net`, `--allow-read`)
- Tool é Python ou Go (nesses casos, runtime não é Deno nem Bun — é Python/Go)

**Na dúvida: Bun.** Padronização > otimização.

## Output shape for chat flows

If the final step returns an object, Windmill chat UI:
- Default: pretty-prints entire JSON as assistant message
- With `windmill_chat_answer: "string"` field: uses that string as message,
  other keys still in flow output but hidden from chat transcript
- With `windmill_chat_answer: null`: no message appended (silent turn)

**Pattern:** always return `{ windmill_chat_answer: <display text>, ... }`
from final step in chat flows. Other fields (debug info, tool counts) stay
in flow output without polluting the chat UI.

## Flow as Chat schema placement

O schema do flow DEVE estar no TOP LEVEL do YAML, NÃO dentro de `value`.

❌ ERRADO (servidor ignora, schema fica null):
```yaml
value:
  modules: [...]
  schema: { type: object, properties: { user_message: { type: string } } }
  chat_input_enabled: true
```

✅ CORRETO:
```yaml
value:
  modules: [...]
  chat_input_enabled: true
schema:
  type: object
  properties:
    user_message: { type: string, description: Auto-injected em chat mode }
    memory_id: { type: string, description: Conversation ID (UUID ou string) }
```

**Diagnóstico:**
```bash
wmill flow get f/path --json | python3 -c "import json,sys; d=json.loads(sys.stdin.read()); print('schema:', 'present' if d.get('schema') else 'NULL')"
# "NULL" = schema mal posicionado
```

## memory_id em script steps (não AI Agent)

Flow as Chat injeta `user_message` automaticamente se o schema declara.
Mas `memory_id` SÓ é injetado se o schema declara o campo E o runtime
consegue injetar. Para script steps NÃO é automático como no AI Agent.

**Pra garantir memória entre turns em script step:**
```yaml
input_transforms:
  memory_id:
    type: static
    value: chat-ui-conversation  # fixo, todas as turns recebem o mesmo
```

Isso SOBRESCREVE o memory_id da UI (UUID). Intencional — o orchestrator
usa esse ID fixo pra salvar/carregar conversa na variável.

**Atenção:** o Flow as Chat runtime PRECISA de um UUID para a conversa
(sidebar, histórico). Ele gera internamente — nosso static não interfere.

## Variable size limit (15K chars) + empty array [] bug

Windmill variables (PG varchar) têm limite de **15000 caracteres**. Se a
conversa salva excede, `setVariable` falha com:
```
SqlErr: value too long for type character varying(15000)
```

Fix: truncar antes de salvar:
```typescript
async function saveConversation(memId: string, msgs: Message[]): Promise<void> {
  const key = `u/admin/conversation/${memId}`;
  const serialized = JSON.stringify(msgs.slice(-50));
  if (serialized.length > 14000) {
    const sys = msgs.find(m => m.role === "system") || msgs[0];
    const tail = msgs.filter(m => m !== sys).slice(-10);
    await setVariable(key, JSON.stringify([sys, ...tail]));
  } else {
    await setVariable(key, serialized);
  }
}
```

**Outro bug sutil:** `[]` (array vazio) é truthy em JavaScript.
```typescript
const raw = await getVariable(key);  // "[]"
if (raw) return JSON.parse(raw) as Message[];
// retorna [] vazio, SEM system prompt!
```

Sempre verificar `raw !== "[]"`:
```typescript
if (raw && raw !== "[]") {
  return JSON.parse(raw) as Message[];
}
// continua => cria conversa nova com system prompt
```

## Direct import elimina taskScript (mais robusto)

Se tools e orchestrator estão no mesmo runtime (Bun), **importe direto**
em vez de usar `taskScript`:

```typescript
import { main as checkAvailability } from "./check_availability.ts";

async function executeTool(fnName: string, args: any): Promise<any> {
  switch (fnName) {
    case "check_availability":
      return await checkAvailability(args.professional_email, args.date, ...);
    default:
      throw new Error(`Unknown tool: ${fnName}`);
  }
}
```

**Vantagens:**
- Zero child job → zero deadlock (sem "Processing..." infinito)
- Erro vira exception normal (catch no orchestrator)
- Não precisa de WAC v2 checkpoint (não há taskScript pra suspender)
- Push mais rápido (tools compilam no bundle do orchestrator)

**Desvantagens:**
- Tools precisam ser MESMO runtime que orchestrator
- Se tool crasha, orchestrator crasha (tools compartilham processo)
- Switch/case dispatcher precisa ser mantido manualmente (~5 linhas por tool)

**Quando usar taskScript ainda:** tools em runtime diferente (Python, Go,
Deno sem conversão). Nesse caso, SEMPRE adicionar timeout:
```typescript
const script = taskScript(`f/path/${fnName}`, { timeout: 30, tag: "deno" });
```
