# MANDATORY: Templ for ALL HTML Rendering

## NEVER Write HTML in Go Source Files

Every line of HTML in this project MUST be in a `.templ` file under `features/<name>/components/`. This is not optional — it prevents a class of bugs that have caused production issues.

### ❌ Anti-pattern (Blocked by CI)

```go
// NEVER do this:
html := fmt.Sprintf(`<div class="card">%s</div>`, value)
fmt.Fprint(w, html)

// NEVER do this with indexed arguments:
html := fmt.Sprintf(`<div data-signals='{"key": %[1]s}'>%[2]s</div>`, json1, html2)

// NEVER use fmt.Sprintf for any HTML fragment over 1 line:
barHTML := fmt.Sprintf(`<button class="%s" data-on:click="...">%s</button>`, class, label)
```

### ✅ Required Pattern

```templ
// features/myfeature/components/mything.templ
package myfeature

templ MyThing(value string) {
    <div class="card">{ value }</div>
}
```

```go
// Handler file — NO HTML here:
func (ns *Service) HandlePage(w http.ResponseWriter, r *http.Request) {
    data := loadData()
    myfeature.MyThing(data.Value).Render(r.Context(), w)
}
```

## CI Enforcement

Add this to your CI/pre-commit hook:

```bash
# Block fmt.Sprintf with HTML
grep -rn 'fmt\.Sprintf.*<' features/ --include="*.go" | grep -v '_test.go' && {
    echo "ERROR: fmt.Sprintf with HTML detected! Use templ components instead."
    echo "See: ~/.agents/skills/cali-boilerplate-go/references/MANDATORY_TEMPL_USAGE.md"
    exit 1
}

# Block percent-indexed arguments (%[N]s — known bug source)
grep -rn '%\[[0-9]' features/ --include="*.go" && {
    echo "ERROR: Indexed format specifiers (%[N]s) detected! These cause confusion bugs."
    exit 1
}
```

## FMT.Sprintf: When It IS Allowed

The ONLY exceptions:
1. **SSE fragment under 5 lines** with 0-1 interpolations, used once. If used 2+ times, extract to `.templ`.
2. **String formatting** that doesn't generate HTML tags (e.g., `fmt.Sprintf("Session %s", id)` for log messages).

## Related References

- [JS_VS_DATASTAR_DECISIONS.md](./JS_VS_DATASTAR_DECISIONS.md) — When to use Datastar vs JavaScript
- [toast.md](./datastar/toast.md) — Backend-driven toast patterns (zero JS) with animated entry/exit

## Why This Rule Exists

This project was bitten by:
- `%[N]s` index confusion (Go resets implicit counter after explicit indices → wrong values in wrong positions)
- LLM context confusion (HTML-in-Go-strings + numbered args = LLMs make errors)
- DRY violations (13 repeated `json.Marshal` + `escJSON` blocks in one file)
- Hard-to-test HTML (can't assert structure, only string equality)
