# lat.md CLI Reference

> **Source:** Installed via `npm install -g lat.md` or `lat init`.

## Commands

| Command | Purpose | When to use |
|---|---|---|
| `lat init` | Initialize `lat.md/` directory, install extensions and hooks | First time setting up a project |
| `lat search "query"` | Semantic search across all sections using embeddings | Before any task — find relevant context |
| `lat search "query" --limit N` | Limit results (default 5) | When too many results |
| `lat section "file#Heading"` | Show full content of a section with incoming/outgoing refs | Deep reading of a specific topic |
| `lat locate "name"` | Find a section by name (exact → stem → fuzzy) | When you know the name but not the file |
| `lat expand "text [[ref]]"` | Resolve `[[wiki links]]` in text to file locations | Before working with user prompts containing refs |
| `lat check` | Validate all wiki links and code refs | **Always** after modifying `lat.md/` or code |
| `lat refs "file#Heading"` | Find what references a section | Before renaming or deleting a section |
| `lat locate --help` | See locate options | When locate is not finding what you expect |

## Common Patterns

### Before Work (always)
```bash
lat search "<task description>"
lat expand "<user prompt with [[refs]]>"
```

### After Work (always)
```bash
# Update lat.md/ with changes
lat check
```

### Freshness Check
```bash
lat check  # fail CI or block agent_end if broken
```

### Code Refs Syntax

| Language | Syntax |
|---|---|
| TypeScript / JS | `// @lat: [[section-id]]` |
| Go | `// @lat: [[section-id]]` |
| Rust | `// @lat: [[section-id]]` |
| C | `// @lat: [[section-id]]` |
| Python | `# @lat: [[section-id]]` |
| HTML / Templ | `<!-- @lat: [[section-id]] -->` |
| Markdown | `[//]: # (@lat: [[section-id]])` |

## MCP Server (for Claude Code, Cursor, etc.)

Started automatically by `lat init` hooks. Provides 6 tools:
`lat_search`, `lat_section`, `lat_check`, `lat_expand`, `lat_refs`, `lat_locate`.

## Pi Extension

Started automatically by `lat init`. Registers tools and lifecycle hooks
(`before_agent_start` + `agent_end`) under `~/.pi/agent/extensions/`.
