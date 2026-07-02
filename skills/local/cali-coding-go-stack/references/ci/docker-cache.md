# Docker Build CI: Fast Go Builds (GitHub Actions)

**Pattern validado jun 2026** (Treinador de Práticas Narrativas — Go 1.26, 79 deps).

## Quick Win

| Build type | Cold | Warm (com cache) |
|---|---|---|
| **Go app puro** (1-90 deps) | ~3-4min | ~1-2min |
| **Rust workspace** (8 crates) | ~8-10min | ~3-5min |
| **Rust workspace SEM cache mounts** | 45-90min | 30-60min |

Cache mounts em Dockerfile são **in-job only** em GHA runners (fresh VM cada job). Precisam de cache-from/cache-to GHA backend pra cross-run persistence.

## Dockerfile Pattern (Go)

```dockerfile
# syntax=docker/dockerfile:1.7
FROM golang:1.26-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /src

# Layer 1: module graph (invalidado só em go.mod/go.sum change)
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    go mod download

# Layer 2: source (invalidado em .go/.templ change)
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    go tool templ generate && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /usr/local/bin/app ./cmd/web

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /usr/local/bin/app /app/app
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]
```

## GitHub Actions Workflow

```yaml
name: Build and Publish Docker Image
on:
  push: { branches: [master] }
  pull_request: {}

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.26'
          cache: true
          cache-dependency-path: go.sum
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/setup-buildx-action@v3
      - uses: docker/build-push-action@v6
        with:
          context: .
          push: ${{ github.event_name == 'push' }}
          platforms: linux/amd64   # drop arm64 unless using Graviton runner
          tags: |
            ghcr.io/${{ github.repository }}:latest
            ghcr.io/${{ github.repository }}:${{ github.sha }}
          # GHA cache: cross-run persistence, mode=max captura todas as stages
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

## Anti-patterns (validated, all slow)

| Anti-pattern | Symptom | Fix |
|---|---|---|
| `BUILDKIT_INLINE_CACHE=1` | useless pra multi-stage | Use `cache-from/cache-to: type=gha,mode=max` |
| `platforms: linux/amd64,linux/arm64` sem Graviton runner | 5-10x slowdown (qemu emulation) | Drop arm64 ou use `runs-on: ubuntu-24.04-arm` |
| Cache mount alone (sem GHA cache) | Cold every run | Combine mount + cache-from/cache-to |
| `docker pull` dentro de `docker buildx build` step | Bloqueia step inteiro (timeout) | NUNCA misturar. Use `docker/login-action` + multi-stage se precisar de deps externas |
| `rust:slim` pra builds com native deps (ort, ring, lopdf) | `cannot find -lstdc++` no link | `rust:1.88-trixie` (full) + `apt install g++` |
| `rust:bookworm` (glibc 2.36) com ort prebuilt 2.0+ | `undefined reference to __isoc23_strtol` | `rust:1.88-trixie` (glibc 2.38) |
| `WORKDIR /build` + `git clone .` | "destination path . already exists" | Clone em `/src/<name>` + `cp -r` |
| `docker buildx build` direto (sem build-push-action) | Sem provenance, sem cache GHA | Use `docker/build-push-action@v6` |

## Rust Sidecar (CloakPipe pattern)

Quando projeto Go tem sidecar Rust pesado (CI lento):

```dockerfile
# syntax=docker/dockerfile:1.7
FROM rust:1.88-trixie AS builder
RUN apt-get update && apt-get install -y --no-install-recommends \
    pkg-config libssl-dev ca-certificates git g++ \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /build
RUN --mount=type=cache,target=/build/target,sharing=locked \
    --mount=type=cache,target=/usr/local/cargo/registry,sharing=locked \
    git clone --depth 1 --branch ${VERSION} https://github.com/X/Y.git /src/y \
    && cp -r /src/y/. /build/ \
    && cargo build --release -p y-cli \
    && cp target/release/y /usr/local/bin/y \
    && rm -rf /src/y
```

Workflow strategy (KISS — não buildar sidecar em todo push):
- **App workflow**: builda Go app, ~3-4min cold / ~1-2min warm
- **Sidecar workflow**: builda Rust sidecar, **APENAS em tag push + manual dispatch**
- **Dependabot**: tracking semanal upstream do sidecar (`docker-compose` ecosystem)

```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "docker-compose"
    directory: "/"
    schedule: { interval: "weekly", day: "monday" }
    labels: ["dependencies"]
  - package-ecosystem: "gomod"
    directory: "/"
    schedule: { interval: "weekly" }
    labels: ["dependencies"]
```

## Tracking pre-built sidecar updates

Quando builda imagem sidecar 1x e usa `:latest` forever:
1. Dependabot (recomendado) — GHCR nativo, zero infra, PR automático semanal
2. Scheduled Action que `gh api releases/latest` e abre issue
3. Renovate `customManagers` regex (overkill pra este caso)

**Trade-off Dependabot**: não detecta imagem que NÃO tem semver tag (ex: `:latest` only). Se upstream usa SHA digest, `versioning-strategy: digest` em dependabot.yml.

## Debugging CI build lento

Checklist em ordem:
1. `gh run view <id> --log-failed` → qual stage demora?
2. Cold cache em todo run? Falta `cache-from/cache-to: type=gha`
3. qemu emulation rodando? Falta `platforms: linux/amd64` only ou Graviton runner
4. Native deps falhando no link? `rust:slim` → `rust:full` + `g++`
5. `docker pull` dentro de build step? Mover pra stage separada
6. Multi-stage build ineficiente? Separar COPY go.mod do COPY . . em layers

## Referências

- Docker BuildKit cache mounts: https://docs.docker.com/build/cache/backends/
- GitHub Actions cache backend: https://docs.docker.com/build/cache/backends/gha/
- Dependabot docker-compose GA (Fev 2025): https://github.blog/changelog/2025-02-25-dependabot-version-updates-now-support-docker-compose-in-general-availability/
- Multi-platform builds + emulation cost: https://github.com/docker/buildx/discussions/1382
