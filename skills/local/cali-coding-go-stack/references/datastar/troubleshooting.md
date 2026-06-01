# Datastar Troubleshooting

## Common Pitfalls

### 1. data-signals JSON Escaping

**❌ Don't:** Manual escaping with strings.ReplaceAll
```go
s = strings.ReplaceAll(s, "\n", "\\n")
s = strings.ReplaceAll(s, "'", "\\'")
```

**✅ Do:** Use json.Marshal
```go
func escapeForJS(s string) string {
    b, _ := json.Marshal(s)
    return string(b)
}
```

**Note:** `json.Marshal` returns the string with quotes included, so use `%s` directly in templates, not `'%s'`.

### 2. Form Data Not Sending

**❌ Don't:** Inputs without name attribute
```html
<input data-bind="model">
```

**✅ Do:** Always include name
```html
<input name="model" data-bind="model">
```

### 3. Textareas Not Syncing

**❌ Don't:** Only data-on:input
```html
<textarea data-on:input="$prompt = el.value">
```

**✅ Do:** Use data-bind
```html
<textarea name="prompt" data-bind="prompt">
```

### 4. Tabs Not Working

**❌ Don't:** Mix class="hidden" with data-show
```html
<div data-show="$tab === 'a'" class="hidden">
```

**✅ Do:** Use only data-show
```html
<div data-show="$tab === 'a'">
```

### 5. Loading State Stuck

**❌ Don't:** Forget to reset signal
```go
// Handler only saves, doesn't reset
repo.Save(data)
```

**✅ Do:** Reset signal after operation
```go
repo.Save(data)
sse.MarshalAndPatchSignals(map[string]bool{"isLoading": false})
```

## Error Messages

### "Invalid or unexpected token" in console

**Cause:** Malformed data-signals JSON
**Fix:**
1. Check json.Marshal escaping
2. View page source and validate JSON
3. Check for unescaped newlines/quotes

### "POST body empty"

**Cause:** Inputs missing name attribute
**Fix:** Add `name="fieldName"` to all form inputs

### "datastar not defined"

**Cause:** Script not loaded
**Fix:**
1. Check 404 for datastar.js
2. Verify static file serving
3. Check script path

### "Signals not updating"

**Cause:** Typically a data-bind or signal definition issue
**Fix:**
1. Ensure `data-bind` attribute is present on form elements
2. Check signal name matches between definition and usage
3. Verify server returns proper SSE response

### Handler never called

**Cause:** Wrong data-on:click syntax
**Fix:** Use `@post('/api/action')` format with quotes and leading slash

### Form data not received at backend

**Cause:** Missing `{contentType: 'form'}` in action
**Fix:** Use `@post('/api/action', {contentType: 'form'})` and call `r.ParseForm()` in handler

## Best Practices

1. **Self-contained features** — Each feature in its own directory
2. **Delta sync** — Send only what changed (not entire objects)
3. **KISS (SSE)** — Prefer SSE over WebSockets for one-way updates
4. **Indicators** — Always show loading states with `data-indicator`
5. **Error boundaries** — try/catch on media and storage operations
6. **View transitions** — Use `@view-transition { navigation: auto; }` for smooth navigation
