#!/bin/bash
# pre-commit-hook.sh — AGENTS.md staleness detection via sem
#
# Install:
#   cp references/pre-commit-hook.sh .git/hooks/pre-commit
#   chmod +x .git/hooks/pre-commit
#
# Or merge with existing pre-commit hook (see instructions below).
#
# Requires: sem (brew install sem-cli)

set -euo pipefail

# --- Configuration ---
# Files to scan for entity changes (add your primary language extensions)
SCAN_EXTENSIONS="${AGMD_SCAN_EXTENSIONS:-.go .ts .tsx .js .jsx .py .rs}"

# Skip if no staged files exist
if ! git diff --cached --quiet 2>/dev/null; then
  # Check if sem is available
  if ! command -v sem &>/dev/null; then
    echo "⚠️  sem not found. Install: brew install sem-cli"
    echo "   Skipping AGENTS.md staleness check."
    exit 0
  fi

  # Check if AGENTS.md exists
  if [ ! -f "AGENTS.md" ]; then
    echo "ℹ️  No AGENTS.md found. Consider creating one."
    echo "   Run: /skill:cali-agents-md"
    exit 0
  fi

  # Run sem diff on staged changes
  SEM_OUTPUT=$(sem diff --staged --format json 2>/dev/null || echo '{"summary":{"total":0}}')

  # Parse with Python (available on macOS by default)
  TOTAL=$(echo "$SEM_OUTPUT" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('summary', {}).get('total', 0))
except:
    print(0)
" 2>/dev/null)

  if [ "$TOTAL" -gt 0 ]; then
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "⚠️  AGENTS.md staleness check: $TOTAL entities changed"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "Changed entities:"
    echo "$SEM_OUTPUT" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    for c in d.get('changes', [])[:15]:
        ct = c.get('changeType', '?')
        name = c.get('entityName', '?')
        etype = c.get('entityType', '?')
        fpath = c.get('filePath', '?')
        print(f'  {ct:>10}  {name} ({etype}) in {fpath}')
    if len(d.get('changes', [])) > 15:
        remaining = len(d['changes']) - 15
        print(f'  ... and {remaining} more')
except:
    pass
" 2>/dev/null

    echo ""
    echo "💡 Review AGENTS.md for outdated content."
    echo "   If architecture/commands changed, update AGENTS.md before committing."
    echo ""
  fi
fi
