#!/bin/bash
# setup-hooks.sh — Auto-configure Go project checks
# Called by cali-coding-go-standards during project scaffold or first edit.
# Detects pi.dev, creates git hooks, configures Air, generates CI workflow.
set -euo pipefail

PROJECT_DIR="${1:-$(pwd)}"
cd "$PROJECT_DIR"

echo "━━━ Setting up Go project hooks ━━━"
echo "Project: $PROJECT_DIR"

PI_HOOKS_DIR="$HOME/.pi/agent/hook"
PI_HOOKS_FILE="$PI_HOOKS_DIR/hooks.yaml"
HOOK_MARKER="# cali-go-hooks"

# ═══════════════════════════════════════════
# 1. pi.dev hooks (global, user machine)
# ═══════════════════════════════════════════
if [ -d "$PI_HOOKS_DIR" ]; then
  echo "→ pi.dev detected — updating global hooks..."
  mkdir -p "$PI_HOOKS_DIR"

  if grep -q "$HOOK_MARKER" "$PI_HOOKS_FILE" 2>/dev/null; then
    echo "  ✓ pi.dev hooks already configured"
  else
    echo "  → Installing pi.dev hooks..."
    # Remove any previous block if marker exists (clean slate)
    if grep -q "cali-go-hooks" "$PI_HOOKS_FILE" 2>/dev/null; then
      sed -i '' '/# cali-go-hooks/,/cali-go-hooks/d' "$PI_HOOKS_FILE"
    fi
    # Append hooks indented under hooks: root key
    cat >> "$PI_HOOKS_FILE" << 'HOOKS'

# cali-go-hooks
  # ── Post-build: inform, never block ──
  - event: tool.after.bash
    actions:
      - bash: |
          grep -qE '"command":"go build' /dev/stdin 2>/dev/null || exit 0
          cd "$PI_PROJECT_DIR"
          [ ! -f go.mod ] && exit 0
          echo "━━━ Post-build Go Checks (info) ━━━"
          any_issue=0
          echo "→ File size (max 500 lines)..."
          while read f; do lines=$(wc -l < "$f")
            [ "$lines" -gt 500 ] && echo "  ⚠ $f ($lines lines, max 500)" && any_issue=1
          done < <(find . -name '*.go' ! -path './vendor/*' ! -path './tmp/*')
          [ "$any_issue" -eq 0 ] && echo "  ✓"
          echo "→ golangci-lint..."
          which golangci-lint >/dev/null 2>&1 && golangci-lint run ./... >/tmp/pi-lint-out.txt 2>&1 && echo "  ✓" || { echo "  ⚠ Issues found (see /tmp/pi-lint-out.txt)"; any_issue=1; }
          echo "→ deadcode..."
          which deadcode >/dev/null 2>&1 && deadcode -test ./... >/tmp/pi-deadcode-out.txt 2>&1 && echo "  ✓" || { echo "  ⚠ Found (see /tmp/pi-deadcode-out.txt)"; any_issue=1; }
          [ "$any_issue" -eq 1 ] && echo "━━━ Issues found (non-blocking) ━━━" || echo "━━━ All checks clean ✓ ━━━"

  # ── Pre-commit: blocking ──
  - event: tool.before.bash
    actions:
      - bash: |
          grep -qE '"command":"git commit' /dev/stdin 2>/dev/null || exit 0
          cd "$PI_PROJECT_DIR"
          [ ! -f go.mod ] && exit 0
          echo "━━━ Pre-commit Go Checks ━━━"
          echo "→ File size (max 500 lines)..."
          while read f; do lines=$(wc -l < "$f")
            [ "$lines" -gt 500 ] && echo "  FAIL: $f ($lines lines, max 500)" && exit 2
          done < <(find . -name '*.go' ! -path './vendor/*' ! -path './tmp/*')
          echo "  ✓"
          echo "→ golangci-lint..."
          which golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || { echo "  installing..."; curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)/bin"; golangci-lint run ./... || { echo "❌ Lint failed"; exit 2; }; }
          echo "  ✓"
          echo "→ go mod tidy..."
          go mod tidy && git diff --exit-code go.mod go.sum || { echo "❌ go.mod/go.sum not tidy. Run 'go mod tidy' and commit."; exit 2; }
          echo "  ✓"
          echo "━━━ All checks passed ✓ ━━━"
      - notify: "Pre-commit Go checks passed"

  # ── Pre-push: vulnerability scan, blocking ──
  - event: tool.before.bash
    actions:
      - bash: |
          grep -qE '"command":"git push' /dev/stdin 2>/dev/null || exit 0
          cd "$PI_PROJECT_DIR"
          [ ! -f go.mod ] && exit 0
          echo "→ govulncheck (pre-push)..."
          which govulncheck >/dev/null 2>&1 && govulncheck ./... || { echo "❌ Vulnerability scan failed"; exit 2; }
          echo "  ✓"
          echo "━━━ Pre-push checks passed ✓ ━━━"
# cali-go-hooks
HOOKS
    echo "  ✓ pi.dev hooks installed"
    # Validate YAML structure (catch missing hooks: root key)
    if command -v python3 &>/dev/null; then
      if python3 -c "import yaml; yaml.safe_load(open('$PI_HOOKS_FILE'))" 2>/dev/null; then
        echo "  ✓ pi.dev hooks YAML valid"
      else
        echo "  ⚠ YAML validation failed — check $PI_HOOKS_FILE manually"
      fi
    fi
  fi
else
  echo "  → pi.dev not detected — using git hooks instead"
fi

# ═══════════════════════════════════════════
# 2. Git hooks (per-project, universal)
# ═══════════════════════════════════════════
echo "→ Setting up git hooks..."
mkdir -p .githooks

cat > .githooks/pre-commit << 'GIT_HOOK'
#!/bin/bash
set -euo pipefail

echo "=== Pre-commit check ==="

# File size check (max 500 lines per .go file)
find . -name '*.go' ! -path './vendor/*' ! -path './tmp/*' | \
  while read f; do
    lines=$(wc -l < "$f")
    if [ "$lines" -gt 500 ]; then
      echo "FAIL: $f ($lines lines, max 500)"
      exit 1
    fi
  done

# golangci-lint
if which golangci-lint >/dev/null 2>&1; then
  golangci-lint run ./... --timeout 5m || exit 1
else
  echo "golangci-lint not installed — skipping"
fi

# go mod tidy check
go mod tidy
git diff --exit-code go.mod go.sum || {
  echo "FAIL: go.mod or go.sum not tidy. Run 'go mod tidy' and commit."
  exit 1
}

echo "=== Pre-commit OK ==="
GIT_HOOK

chmod +x .githooks/pre-commit
git config core.hooksPath .githooks 2>/dev/null || true
echo "  ✓ .githooks/pre-commit created and installed"

# ═══════════════════════════════════════════
# 3. Air post_cmd (per-project)
# ═══════════════════════════════════════════
if [ -f .air.toml ]; then
  echo "→ Updating .air.toml with post_cmd..."
  # Inject post_cmd into [build] section if not present
  if grep -q "post_cmd" .air.toml 2>/dev/null; then
    echo "  ✓ post_cmd already configured"
  else
    # Add post_cmd after the bin line (heuristic: insert before [log])
    if grep -q "^\[log\]" .air.toml 2>/dev/null; then
      sed -i '' '/^\[log\]/i\
  post_cmd = ["golangci-lint run ./... --timeout 3m --fast || true"]' .air.toml
      echo "  ✓ post_cmd added to .air.toml"
    else
      echo "  ⚠ could not find [log] section in .air.toml — add manually:"
      echo '  post_cmd = ["golangci-lint run ./... --timeout 3m --fast || true"]'
    fi
  fi
fi

# ═══════════════════════════════════════════
# 4. CI workflow (per-project)
# ═══════════════════════════════════════════
echo "→ Setting up CI workflow..."
mkdir -p .github/workflows

cat > .github/workflows/check.yml << 'CI_YML'
name: Go Checks
on: [push, pull_request]
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: stable
      - run: go mod tidy && git diff --exit-code go.mod go.sum
      - run: find . -name '*.go' ! -path './vendor/*' -exec sh -c 'test $(wc -l < "$1") -le 500' _ {} \;
      - run: go vet ./...
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          args: --timeout 5m
      - run: go test -race ./...
CI_YML
echo "  ✓ .github/workflows/check.yml created"

# ═══════════════════════════════════════════
# 5. Makefile lint target
# ═══════════════════════════════════════════
if [ -f Makefile ]; then
  echo "→ Checking Makefile for lint target..."
  if grep -q "^lint:" Makefile 2>/dev/null; then
    echo "  ✓ lint target exists"
  else
    cat >> Makefile << 'MAKEFILE'

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
MAKEFILE
    echo "  ✓ lint target added to Makefile"
  fi
fi

echo ""
echo "━━━ Hook setup complete ━━━"
echo "  • pi.dev hooks:     $( [ -d "$PI_HOOKS_DIR" ] && echo '✅' || echo '❌ (not detected)' )"
echo "  • Git hooks:         ✅ (.githooks/pre-commit)"
echo "  • Air post_cmd:     $( [ -f .air.toml ] && grep -q 'post_cmd' .air.toml 2>/dev/null && echo '✅' || echo '❌' )"
echo "  • CI workflow:      ✅ (.github/workflows/check.yml)"
echo "  • Makefile lint:    $( [ -f Makefile ] && grep -q '^lint:' Makefile 2>/dev/null && echo '✅' || echo '❌' )"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
