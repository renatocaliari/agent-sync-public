# Scaffold Template

<!-- Replace module name and project-specific notes after copying this scaffold. -->

## Project Overview

Go web app scaffold using Templ, Datastar, DaisyUI, TailwindCSS, and NATS.

Module: `northstar` (replace with your project module).

## Stack

Go 1.26 | Templ | Datastar | DaisyUI + TailwindCSS | NATS

## Commands

| Command | Description |
|---------|-------------|
| `make dev` / `air` | Dev server with live reload |
| `go run -tags=dev ./cmd/web` | Run once without watcher |
| `go test ./...` | Run tests |
| `make build` | Build dev binary |

## Skills

| Skill | When |
|-------|------|
| `/skill:cali-coding-standards` | Universal coding principles |
| `/skill:cali-coding-go-standards` | Go backend, modules, linting, local dev workflow |
| `/skill:cali-coding-go-stack` | Templ, Datastar, DaisyUI, TailwindCSS, NATS patterns |

## Don'ts

- Do NOT use HTMX/Alpine.js — use Datastar
- Do NOT use `fmt.Sprintf` for HTML — use Templ
- Do NOT use global mutable state
- Do NOT add dependencies without asking

## Architecture

```text
cmd/web/          → Entry point
config/           → Config loading
router/           → Route registration
nats/             → Embedded NATS setup
features/         → Self-contained features
web/resources/    → Static assets
```

## Hot Reload

Air is the default local dev loop.

- `.templ` changes run `go tool templ generate`
- Go changes rebuild and restart the server
- CSS/JS static asset changes rebuild so embedded assets stay fresh
- DaisyUI/Tailwind via CDN need no local CSS watcher

Local overrides may use `.air.toml.local`.
