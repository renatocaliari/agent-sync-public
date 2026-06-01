# ko Build Pipeline Reference

## What is ko?

ko compiles Go binaries directly into minimal OCI images — no Dockerfile needed.
~8 MB images, native cross-compilation, distroless base.

## Local Build

```bash
# Build OCI tarball (no Docker daemon needed)
ko build --platform=linux/amd64 --tarball=/tmp/app.tar --tags dev --bare ./cmd/web/

# Build and push
KO_DOCKER_REPO=ghcr.io/org/repo ko build --platform=linux/amd64,linux/arm64 --bare ./cmd/web/
```

## Configuration (.ko.yaml)

```yaml
defaultBaseImage: gcr.io/distroless/static:nonroot

builds:
  - id: web
    main: ./cmd/web
    env:
      - CGO_ENABLED=0
    ldflags:
      - -s -w
      - -X main.version={{.Env.VERSION}}
      - -X main.commit={{.Env.COMMIT}}
```

## GitHub Actions

```yaml
- uses: actions/setup-go@v5
  with: { go-version: '1.26' }
- run: templ generate          # if using Templ
- uses: ko-build/setup-ko@v0.2
- env: { KO_DOCKER_REPO: ghcr.io/org/repo }
  run: |
    ko build \
      --platform=linux/amd64,linux/arm64 \
      --tags latest,${{ steps.version.outputs.version }} \
      --bare ./cmd/web/
```

## Requirements

- `//go:embed` for all static files (JS, CSS, templates)
- Binary must be self-contained
- No runtime file copies from source tree

## When NOT to use ko

- Project uses CGo
- Needs OS packages (ca-certificates, tzdata)
- In that case: multi-stage Dockerfile with `--mount=type=cache`
