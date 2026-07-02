# Troubleshooting

## "Invalid or unexpected token" in console

**Cause:** Malformed data-signals JSON
**Fix:**
1. Check json.Marshal escaping
2. View page source and validate JSON
3. Check for unescaped newlines/quotes

## "POST body empty"

**Cause:** Inputs missing name attribute
**Fix:** Add `name="fieldName"` to all form inputs

## "datastar not defined"

**Cause:** Script not loaded
**Fix:**
1. Check 404 for datastar.js
2. Verify static file serving
3. Check script path

## "Signals not updating"

**Cause:** data-bind or signal definition issue
**Fix:**
1. Ensure `data-bind` attribute is present on form elements
2. Check signal name matches between definition and usage
3. Verify server returns proper SSE response

## Handler never called

**Cause:** Wrong data-on:click syntax
**Fix:** Use `@post('/api/action')` format with quotes and leading slash

## Form data not received at backend

**Cause:** Missing `{contentType: 'form'}` in action
**Fix:** Use `@post('/api/action', {contentType: 'form'})` and call `r.ParseForm()` in handler
