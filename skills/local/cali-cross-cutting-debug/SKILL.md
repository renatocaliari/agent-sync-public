---
name: cali-cross-cutting-debug
description: >
  [Cali] Cross-cutting debugging rules: framework JS API verification and
  binary freshness checks. Activates when debugging reactive framework
  interactions (Datastar, Alpine, Vue, Svelte, Solid) or diagnosing code
  changes that don't take effect after restart.
metadata:
  frequency: weekly
  category: code
  context-cost: low
---

# Cross-Cutting Debugging Rules

## Framework JS API — Read First, Never Assume

Before writing JavaScript that interacts with a reactive framework (Datastar, Alpine, Vue, Svelte, Solid, etc.), read the framework's API reference FIRST.

ES module frameworks often expose **no global `window.Framework` API**. Do not assume `window.datastar`, `window.Alpine`, etc. exist without confirmation. Many modern frameworks are pure ES modules — the API is only available through imports, not global scope.

## Verify Binary Freshness After a Fix

When a code change doesn't take effect, check these timestamps:

1. **Binary** must be newer than **source** (or regenerated files like `*_templ.go`)
2. **Server process start time** must be newer than **binary**

If stale:
```bash
# Kill the process using the port
lsof -ti:<PORT> | xargs kill -9 2>/dev/null

# Rebuild and restart
templ generate && go build -o ./app ./cmd/web && ./app
```

If using Air, tell the user to save the file — Air auto-rebuilds.
