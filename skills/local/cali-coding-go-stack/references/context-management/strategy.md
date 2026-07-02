# Gerenciamento de Contexto para Prompts LLM

Estratégia híbrida para sessões longas no Treinador de Práticas Narrativas.

## Arquitetura

```
Prompt final = [ENCONTROS] (sumários) + [CONTEXTO-ATUAL] (termos-chave) + tools (FTS5, opcional)
                ↕                        ↕                            ↕
         sliding window          extractKeyTerms                 pb-plugin-fts
         (buildMeetingSummaries)  (bbalet/stopwords)             (SQLite FTS5)
```

## Camadas

### 1. Sliding Window + Sumarização (sempre ativo)

- **Mensagens recentes** (últimas N): vão integrais no prompt
- **Mensagens antigas**: comprimidas em sumários narrativos por encontro (`[ENCONTROS]`)
- **Termos-chave**: extraídos deterministicamente via `extractKeyTerms` com `bbalet/stopwords` (`[CONTEXTO-ATUAL]`)
- Implementado em: `buildHistoryWithMeetings`, `history_enrichment.go`, `key_terms.go`

### 2. FTS5/BM25 (opcional, quando pb-plugin-fts estiver ativo)

- Indexa `messages.content` em tabela virtual SQLite FTS5
- `extractKeyTerms` gera os termos de busca
- Resultados injetados no prompt como contexto adicional (busca **proativa**, sem tool)
- Alternativa tool: `buscar_mensagens` em `tools.go` pode usar FTS5 em vez de `LIKE` se reativada

### 3. superfly/contextwindow (opcional, avaliação futura)

- https://github.com/superfly/contextwindow
- Biblioteca Go para gerenciamento automático de contexto com contagem de tokens, compressão via sumário, persistência SQLite
- Avaliar quando v1.0+. Pode substituir `buildHistory*` no futuro se a estratégia atual mostrar limitação.

## Restrição Conhecida: Groq + Tools

A API da Groq **não aceita** `response_format` (JSON mode/schema) combinado com `tools`:

```
response_format json_object cannot be combined with tool/function calling
```

Confirmado em: LiteLLM [#15761](https://github.com/BerriAI/litellm/issues/15761), Agno [#2870](https://github.com/agno-agi/agno/issues/2870), OpenAI Agents Python [#2140](https://github.com/openai/openai-agents-python/issues/2140)

**Workarounds possíveis se tools forem reativadas:**
- **Two-step**: 1ª chamada com tools (sem response_format), 2ª com response_format (sem tools)
- **Function-as-output**: tool_choice=required com função cujo schema é o JSON de saída

## Código Existente

| Arquivo | Propósito | Status |
|---------|-----------|--------|
| `key_terms.go` | Extração determinística de termos | ✅ Ativo |
| `history_enrichment.go` | Montagem do bloco `[CONTEXTO-ATUAL]` | ✅ Ativo |
| `session.go:buildHistoryWithMeetings` | Montagem do prompt completo | ✅ Ativo |
| `tools.go` | `makeSearchMessagesTool`, `searchMessages`, `toolInstructionsFor*` | 📦 Dead code (reativar quando tools forem suportadas) |
