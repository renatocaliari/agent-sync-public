# Air — Local Development

## Installation

```bash
go install github.com/air-verse/air@latest
```

## Project Policy

- This skill is the single source of truth for Air policy.
- `cali-coding-go-stack` must not duplicate Air policy; it may only reference this skill.
- Project `AGENTS.md` must list `air` as default local dev command.
- `Makefile dev` must start Air.
- Commit `.air.toml` when it is the project's standard dev config.
- Ignore only `.air.toml.local` for user-specific overrides.
- Keep `.air.toml.example` only when `.air.toml` is intentionally local and not committed.
- Do not duplicate Air processes in `solo.yml` (`air` and `make dev` are the same workflow).

## `.air.toml` — Standard Config (Projects with Templ)

```toml
[build]
  pre_cmd        = ["templ generate"]
  cmd            = "go build -ldflags='-w -X main.Version=dev' -o ./tmp/app ./cmd/web"
  bin            = "./tmp/app"
  include_ext    = ["go", "templ", "yaml", "css", "js"]
  exclude_dir    = ["tmp", "vendor", "node_modules", "testdata"]
  exclude_regex  = ["_templ\\.go$"]
  kill_delay     = "500ms"
  send_interrupt = true
  stop_on_error  = true

[log]
  time = true

[misc]
  clean_on_exit = true
```

For projects **without Templ**, remove `pre_cmd`. For projects using Go tool dependencies, `go tool templ generate` is also acceptable.

### Lint on Every Save (Recommended)

Add `post_cmd` to catch issues immediately during development. This runs after every successful build but does NOT block the restart (failures are warnings, not errors):

```toml
[build]
  # ... pre_cmd, cmd, bin, etc. ...
  post_cmd = [
    "golangci-lint run ./... --timeout 3m --fast || true",
    "find . -name '*.go' ! -path './vendor/*' ! -path './tmp/*' -exec sh -c 'test $(wc -l < \"$1\") -le 500' _ {} \\; || echo 'WARNING: file exceeds 500 lines'"
  ]
```

The pre-commit hook (see `cali-coding-go-standards` SKILL.md) is BLOCKING — `post_cmd` here is non-blocking feedback during dev.

## `.gitignore`

```
tmp/
.air.toml.local
```

Do **not** ignore `.air.toml` when it is the committed project config.

## Running Air

```bash
air   # run from project root; keep terminal open while developing
```

Stop with `Ctrl+C` when done.

## Frontend Hot Reload

- `.templ` changes handled by `templ generate` in Air `pre_cmd`.
- Go changes rebuild and restart.
- DaisyUI/Tailwind via CDN: no local watcher needed.
- Local CSS/JS/static assets: need explicit watcher or include rule.
- Do not switch to Templier unless Air cannot cover required hot reload.

## Agent Rules

- Never instruct user to run `go build` + `./app` manually during dev session.
- Never ask user to restart server manually when Air is active.
- Check if Air is running: `pgrep -x air`
  - If running: tell user to save the file — Air rebuilds automatically.
  - If not running: tell user to start it with `air` or `make dev`.

## Binary on Disk vs Process in Memory

Air rebuilds `tmp/app` on disk, but the running process holds the OLD code in memory (Go has no hot-reload). If a change "doesn't show up", check whether the running PID has the latest binary:

```bash
stat -f "%Sm" tmp/app                       # binary mtime
ps -o lstart= -p $(pgrep -f tmp/app | head -1)  # when process started
```

If binary mtime is later than process start → process is stale. `pkill -9 -f tmp/app && ./tmp/app > /tmp/app.log 2>&1 &` to restart fresh.

**Tip**: when debugging Go server behavior, run with logs to file (`> /tmp/app.log 2>&1 &`) instead of letting Air pipe to a tty — makes stack traces inspectable.

## Staging and Production

Never use Air outside of local development. For deploying:

1. If the project has a deploy mechanism (script, Makefile, Dockerfile, process manager), use it.
2. If no deploy mechanism exists:

```bash
templ generate
lsof -ti:<PORT> | xargs kill -9 2>/dev/null || true
sleep 1
go build -ldflags="-w -X main.Version=$(git describe --tags --abbrev=0 2>/dev/null || echo dev) -X main.CommitHash=$(git rev-parse --short HEAD 2>/dev/null || echo unknown) -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o ./app ./cmd/web
nohup ./app > server.log 2>&1 &
echo "PID: $!"
```

Drop `-X` flags for version/commit/build time if the project does not expose them.

| Tool | Context | Builds with ldflags | Watches files | Long-running |
|------|---------|-------------------|---------------|-------------|
| `air` | Local dev | ❌ (uses `dev`) | ✅ | No |
| deploy steps | Staging/prod | ✅ | ❌ | Yes (nohup) |
