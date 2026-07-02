# Semantic Diff Enhancement for Releases

If `sem` is installed, enhance 3 release steps. If not, fall back to commit-prefix parsing.

Install: `brew install sem-cli` (macOS).

## Enhancement 1: Auto-detect bump type

`sem diff` sees structural changes (function signatures, class interfaces, removed exports) that commit message prefixes miss:

```bash
# sem diff between last tag and HEAD — structured output
sem diff "$LAST_TAG" HEAD --format json 2>/dev/null || true
```

How to interpret for version bump:
- Function **signatures changed** or **exports removed** → **major** bump
- New functions/classes/exports added, nothing removed → **minor** bump
- Only implementation changes (no API surface change) → **patch** bump
- This overrides commit-prefix parsing when `sem` is available

## Enhancement 2: Rich changelog

Replace `git log --oneline` with symbol-level changelog:

```bash
# Replace raw git log with structured diff summary
NOTES_FILE=$(mktemp)
cat > "$NOTES_FILE" << EOF
## Changes since $LAST_TAG

EOF

# Append sem summary if available
if command -v sem &>/dev/null; then
  echo '### Symbols changed' >> "$NOTES_FILE"
  sem diff "$LAST_TAG" HEAD 2>/dev/null >> "$NOTES_FILE" || true
  echo '' >> "$NOTES_FILE"
fi

echo '### Commits' >> "$NOTES_FILE"
CHANGES=$(git log "$LAST_TAG"..HEAD --oneline --no-decorate)
echo "$CHANGES" >> "$NOTES_FILE"

gh release create "$NEW_TAG" --title "$NEW_TAG" --notes-file "$NOTES_FILE"
rm -f "$NOTES_FILE"
```

## Enhancement 3: Pre-release sanity check

Before creating a release, confirm that actual code changes exist:

```bash
if command -v sem &>/dev/null; then
  DIFF_COUNT=$(sem diff "$LAST_TAG" HEAD --format json 2>/dev/null | jq '.changes | length' 2>/dev/null || echo "0")
  if [ "$DIFF_COUNT" = "0" ]; then
    echo "❌ sem diff shows 0 symbol changes since $LAST_TAG — nothing to release"
    exit 1
  fi
  echo "✅ sem reports $DIFF_COUNT symbol changes"
fi
```
