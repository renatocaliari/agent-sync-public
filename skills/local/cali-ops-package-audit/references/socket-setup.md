# Socket.dev Setup Reference

## Installation

```bash
npm install -g socket
# or
pip install socket
```

## First-time Setup

```bash
# Authenticate (opens browser)
socket login

# Enable wrapper mode (intercepts package installs)
socket wrapper on
```

## CI Integration

### GitHub Actions

```yaml
# .github/workflows/socket.yml
name: Socket Security Scan
on: [pull_request]
jobs:
  socket:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: SocketDev/socket-scans@main
        with:
          org-id: ${{ secrets.SOCKET_ORG_ID }}
```

### GitLab CI

```yaml
socket-scan:
  stage: test
  script:
    - npm install -g socket
    - socket ci
```

## What Socket Detects

| Category | Example |
|----------|---------|
| Malicious code | Obfuscated install scripts |
| Network access | Unexpected outbound calls |
| Typosquatting | `lod-ash` vs `lodash` |
| Hidden files | Files not in git but in package |
| Shell injection | `postinstall` scripts with `eval` |

## Excluding Packages

```bash
# Ignore specific package
socket ignore <package-name>

# Ignore in config
# .socketrc.json
{
  "ignore": ["some-package"]
}
```
