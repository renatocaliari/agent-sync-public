# Deploy & Versioning

Loaded on demand when user confirms a production deploy target.

## Target Options

| Target | Action |
|--------|--------|
| `your-server.com` | Generate full CI/CD pipeline with ghcr.io + cron |
| `other` | Generate pipeline with placeholders |
| `none` | Skip entirely |

## CI/CD: `.github/workflows/deploy.yml`

```yaml
name: Build and Publish Docker Image
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  workflow_dispatch:
permissions:
  packages: write
  contents: read
jobs:
  build-and-push:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: docker/setup-buildx-action@v3
      - run: |
          echo "${{ secrets.GITHUB_TOKEN }}" | docker login ghcr.io -u ${{ github.actor }} --password-stdin
      - id: version
        run: |
          VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
          echo "version=$VERSION" >> $GITHUB_OUTPUT
      - uses: docker/metadata-action@v5
        with:
          images: ghcr.io/{{GITHUB_REPO}}
          tags: |
            type=ref,event=branch
            type=sha
            type=raw,value=latest
      - uses: docker/build-push-action@v6
        with:
          context: .
          push: true
          platforms: linux/amd64,linux/arm64
          tags: ${{ steps.meta.outputs.tags }}
          build-args: |
            CGO_ENABLED=0
            VERSION=${{ steps.version.outputs.version }}
```

## Release: `.github/workflows/release.yml`

```yaml
name: Release
on:
  push:
    branches: [main]
concurrency:
  group: release-please
  cancel-in-progress: false
permissions:
  contents: write
  pull-requests: write
jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: googleapis/release-please-action@v4
        id: release
        with:
          release-type: go
          config-file: release-please-config.json
          manifest-file: .release-please-manifest.json
          target-branch: ${{ github.ref_name }}
          token: ${{ secrets.GITHUB_TOKEN }}
      - name: Auto-merge Release PR
        if: ${{ steps.release.outputs.release_created != 'true' }}
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          PR_NUMBER="${{ fromJSON(steps.release.outputs.pr).number }}"
          if [ -n "$PR_NUMBER" ]; then
            gh pr merge --merge "$PR_NUMBER"
          fi
```

Also create `release-please-config.json`:
```json
{
  "packages": {
    ".": {
      "release-type": "go"
    }
  }
}
```

And `.release-please-manifest.json`:
```json
{
  ".": "0.0.0"
}
```

**Important**: release-please requires **conventional commits** format. Use:
- `fix: desc` for patch bumps
- `feat: desc` for minor bumps
- `chore:`, `docs:`, etc. for no bump
- Avoid emoji prefixes like `:bug: fix:` — release-please ignores these

**⚠️ Critical `prs_created` trap**: Always use `steps.release.outputs.pr` (actual PR number), not `prs_created`.

## Server Update Script: `update.sh`

Place at `/opt/{{APP_NAME}}/update.sh`:

```bash
#!/bin/bash
set -e
IMAGE="ghcr.io/{{GITHUB_REPO}}"
CONTAINER_NAME="{{APP_NAME}}"
TOKEN_FILE="/opt/{{APP_NAME}}/.gh_token"

log() { echo "$(date): $1" | tee -a /opt/{{APP_NAME}}/update.log; }

log "Checking for updates..."
echo "$(cat $TOKEN_FILE)" | docker login ghcr.io -u ${{ GITHUB_USERNAME }} --password-stdin > /dev/null 2>&1

docker pull "$IMAGE:latest" > /dev/null 2>&1

REMOTE_ID=$(docker inspect "$IMAGE:latest" --format="{{.Id}}" 2>/dev/null || echo "")
LOCAL_ID=$(docker inspect "$CONTAINER_NAME" --format="{{.Image}}" 2>/dev/null || echo "")

if [ -z "$LOCAL_ID" ]; then
    log "Container not running, will start..."
fi

if [ "$REMOTE_ID" != "$LOCAL_ID" ]; then
    log "New version detected ($REMOTE_ID), deploying..."
    docker stop "$CONTAINER_NAME" 2>/dev/null || true
    docker rm "$CONTAINER_NAME" 2>/dev/null || true
    SESSION_SECRET=$(openssl rand -base64 32 2>/dev/null || head -c32 /dev/urandom | base64)
    docker run -d \
        --name "$CONTAINER_NAME" \
        --restart unless-stopped \
        -p 127.0.0.1:8080:8080 \
        -e SESSION_SECRET="$SESSION_SECRET" \
        -v /opt/{{APP_NAME}}/data:/app/data \
        "$IMAGE:latest"
    log "SUCCESS: Deployed new version (image: $REMOTE_ID)"
else
    log "No changes detected, skipping restart"
fi
```

## Dockerfile (multi-stage)

- Builder: `golang:{{GOVSN}}-alpine` + `templ generate` + `go build -ldflags="-X main.Version=${VERSION}"`
- Runtime: `alpine:3.21` (no root, healthcheck via TCP port check)
- Expose: `8080`

## Go Configuration & Version Display

Use environment variables directly (no `.env` by default). For production, pass via `docker run -e VAR=value`.

Inject version via ldflags:
```bash
go build -ldflags="-X main.Version=${VERSION}" -o app ./cmd/web/
```

In Go code:
```go
var Version = "dev"  // set via ldflags

type Config struct {
    Version string
}

// Pass to templates via page data structs
type HomePageData struct {
    Version string
}
```
