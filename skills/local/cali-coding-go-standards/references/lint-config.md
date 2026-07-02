# Static Analysis & Linting

## Installation

```bash
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

## Recommended `.golangci.yml`

```yaml
linters:
  enable:
    - errcheck        # unchecked errors
    - govet           # suspicious constructs (shadow, printf misuse, etc.)
    - staticcheck     # comprehensive static analysis (SA*, S*, QF*)
    - gosec           # security anti-patterns (hardcoded creds, weak crypto, etc.)
    - revive          # idiomatic style (golint successor)
    - gocritic        # opinionated but high-signal checks
    - funlen          # function length limit (enforced via lines/statements)
    - gocyclo         # cyclomatic complexity (enforced via min-complexity)
    - ineffassign     # assignments whose value is never used
    - unused          # unexported identifiers never referenced
    - goimports       # import ordering and formatting
    - gofumpt         # stricter formatting superset of gofmt
    - noctx           # http requests without context.Context
    - lll             # long line detection

linters-settings:
  gosec:
    severity: medium
  gocyclo:
    min-complexity: 15
  funlen:
    lines: 100
    statements: 60
  lll:
    line-length: 120
    tab-width: 1
  revive:
    rules:
      - name: exported
      - name: var-naming
      - name: error-return
      - name: context-as-argument
      - name: context-keys-type
```

## File Size Check

Enforce 500-line max per `.go` file via Makefile (no linter exists for file-level line count):

```makefile
lint: check-file-size
	golangci-lint run ./...

check-file-size:
	@echo "Checking file size limits..."
	@find . -name '*.go' ! -path './vendor/*' ! -path './tmp/*' | \
	  while read f; do \
	    lines=$$(wc -l < "$$f"); \
	    if [ $$lines -gt 500 ]; then \
	      echo "FAIL: $$f ($$lines lines, max 500)"; \
	      exit 1; \
	    fi; \
	  done
	@echo "All files under 500 lines ✓"
```

The agent must add/update this target whenever a `Makefile` exists.

## Ruleguard Custom Rules

**READ `references/gorules.go`** for Datastar + Go LLM patterns that standard linters miss. Covers: context.Background in HTTP handlers, nil context passed to ctx-accepting APIs (jobs.Enqueue, goqite/jobs.Create, http.NewRequestWithContext — silent panic when framework has panicRecover), loop variable capture, mutex copy by value, ListenAndServe error, SSE patterns, JSON marshaling shortcuts.

**Setup:** Copy `references/gorules.go` → `rules/gorules.go` in your project, then add to `.golangci.yml`:

```yaml
linters-settings:
  gocritic:
    enabled-tags:
      - ruleguard
    settings:
      ruleguard:
        rules: "rules/gorules.go"
```

## Run Linting

```bash
golangci-lint run ./...
```
