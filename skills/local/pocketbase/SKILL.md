---
name: pocketbase
description: PocketBase v0.39+ development - API rules, auth, collections, SDK, realtime, files, Go/JS extending, deployment, production tuning.
---

# PocketBase v0.39+ — Merged Reference

**Auto-invoke:** triggers on PocketBase, pb\_hooks, pb\_migrations topics. No external files — all rules inline.

**Target:** v0.39.1. Verified against current docs. Breaking changes noted inline.

---

## 1. Quick Start / CLI

```
./pocketbase serve          # dev on :8090
./pocketbase superuser create admin@x.com pass123
./pocketbase --dir=./mydata serve   # custom pb_data
./pocketbase --encryptionEnv=PB_ENCRYPTION_KEY serve
```

Routes: `:8090/` (static pb\_public), `:8090/_/` (admin), `:8090/api/` (REST).

---

## 2. Collection Design (CRITICAL)

### 2.1 Auth vs Base vs View

| Type | Use |
|------|-----|
| **auth** | User accounts, any entity that logs in (built-in password, OAuth2, OTP, MFA, token mgmt) |
| **base** | Regular data: posts, products, orders |
| **view** | Read-only SQL aggregations (must return `id` column) |

### 2.2 Field Types

| Field | When |
|-------|------|
| `text` | Strings, optional min/max/regex |
| `number` | Integers/floats, min/max |
| `bool` | true/false |
| `email` | Email validation |
| `url` | URL validation |
| `date` / `autodate` | Dates (auto on create/update) |
| `select` | Single/multi predefined values |
| `json` | Arbitrary JSON |
| `file` | Attachments, maxSelect/maxSize/mimeTypes/thumbs |
| `relation` | Ref to another collection, cascadeDelete |
| `editor` | Rich text HTML |
| `geopoint` | `{lon, lat}` — supports `geoDistance()` in filters |

### 2.3 Indexes

Index every field used in `filter`, `sort`, or API rules. Composite indexes order LTR.
```sql
CREATE INDEX idx_posts_author_status ON posts(author, status);
```

### 2.4 Relations & Cascade

- `cascadeDelete: true` for dependent data (comments on post)
- `cascadeDelete: false` for important data (orders, transactions) — blocks deletion
- Relation IDs auto-indexed

### 2.5 GeoPoint

```js
// Create
await pb.collection('places').create({
  location: { lon: -73.9857, lat: 40.7484 }  // lon FIRST
});
// Query nearby
await pb.collection('places').getList(1, 50, {
  filter: pb.filter('geoDistance(location, {:point}) <= {:km}', { point: { lon, lat }, km: 5 }),
  sort: pb.filter('geoDistance(location, {:point})', { point: { lon, lat } })
});
```

---

## 3. API Rules & Security (CRITICAL)

### 3.1 Rule Value Meanings

| Value | Access | Use |
|-------|--------|-----|
| `null` | Locked — superusers only | Admin data, system tables |
| `""` (empty) | Open to all | Public content |
| `"expression"` | Conditional | Ownership, role checks |

### 3.2 @request Fields

| Field | Use |
|-------|-----|
| `@request.auth.id` | Authed user ID (empty if guest) |
| `@request.auth.*` | Any auth record field (role, verified) |
| `@request.body.*` | Request body (create/update only) |
| `@request.query.*` | URL params (user-controlled — don't use for authz!) |
| `@request.context` | `default`, `oauth2`, `otp`, `password`, `realtime`, `protectedFile` |
| `@request.method` | HTTP method |

### 3.3 Body Field Modifiers

- `@request.body.field:isset` — true if field being sent
- `@request.body.field:changed` — true if changed from current
- `@request.body.field:length` — array/string length

### 3.4 Common Patterns

```js
// Owner-only
listRule: 'owner = @request.auth.id'
viewRule: 'owner = @request.auth.id'
createRule: '@request.auth.id != "" && @request.body.owner = @request.auth.id'
updateRule: 'owner = @request.auth.id && @request.body.owner:isset = false'
deleteRule: 'owner = @request.auth.id'

// Public read, authenticated write
listRule: ''  viewRule: ''  createRule: '@request.auth.id != ""'

// Role-based (@collection cross-lookup)
listRule: '@collection.team_members.user ?= @request.auth.id && @collection.team_members.team ?= team'

// Admin only
listRule: '@request.auth.role = "admin"'
```

### 3.5 Filter Functions

- `strftime(fmt, dt)` — date part extraction (v0.36+): `strftime('%Y-%m', created) = "2026-03"`
- `length(field)` — multi-value count
- `each(field, expr)` — iterate multi-value: `each(tags, ? ~ "urgent")`
- `issetIf(field, val)` — conditional presence check
- `geoDistance(geopoint, point)` — km distance

### 3.6 Error Codes by Rule

| Rule fail | HTTP |
|-----------|------|
| ListRule | 200 empty items |
| View/Update/DeleteRule | 404 (hides existence) |
| CreateRule | 400 |
| Locked rule | 403 |

---

## 4. Authentication (CRITICAL)

### 4.1 Password Auth

```js
await pb.collection('users').authWithPassword('email', 'password');
// Generic error on fail — never leak email existence
```

### 4.2 OTP (Email Code) — v0.36+

```js
const { otpId } = await pb.collection('users').requestOTP('user@example.com');
// Always returns otpId (even if email missing) — do not break enumeration protection
const authData = await pb.collection('users').authWithOTP(otpId, '12345678');
```

- Rate-limit `requestOTP` (built-in rate limiter, label `*:requestOTP`)
- OTP: 8 digits, 3min TTL
- `authWithOTP` consumes the code, invalidates otpId
- To use OTP without password: disable Password option, enable OTP

### 4.3 OAuth2

```js
// All-in-one (recommended for web)
await pb.collection('users').authWithOAuth2({ provider: 'google', createData: { name: '' } });

// Manual code exchange (React Native, deep links)
const methods = await pb.collection('users').listAuthMethods();
const provider = methods.oauth2.providers.find(p => p.name === 'google');
// Store provider.state and provider.codeVerifier, redirect to provider.authURL + redirectURL
// In callback: pb.collection('users').authWithOAuth2Code(provider.name, code, codeVerifier, redirectURL)
```

Redirect URL: `https://yourdomain.com/api/oauth2-redirect`

### 4.4 MFA (v0.23+)

```js
// Step 1: auth method A → catches 401 with mfaId
try { await pb.collection('users').authWithPassword('email', 'pass'); }
catch (err) {
  const mfaId = err.response?.mfaId;
  // Step 2: auth method B + mfaId
  await pb.collection('users').authWithOTP(otpId, '12345678', { mfaId });
}
```

### 4.5 Impersonation

```js
// Superuser generates token for another user
const client = await adminPb.collection('users').impersonate(userId, 3600);
// Non-renewable token! Generate fresh when expired.
```

### 4.6 Token Management

- `pb.authStore.isValid` — client-side (JWT expiry)
- `authRefresh()` — server verification, returns fresh token
- SSR: `pb.authStore.loadFromCookie(cookie)` / `exportToCookie({ httpOnly, secure })`
- AuthStore types: `LocalAuthStore` (browser), `AsyncAuthStore` (React Native), `BaseAuthStore` (custom)

---

## 5. SDK Usage (HIGH)

### 5.1 Client Init

```js
import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');
// Node.js: global.EventSource = EventSource from 'eventsource'
```

### 5.2 Auto-Cancellation

- SDK auto-cancels duplicate pending requests to same path
- Disable per-request: `{ requestKey: null }`
- Custom key: `{ requestKey: 'search' }`
- Cancel: `pb.cancelRequest('key')` / `pb.cancelAllRequests()`
- `autoCancellation(false)` globally disables

### 5.3 Error Handling

```js
import { ClientResponseError } from 'pocketbase';
try { ... }
catch (error) {
  if (error instanceof ClientResponseError) {
    error.status, error.response, error.isAbort;
    // Handle 400 (validation), 401, 403, 404, 429
  }
}
```

### 5.4 Field Modifiers (Atomic Updates)

| Modifier | Field Types | Effect |
|----------|-------------|--------|
| `field+` / `+field` | relation, file | Append/prepend |
| `field-` | relation, file | Remove |
| `field+` | number | Increment |
| `field-` | number | Decrement |

```js
await pb.collection('posts').update(id, { 'views+': 1, 'tags+': newTagId });
```

### 5.5 Safe Parameter Binding

```js
// ALWAYS use pb.filter() with params — never concatenate
pb.filter('title ~ {:q} && status = {:s}', { q: input, s: 'published' });
```

### 5.6 Send Hooks

```js
pb.beforeSend = (url, options) => { /* modify request */ return { url, options }; };
pb.afterSend = (response, data) => { /* process */ return data; };
```

---

## 6. Query Performance (HIGH)

### 6.1 Expand Relations

```js
// Single request, eliminates N+1
await pb.collection('posts').getList(1, 20, { expand: 'author,category,tags' });
// Nested: { expand: 'author.profile,category.parent' }
// Back-relation: { expand: 'comments_via_post,comments_via_post.author' }
// Back-relation syntax: {referencing_collection}_via_{relation_field}
```

### 6.2 Field Selection

```js
await pb.collection('posts').getList(1, 20, {
  fields: 'id,title,created,expand.author.name'
});
// Excerpt: content:excerpt(200,true) — truncate with ellipsis
```

### 6.3 getFirstListItem

```js
// Single record by any field
const user = await pb.collection('users').getFirstListItem(
  pb.filter('email = {:e}', { e: email })
);
// Throws 404 if not found
```

### 6.4 Batch Operations (Atomic)

```js
const batch = pb.createBatch();
batch.collection('orders').create({ ...order, id: orderId });
items.forEach(item => batch.collection('order_items').create({ ...item, order: orderId }));
const results = await batch.send(); // all or nothing
```

⚠ Must be enabled in Settings → Application first.

### 6.5 Pagination

```js
// Cursor-based (infinite scroll)
const posts = await pb.collection('posts').getList(1, 20, {
  skipTotal: true, sort: '-created'
});
const nextCursor = posts.items.at(-1)?.created;
```

### 6.6 Prevent N+1

- Use `expand` for relations
- Batch-fetch unique IDs with `Promise.all`
- Use view collections for complex joins

---

## 7. Realtime (MEDIUM)

### 7.1 Subscribe / Unsubscribe

```js
// Collection-wide
const unsub = await pb.collection('posts').subscribe('*', (e) => {
  e.action; // 'create' | 'update' | 'delete'
  e.record;
}, { expand: 'author', fields: 'id,title' });

// Specific record
await pb.collection('posts').subscribe('RECORD_ID', handler);

// Unsubscribe
unsub();                                // single callback
pb.collection('posts').unsubscribe();   // all in collection
pb.realtime.unsubscribe();              // all subscriptions
```

### 7.2 Auth

- Subscribe AFTER auth — subscription uses current auth context
- Re-subscribe on auth change (`pb.authStore.onChange`)
- ListRule checked for `*`, ViewRule for specific record

### 7.3 Connection Events

```js
pb.realtime.subscribe('PB_CONNECT', (e) => { /* connected, clientId: e.clientId */ });
pb.realtime.onDisconnect = (subscriptions) => { /* handle reconnect */ };
```

### 7.4 Performance

- Prefer `subscribe(recordId)` over `subscribe('*')` for high-traffic
- Subscription options: `filter`, `fields`, `expand` to reduce payload
- Server disconnects after 5min idle — SDK auto-reconnects

---

## 8. File Handling (MEDIUM)

### 8.1 URLs

```js
pb.files.getURL(record, record.image);
pb.files.getURL(record, record.image, { thumb: '100x100' });
// Protected files: get token first
const token = await pb.files.getToken();
pb.files.getURL(record, record.file, { token });
```

Thumbnail formats: `WxH` (fit), `Wx0` (fit width), `0xH` (fit height), `WxHt` (top crop), `WxHb` (bottom), `WxHf` (force).

### 8.2 Uploads

```js
// Plain object (auto-FormData)
await pb.collection('albums').create({ image: fileInput.files[0] });

// Multiple files
await pb.collection('albums').update(id, { images: files });

// Remove
await pb.collection('albums').update(id, { 'images-': filename });
await pb.collection('albums').update(id, { image: null }); // clear all
```

### 8.3 Server-Side Validation (Collection Settings)

```js
// Configure in Admin UI or via API:
{
  maxSelect: 1,
  maxSize: 5242880,  // 5MB
  mimeTypes: ['image/jpeg', 'image/png'],
  thumbs: ['100x100', '200x200']
}
```

### 8.4 S3 Optimization

- For public files, construct CDN URLs directly: `{cdnBase}/{collectionId}/{recordId}/{filename}`
- Protected files must go through PocketBase for token validation

---

## 9. Production & Deployment (LOW-MEDIUM)

### 9.1 Rate Limiting

```bash
./pocketbase serve --rateLimiter=true --rateLimiterRPS=10
```

Strategy (v0.36.7+): fixed-window. Worst case = 2x maxRequests at window boundary. Use Nginx/Caddy in front for defense in depth.

### 9.2 Reverse Proxy (Caddy)

```caddyfile
myapp.com {
  reverse_proxy 127.0.0.1:8090 { flush_interval -1 }
  header { X-Content-Type-Options "nosniff"; Strict-Transport-Security "..."; }
}
```

Nginx: proxy_buffering off, proxy_read_timeout 3600s (SSE).

### 9.3 Backups

```js
await adminPb.backups.create('daily-backup');
await adminPb.backups.getFullList();
await adminPb.backups.restore(key);
```

Automate with cron; store off-site (S3).

### 9.4 OS/Runtime Tuning

```bash
ulimit -n 4096          # open files for realtime connections
GOMEMLIMIT=512MiB       # Go GC memory cap in containers
PB_ENCRYPTION="..."     # 32-char key for _params encryption at rest
```

### 9.5 SMTP

```bash
export SMTP_HOST=smtp.sendgrid.net SMTP_PORT=587 SMTP_USER=... SMTP_PASS=...
```
Never ship `no-reply@example.com` — configure `Meta.senderAddress`.

### 9.6 SQLite Optimization

PocketBase defaults (auto-configured): WAL mode, busy_timeout=10s, cache_size=-32MB, foreign_keys=ON.

- Use batches for multiple writes (single transaction)
- Separate large content into dedicated collections
- Index all filtered/sorted fields
- Verify with `EXPLAIN QUERY PLAN`

---

## 10. Server-Side Extending (HIGH)

### 10.1 Go Extension Setup

Requires Go 1.25+ (v0.36.7+). No CGO by default.

```go
package main
import (
  "github.com/pocketbase/pocketbase"
  "github.com/pocketbase/pocketbase/core"
  "github.com/pocketbase/pocketbase/apis"
)
func main() {
  app := pocketbase.New()
  app.OnServe().BindFunc(func(se *core.ServeEvent) error {
    se.Router.GET("/api/myapp/hello/{name}", func(e *core.RequestEvent) error {
      return e.JSON(200, map[string]string{"msg": "hello " + e.Request.PathValue("name")})
    }).Bind(apis.RequireAuth())
    return se.Next()
  })
  if err := app.Start(); err != nil { log.Fatal(err) }
}
```

### 10.2 JSVM (pb_hooks)

```js
// pb_hooks/main.pb.js  — filename must end in .pb.js
/// <reference path="../pb_data/types.d.ts" />
routerAdd("GET", "/api/myapp/hello/{name}", (e) => {
  return e.json(200, { message: "Hello " + e.request.pathValue("name") });
}, $apis.requireAuth());
```

- Auto-reload on file change (UNIX)
- JS camelCase equivalents of Go methods
- Globals: `$app`, `$apis`, `$os`, `$security`, `$filesystem`, `$dbx`, `$mails`, `__hooks`
- ⚠ Variables outside handlers are undefined at runtime (handler serialization). Load via `require()` inside handler.

### 10.3 Module Loading (JSVM)

- Only CommonJS: `require()` works, ESM `import` does not
- `require` modules from `node_modules/` or `${__hooks}/` paths
- Bundle ESM packages to CJS first via rollup
- No `setTimeout`/`setInterval`, no Node.js `fs`, no `fetch` — use `$app.newHttpClient()`
- Pool of 15 JS runtimes by default (`--hooksPool=N`)

### 10.4 Event Hooks

**Always call `e.Next()` (Go) / `e.next()` (JS)** — skipping it silently breaks the chain.

```go
app.OnRecordAfterCreateSuccess("posts").BindFunc(func(e *core.RecordEvent) error {
  // Use e.App (not captured outer `app`) — inside a tx, e.App IS txApp
  return e.Next()
})
// Bind with Id for later Unbind:
app.OnRecordAfterCreateSuccess("posts").Bind(&hook.Handler{Id: "my-hook", Func: func(e *core.RecordEvent) error {
  return e.Next()
}})
app.OnRecordAfterCreateSuccess("posts").Unbind("my-hook")
```

### 10.5 Record Hook Families

| Family | Examples | Has request context? |
|--------|----------|---------------------|
| **Model** | `OnRecordCreate`, `OnRecord*Success` | No — any save path |
| **Request** | `OnRecord*Request`, `OnRecordsListRequest` | Yes — HTTP only |
| **Enrich** | `OnRecordEnrich` | Yes — response serialization + realtime SSE |

⚠ Use `OnRecordEnrich` to hide/redact fields — it applies to realtime events too. Model/request hooks alone leak.

### 10.6 Composition Flow — "Which App Am I Holding?"

The single most common extending bug: using the wrong app instance. Here is which one is active at each layer of a request.

| Where you are | Use | Why |
|---|---|---|
| Route handler (`func(e *core.RequestEvent)`) | `e.App` | Top-level app; same object server started with |
| Inside `RunInTransaction(func(txApp) { ... })` | **`txApp` only** | Capturing outer app deadlocks on SQLite writer lock |
| Inside a hook fired from a `Save` inside a tx | `e.App` | Framework rebinds `e.App` to `txApp` automatically |
| Inside a hook fired from a non-tx `Save` | `e.App` | Points to top-level app (no tx active) |
| Inside `OnRecordEnrich` | `e.App` | Runs **after** tx committed — no tx context |
| Cron callback | captured `app` / `se.App` | No per-run scoped app; wrap in `RunInTransaction` if atomicity needed |
| Migration function | the `app` argument | Already transactional |

**Error propagation:**
- `return err` inside `RunInTransaction` → rolls back everything, including audit records written by hooks fired from nested `Save` calls
- `return err` from a hook → propagates back through `Save` → rolls back the tx
- **Not calling `e.Next()`** → chain broken silently (no error, no realtime broadcast, no enrich pass)
- Panic inside tx closure → recovered by PB, rolls back, returns 500

### 10.7 Transactions

```go
err := app.RunInTransaction(func(txApp core.App) error {
  // Use ONLY txApp inside — outer app deadlocks (writer lock)
  txApp.Save(record1)
  txApp.Save(record2)
  return nil // commit
  // return err // rollback
})
```

- `e.App` inside hooks is the transactional app when inside a tx
- Keep txs short — no HTTP calls or email sends inside

### 10.8 Route Middlewares

```go
g := se.Router.Group("/api/myapp")
g.Bind(apis.RequireAuth())
g.Bind(apis.Gzip())
g.POST("/admin/rebuild", handler).Bind(apis.RequireSuperuserAuth())
```

JS: `routerUse("/api/myapp", $apis.requireAuth())`

| Middleware | Use |
|---|---|
| `RequireGuestOnly()` | Reject authed clients |
| `RequireAuth(...collections)` | Require auth (opt. restrict to specific auth coll) |
| `RequireSuperuserAuth()` | Superuser only |
| `RequireSuperuserOrOwnerAuth("id")` | Superuser or matching path param |
| `Gzip()` | Compress responses |
| `BodyLimit(bytes)` | Override 32MB default |
| `SkipSuccessActivityLog()` | Suppress activity log |

### 10.9 Custom SQLite Driver

Default: pure-Go `modernc.org/sqlite` (no CGO). Use `DBConnect` for FTS5/ICU/extensions.

⚠ `DBConnect` called twice: `data.db` + `auxiliary.db`.

**Option A: ncruces/go-sqlite3 (recommended, no CGO)** — pure Go via WASM→Go, suporta ext/unicode, ext/vec1, ext/spellfix1, ext/regexp, FTS5 nativo.

```go
import (
    _ "github.com/ncruces/go-sqlite3/driver"
    "github.com/ncruces/go-sqlite3"
    "github.com/ncruces/go-sqlite3/ext/unicode"  // unaccent(), collation
    // "github.com/ncruces/go-sqlite3/ext/vec1"  // busca vetorial
    // "github.com/ncruces/go-sqlite3/ext/spellfix1" // fuzzy matching
)

func init() {
    // Auto-registra extensions em toda nova conexão
    sqlite3.AutoExtension(unicode.Register)
}

app := pocketbase.NewWithConfig(pocketbase.Config{
    DBConnect: func(dbPath string) (*dbx.DB, error) {
        return dbx.Open("sqlite3", "file:"+dbPath+";_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=journal_size_limit(200000000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)&_pragma=temp_store(MEMORY)&_pragma=cache_size(-32000)")
    },
})
```

**Option B: mattn/go-sqlite3 (CGO)** — requer `CGO_ENABLED=1`, perde cross-compilation estática.

```go
import _ "github.com/mattn/go-sqlite3" // CGO
app := pocketbase.NewWithConfig(pocketbase.Config{
  DBConnect: func(dbPath string) (*dbx.DB, error) {
    return dbx.Open("pb_sqlite3", dbPath)
  },
})
```

**Trade-offs ncruces vs modernc:**
- ✅ GANHA: unaccent(), collation pt_BR, vec1, spellfix1, regexp SQL, large reads 2x mais rápido
- ❌ PERDE: insert bulk ~12% mais lento (irrelevante para inserts 1x1)
- ❌ MEMÓRIA: ~2-3x por conexão (sandbox Wasm)
- ❌ BINÁRIO: +4MB
- FTS5 nativo no WASM base (sem registro extra)

### 10.10 Cron Jobs

```go
app.Cron().MustAdd("cleanup-drafts", "0 3 * * *", func() {
  app.Logger().Info("cleaning drafts...")
})
// Don't use __pb*__ prefix (reserved for system jobs)
```

JS: `cronAdd("id", "*/30 * * * *", () => { ... })`

### 10.11 Filesystem Handle

❌ Leaking `NewFilesystem()` leaks S3 connections.
✅ Always `defer fs.Close()` (Go) / `try { ... } finally { fs.close() }` (JS).

```go
fs, err := app.NewFilesystem()
if err != nil { return err }
defer fs.Close()
```

Prefer record field API (`record.Set("file", f)` + `app.Save`) over direct `fs.Upload`.

### 10.12 Server-Side Filter Binding

```go
// ❌ String interpolation: filter injection
// ✅ Placeholder binding:
record, err := app.FindFirstRecordByFilter("users", "email = {:email} && verified = true",
  dbx.Params{"email": email})
```

JS: `$app.findFirstRecordByFilter("users", "email = {:e}", { e: email })`

### 10.13 Migrations

**Go:** `m.Register(upFn, downFn)` in `migrations/` package. Auto-discovered when `migratecmd.MustRegister(app, ...)` is called in main.

**JS (pb\_migrations):**
```js
// pb_migrations/1712500000_add_collection.js
migrate((app) => {
  const col = new Collection({ type: "base", name: "audit", fields: [...] });
  app.save(col);
}, (app) => {
  const col = app.findCollectionByNameOrId("audit");
  app.delete(col);
});
```

- Filename format: `<unix>_<description>.js`
- `app` in up/down is transactional
- Use collection API, not raw SQL (cache invalidation)
- Commit `pb_migrations/`, never `pb_data/`

### 10.14 Sending Email

```go
meta := e.App.Settings().Meta
from := mail.Address{Name: meta.SenderName, Address: meta.SenderAddress}
msg := &pbmail.Message{From: from, To: []mail.Address{{Address: to}}, Subject: "...", HTML: "..."}
if err := e.App.NewMailClient().Send(msg); err != nil {
  e.App.Logger().Error("email failed", "err", err)
}
// Don't return email errors from hooks — don't roll back business txn
```

### 10.15 Testing

```go
app, _ := tests.NewTestApp("test_pb_data")
defer app.Cleanup()

tests.ApiScenario{
  Name: "create post",
  Method: http.MethodPost,
  URL: "/api/collections/posts/records",
  Body: strings.NewReader(`{"title":"hi"}`),
  ExpectedStatus: 200,
  ExpectedEvents: map[string]int{
    "OnRecordCreateRequest": 1,
    "OnRecordAfterCreateSuccess": 1,
  },
  TestAppFactory: func(t testing.TB) *tests.TestApp { return app },
}.Test(t)
```

### 10.16 Settings & Encryption

- Read at call time: `app.Settings()`, never capture at startup
- Mutate: `settings := app.Settings()` → `app.Save(settings)`
- Encrypt at rest: `export PB_ENCRYPTION="32-char-exactly"` — AES encrypts `_params`
- Losing the key = unrecoverable
- `OnSettingsReload` hook for in-memory cache invalidation

---

## 11. Schema Templates

### 11.1 Blog

**posts** (base): title(text), slug(text,unique), content(text), excerpt(text), featured_image(file), author(→users), category(→categories), tags(→tags). Rules: public read, author write.

**categories** (base): name(text,unique), slug(text,unique), description(text).

**tags** (base): name(text,unique).

### 11.2 E-commerce

**products** (base): name(text), slug(unique), description(editor), price(number), compare_price(number), sku(text,unique), stock(number), images(file,many), categories(→categories,many). Rules: public read, admin write.

**orders** (base): orderNumber(text,unique), status(select: pending/paid/shipped/delivered/cancelled), customer(→users), items(json), total(number,required), shippingAddress(editor), notes(editor). Rules: owner view, admin manage.

**reviews** (base): product(→products), author(→users), rating(number,1-5), content(text). Rules: author write own, public read.

### 11.3 Social Network

**profiles** (auth): extends users. displayName(text), bio(text), avatar(file). Rules: owner edit, public view.

**posts** (base): content(editor), author(→profiles), media(file,many), likes(json). Rules: public read, auth write.

**comments** (base): post(→posts,cascadeDelete), author(→profiles), content(text). Rules: public read, auth create, author edit.

**follows** (base): follower(→profiles), following(→profiles), unique index on pair. Rules: auth create own.

### 11.4 Task Management

**projects** (base): name(text), description(text), owner(→users), members(→users,many), status(select: active/archived). Rules: member view, owner manage.

**tasks** (base): title(text), description(editor), status(select: todo/in_progress/done), priority(select: low/medium/high/critical), assignee(→users), project(→projects,cascadeDelete), dueDate(date). Rules: project member view, assignee/owner update.

### 11.5 Forum

**categories** (base): name(text,unique), description(text), sortOrder(number). Rules: public read, admin manage.

**threads** (base): title(text), content(editor), author(→users), category(→categories,cascadeDelete), pinned(bool), views(number). Rules: public read, auth create, author edit.

**posts** (base): content(editor), author(→users), thread(→threads,cascadeDelete). Rules: public read, auth create, author edit.

---

## 12. Security Rules Reference

### 12.1 Common Rule Builder

| Pattern | listRule | viewRule | createRule | updateRule | deleteRule |
|---------|----------|----------|------------|------------|------------|
| Public read, auth write | `""` | `""` | `@request.auth.id != ""` | *owner check* | *owner check* |
| Owner only | `owner = @request.auth.id` | same | `@request.auth.id != "" && @request.body.owner = @request.auth.id` | `owner = @request.auth.id` | `owner = @request.auth.id` |
| Auth users | `@request.auth.id != ""` | same | same | same | same |
| Admin only | `@request.auth.role = "admin"` | same | same | same | same |
| Public read-only | `""` | `""` | `null` | `null` | `null` |

### 12.2 Role-Based Access via @collection

```js
// roles collection: { user(→users), role(text) }
// resource rules:
listRule: '@collection.roles.user = @request.auth.id && @collection.roles.role = "admin"'
```

### 12.3 Field-Level Security

- `Hide("fieldName")` — in `OnRecordEnrich` hook (Go) or `e.record.hide()` (JS)
- `@request.body.field:isset = false` — prevent field changes via API rules

### 12.4 File Security

- Protected files require `pb.files.getToken()` — valid for limited time
- Superusers bypass all rules
- Realtime subscriptions respect ListRule/ViewRule

---

## 13. Dart SDK Quick Ref

```dart
final pb = PocketBase('http://127.0.0.1:8090');
await pb.collection('users').authWithPassword('email', 'pass');
final records = await pb.collection('posts').getList(page: 1, perPage: 20);
final unsub = await pb.collection('posts').subscribe('*', (e) { print(e.record); });
```

Same API shape as JS SDK. Use `AsyncAuthStore` for Flutter.

---

## 14. Data Migration Workflows

### 14.1 Import/Export

```bash
# Export all collections schema as JSON
# (Admin UI > Settings > Export)
# Import via Admin UI or API
```

### 14.2 Scripted Migration (JS)

```js
// pb_migrations/1712500000_seed_data.js
migrate((app) => {
  const col = app.findCollectionByNameOrId("products");
  const data = JSON.parse($os.readFile(`${__hooks}/seed/products.json`));
  data.forEach(item => {
    const r = new Record(col);
    Object.entries(item).forEach(([k, v]) => r.set(k, v));
    app.save(r);
  });
}, (app) => {
  // reverse
});
```

---

## Notes

- **v0.38+**: Superuser CIDR whitelist (Settings > Application > Superuser IPs)
- **v0.38.1**: Auth state force-cleared on password/collection secret change
- **v0.39+**: SQL console (Settings > Debug), system email alerts for backup errors
- Min Go version: 1.26.4 (v0.39.1)
- Prebuilt executable uses pure-Go SQLite (no CGO)
- Always use `pb.filter()` with named params — never string concatenation
- New collections: start locked (null), open explicitly
- For production: HTTPS + rate limiter + backups + encryption key + ulimit
