# Deploy & Versioning

## Target Options

| Target | When to Use |
|---|---|
| `your-server.com` | Default - Tailscale SSH + ko build |
| `other` | Custom CI/CD (AWS, GCP, Azure, etc.) |

## Server Deploy (your-server.com)

### Prerequisites
- Tailscale installed on server
- OpenSSH configured
- ko installed locally

### Setup
```bash
# On server: install ko
go install github.com/google/ko@latest

# On server: setup deploy user
sudo useradd -m -s /bin/bash deploy
sudo usermod -aG docker deploy
```

### GitHub Actions Workflow
```yaml
name: Deploy
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.26'
      - run: templ generate
      - uses: ko-build/setup-ko@v0.2
      - env:
          KO_DOCKER_REPO: ghcr.io/${{ github.repository }}
        run: |
          ko build \
            --platform=linux/amd64,linux/arm64 \
            --tags latest,${{ steps.version.outputs.version }} \
            --bare ./cmd/web/
```

### Server Update Script
```bash
#!/bin/bash
set -e

IMAGE="ghcr.io/your-org/your-app:latest"
CONTAINER="your-app"

docker pull $IMAGE
docker stop $CONTAINER || true
docker rm $CONTAINER || true
docker run -d \
  --name $CONTAINER \
  --restart unless-stopped \
  -p 8080:8080 \
  $IMAGE
```

## Other Deploy Targets

For AWS, GCP, Azure, or other cloud providers, adapt the CI/CD pipeline accordingly.

## Go Configuration Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Server port |
| `ENV` | `development` | Environment (development/production) |
| `NATS_URL` | `nats://localhost:4222` | NATS connection URL |
| `DB_PATH` | `./data/app.db` | SQLite database path |

## Version Display in UI

The version is set at build time via `-ldflags`:

```bash
go build -ldflags "-X main.Version=$(git describe --tags --always)" ./cmd/web/
```

In your templ component:

```go
templ VersionDisplay(version string) {
    <div class="text-xs text-base-content/50">
        v{ version }
    </div>
}
```

## Generated Structure

```
your-project/
├── cmd/web/
│   └── main.go           # Entry point
├── features/
│   ├── chat/             # Chat feature
│   │   ├── handler.go    # HTTP handlers
│   │   ├── service.go    # Business logic
│   │   └── chat.templ    # UI components
│   └── settings/         # Another feature
├── db/
│   └── repository.go     # Database layer
├── static/
│   ├── css/
│   └── js/
├── datastar/             # Datastar SDK (vendored)
├── go.mod
├── go.sum
└── Makefile
```
