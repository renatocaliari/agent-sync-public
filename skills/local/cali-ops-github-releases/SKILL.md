---
name: cali-ops-github-releases
description: "Create GitHub releases following project conventions. Triggers when: user says 'release', 'create release', 'push release', 'deploy to main', 'merge to main', user merges a PR to main, or when git push to main is detected. Also triggers on mentions of: gh release, semver, version bump, changelog, release-please. Covers: automated (release-please) and manual (gh CLI) release flows, versioning rules, tag management, and the mandatory release-on-merge convention."

metadata:
  frequency: weekly
  category: infra
  context-cost: low
---

# Releases

**Rule:** Any merge/push to `main` MUST include a GitHub release.
Do not wait to be asked — generate release automatically.

**Read first:** `{project}/.pi/instructions.md` (if exists)

## When to use

Use this skill when:
- Merging a PR to `main`
- Pushing commits to `main`
- Creating a release manually
- Setting up release automation
- Handling hotfix releases
- Debugging release-please issues

## Contents

| Section | What | When to consult |
|---------|------|------------------|
| [Release Strategies](#release-strategies) | Automated vs Manual | Before any release |
| [Versioning Rules](#versioning-rules) | Alpha/beta convention, version increments | Before creating any release |
| [Automated Flow (release-please)](#automated-flow-release-please) | PR-based automation | When project uses release-please |
| [Manual Flow (gh CLI)](#manual-flow-gh-cli) | Direct release creation | When project uses manual releases |
| [Refactor Commit Handling](#refactor-commit-handling) | Releases for non-conventional commits | When refactor commits need releases |
| [CI/CD Integration](#cicd-integration) | Build triggers, deploy patterns | When project uses CI/CD |
| [Edge Cases](#edge-cases) | Common problems and solutions | When issues arise |

## Release Strategies

Choose based on project setup:

| Strategy | When to use | Pros | Cons |
|----------|-------------|------|------|
| **release-please** | Teams, conventional commits | Auto versioning, changelogs, PRs | Setup complexity |
| **Manual (gh CLI)** | Solo projects, simple needs | Full control, simple | Manual versioning |

**Detection:** Check for `.release-please-manifest.json` or `release.yml` workflow.

## Versioning Rules

| Rule | Example |
|------|---------|
| Use alpha/beta only | `0.2.0-alpha`, `0.3.1-beta` |
| NEVER `1.0.0` without owner confirmation | ❌ `1.0.0` → ✅ `0.9.0` |
| Increment patch for fixes | `0.2.0` → `0.2.1` |
| Increment minor for features | `0.2.0` → `0.3.0` |
| Increment major for breaking | `0.2.0` → `1.0.0` (with approval) |

## Automated Flow (release-please)

When project uses `googleapis/release-please-action`:

### Workflow Pattern

```yaml
name: Release
on:
  push:
    branches: [master]

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: googleapis/release-please-action@v4
        id: release
        with:
          config-file: release-please-config.json
          manifest-file: .release-please-manifest.json
          target-branch: master
          token: ${{ secrets.GITHUB_TOKEN }}

      - name: Auto-merge Release PR
        if: ${{ steps.release.outputs.release_created != 'true' }}
        env:
          GH_TOKEN: ${{ secrets.GH_PAT || secrets.GITHUB_TOKEN }}
          RELEASE_PR: ${{ steps.release.outputs.pr }}
        run: |
          if [ -z "$RELEASE_PR" ] || [ "$RELEASE_PR" = "null" ]; then
            echo "No release PR created, skipping merge"
            exit 0
          fi
          PR_NUMBER=$(echo "$RELEASE_PR" | jq -r '.number // empty')
          if [ -n "$PR_NUMBER" ]; then
            echo "Merging release PR #$PR_NUMBER"
            gh pr merge --merge "$PR_NUMBER"
          else
            echo "No release PR to merge"
          fi
```

### Critical: Safe JSON Handling in Bash

**Problem:** GitHub Actions context outputs contain JSON with special characters (parentheses, quotes) that break bash syntax when expanded inline.

**Wrong:**
```bash
# ❌ BREAKS when JSON contains parentheses or special chars
PR_NUMBER="${{ fromJSON(steps.release.outputs.pr).number }}"
```

**Correct:**
```bash
# ✅ Pass JSON via env var, parse with jq
env:
  RELEASE_PR: ${{ steps.release.outputs.pr }}
run: |
  if [ -z "$RELEASE_PR" ] || [ "$RELEASE_PR" = "null" ]; then
    echo "No release PR"
    exit 0
  fi
  PR_NUMBER=$(echo "$RELEASE_PR" | jq -r '.number // empty')
```

**Why:**
- Inline `${{ }}` expansion happens before bash parsing
- JSON with `()`, `"`, `$` causes syntax errors
- Env vars preserve content integrity
- `jq` safely extracts values

### Workflow Redundancy

**Rule:** One workflow per trigger type. Avoid duplicate workflows for same event.

**Check:** If you have both `release.yml` and `auto-merge-release.yml`:
- `release.yml` on `push` → handles release-please + auto-merge
- `auto-merge-release.yml` on `pull_request_target` → redundant if release.yml already merges

**Fix:** Remove redundant workflow. Keep the one that handles the full flow.

## Manual Flow (gh CLI)

### Pre-release Checks

```bash
# Run tests
npm run test        # or: go test ./... / make test

# Run linter
npm run lint        # or: golangci-lint run

# Check for uncommitted changes
git status
```

### Determine Version

```bash
# Check current version
git tag --sort=-v:refname | head -5

# Or check package.json / go module version
```

### Create Release

```bash
# Create release (interactive — pick or type version)
gh release create

# Or explicit:
gh release create v0.X.Y-alpha \
  --title "v0.X.Y-alpha: <summary>" \
  --notes "## Changes

- Feature: description
- Fix: description

## Breaking Changes

- None (or list)

## Testing

- [ ] All tests pass
- [ ] Manual verification done"
```

### Push Tag (if not auto-pushed)

```bash
git push origin v0.X.Y-alpha
```

### Verify

```bash
# Check release appears
gh release list

# Check tag exists
git tag -l "v0.X.Y*"
```

## Refactor Commit Handling

When commits use `refactor:` prefix (not conventional commit types), release-please won't create a release. Handle manually:

### Detection Script

```bash
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
COMMITS_SINCE=$(git log "${LATEST_TAG}"..HEAD --format="%s" 2>/dev/null || echo "")

if echo "$COMMITS_SINCE" | grep -qE "^refactor:"; then
  echo "Refactor commits detected — manual release needed"
fi
```

### Auto-create Release for Refactors

```bash
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
CURRENT=${LATEST_TAG#v}
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT"
NEW_PATCH=$((PATCH + 1))
NEW_TAG="v${MAJOR}.${MINOR}.${NEW_PATCH}"

gh release create "$NEW_TAG" \
  --title "$NEW_TAG" \
  --notes "Code refactoring changes" \
  --target master
```

### Workflow Integration

Add to release workflow after release-please step:

```yaml
- name: Create release for refactor commits
  if: ${{ steps.release.outputs.release_created != 'true' }}
  env:
    GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  run: |
    LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    COMMITS_SINCE=$(git log "${LATEST_TAG}"..HEAD --format="%s" 2>/dev/null || echo "")
    if echo "$COMMITS_SINCE" | grep -qE "^refactor:"; then
      CURRENT=${LATEST_TAG#v}
      IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT"
      NEW_PATCH=$((PATCH + 1))
      NEW_TAG="v${MAJOR}.${MINOR}.${NEW_PATCH}"
      gh release create "$NEW_TAG" \
        --title "$NEW_TAG" \
        --notes "Code refactoring changes" \
        --target master
    fi
```

## CI/CD Integration

### ko Build Trigger

If project uses `ko` for builds, the release can trigger image builds:

```yaml
# .github/workflows/deploy.yml (triggered on release)
on:
  release:
    types: [published]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: ko-build/setup-ko@v0.2
      - run: ko build --platform=linux/amd64,linux/arm64 --bare ./cmd/web/
```

### Deploy on Push (Alternative)

```yaml
on:
  push:
    branches: [master]

jobs:
  build-and-deploy:
    steps:
      - uses: actions/checkout@v4
      - name: Build and push
        run: ko build --platform=linux/amd64,linux/arm64 --bare ./cmd/web/

      - name: Deploy via SSH
        env:
          SSH_KEY: ${{ secrets.DEPLOY_SSH_KEY }}
        run: |
          echo "$SSH_KEY" > /tmp/deploy_key && chmod 600 /tmp/deploy_key
          ssh -i /tmp/deploy_key deploy@server 'cd /opt/app && docker compose up -d'
```

## Environment Variables Management

### CI/CD Secrets

**Never** put secrets in workflow files. Use GitHub Secrets:

```yaml
env:
  GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  SSH_KEY: ${{ secrets.DEPLOY_SSH_KEY }}
```

### Local Development

Use `dotenvx` for encrypted `.env` files (see `cali-ops-package-audit`):

```bash
# Run with env vars injected
dotenvx run -- go run ./cmd/web/

# Encrypt .env for safe commit
dotenvx encrypt

# Decrypt when needed
dotenvx decrypt
```

### Safe Patterns

| Pattern | Use case |
|---------|----------|
| `${{ secrets.X }}` | GitHub Actions secrets |
| `env: VAR: ${{ ... }}` | Pass context to bash safely |
| `jq` parsing | Extract values from JSON |
| `dotenvx` | Local dev encrypted envs |

## Emergency Hotfix Release

```bash
# 1. Create hotfix branch
git checkout -b hotfix/fix-critical main

# 2. Make fix, commit
git commit -m "fix: critical issue description"

# 3. Merge to main
git checkout main
git merge hotfix/fix-critical

# 4. Create release immediately
gh release create v0.X.Y-alpha \
  --title "v0.X.Y-alpha: Hotfix - description" \
  --notes "Critical fix for..."

# 5. Clean up
git branch -d hotfix/fix-critical
```

## Examples

### Example 1: release-please project

**Input:** "Pushed fix to main"

**Steps:**
1. Workflow triggers release-please
2. release-please creates PR with version bump
3. Workflow auto-merges PR
4. Tag created, release published

**Output:** "Release v0.24.8 created via release-please"

### Example 2: Manual release after feature merge

**Input:** "Just merged the auth feature to main"

**Steps:**
1. Run: `git tag --sort=-v:refname | head -3` → last tag was `v0.2.0-alpha`
2. Run: `gh release create v0.3.0-alpha --title "v0.3.0-alpha: Auth feature" --notes "..."`
3. Run: `git push origin v0.3.0-alpha`

**Output:** "Released v0.3.0-alpha with auth feature. Tag pushed."

### Example 3: Refactor commit release

**Input:** "Merged refactor commits, need release"

**Steps:**
1. Run: `git log v0.24.7..HEAD --oneline` → shows `refactor: ...` commits
2. Run: `gh release create v0.24.8 --title "v0.24.8" --notes "Code refactoring changes"`

**Output:** "Released v0.24.8 for refactor commits"

### Example 4: Emergency hotfix

**Input:** "Production is down, critical bug in auth"

**Steps:**
1. Run: `git checkout -b hotfix/fix-auth-crash main`
2. Make fix, commit
3. Merge to main
4. Run: `gh release create v0.2.1-alpha --title "v0.2.1-alpha: Hotfix auth crash" --notes "..."`

**Output:** "Hotfix released v0.2.1-alpha. Production should recover."

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
- If not logged in: `gh auth login` (interactive)
- Tell user they need to authenticate first

### Tag already exists
- Check: `git tag -l "v0.X.Y*"`
- If tag exists: "Tag v0.X.Y-alpha already exists. Use v0.X.Y.1-alpha or delete the tag first."

### release-please not creating releases
- Check: commit messages follow conventional commits (`feat:`, `fix:`, etc.)
- `refactor:` commits don't trigger releases — use manual flow
- Check `.release-please-manifest.json` exists

### JSON parsing errors in workflow
- Symptom: `Error reading JToken from JsonReader` or syntax errors
- Fix: Pass JSON via env var, parse with `jq`
- See: [Critical: Safe JSON Handling](#critical-safe-json-handling-in-bash)

## Test Cases

### Should activate
- "Create a release"
- "Push a release to main"
- "We merged to main, release now"
- "What version should this be?"
- "Emergency hotfix needed"
- "Release-please not working"

### Should NOT activate
- "Write release notes" (just writing, not releasing)
- "What's the current version?" (just asking, not releasing)
- "Run the tests" (testing, not releasing)

## References

- `references/ko-build.md` — ko build pipeline reference
- `cali-ops-package-audit` — dotenvx for env management
- `cali-ops-deploy-github-tailscale` — deploy pipeline patterns
