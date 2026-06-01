# Scaffold Template

This directory contains a complete project template ready to be copied for new projects.

## Structure

```
scaffold/
├── cmd/web/main.go          # Application entrypoint
├── config/
│   ├── config.go            # Config types and loader
│   ├── config_dev.go        # Dev-specific config
│   └── config_prod.go       # Prod-specific config
├── router/router.go         # Route registration
├── nats/nats.go             # Embedded NATS setup
├── features/
│   ├── common/
│   │   ├── layouts/base.templ
│   │   └── components/
│   └── whiteboard/
│       ├── routes.go
│       ├── handlers.go
│       ├── pages/whiteboard.templ
│       └── services/
│           ├── fabric_service.go
│           └── cursor_service.go
├── web/resources/           # Static assets (empty, generated)
├── Taskfile.yml             # Build tasks
├── Dockerfile
├── go.mod
└── .gitignore
```

## Usage

1. Copy entire `scaffold/` directory to new project location
2. Update module name in `go.mod`
3. Update import paths in all Go files
4. Run `go mod tidy`
5. Run `go tool task live` to start development

## What's Included

- **Embedded NATS**: In-process messaging and state persistence
- **Hot Reload**: Air + templ watch + tailwind watch
- **Chi Router**: HTTP routing with middleware
- **Session Management**: Gorilla sessions
- **Templ Templates**: Type-safe HTML
- **TailwindCSS + BasecoatUI**: Styling
- **Datastar Integration**: Real-time hypermedia
- **Fabric.js Whiteboard**: Real-time collaborative whiteboard with:
  - Drawing tools (pencil, shapes, text)
  - Image upload and clipboard paste
  - Multi-user cursors with funny names
  - Live sync while moving/resizing objects
  - NATS JetStream KV for persistence
  - Delta sync optimization (50-150ms debounce)

## Whiteboard Features

| Feature | Description |
|---------|-------------|
| **Select** | Move, resize, rotate objects |
| **Pencil** | Free-hand drawing |
| **Line** | Draw lines |
| **Rectangle** | Draw rectangles |
| **Circle** | Draw circles |
| **Text** | Add text labels |
| **Upload** | Upload images (button) |
| **Paste** | Ctrl+V images from clipboard |
| **Delete** | Remove selected objects |
| **Clear** | Clear entire canvas |

## Development Commands

```bash
go tool task live      # Start with hot reload
go tool task run       # Build and run
go tool task build     # Production build
go tool task debug     # Run with debugger
```
