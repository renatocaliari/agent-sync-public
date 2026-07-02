---
name: cali-ops-github-releases
description: "Create GitHub releases following project conventions. Triggers when: user says 'release', 'create release', 'push release', 'deploy to main', 'merge to main', user merges a PR to main, or when git push to main is detected. Also triggers on mentions of: gh release, semver, version bump, changelog, release-please. Covers: config-driven (read .release.yml and execute) and fallback (gh CLI) release flows, versioning rules, tag management, and the mandatory release-on-merge convention."

metadata:
  frequency: weekly
  category: infra
  context-cost: low
---

# Releases

**Rule:** Any merge/push to `main` MUST include a GitHub release.
Do not wait to be asked — generate release automatically.

## When to use

Merging to main, pushing to main, creating release manually, setting up automation, hotfix releases, debugging release-please.

## Contents

| Section | When to consult |
|---------|------------------|
| [Release Strategies + sem](#release-strategies) | Before any release |
| [Versioning Rules](#versioning-rules) | Before bumping version |
| [Config-Driven (.release.yml)](#config-driven-release-releaseyml) | Project has `.release.yml` |
| [Fallback: Raw gh CLI](#fallback-raw-gh-cli) | No config file |
| [Dry-Run / Preview](#dry-run--preview) | Preview before executing |
| [Examples](#examples) | First time using skill |
| [Edge Cases](#edge-cases) | Problems arise |
| [Test Cases](#test-cases) | Activation reference |

## Release Strategies

**Discovery order — always check in this sequence:**
1. `.release-please-manifest.json` exists → project uses release-please (see `references/release-please-flow.md`)
2. `.release.yml` exists → use [Config-Driven Release](#config-driven-release-releaseyml) below
3. Neither exists → use [Fallback: Raw gh CLI](#fallback-raw-gh-cli)

## Semantic Diff (Optional Enhancement)

If `sem` is installed (`command -v sem`), enhance the release with structural analysis.
If not installed, the release proceeds normally using commit-prefix parsing.
**Never block a release on `sem` being absent.**

**Prefer `sem diff` over `git diff`** for any entity-level diff (functions, types, methods).
Use `git diff` only when `sem` is unavailable — it shows raw lines, not meaningful symbols.

Check: `command -v sem && echo "enhanced mode" || echo "commit-parsing mode"`

| Enhancement | What `sem` adds | Fallback |
|-------------|-----------------|----------|
| Bump type detection | Sees signature changes, removed exports → catches breaking changes commit prefixes miss | Parse `feat:`/`fix:`/`BREAKING:` in commit messages |
| Rich changelog | Symbol-level diff (`sem diff $LAST_TAG HEAD`) alongside commit list | `git log --oneline` only |
| Pre-release sanity | Confirms actual symbol changes exist before tagging | Trust commit list |

Full commands and integration into Steps 2-3, 10-11, and raw Steps 3-4:
`references/sem-enhancement.md`

## Versioning Rules

Follow [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html):

| Rule | Example |
|------|---------|
| Use alpha/beta only | `0.2.0-alpha`, `0.3.1-beta` |
| NEVER `1.0.0` without owner confirmation | ❌ `1.0.0` → ✅ `0.9.0` |
| Patch for backwards-compatible fixes | `0.2.0` → `0.2.1` |
| Minor for new features (backwards-compatible) | `0.2.0` → `0.3.0` |
| Major for breaking changes | `0.2.0` → `1.0.0` (with approval) |

## Pre-Flight Checklist (ALL flows)

Run these BEFORE any release, regardless of which flow is used:

```bash
# 1. Check gh CLI is authenticated
gh auth status || { echo "❌ Run: gh auth login"; exit 1; }

# 2. Check clean git state — auto-commit uncommitted changes
if ! git diff --quiet HEAD; then
  echo "📝 Uncommitted changes found — auto-committing before release"
  git add -A

  # Build a descriptive commit message
  FILE_COUNT=$(git diff --cached --name-only | wc -l | tr -d ' ')
  ADD_DEL=$(git diff --cached --numstat | awk '{a+=$1; d+=$2} END {printf "+%d -%d", a, d}')

  if command -v sem &> /dev/null; then
    SEM_OUT=$(sem diff 2>/dev/null | head -3 | tr '\n' ' ' | xargs)
  fi

  if [ -n "$SEM_OUT" ]; then
    COMMIT_MSG="chore: ${SEM_OUT} (${FILE_COUNT} files, ${ADD_DEL})"
  else
    COMMIT_MSG="chore: prepare release (${FILE_COUNT} files, ${ADD_DEL})"
  fi

  git commit -m "${COMMIT_MSG}"
  echo "✅ Changes committed: $(git rev-parse HEAD | head -c 7) — ${COMMIT_MSG}"
fi

# 3. Check we're on main/master (block feature branches)
BRANCH=$(git branch --show-current)
if [ "$BRANCH" != "main" ] && [ "$BRANCH" != "master" ]; then
  echo "❌ Not on main/master branch. Current: $BRANCH"
  echo "   git checkout main && git pull   # switch to main first"
  exit 1
fi
echo "✅ On branch: $BRANCH"
```

## Config-Driven Release (`.release.yml`)

When a project has `.release.yml`, the skill reads it and executes every step.
**No project-level scripts needed** — the skill is the executor.

### `.release.yml` format

```yaml
# Suffix appended to version numbers (e.g. 0.11.1-alpha)
# Default: ""  (no suffix)
suffix: alpha

# Version files to sync when bumping
# The skill will update the version field in each file.
# Default: [] (auto-detect package.json, pyproject.toml, Cargo.toml)
version_files:
  - package.json

# Command to run tests before release
# Default: "" (no tests)
test_command: npm test

# Command to run lint/typecheck before release
# Default: "" (no lint)
lint_command: npm run typecheck
```

### Execution Steps

The skill follows this exact flow:

```
Step 0:  Read .release.yml
Step 1:  Run Pre-Flight Checklist (gh auth + clean git)
Step 2:  Detect last tag (git describe --tags)
Step 3:  Parse version and bump it (preserving suffix)
Step 4:  Confirm major release if applicable
Step 5:  Run test_command (if defined)
Step 6:  Run lint_command (if defined)
Step 7:  Sync version in all version_files
Step 8:  Create annotated git tag
Step 9:  Push tag to GitHub
Step 10: Build changelog from commits since last tag
Step 11: Create GitHub Release (gh release create)
Step 12: Verify and report
```

### Commands to execute (per step)

**Step 0 — Read config:**
```bash
# Read .release.yml values. Defaults if missing:
SUFFIX=""                 # default: no suffix
VERSION_FILES=""          # auto-detect
TEST_CMD=""
LINT_CMD=""
```

**Step 2-3 — Detect & bump version:**
```bash
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
RAW="${LAST_TAG#v}"

# Parse: major.minor.patch (suffix comes from .release.yml, ignore tag suffix)
IFS='-' read -r VERSION_PART _REST <<< "$RAW"
IFS='.' read -r MAJOR MINOR PATCH <<< "${VERSION_PART}"

# Bump — TYPE is patch/minor/major determined from commits
case "$TYPE" in
  patch) PATCH=$((PATCH + 1)) ;;
  minor) MINOR=$((MINOR + 1)); PATCH=0 ;;
  major) MAJOR=$((MAJOR + 1)); MINOR=0; PATCH=0 ;;
esac

# Apply suffix from .release.yml (populated by agent reading the config)
CONFIG_SUFFIX="${YML_SUFFIX:-}"
if [ -n "$CONFIG_SUFFIX" ]; then
  NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}-${CONFIG_SUFFIX}"
else
  NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"
fi
NEW_TAG="v${NEW_VERSION}"
```

**Step 4 — Major confirmation:**
```bash
if [ "$TYPE" = "major" ]; then
  echo "⚠  MAJOR release: $LAST_TAG → $NEW_TAG"
  echo "This requires owner confirmation. Ask the user before proceeding."
fi
```

**Step 5-6 — Run tests/lint:**
```bash
if [ -n "$TEST_CMD" ]; then
  echo "➜ Running tests: $TEST_CMD"
  eval "$TEST_CMD" || { echo "Tests failed — aborting release"; exit 1; }
fi

if [ -n "$LINT_CMD" ]; then
  echo "➜ Running lint: $LINT_CMD"
  eval "$LINT_CMD" || { echo "Lint failed — aborting release"; exit 1; }
fi
```

**Step 5b — Signoff gate (CI fallback):**
If the repo has **no GitHub Actions CI**, sign the release commit cryptographically.
```bash
if [ -n "$TEST_CMD" ] && ! ls .github/workflows/ 2>/dev/null | grep -q .; then
  echo "➜ No GitHub Actions CI detected — running gh signoff..."
  gh signoff || { echo "Signoff failed — aborting release"; exit 1; }
fi
```
Requires [`gh-signoff`](https://github.com/basecamp/gh-signoff): `gh extension install basecamp/gh-signoff`

**Step 7 — Sync version files:**
```bash
# Read version_files from .release.yml
# For each file, update the version field
for f in "${VERSION_FILES[@]}"; do
  case "$f" in
    *.json)
      node -e "
        const fs = require('fs');
        const p = JSON.parse(fs.readFileSync('./$f', 'utf8'));
        if (p.version) {
          p.version = '$NEW_VERSION';
          fs.writeFileSync('./$f', JSON.stringify(p, null, 2) + '\n');
        } else {
          console.error('WARNING: no version field in $f');
        }
      "
      git add "$f"
      ;;
    *.toml)
      sed -i.bak 's/^version = ".*"/version = "'"$NEW_VERSION"'"/' "$f" && rm -f "$f.bak"
      git add "$f"
      ;;
    *.go)
      # Go version file: var/const Version = "x.y.z"
      sed -i.bak 's/\(Version\\s*=\\s*\)".*"/\1"'"$NEW_VERSION"'"/' "$f" && rm -f "$f.bak"
      git add "$f"
      ;;
  esac
done

# Commit version bump BEFORE tagging (tag must point to the bumped commit)
if ! git diff --cached --quiet; then
  git commit -m "chore: bump version to $NEW_VERSION"
fi
```

**Step 8-9 — Tag + push:**
```bash
git tag -a "$NEW_TAG" -m "Release $NEW_TAG"
git push origin "$NEW_TAG"
```

**Step 10-11 — Create GitHub Release:**

Follow [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/) format:

```bash
# Build changelog — handle root commit / first tag gracefully
PREV_TAG=$(git describe --tags --abbrev=0 "$NEW_TAG"^ 2>/dev/null || true)

CATEGORIZED=$(mktemp)
CHANGELOG=$(mktemp)

# Categorize commits by conventional commit prefix
if [ -n "$PREV_TAG" ]; then
  LOG_RANGE="$PREV_TAG..$NEW_TAG"
  SCOPE_LABEL="$PREV_TAG"
else
  LOG_RANGE="HEAD"
  SCOPE_LABEL="initial commit"
fi

# Build categorized changelog following Keep a Changelog format
{
  echo "# Release $NEW_TAG"
  echo ""
  echo "## What's Changed"
  echo ""

  # Added (feat:)
  ADDED=$(git log "$LOG_RANGE" --oneline --no-decorate --grep="^feat" 2>/dev/null | sed 's/^/* /')
  if [ -n "$ADDED" ]; then
    echo "### Added"
    echo "$ADDED"
    echo ""
  fi

  # Fixed (fix:)
  FIXED=$(git log "$LOG_RANGE" --oneline --no-decorate --grep="^fix" 2>/dev/null | sed 's/^/* /')
  if [ -n "$FIXED" ]; then
    echo "### Fixed"
    echo "$FIXED"
    echo ""
  fi

  # Changed (refactor:, perf:, chore:)
  CHANGED=$(git log "$LOG_RANGE" --oneline --no-decorate --grep="^refactor\|^perf\|^chore" 2>/dev/null | sed 's/^/* /')
  if [ -n "$CHANGED" ]; then
    echo "### Changed"
    echo "$CHANGED"
    echo ""
  fi

  # Security (security:)
  SECURITY=$(git log "$LOG_RANGE" --oneline --no-decorate --grep="^security\|^sec" 2>/dev/null | sed 's/^/* /')
  if [ -n "$SECURITY" ]; then
    echo "### Security"
    echo "$SECURITY"
    echo ""
  fi

  # Remaining uncategorized commits
  CATEGORIZED_COM=$( (git log "$LOG_RANGE" --oneline --no-decorate --grep="^feat\|^fix\|^refactor\|^perf\|^chore\|^security\|^sec" 2>/dev/null) | wc -l | tr -d ' ')
  TOTAL_COM=$(git log "$LOG_RANGE" --oneline --no-decorate 2>/dev/null | wc -l | tr -d ' ')
  if [ "$TOTAL_COM" -gt "$CATEGORIZED_COM" ]; then
    echo "### Other"
    git log "$LOG_RANGE" --oneline --no-decorate --invert-grep --grep="^feat\|^fix\|^refactor\|^perf\|^chore\|^security\|^sec" 2>/dev/null | sed 's/^/* /'
    echo ""
  fi

  echo "---"
  echo "_Full diff: [$SCOPE_LABEL...$NEW_TAG](https://github.com/$(gh repo view --json nameWithOwner --jq .nameWithOwner 2>/dev/null || echo 'repo')/compare/$SCOPE_LABEL...$NEW_TAG)_"
} > "$CATEGORIZED"

# Use --notes-file to avoid shell escaping bugs
gh release create "$NEW_TAG" --title "$NEW_TAG" --notes-file "$CATEGORIZED"
rm -f "$CATEGORIZED" "$CHANGELOG"
```

## Fallback: Raw gh CLI

Use this when the project has **neither** `.release-please-manifest.json` nor `.release.yml`.

### Step 1 — Run Pre-Flight Checklist

```bash
# Execute the checks in [Pre-Flight Checklist](#pre-flight-checklist-all-flows)
# If any check fails, stop and tell the user what to fix.
```

### Step 1b — Test + Signoff gate (no CI fallback)

If the repo has **no GitHub Actions CI**, run tests and sign the release.
```bash
TEST_CMD=""
[ -f package.json ] && TEST_CMD="npm test"
[ -f go.mod ] && TEST_CMD="go test ./..."
[ -f Cargo.toml ] && TEST_CMD="cargo test"
[ -f pyproject.toml ] && TEST_CMD="pytest"

if [ -n "$TEST_CMD" ] && ! ls .github/workflows/ 2>/dev/null | grep -q .; then
  echo "➜ No GitHub Actions CI detected — running tests..."
  eval "$TEST_CMD" || { echo "Tests failed — aborting release"; exit 1; }
  echo "➜ Running gh signoff..."
  gh signoff || { echo "Signoff failed — aborting release"; exit 1; }
fi
```
Requires [`gh-signoff`](https://github.com/basecamp/gh-signoff): `gh extension install basecamp/gh-signoff`

### Step 2 — Determine what changed since last release

```bash
# Find last tag
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

if [ -z "$LAST_TAG" ]; then
  echo "No previous tags found. This is the first release."
  # Check for existing version indicators
  cat package.json 2>/dev/null | grep '"version"' || true
  grep 'version =' Cargo.toml 2>/dev/null || true
else
  echo "Last tag: $LAST_TAG"
  echo ""
  echo "Commits since last release:"
  git log "$LAST_TAG"..HEAD --oneline --no-decorate
fi
```

### Step 3 — Decide version bump type

Based on the commits found:
- `fix:` commits → **patch** bump
- `feat:` commits → **minor** bump
- `BREAKING CHANGE:` or `!:` → **major** bump (requires user confirmation)
- `refactor:` / `chore:` / `docs:` → **patch** bump (safe default)
- Mixed feat + fix → **minor** bump
- No conventional commit prefixes → ask the user with this template:

  > "The commits don't follow conventional commit format (no `feat:`/`fix:` prefixes).
  >  Here are the changes since the last release:
  >  ```
  >  [paste git log output]
  >  ```
  >  What kind of version bump should this be? `patch` (fixes) / `minor` (features) / `major` (breaking)"

```bash
# Compute next version
RAW="${LAST_TAG#v}"
IFS='.' read -r MAJOR MINOR PATCH <<< "$RAW"
# Bump based on decision above
NEW_TAG="v${MAJOR}.${MINOR}.$((PATCH + 1))"  # default: patch
echo "Proposed: $LAST_TAG → $NEW_TAG"
```

### Step 4 — Create the release

```bash
# Build categorized changelog following Keep a Changelog format
CHANGELOG=$(mktemp)

if [ -n "$LAST_TAG" ]; then
  LOG_RANGE="$LAST_TAG..HEAD"
else
  LOG_RANGE="HEAD"
fi

{
  echo "# Release $NEW_TAG"
  echo ""
  echo "## What's Changed"
  echo ""

  # Added (feat:)
  ADDED=$(git log "$LOG_RANGE" --oneline --no-decorate --grep="^feat" 2>/dev/null | sed 's/^/* /')
  if [ -n "$ADDED" ]; then
    echo "### Added"
    echo "$ADDED"
    echo ""
  fi

  # Fixed (fix:)
  FIXED=$(git log "$LOG_RANGE" --oneline --no-decorate --grep="^fix" 2>/dev/null | sed 's/^/* /')
  if [ -n "$FIXED" ]; then
    echo "### Fixed"
    echo "$FIXED"
    echo ""
  fi

  # Changed (refactor:, perf:, chore:)
  CHANGED=$(git log "$LOG_RANGE" --oneline --no-decorate --grep="^refactor\|^perf\|^chore" 2>/dev/null | sed 's/^/* /')
  if [ -n "$CHANGED" ]; then
    echo "### Changed"
    echo "$CHANGED"
    echo ""
  fi

  # Security (security:)
  SECURITY=$(git log "$LOG_RANGE" --oneline --no-decorate --grep="^security\|^sec" 2>/dev/null | sed 's/^/* /')
  if [ -n "$SECURITY" ]; then
    echo "### Security"
    echo "$SECURITY"
    echo ""
  fi

  # Remaining uncategorized commits
  CATEGORIZED_COM=$( (git log "$LOG_RANGE" --oneline --no-decorate --grep="^feat\|^fix\|^refactor\|^perf\|^chore\|^security\|^sec" 2>/dev/null) | wc -l | tr -d ' ')
  TOTAL_COM=$(git log "$LOG_RANGE" --oneline --no-decorate 2>/dev/null | wc -l | tr -d ' ')
  if [ "$TOTAL_COM" -gt "$CATEGORIZED_COM" ]; then
    echo "### Other"
    git log "$LOG_RANGE" --oneline --no-decorate --invert-grep --grep="^feat\|^fix\|^refactor\|^perf\|^chore\|^security\|^sec" 2>/dev/null | sed 's/^/* /'
    echo ""
  fi
} > "$CHANGELOG"

# Create the release
gh release create "$NEW_TAG" \
  --title "$NEW_TAG" \
  --notes-file "$CHANGELOG"

rm -f "$CHANGELOG"
```

### Step 5 — Verify

```bash
gh release view "$NEW_TAG"
echo "✅ Release $NEW_TAG created"
```

## Examples

### Example 1: Config-driven release (.release.yml)

**Input:** "Just merged the auth feature to main"

**Discovery:** `.release.yml` exists → use config-driven flow

**Execution:**
1. Run Pre-Flight Checklist → clean ✅
2. Read `.release.yml`: suffix=alpha, test_command=`npm test`, version_files=`[package.json]`
3. Detect last tag: `v0.2.0-alpha`
4. Bump to: `v0.3.0-alpha` (feature = minor)
5. Run: `npm test` ✅
6. Sync: `package.json` → `0.3.0-alpha`
7. Create tag: `v0.3.0-alpha`
8. Push + create GitHub Release

**Output:** "Released v0.3.0-alpha. Tests passed. Version synced to package.json."

### Example 2: Raw gh CLI fallback (no config)

**Input:** "Release this"

**Discovery:** No `.release.yml`, no `release-please` → fallback to raw gh CLI

**Execution:**
1. Run Pre-Flight Checklist → clean ✅
2. Find last tag: `v0.1.0`
3. Show commits: `feat: add login`, `fix: button color`
4. Decision: minor bump → `v0.2.0`
5. Create release: `gh release create v0.2.0 --title "v0.2.0" --notes "..."`

**Output:** "Released v0.2.0"

### Example 3: Emergency hotfix

**Input:** "Production is down, critical bug in auth"

**Execution:**
1. Create hotfix branch: `git checkout -b hotfix/fix-auth-crash main`
2. Make fix, commit
3. Merge to main
4. Run pre-flight checklist
5. Create release (config-driven if `.release.yml` exists, raw gh CLI otherwise)

**Output:** "Hotfix released. Production should recover."

## Dry-Run / Preview

To preview what would happen without making changes, the agent should:
1. Run discovery and version detection
2. If `sem` available, run `sem diff` to enhance the preview
3. Print: `Would create tag vX.Y.Z from branch <branch> with N commits`
4. Show the changelog that would be included (enriched with `sem diff` if available)
5. NOT create any tags, commits, or releases

Example output:
```
[Dry-Run] sem: 3 symbols changed (1 new function, 2 modified)
[Dry-Run] Would bump v0.2.0-alpha → v0.3.0-alpha (minor: 2 feat commits)
[Dry-Run] Changelog:
  feat: add login page
  feat: add dashboard
  fix: button hover color
[Dry-Run] No actions taken. Run without --dry-run to execute.
```

## Edge Cases

### Tests fail before release
- BLOCK the release. Do not create.
- Tell user: "Tests failing — fix first, then release"
- Offer to run tests: `go test ./...` or `npm test`

### No changes since last release
- Check: `git diff v0.X.Y-alpha..HEAD --stat`
- If empty: "No changes since last release. Nothing to release."
- Don't create empty releases

### gh CLI not authenticated
- Check: `gh auth status`
- If not logged in: tell user to run `gh auth login`
- Do not proceed until authenticated

### Tag already exists
- Check: `git tag -l "v0.X.Y*"`
- If tag exists: "Tag v0.X.Y already exists. Delete the tag first or use a different version."
- **Never** force-overwrite tags

### Uncommitted changes
- Auto-committed by Pre-Flight Checklist in ALL flows (git add -A + commit)
- Message: "📝 Uncommitted changes found — auto-committing before release"
- Builds a descriptive commit message:
  - If `sem` is installed: `chore: <sem diff summary> (N files, +M -K)`
  - Fallback: `chore: prepare release (N files, +M -K)`
  - Always includes file count and +/- diffstat
- If user wants to abort after the fact: `git reset --soft HEAD~1`
- **Never** leave uncommitted work behind — the release tag must capture all changes

### GPG-signed tags required
- Some orgs require signed tags (`git tag -s` instead of `-a`)
- If the project has `tag.gpgSign = true` in git config, use `-s` flag
- Ask user: "Does this project require GPG-signed tags?"
- Fail gracefully if GPG key is missing: `error: gpg failed to sign the tag`

### Protected branches prevent tag push
- If `git push origin "$NEW_TAG"` is rejected: "Tag push rejected — main branch may be protected"
- Check: GitHub branch protection rules for `main`
- Fix: user must temporarily allow tag pushes or create via GitHub UI

### Monorepo with independent versions
- If project has multiple packages (`packages/*`), `.release.yml` needs `version_files` per package
- Each package gets its own tag (e.g. `pkg-a-v0.3.0`, `pkg-b-v0.1.5`)
- Ask user which packages to release

### release-please not creating releases
- Check: commit messages follow conventional commits (`feat:`, `fix:`, etc.)
- `refactor:` commits don't trigger releases — use manual flow
- Check `.release-please-manifest.json` exists
- See `references/release-please-flow.md`

### JSON parsing errors in workflow
- Symptom: `Error reading JToken from JsonReader` or syntax errors
- Fix: Pass JSON via env var, parse with `jq`
- See: `references/release-please-flow.md`

### Empty repository (no tags, no commits)
- First release: start at `v0.1.0` (or `v0.1.0-alpha` with suffix)
- Confirm version with user before creating

## Test Cases

### Should activate
- "Create a release", "Push a release to main", "release now", "Emergency hotfix needed", "Release-please not working"

### Should NOT activate
- "Write release notes" (just writing, not releasing)
- "What's the current version?" or "Run the tests"

## References

### Internal
- `references/release-please-flow.md` — release-please automation, refactor handling, JSON safety, config
- `references/sem-enhancement.md` — sem diff: bump detection, rich changelog, sanity check
- `cali-ops-deploy-github-tailscale` — deploy pipeline (trigger after release)

### External
- [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/) — changelog format standard (Added/Changed/Deprecated/Removed/Fixed/Security)
- [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html) — version number semantics (MAJOR.MINOR.PATCH)
