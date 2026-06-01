# Working with Existing Projects

## Migration Checklist

When adding this stack to an existing Go project:

- [ ] Add `go-chi/chi` as router
- [ ] Add Datastar JS to your HTML template
- [ ] Create `static/` directory for CSS/JS
- [ ] Add `go:embed` for static files
- [ ] Create `features/` directory structure
- [ ] Add NATS connection (if using real-time)
- [ ] Add SQLite/PocketBase (if using database)
- [ ] Update `Makefile` with dev/build commands

## Common Migration Issues

### "Cannot use templ components"
**Cause:** templ not installed or generated
**Fix:**
```bash
go install github.com/a-h/templ/cmd/templ@latest
templ generate
```

### "Static files not found"
**Cause:** go:embed path incorrect
**Fix:** Ensure `//go:embed static/*` is in your main.go

### "NATS connection refused"
**Cause:** NATS not running
**Fix:**
```bash
# Install NATS
go install github.com/nats-io/nats-server/v2@latest

# Run NATS
nats-server
```
