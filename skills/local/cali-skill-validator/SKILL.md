---
name: cali-skill-validator
description: "Validate agent skills against industry best practices from AgentSkills Spec, Microsoft Learn, Claude API Docs, Perplexity Research, and Trail of Bits. Triggers when: user says 'validate skills', 'check skills', 'skill quality', 'skill audit', 'are my skills good', 'skill criteria', or when creating/publishing a new skill. Covers: 10-rule validation of frontmatter, description, progressive disclosure, examples, edge cases, test cases, and directory structure."

metadata:
  frequency: monthly
  category: meta
  context-cost: low
---

# Skill Validator

Validate skills against 10 criteria from industry best practices.

## Contents

| Section | What | When to consult |
|---------|------|------------------|
| [Sources](#sources) | Which standards this implements | When citing criteria |
| [10 Rules](#10-rules) | The validation criteria | When understanding checks |
| [Run Validator](#run-validator) | Execute the validation script | Before publishing skills |
| [Fix Common Issues](#fix-common-issues) | How to fix each failure | When a check fails |
| [Test Cases](#test-cases) | Prompts that should/shouldn't activate | When testing trigger accuracy |

## Sources

| Source | What it defines |
|--------|----------------|
| [AgentSkills Spec](https://github.com/agentskills/agentskills) (19.5k ⭐) | Open specification: frontmatter, directory structure, progressive disclosure |
| [Microsoft Learn](https://learn.microsoft.com/en-us/agent-framework/agents/skills) | Enterprise guidance: when to create skills, portable packages |
| [Addy Osmani](https://github.com/addyosmani/agent-skills) | Directory layout, SKILL.md anatomy, references/ pattern |
| [Perplexity Research](https://research.perplexity.ai/articles/designing-refining-and-maintaining-agent-skills-at-perplexity) | 3-tier disclosure, description optimization, testing with evals |
| [OpenAI Skills](https://github.com/openai/skills) | name + description as routing mechanism |
| [Claude API Docs](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices) | Trigger conditions, degrees of freedom, test with all models |
| [Trail of Bits](https://github.com/trailofbits/skills) | 500-line rule, what goes where |
| [Lalit Madan](https://lalitmadan.com/post/the-skill-md-guide) | "Operating manual for repeatable work" |
| [agentskills.io](https://agentskills.io/skill-creation/optimizing-descriptions) | Description optimization |

## 10 Rules

| # | Rule | Source | Severity |
|---|------|--------|----------|
| R1 | SKILL.md exists | AgentSkills spec | FAIL |
| R2 | Frontmatter has `name` + `description` | AgentSkills spec, OpenAI | FAIL |
| R3 | Description has keywords, <1024 chars | agentskills.io, Perplexity | FAIL/WARN |
| R4 | SKILL.md < 500 lines | Trail of Bits | FAIL |
| R5 | `references/` exists when SKILL.md > 50 lines | Addy Osmani | WARN |
| R6 | Examples section with I/O | Claude API, Open Agent Skills | FAIL |
| R7 | Edge cases / failure handling | Claude API, Perplexity | FAIL |
| R8 | Test cases (should/shouldn't activate) | Perplexity, agentskills.io | WARN |
| R9 | "When to use" section in body | Claude API | WARN |
| R10 | Clean directory (no loose files) | AgentSkills spec | WARN |

## Run Validator

```bash
# Validate all skills in default directory
bash references/validate.sh

# Validate specific directory
bash references/validate.sh /path/to/skills

# Validate single skill
bash references/validate.sh /path/to/skills/my-skill
```

Output: per-skill checks with ✅/❌/⚠️ and summary with pass/fail/warn counts.

## Fix Common Issues

### R3: Description too short / no keywords

**Bad:** `description: Helps with PDFs.`
**Good:** `description: Extracts text and tables from PDF files, fills PDF forms, and merges multiple PDFs. Triggers when: user says "extract PDF", "fill PDF form", "merge PDFs", or mentions PDF documents.`

### R4: SKILL.md > 500 lines

Move detailed content to `references/`:
```
my-skill/
├── SKILL.md              # < 500 lines (core workflow)
└── references/
    ├── detailed-guide.md  # Heavy docs loaded on demand
    └── recipes.md         # Concrete examples
```

### R6: No examples

Add section with input → output:
```markdown
## Examples

### Example 1: Extract text from PDF

**Input:** "Extract text from report.pdf"

**Steps:**
1. Run: `python scripts/extract_text.py report.pdf`
2. Return extracted text

**Output:** "Extracted 3 pages, 2,847 characters"
```

### R7: No edge cases

Add section with failure modes:
```markdown
## Edge Cases

### File doesn't exist
- Check if file path is absolute or relative
- Ask user to verify path

### PDF is password-protected
- Ask user for password
- Run: `python scripts/extract_text.py --password <pwd> file.pdf`
```

### R8: No test cases

Add section with activation prompts:
```markdown
## Test Cases

### Should activate
- "Extract text from report.pdf"
- "Fill out the tax form"
- "Merge invoice1.pdf and invoice2.pdf"

### Should NOT activate
- "Convert PDF to Word" (different skill)
- "Read this text file" (not a PDF)
```

## Test Cases

### Should activate
- "Validate my skills"
- "Check if my skills are good"
- "Run skill audit"
- "Are my skills following best practices?"

### Should NOT activate
- "Create a skill" (use skill-creator)
- "Edit SKILL.md" (use edit tool)
- "What is a skill?" (general question)
