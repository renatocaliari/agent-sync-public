---
name: cali-agents-md-validator
description: >
  [Cali] Validate project AGENTS.md files against best practices and the canonical
  template. Use when: user says 'validate agents md', 'check agents md quality',
  'audit my agents md', or after creating/updating AGENTS.md. Checks: structure,
  size, content quality, template compliance, and provides fix recommendations.
  Can offer automatic corrections via ask tool.

metadata:
  frequency: monthly
  category: meta
  context-cost: low
---

# AGENTS.md Validator

Validates AGENTS.md files against 10 criteria from industry research and the canonical nyosegawa template.

## When to Use

- After creating a new AGENTS.md
- After updating an existing AGENTS.md
- When user says "validate agents md", "check agents md quality", "audit my agents md"
- As part of project setup review
- Before committing changes to AGENTS.md

## Validation Rules

### R1: AGENTS.md Exists (FAIL if missing)
```bash
test -f AGENTS.md && echo "pass" || echo "fail"
```

### R2: File Size ≤150 Lines (FAIL if over)
```bash
lines=$(wc -l < AGENTS.md)
if [ "$lines" -gt 150 ]; then echo "fail: $lines lines (max 150)"; else echo "pass"; fi
```
Source: ETH Zurich research shows >150 lines causes "silent rule dropout"

### R3: Has "Commands" Section (FAIL if missing)
```bash
grep -q "^## Commands\|^##commands\|^# Commands" AGENTS.md && echo "pass" || echo "fail"
```
Source: OpenAI Codex, GitHub analysis — copy-pasteable commands are essential

### R4: Has "Don'ts" Section (FAIL if missing)
```bash
grep -q "^## Don'ts\|^##don'ts\|^# Don'ts" AGENTS.md && echo "pass" || echo "fail"
```
Source: GitHub 2500+ repos analysis — explicit prohibitions prevent common errors

### R5: No Secrets in File (FAIL if found)
```bash
grep -qiE "(api_key|secret|password|token|credential).*=.*['\"][a-zA-Z0-9]{20,}" AGENTS.md && echo "fail" || echo "pass"
```
Source: GitHub analysis — #1 most common constraint in AGENTS.md files

### R6: Stack with Exact Versions (WARN if vague)
```bash
grep -qE "(React|Vue|Angular|Go|Node|Python|Java) *[0-9]+\.[0-9]+" AGENTS.md && echo "pass" || echo "warn: no exact versions found"
```
Source: Augment Code, Inngest repo — "React 19" not just "React"

### R7: No Architectural Overview Duplicating README (WARN)
```bash
# Check if AGENTS.md has long prose sections (>20 lines of description)
awk '/^##.*Architecture|^##.*Overview/,/^##[^#]/' AGENTS.md | wc -l | \
  awk '{if ($1 > 20) print "warn: consider moving architectural details to README"; else print "pass"}'
```
Source: ETH Zurich — arch sections increase cost without improving success

### R8: Critical Rules in First 50 Lines (WARN)
```bash
head -50 AGENTS.md | grep -q "^## Don'ts\|^## Commands\|^## Rule\|^## Rule #1" && echo "pass" || echo "warn: critical rules may be buried"
```
Source: "Lost in the middle" phenomenon — LLMs pay most attention to first/last 25%

### R9: References Use /skill:name or Links (WARN)
```bash
# Check for inline content that should be references
grep -c "@file:\|references/" AGENTS.md | \
  awk '{if ($1 == 0) print "warn: consider using /skill:name for extended docs"; else print "pass"}'
```
Source: Progressive disclosure pattern — reference skills, don't paste content

### R10: Template Compliance (WARN)
```bash
# Check for required sections: Commands + Don't + (Architecture OR Stack)
has_commands=$(grep -q "^## Commands" AGENTS.md && echo 1 || echo 0)
has_donts=$(grep -q "^## Don'ts" AGENTS.md && echo 1 || echo 0)
has_arch=$(grep -q "^## Architecture\|^## Stack" AGENTS.md && echo 1 || echo 0)

if [ "$has_commands" -eq 1 ] && [ "$has_donts" -eq 1 ] && [ "$has_arch" -eq 1 ]; then
  echo "pass"
else
  missing=""
  [ "$has_commands" -eq 0 ] && missing="$missing Commands"
  [ "$has_donts" -eq 0 ] && missing="$missing Don'ts"
  [ "$has_arch" -eq 0 ] && missing="$missing Architecture/Stack"
  echo "warn: missing:$missing"
fi
```
Source: nyosegawa template — Commands + Don'ts + Architecture are the core sections

## Workflow

### Step 1: Run Validation

Execute the validation script:
```bash
bash references/validate-agents-md.sh
```

Or run manually:
```bash
# Count lines
lines=$(wc -l < AGENTS.md)
echo "Lines: $lines"

# Check sections
grep -E "^## " AGENTS.md
```

### Step 2: Present Results

Show the user a clear report:
```
📊 AGENTS.md Validation Report

✅ R1: File exists
✅ R2: 75 lines (max 150)
✅ R3: Commands section found
✅ R4: Don'ts section found
✅ R5: No secrets detected
✅ R6: Stack has exact versions (Go 1.26, Node >=20)
✅ R7: No architectural bloat
✅ R8: Critical rules in first 50 lines
⚠️ R9: No /skill:name references found
✅ R10: Template compliance: Commands + Don'ts + Architecture

Result: 9/10 passed, 1 warning
```

### Step 3: Offer Corrections (if warnings or failures)

Use the ask tool to offer corrections:
```typescript
ask_user_question({
  questions: [{
    question: "How would you like to proceed?",
    header: "Action",
    options: [
      { label: "Auto-fix warnings", description: "I'll apply recommended fixes automatically" },
      { label: "Show recommendations", description: "I'll explain each warning with specific suggestions" },
      { label: "Skip fixes", description: "Just record the report, no changes needed" }
    ]
  }]
})
```

### Step 4: Apply Fixes (if user confirms)

For each warning/failure, generate a fix:

**R2 (too long):** Offer to trim sections, move content to references
**R3 (no Commands):** Generate Commands section from package.json/Makefile
**R4 (no Don'ts):** Generate Don'ts section from common project errors
**R5 (secrets found):** Immediately remove and warn
**R6 (no versions):** Detect versions from go.mod/package.json
**R7 (arch bloat):** Suggest moving to README
**R8 (rules buried):** Offer to reorder sections
**R9 (no skill refs):** Add /skill:name references for extended docs
**R10 (template non-compliant):** Restructure to match template

### Step 5: Verify Fix

After applying fixes, re-run validation to confirm improvement.

## Template Reference

The canonical template structure (from nyosegawa/agents-md-generator):

```markdown
# AGENTS.md

## Project Overview
[Brief description]

## Commands
| Command | Description |
|---------|-------------|
| `npm run dev` | Start dev server |

## Architecture
[Key decisions and patterns]

## Conventions
- Code style rules
- File naming

## Don'ts
- Things to never do
```

**Our extensions (valid in our ecosystem):**
- RULE #1 in Project Overview (trigger rules)
- Tool Rules in Architecture (tool decision matrix)
- Skills Reference in Architecture (skill routing)
- Language Convention in Conventions (English-only)
- Semantic Diff in Commands (sem commands)

## Edge Cases

### AGENTS.md doesn't exist
- Report: "❌ R1: No AGENTS.md found"
- Offer: "Shall I create one? Use `/skill:cali-agents-md-generator`"

### File is exactly 150 lines
- Report: "✅ R2: 150 lines (at limit)"
- No action needed, but warn: "Consider trimming for better LLM adherence"

### Multiple Don'ts sections
- Report: "⚠️ R4: Multiple 'Don'ts' sections found"
- Offer to consolidate into one section

## Test Cases

### Should activate
- "Validate my AGENTS.md"
- "Check if agents md is good quality"
- "Audit the AGENTS.md file"
- "Is my agents md following best practices?"

### Should NOT activate
- "Create AGENTS.md" (use generator instead)
- "Read AGENTS.md" (just reading, not validating)
- "Edit AGENTS.md" (direct edit, not validation workflow)

## References

- `references/validate-agents-md.sh` — Bash validation script
- nyosegawa/agents-md-generator — Canonical template
- agents.md — AGENTS.md standard
- ETH Zurich research — Size limits and rule dropout
- GitHub analysis — 2500+ repos constraint patterns
