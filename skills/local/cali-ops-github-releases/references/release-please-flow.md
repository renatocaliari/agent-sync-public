# Release-Please Automated Flow

When project uses `googleapis/release-please-action`.

## Workflow Pattern

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
        env:
          GH_TOKEN: ${{ secrets.GH_PAT || secrets.GITHUB_TOKEN }}
          RELEASE_CREATED: ${{ steps.release.outputs.release_created }}
          RELEASE_PR: ${{ steps.release.outputs.pr }}
        run: |
          if [ -z "$RELEASE_PR" ] || [ "$RELEASE_PR" = "null" ]; then
            echo "No release PR created, skipping merge"
            exit 0
          fi
          PR_NUMBER=$(echo "$RELEASE_PR" | jq -r '.number // empty')
          if [ -n "$PR_NUMBER" ]; then
            echo "Merging release PR #$PR_NUMBER (release_created=$RELEASE_CREATED)"
            gh pr merge --merge "$PR_NUMBER"
          else
            echo "No release PR to merge"
          fi
```

## Critical: Safe JSON Handling in Bash

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

## Workflow Redundancy

**Rule:** One workflow per trigger type. Avoid duplicate workflows for same event.

**Check:** If you have both `release.yml` and `auto-merge-release.yml`:
- `release.yml` on `push` → handles release-please + auto-merge
- `auto-merge-release.yml` on `pull_request_target` → redundant if release.yml already merges

**Fix:** Remove redundant workflow. Keep the one that handles the full flow.

## Config File (`release-please-config.json`)

Minimal configuration:

```json
{
  "packages": {
    ".": {
      "release-type": "simple",
      "changelog-path": "CHANGELOG.md",
      "bump-minor-pre-major": true,
      "bump-patch-for-minor-pre-major": true,
      "include-v-in-tag": true
    }
  }
}
```

Common `release-type` values:
- `simple` — no language-specific logic
- `node` — Node.js (updates package.json)
- `go` — Go (updates version in source)
- `python` — Python (updates pyproject.toml)

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
