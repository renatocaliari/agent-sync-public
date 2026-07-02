#!/usr/bin/env bash
# cali-agents-md-validator: Validate AGENTS.md against 10 criteria
# Sources: agents.md spec, ETH Zurich, Vercel, aihackers.net, nyosegawa template
#
# Usage: bash validate-agents-md.sh [path-to-AGENTS.md]
# Default: ./AGENTS.md

set -euo pipefail

AGENTS_MD="${1:-./AGENTS.md}"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

total=0
passed=0
failed=0
warnings=0

check() {
  local rule="$1"
  local status="$2"  # pass, fail, warn
  local detail="$3"
  total=$((total + 1))
  
  if [ "$status" = "pass" ]; then
    echo -e "  ${GREEN}✅${NC} $rule: $detail"
    passed=$((passed + 1))
  elif [ "$status" = "fail" ]; then
    echo -e "  ${RED}❌${NC} $rule: $detail"
    failed=$((failed + 1))
  else
    echo -e "  ${YELLOW}⚠️${NC} $rule: $detail"
    warnings=$((warnings + 1))
  fi
}

echo "🔍 cali-agents-md-validator"
echo "Checking: $AGENTS_MD"
echo ""

if [ ! -f "$AGENTS_MD" ]; then
  check "R1: AGENTS.md exists" "fail" "File not found: $AGENTS_MD"
  echo ""
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "📊 Summary"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo -e "  ${RED}❌ No AGENTS.md found — use /skill:cali-agents-md-generator to create one${NC}"
  exit 1
fi

check "R1: AGENTS.md exists" "pass" "Found"

# R2: File size ≤150 lines
lines=$(wc -l < "$AGENTS_MD")
if [ "$lines" -gt 150 ]; then
  check "R2: File size (≤150 lines)" "fail" "$lines lines — too long, LLMs may ignore rules"
elif [ "$lines" -gt 120 ]; then
  check "R2: File size (≤150 lines)" "warn" "$lines lines — approaching limit"
else
  check "R2: File size (≤150 lines)" "pass" "$lines lines"
fi

# R3: Has "Commands" section
if grep -q "^## Commands\|^##commands\|^# Commands" "$AGENTS_MD"; then
  check "R3: Commands section" "pass" "Found"
else
  check "R3: Commands section" "fail" "No Commands section — add copy-pasteable commands"
fi

# R4: Has "Don'ts" section
if grep -q "^## Don'ts\|^##don'ts\|^# Don'ts" "$AGENTS_MD"; then
  check "R4: Don'ts section" "pass" "Found"
else
  check "R4: Don'ts section" "fail" "No Don'ts section — add explicit prohibitions"
fi

# R5: No secrets in file
if grep -qiE "(api_key|secret|password|token|credential).*=.*['\"][a-zA-Z0-9]{20,}" "$AGENTS_MD"; then
  check "R5: No secrets" "fail" "Possible secret detected — remove immediately"
else
  check "R5: No secrets" "pass" "No secrets found"
fi

# R6: Stack with exact versions
if grep -qE "(React|Vue|Angular|Go|Node|Python|Java|TypeScript) *[0-9]+\.[0-9]+" "$AGENTS_MD"; then
  check "R6: Exact versions" "pass" "Stack has version numbers"
else
  check "R6: Exact versions" "warn" "No exact versions found — add specific versions"
fi

# R7: No architectural bloat (>20 lines of description in any section)
bloat=$(awk '/^## [^#]/{section=$0; lines=0} /^## [^#]/{if(lines>20) print section": "lines" lines"} !/^## [^#]/{lines++}' "$AGENTS_MD" | head -1)
if [ -n "$bloat" ]; then
  check "R7: No arch bloat" "warn" "$bloat — consider moving details to README"
else
  check "R7: No arch bloat" "pass" "No oversized sections"
fi

# R8: Critical rules in first 50 lines
if head -50 "$AGENTS_MD" | grep -q "^## Don'ts\|^## Commands\|^## Rule\|^## RULE"; then
  check "R8: Critical rules early" "pass" "Critical rules in first 50 lines"
else
  check "R8: Critical rules early" "warn" "Critical rules may be buried — move to top"
fi

# R9: References use /skill:name or links
if grep -q "/skill:\|@file:\|references/" "$AGENTS_MD"; then
  check "R9: Skill references" "pass" "Uses /skill:name or references"
else
  check "R9: Skill references" "warn" "No skill references — use /skill:name for extended docs"
fi

# R10: Template compliance (Commands + Don'ts + Architecture/Stack)
has_commands=$(grep -q "^## Commands" "$AGENTS_MD" && echo 1 || echo 0)
has_donts=$(grep -q "^## Don'ts" "$AGENTS_MD" && echo 1 || echo 0)
has_arch=$(grep -q "^## Architecture\|^## Stack" "$AGENTS_MD" && echo 1 || echo 0)

if [ "$has_commands" -eq 1 ] && [ "$has_donts" -eq 1 ] && [ "$has_arch" -eq 1 ]; then
  check "R10: Template compliance" "pass" "Commands + Don'ts + Architecture"
else
  missing=""
  [ "$has_commands" -eq 0 ] && missing="$missing Commands"
  [ "$has_donts" -eq 0 ] && missing="$missing Don'ts"
  [ "$has_arch" -eq 0 ] && missing="$missing Architecture/Stack"
  check "R10: Template compliance" "warn" "Missing:$missing"
fi

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "  Total checks: $total"
echo -e "  ${GREEN}Passed: $passed${NC}"
echo -e "  ${RED}Failed: $failed${NC}"
echo -e "  ${YELLOW}Warnings: $warnings${NC}"
echo ""

if [ "$failed" -gt 0 ]; then
  echo -e "${RED}❌ $failed checks failed — fix before publishing${NC}"
  exit 1
elif [ "$warnings" -gt 0 ]; then
  echo -e "${YELLOW}⚠️ $warnings warnings — consider fixing${NC}"
  exit 0
else
  echo -e "${GREEN}✅ All checks passed!${NC}"
  exit 0
fi
