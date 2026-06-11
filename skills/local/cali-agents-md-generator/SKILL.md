---
name: cali-agents-md-generator
description: "[Cali] Generate and maintain project AGENTS.md files using semantic analysis. Triggers when: a project has no AGENTS.md, user asks to 'create AGENTS.md', 'generate agents md', 'setup agents md', or when starting work on a project that lacks one. Also triggers when user asks to 'check if AGENTS.md is stale', 'update agents md', or when evolving an existing AGENTS.md. Covers: initial generation from codebase analysis, sem-based staleness detection, pre-commit hooks, and CI guards. Always offer hook installation when invoked for any reason."

metadata:
  frequency: monthly
  category: meta
  context-cost: low
---

# AGENTS.md Generator

Creates and maintains lean, high-signal AGENTS.md files for AI coding agents.
Uses `sem` (semantic version control) for staleness detection on ongoing commits.

## Prerequisites: `sem`

`sem` is required for staleness detection. Check if installed:

```bash
sem --version
```

If not installed:

```bash
brew install sem-cli
```

Verify it works in the project:

```bash
cd /path/to/project
sem diff --format json  # Should return JSON with entity changes
```

## Workflow

### Phase 1: Check

Does the project root have `AGENTS.md`?

```bash
ls AGENTS.md 2>/dev/null
```

- **If exists**: Ask user if they want to review/update it. If yes, go to Phase 3.
- **If missing**: Ask user: "This project has no AGENTS.md. Shall I create one?"
  - If no: stop.
  - If yes: go to Phase 2.

### Phase 2: Scaffold

Generate a minimal AGENTS.md using the template from `references/agmd-template.md`.
Write it to the project root.

### Phase 3: Investigate & Fill

Analyze the full codebase to fill in placeholders:

1. **Directory structure**: `find . -type f -name '*.go' | head -30` (or relevant extensions)
2. **Entry points**: Look for `main.go`, `cmd/`, `index.ts`, etc.
3. **Build system**: Check `Makefile`, `go.mod`, `package.json`, `Dockerfile`
4. **Test setup**: Find test files, check test runner config
5. **Dependencies**: Read `go.mod`, `package.json`, `requirements.txt`
6. **Conventions**: Scan existing code for naming patterns, error handling, project structure

Fill every `[To be determined]` and `[Add your ... here]` placeholder.
Remove any section that stays empty after investigation.

**Target: 20-30 lines max.** If over 30 lines, move detailed content to separate docs.

### Phase 4: Review

Present the generated AGENTS.md to the user for review.
Suggest removing any section the agent can infer from code.

### Phase 5: Offer Staleness Guard (ALWAYS)

**Every time this skill is invoked** — whether creating new, reviewing, or evolving —
check if the sem pre-commit hook is installed and offer to install it.

Check if hook exists:

```bash
test -f .git/hooks/pre-commit && grep -q "sem diff" .git/hooks/pre-commit 2>/dev/null && echo "installed" || echo "not installed"
```

**If not installed**, ask the user:

> The `sem` pre-commit hook detects when code entities (functions, classes, methods)
> change and warns if AGENTS.md may be stale. This keeps your project docs accurate
> without manual checking.
>
> Shall I install the sem staleness hook in `.git/hooks/pre-commit`?

**If user confirms**, install:

```bash
# If no existing pre-commit hook:
cp references/pre-commit-hook.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

# If existing hook needs merging: append sem check to end of existing hook
cat references/pre-commit-hook.sh >> .git/hooks/pre-commit
```

**If user declines**, skip silently. Don't ask again in this session.

**If already installed**, confirm silently:

```
✅ sem staleness hook already installed in .git/hooks/pre-commit
```

## Template Design Principles

- **20-30 line budget** — LLMs degrade with too many instructions
- **No backward compatibility by default** — Bold refactoring over legacy support
- **Placeholder sections** — Fill as project grows, then remove placeholders
- **Section protection via HTML comments** — `<!-- Do not restructure or delete sections -->`
- **Only document non-obvious things** — Code structure belongs in code, not AGENTS.md

## Staleness Detection

The pre-commit hook uses `sem diff --format json` to detect entity-level changes
(functions, classes, methods, headings) and warns when AGENTS.md may need updating.

**What `sem` provides:**
- Entity-level diffs (not just line numbers)
- Rename detection
- Structural change classification
- JSON output for script consumption

**What the hook checks:**
- Whether any code entities changed in the staged files
- Lists the changed entities with type and file
- Reminds to review AGENTS.md

## CI Guard (Optional)

For GitHub Actions, add the workflow from `references/github-action-guard.yml`
to `.github/workflows/agents-md-guard.yml`. It runs `sem diff` in CI and
comments on PRs when code entities change without corresponding AGENTS.md updates.

## Examples

### Example 1: New project without AGENTS.md

**Input:** "This project has no AGENTS.md, create one"

**Steps:**
1. Check: `ls AGENTS.md 2>/dev/null` → not found
2. Ask: "Shall I create one?" → User says yes
3. Scaffold from template
4. Investigate: entry points, build system, tests, conventions
5. Fill placeholders, target 20-30 lines
6. Present for review
7. Offer hook installation

**Output:** "Created AGENTS.md with 24 lines covering: entry point, build, test, conventions. Hook installed."

### Example 2: Existing AGENTS.md needs update

**Input:** "Check if our AGENTS.md is still accurate"

**Steps:**
1. Check: `ls AGENTS.md 2>/dev/null` → found
2. Ask: "Want to review/update it?" → User says yes
3. Run: `sem diff --format json` → shows 5 entity changes since last AGENTS.md update
4. Investigate changed entities
5. Update relevant sections
6. Offer hook installation (if not installed)

**Output:** "Found 5 entity changes since last update. Updated 2 sections. Hook already installed."

### Example 3: Hook installation

**Input:** "Install the staleness hook"

**Steps:**
1. Check: `test -f .git/hooks/pre-commit && grep -q "sem diff" .git/hooks/pre-commit` → not found
2. Explain benefit
3. Ask: "Shall I install?" → User says yes
4. Install: `cp references/pre-commit-hook.sh .git/hooks/pre-commit`
5. Verify: `chmod +x .git/hooks/pre-commit`

**Output:** "✅ Sem staleness hook installed. It will warn when code entities change without AGENTS.md updates."

## Edge Cases

### Project has no .git directory
- Cannot install pre-commit hook
- Tell user: "This project has no git repo. Initialize git first, then install the hook."
- Still generate AGENTS.md (it's useful without the hook)

### sem is not installed
- Check: `sem --version` → not found
- Offer to install: `brew install sem-cli`
- If user declines: skip hook installation, still generate AGENTS.md
- Note: staleness detection won't work without sem

### AGENTS.md is already > 30 lines
- Offer to trim it down
- Suggest moving detailed content to separate docs
- Apply 20-30 line budget principle

### User declines hook installation
- Skip silently
- Don't ask again in this session
- AGENTS.md is still useful without the hook

## Test Cases

### Should activate
- "Create AGENTS.md for this project"
- "Generate agents md"
- "Setup agents md"
- "Check if AGENTS.md is stale"
- "Update the agents md"
- "Install the sem hook"

### Should NOT activate
- "Read AGENTS.md" (just reading, not updating)
- "What is AGENTS.md?" (general question)
- "Edit the AGENTS.md file" (direct edit, not skill workflow)

## References

- `references/agmd-template.md` — Minimal AGENTS.md template (20-30 lines)
- `references/pre-commit-hook.sh` — sem-based staleness detection hook
- `references/github-action-guard.yml` — Optional CI guard workflow
