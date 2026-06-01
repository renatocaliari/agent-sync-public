#!/usr/bin/env bash
# detect-pw-artifacts.sh
# Detects cali-product-workflow artifacts in the current project.
# Prints found paths to stdout, one per line, prefixed by type.
# Exit code 0 if any found, 1 if none.
#
# Usage: cd /project && scripts/detect-pw-artifacts.sh

set -euo pipefail

ROOT="${1:-.}"
FOUND=0

# Define search paths
SEARCH_PATHS=(
  ".cali-product-workflow"
)

for dir in "${SEARCH_PATHS[@]}"; do
  # spec-product (business rules)
  while IFS= read -r -d '' f; do
    echo "spec-product|$f"
    FOUND=1
  done < <(find "$ROOT/$dir" -path '*/plans/specs/spec-product*.md' -type f -print0 2>/dev/null || true)

  # spec-tech (architecture)
  while IFS= read -r -d '' f; do
    echo "spec-tech|$f"
    FOUND=1
  done < <(find "$ROOT/$dir" -path '*/plans/spec-tech*.md' -type f -print0 2>/dev/null || true)

  # tech-planning (scopes, sequencing)
  while IFS= read -r -d '' f; do
    echo "tech-planning|$f"
    FOUND=1
  done < <(find "$ROOT/$dir" -path '*/plans/tech-planning*.md' -type f -print0 2>/dev/null || true)

  # critique reports
  while IFS= read -r -d '' f; do
    echo "critique|$f"
    FOUND=1
  done < <(find "$ROOT/$dir" -path '*/critiques/critique-report*.md' -type f -print0 2>/dev/null || true)

  # interfaces
  while IFS= read -r -d '' f; do
    echo "interfaces|$f"
    FOUND=1
  done < <(find "$ROOT/$dir" -path '*/interfaces/interfaces*.md' -type f -print0 2>/dev/null || true)

  # scopes
  while IFS= read -r -d '' f; do
    echo "scope|$f"
    FOUND=1
  done < <(find "$ROOT/$dir" -path '*/plans/scopes/*.md' -type f -print0 2>/dev/null || true)

  # verification / execution-critique
  while IFS= read -r -d '' f; do
    echo "verification|$f"
    FOUND=1
  done < <(find "$ROOT/$dir" -path '*/plans/verification*.md' -type f -print0 2>/dev/null || true)

  while IFS= read -r -d '' f; do
    echo "execution-critique|$f"
    FOUND=1
  done < <(find "$ROOT/$dir" -path '*/plans/execution-critique*.md' -type f -print0 2>/dev/null || true)
done

exit $((1 - FOUND))
