#!/usr/bin/env bash
# cali-skill-validator: Validate skills against 10 criteria from industry best practices
# Sources: AgentSkills Spec, Microsoft Learn, Addy Osmani, Perplexity Research,
#          OpenAI Skills, Claude API Docs, Trail of Bits, Lalit Madan, agentskills.io
#
# Usage: bash validate.sh [skills-directory]
# Default: ~/.agents/skills

set -euo pipefail

SKILLS_DIR="${1:-$HOME/.agents/skills}"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

total=0
passed=0
failed=0
warnings=0

check() {
  local skill="$1"
  local rule="$2"
  local status="$3"  # pass, fail, warn
  local detail="$4"
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

validate_skill() {
  local skill_dir="$1"
  local skill_name=$(basename "$skill_dir")
  local skill_md="$skill_dir/SKILL.md"
  
  echo ""
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "📦 Validating: $skill_name"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  
  # Rule 1: SKILL.md exists
  if [ ! -f "$skill_md" ]; then
    check "$skill_name" "R1: SKILL.md exists" "fail" "Missing $skill_md"
    return
  fi
  check "$skill_name" "R1: SKILL.md exists" "pass" "Found"
  
  # Rule 2: Frontmatter has name + description
  local has_frontmatter=$(head -1 "$skill_md" | grep -c "^---" || true)
  if [ "$has_frontmatter" -eq 0 ]; then
    check "$skill_name" "R2: YAML frontmatter" "fail" "No --- frontmatter block"
    return
  fi
  
  local has_name=$(sed -n '/^---$/,/^---$/p' "$skill_md" | grep -c "^name:" || true)
  local has_desc=$(sed -n '/^---$/,/^---$/p' "$skill_md" | grep -c "^description:" || true)
  
  if [ "$has_name" -eq 0 ] || [ "$has_desc" -eq 0 ]; then
    check "$skill_name" "R2: Frontmatter fields" "fail" "Missing name or description"
  else
    check "$skill_name" "R2: Frontmatter fields" "pass" "name + description present"
  fi
  
  # Rule 3: Description has keywords + <1024 chars
  local desc_line=$(sed -n '/^description:/,/^---$/p' "$skill_md" | head -1)
  local desc_len=${#desc_line}
  if [ "$desc_len" -gt 1024 ]; then
    check "$skill_name" "R3: Description quality" "fail" "Description too long ($desc_len chars, max 1024)"
  elif [ "$desc_len" -lt 30 ]; then
    check "$skill_name" "R3: Description quality" "warn" "Description very short ($desc_len chars) — may lack keywords"
  else
    check "$skill_name" "R3: Description quality" "pass" "$desc_len chars"
  fi
  
  # Rule 4: SKILL.md < 500 lines
  local line_count=$(wc -l < "$skill_md")
  if [ "$line_count" -gt 500 ]; then
    check "$skill_name" "R4: Progressive disclosure (<500 lines)" "fail" "$line_count lines — move content to references/"
  else
    check "$skill_name" "R4: Progressive disclosure (<500 lines)" "pass" "$line_count lines"
  fi
  
  # Rule 5: References directory (if SKILL.md > 50 lines)
  if [ "$line_count" -gt 50 ] && [ ! -d "$skill_dir/references" ]; then
    check "$skill_name" "R5: references/ directory" "warn" "SKILL.md has $line_count lines but no references/ dir — consider splitting"
  elif [ -d "$skill_dir/references" ]; then
    local ref_count=$(find "$skill_dir/references" -type f | wc -l)
    check "$skill_name" "R5: references/ directory" "pass" "$ref_count reference files"
  else
    check "$skill_name" "R5: references/ directory" "pass" "Small skill, no refs needed"
  fi
  
  # Rule 6: Examples section with I/O
  local has_examples=$(grep -ciE "## .*example|### .*example|input.*output|\*\*Input:\*\*|\*\*Output:\*\*" "$skill_md" || true)
  if [ "$has_examples" -eq 0 ]; then
    check "$skill_name" "R6: Examples with I/O" "fail" "No examples section or input/output examples"
  else
    check "$skill_name" "R6: Examples with I/O" "pass" "Examples section found"
  fi
  
  # Rule 7: Edge cases section
  local has_edge=$(grep -ciE "## .*edge|## .*fail|## .*error|when.*fail|when.*error" "$skill_md" || true)
  if [ "$has_edge" -eq 0 ]; then
    check "$skill_name" "R7: Edge cases" "fail" "No edge cases or failure handling section"
  else
    check "$skill_name" "R7: Edge cases" "pass" "Edge cases section found"
  fi
  
  # Rule 8: Test cases section (prompts that should / should not activate)
  local has_test=$(grep -ciE "## .*test|### .*test|## .*trigger|should activate|should not activate" "$skill_md" || true)
  if [ "$has_test" -eq 0 ]; then
    check "$skill_name" "R8: Test cases" "warn" "No test cases section — add prompts that should/shouldn't activate"
  else
    check "$skill_name" "R8: Test cases" "pass" "Test cases section found"
  fi
  
  # Rule 9: "When to use" section in body
  local has_when=$(grep -ciE "## .*when to use|## .*when.*trigger|## .*use when" "$skill_md" || true)
  if [ "$has_when" -eq 0 ]; then
    check "$skill_name" "R9: When to use (body)" "warn" "No 'When to use' section in body — description covers but body should reinforce"
  else
    check "$skill_name" "R9: When to use (body)" "pass" "When to use section found"
  fi
  
  # Rule 10: Directory structure (no loose files)
  local loose_files=$(find "$skill_dir" -maxdepth 1 -type f ! -name "SKILL.md" ! -name "README.md" ! -name "LICENSE" | wc -l)
  if [ "$loose_files" -gt 0 ]; then
    check "$skill_name" "R10: Clean directory" "warn" "$loose_files unexpected files at root level"
  else
    check "$skill_name" "R10: Clean directory" "pass" "No loose files"
  fi
}

# Main
echo "🔍 cali-skill-validator"
echo "Scanning: $SKILLS_DIR"
echo ""

if [ ! -d "$SKILLS_DIR" ]; then
  echo "❌ Directory not found: $SKILLS_DIR"
  exit 1
fi

for skill_dir in "$SKILLS_DIR"/*/; do
  [ -d "$skill_dir" ] || continue
  validate_skill "$skill_dir"
done

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
