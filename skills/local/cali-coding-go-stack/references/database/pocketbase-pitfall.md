# PocketBase Startup Pitfall: EnsureCollections + backfillTimestampFields

## Symptoma

No startup do servidor, dezenas de `UPDATE` queries pesadas no log, cada uma contendo
blobs grandes (`snapshot_context_blocks` com textos de exercícios, `snapshot_prompts`
com templates de prompt). Exemplo:

```
UPDATE `sessions` SET `snapshot_context_blocks`='[{"content":"traduzido e adaptado...", ...}]', ... WHERE `id`='...'
```

## Causa

Duas fontes:

### 1. `EnsureCollections` chamado duas vezes

O código frequentemente acopla `EnsureCollections` em dois lugares:

```go
// Em InitPocketBase (via OnServe hook)
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
    if err := EnsureCollections(se.App); err != nil { ... }
    return se.Next()
})

// E também no main.go
appdb.EnsureCollections(pbApp)
```

Isso faz toda a lógica de migração/validação rodar em dobro.

### 2. `backfillTimestampFields` salva toda sessão todo startup

Função de migração one-shot que preenche `created_at`/`updated_at` em registros
que ainda não têm. O problema:

```go
session.Set("created_at", createdAt)
session.Set("updated_at", updatedAt)
app.Save(session)  // ← sempre salva, mesmo sem mudança
```

`Set()` marca o record como dirty → PocketBase persiste → o `Save()` grava **todos**
os campos, incluindo `snapshot_context_blocks` e `snapshot_prompts` (blobs grandes).

Depois que todos os registros já têm timestamp, essa função não deveria mais tocar
neles — mas sem o `continue` early, ela continua carregando e salvando.

## Correção

### 1. Remover duplicação

Manter `EnsureCollections` em **um** lugar só. Preferir o call explícito no `main.go`
e remover o `OnServe` hook:

```go
// InitPocketBase — sem OnServe hook
func InitPocketBase(...) (*pocketbase.PocketBase, error) {
    app := pocketbase.NewWithConfig(...)
    PocketBaseApp = app
    return app, nil
}

// main.go — call único
appdb.EnsureCollections(pbApp)
```

### 2. Skip early em `backfillTimestampFields`

```go
for _, session := range sessions {
    oldCreated := session.GetString("created_at")
    oldUpdated := session.GetString("updated_at")

    // Já preenchido → pula. Migração one-shot.
    if oldCreated != "" && oldUpdated != "" {
        continue
    }
    // ... só salva se algo mudou
}
```

## Prevenção em projetos futuros

- `EnsureCollections`/`Seed*` devem ser **idempotentes** e de preferência só
  escrever no DB quando realmente necessário.
- Migrações one-shot (`backfill*`, `migrateFieldRenames`) devem ter guard clause
  pra não rodar lógica pesada em startups subsequentes.
- Desconfie de `app.Save(record)` dentro de loops sem verificar se o registro
  precisa ser salvo.
