// gorules — Ruleguard rules for Go + Datastar projects.
//
// Add this file to your project as rules/gorules.go and configure
// golangci-lint to use it:
//
//   linters-settings:
//     gocritic:
//       enabled-tags:
//         - ruleguard
//       settings:
//         ruleguard:
//           rules: "rules/gorules.go"
//
// Install: go install github.com/quasilyte/go-ruleguard@latest
//
// These rules catch patterns that LLMs systematically get wrong
// and that standard linters (govet, staticcheck, errcheck) miss.
package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

// ----------------------------------------------------------------
//  GENERAL GO MISTAKES
// ----------------------------------------------------------------

// contextBackgroundInHandler detects context.Background() used inside
// HTTP handler functions or anywhere net/http is imported. Use
// r.Context() instead — avoids cancellation leaks.
func contextBackgroundInHandler(m dsl.Matcher) {
	m.Match(`context.Background()`).
		Where(m.File().PkgPath.Matches(`.*`) && m.File().Imports(`net/http`)).
		Report(`use r.Context() instead of context.Background() inside HTTP handlers — ` +
			`context.Background() never cancels, leaking goroutines on client disconnect`).
		Suggest(`r.Context()`)
}

// nilContextInCall detects nil passed as the first context.Context arg
// of known ctx-accepting APIs in this stack (goqite jobs, http.WithContext).
// Causes silent nil-pointer panic in libs like database/sql,
// goqite.InTx, etc. — often masked by panicRecover middleware.
//
// Rule: ALWAYS pass context.Background() or a derived context. NEVER nil.
// To extend: add more m.Match patterns for ctx-accepting APIs your project uses.
func nilContextInCall(m dsl.Matcher) {
	m.Match(`jobs.Enqueue(nil, $_)`).
		Report(`jobs.Enqueue does not accept nil context — use context.Background()`).
		Suggest(`context.Background()`)

	m.Match(`goqite/jobs.Create(nil, $_, $_, $_)`).
		Report(`goqite/jobs.Create does not accept nil context — use context.Background()`).
		Suggest(`context.Background()`)

	m.Match(`http.NewRequestWithContext(nil, $_, $_, $_)`).
		Report(`http.NewRequestWithContext does not accept nil context — use context.Background()`).
		Suggest(`context.Background()`)
}

// loopVariableCapturedInGoroutine detects classic Go pitfall where
// loop variable is captured by reference in a goroutine closure.
func loopVariableCapturedInGoroutine(m dsl.Matcher) {
	m.Match(`for $_, $v := range $_ { go func($_{}) { $*_($v) }($*_) }`).
		Report(`loop variable '$v' captured in goroutine — value will be last iteration's value ` +
			`when goroutine executes. Pass as argument: go func(v $type) { ... }($v)`)

	m.Match(`for $_, $v := range $_ { go func() { $*_($v) }() }`).
		Report(`loop variable '$v' captured by reference in goroutine — use parameter passing`)
}

// mutexCopiedByValue detects sync.Mutex passed by value (copied).
// Mutex must always be passed by pointer.
func mutexCopiedByValue(m dsl.Matcher) {
	m.Match(`func $_(m $typ) $*_{
		$*_
	}`).
		Where(m["typ"].Type.Implements(`sync.Locker`)).
		Report(`'$typ' is passed by value — mutex copy breaks synchronization. Use *$typ`)

	m.Match(`type $t struct {
		$*_
		$f $typ
		$*_
	}`).
		Where(m["typ"].Type.Is(`sync.Mutex`) || m["typ"].Type.Is(`sync.RWMutex`)).
		Report(`'$f' is '$typ' embedded by value — if '$t' is copied, mutex is copied. ` +
			`Prefer *$typ or ensure $t is always used by pointer`)
}

// directMutexCopy detects mutex copy via struct assignment or initialization.
func directMutexCopy(m dsl.Matcher) {
	m.Match(`$x = $y`).
		Where(m["x"].Type.Is(`sync.Mutex`) || m["x"].Type.Is(`sync.RWMutex`)).
		Report(`mutex value copy via assignment — '$x' is a '${typ}' which should not be copied`).
		At(m["x"])

	m.Match(`$x := $y`).
		Where(m["x"].Type.Is(`sync.Mutex`) || m["x"].Type.Is(`sync.RWMutex`)).
		Report(`mutex copy via short variable declaration — '$x' is a '${typ}' which should not be copied. ` +
			`Use &$y{ } or embed as *sync.Mutex`).
		At(m["x"])
}

// listenAndServeNoErr detects http.ListenAndServe without error check.
func listenAndServeNoErr(m dsl.Matcher) {
	m.Match(`http.ListenAndServe($*_)`).
		Where(m["_"].Const).
		Report(`http.ListenAndServe error is ignored — server may fail silently. ` +
			`Always check: log.Fatal(http.ListenAndServe(...))`)
}

// errGroupNoWait detects errgroup.Group created but Wait() never called.
func errGroupNoWait(m dsl.Matcher) {
	m.Match(`$g, $ctx := errgroup.WithContext($_)`).
		Where(m.File().Imports(`golang.org/x/sync/errgroup`)).
		Report(`errgroup.Group created but g.Wait() may not be called — ` +
			`goroutines leak if Wait() is not invoked`)
}

// jsonUnmarshalWithoutCheck detects json.Unmarshal where error is explicitly
// discarded with '_'. errcheck already catches unchecked errors in general;
// this rule provides a more specific message for json.Unmarshal.
func jsonUnmarshalWithoutCheck(m dsl.Matcher) {
	m.Match(`$_, _ := json.Unmarshal($_, $_)`).
		Report(`json.Unmarshal error explicitly discarded with '_' — ` +
			`invalid input leaves target uninitialized. ` +
			`Always check: if err := json.Unmarshal(...); err != nil { ... }`)
}

// strconvWithoutErr detects strconv.Atoi/ParseInt/ParseFloat where
// error is explicitly discarded with '_'. errcheck already catches
// unchecked errors in general; this rule adds specific messaging.
func strconvWithoutErr(m dsl.Matcher) {
	m.Match(`$_, _ := strconv.Atoi($_)`).
		Report(`strconv.Atoi error explicitly discarded with '_' — ` +
			`invalid input returns 0 silently. Check error or use comma-ok`)

	m.Match(`$_, _ := strconv.ParseInt($*_)`).
		Report(`strconv.ParseInt error explicitly discarded with '_'`)

	m.Match(`$_, _ := strconv.ParseFloat($*_)`).
		Report(`strconv.ParseFloat error explicitly discarded with '_'`)
}

// deferCloseNoCheck detects deferred Close() calls without error handling.
func deferCloseNoCheck(m dsl.Matcher) {
	m.Match(`defer $x.Close()`).
		Where(m["x"].Type.Is(`io.Closer`)).
		Report(`deferred $x.Close() error is ignored — use defer func() { ` +
			`if err := $x.Close(); err != nil { log.Error(err) } }()`)

	m.Match(`defer resp.Body.Close()`).
		Report(`deferred resp.Body.Close() without error check — wrap in closure: ` +
			`defer func() { _ = resp.Body.Close() }()`)
}

// httpHandlerWrongSignature detects incorrect HTTP handler signatures.
func httpHandlerWrongSignature(m dsl.Matcher) {
	m.Match(`http.HandleFunc($_, func(w http.ResponseWriter, r *http.Request) $*_)`).
		Report(`http.HandleFunc handler has wrong parameter order — must be ` +
			`func(w http.ResponseWriter, r *http.Request)`)
}

// timeSleepInTest detects time.Sleep in tests (flaky). Use
// require.Eventually or clock mocking.
func timeSleepInTest(m dsl.Matcher) {
	m.Match(`time.Sleep($_)`).
		Where(m.File().Name().Matches(`_test.go`)).
		Report(`time.Sleep in test creates flakiness — use ` +
			`require.Eventually, retry with backoff, or clock mocking`)
}

// ------------------------------nested json anonymous struct
func noDeferInLoop(m dsl.Matcher) {
	m.Match(`for $*_ {
		$*_(${defer}(*_))
		$*_
	}`).
		Where(m["defer"].Text == "defer").
		Report(`defer inside a loop — deferred calls accumulate until function returns. ` +
			`Use a closure or move the operation outside the loop`)
}

// ----------------------------------------------------------------
//  DATASTAR SDK MISTAKES — Go backend patterns
// ----------------------------------------------------------------

// newSSEBeforeReadSignals detects potential body consumption order.
// Datastar's NewSSE reads headers but not the body. ReadSignals reads
// the body for POST/PUT/PATCH requests and the query string for GET/DELETE.
// The order should be: NewSSE first, then ReadSignals (or vice versa
// with caution). Actually ReadSignals reads r.Body, which can only be
// read once. NewSSE doesn't read the body. So ReadSignals must be called
// BEFORE any other body reading, but NewSSE can be created before.
// This rule warns when ReadSignals is called after something that might
// have consumed the body.
func readSignalsAfterBodyRead(m dsl.Matcher) {
	m.Match(`$*_

		$*_(${func}(*_))
		$*_
		datastar.ReadSignals($_, $_)
		$*_
	`).
		Where(m["func"].Text.Matches(`(json\.Decode|io\.ReadAll|ioutil\.ReadAll|r\.Body\.|fmt\.Fscan)`)).
		Report(`datastar.ReadSignals called AFTER $func may fail — ` +
			`r.Body can only be read once. Call ReadSignals FIRST`)

	m.Match(`$*_
		io.ReadAll($_)
		$*_
		datastar.ReadSignals($_, $_)
	`).
		Report(`io.ReadAll may consume r.Body before datastar.ReadSignals — ` +
			`reorder to call ReadSignals first`)
}

// parseFormMissing detects r.FormValue() called without prior r.ParseForm()
// when contentType: 'form' is used.
func parseFormMissing(m dsl.Matcher) {
	m.Match(`$*_
		$_{r}.FormValue($_)
		$*_
	`).
		Where(m["r"].Type.Is(`*http.Request`) &&
			!m.File().Text.Matches(`(?m).*ParseForm\b.*`)).
		Report(`r.FormValue() used but r.ParseForm() is not called in this file — ` +
			`form values are empty without r.ParseForm(). ` +
			`Add r.ParseForm() before calling r.FormValue()`)
}

// readSignalsErrorIgnored detects datastar.ReadSignals error not checked.
// Unlike errcheck which catches all unchecked errors, this rule provides
// a Datastar-specific message explaining why the error matters.
func readSignalsErrorIgnored(m dsl.Matcher) {
	m.Match(`$_, _ := datastar.ReadSignals($_, $_)`).
		Report(`datastar.ReadSignals error discarded with '_' — invalid JSON signals ` +
			`silently produce zero-valued struct. Handle the error: ` +
			`if err := datastar.ReadSignals(r, &store); err != nil { return err }`)

	m.Match(`datastar.ReadSignals($_, $_)`).
		Where(m["_"].Text != "").
		Report(`datastar.ReadSignals return value (error) must be checked — ` +
			`invalid JSON signals silently produce empty struct`)
}

// newSSECalledTwice detects datastar.NewSSE called more than once in
// the same function. Each request should create exactly one SSE generator.
func newSSECalledTwice(m dsl.Matcher) {
	// Match first occurrence.
	m.Match(`$sse1 := datastar.NewSSE($*_)
		$*_
		$sse2 := datastar.NewSSE($*_)`).
		Where(m["sse1"].Type.Is(`*datastar.ServerSentEventGenerator`) ||
			m["sse2"].Type.Is(`*datastar.ServerSentEventGenerator`)).
		Report(`datastar.NewSSE called twice in the same handler — ` +
			`only one SSE connection per request is needed. ` +
			`Reuse the first *ServerSentEventGenerator`)
}

// marshalAndPatchSignalsPreferred detects manual json.Marshal + PatchSignals
// where MarshalAndPatchSignals should be used instead.
func marshalAndPatchSignalsPreferred(m dsl.Matcher) {
	m.Match(`$b, err := json.Marshal($_); $*_ $sse.PatchSignals($b, $*_)`).
		Where(m["sse"].Type.Is(`*datastar.ServerSentEventGenerator`)).
		Report(`use sse.MarshalAndPatchSignals() instead of manual json.Marshal + PatchSignals`).
		Suggest(`sse.MarshalAndPatchSignals(${"signals"})`)
}

// marshalAndPatchSignalsPreferredAlt catches the two-line form.
func marshalAndPatchSignalsPreferredAlt(m dsl.Matcher) {
	m.Match(`$b := $_; $sse.PatchSignals($b)`).
		Where(m["sse"].Type.Is(`*datastar.ServerSentEventGenerator`)).
		Report(`use sse.MarshalAndPatchSignals() instead of manual marshaling`)
}

// patchElementsErrorIgnored checks sse.PatchElements() error is checked.
// (Complement to errcheck which catches the basic case — this catches
// when error is explicitly discarded with _)
func patchElementsErrorDiscarded(m dsl.Matcher) {
	m.Match(`_ = $sse.PatchElements($*_)`).
		Where(m["sse"].Type.Is(`*datastar.ServerSentEventGenerator`)).
		Report(`error from sse.PatchElements() is discarded — client may not receive update. ` +
			`At minimum log the error`)

	m.Match(`$sse.PatchElements($*_)`).
		Where(m["sse"].Type.Is(`*datastar.ServerSentEventGenerator`) &&
			m["_"].Text != "").
		Report(`sse.PatchElements() error must be checked — client may not receive update`)
}

// // Using http.Error after NewSSE — headers already flushed.
// // ruleguard has limited flow analysis so this catches simple cases.
func httpErrorAfterNewSSE(m dsl.Matcher) {
	m.Match(`$*_
		$sse := datastar.NewSSE($*_)
		$*_
		http.Error($*_)
	`).
		Report(`http.Error() called after datastar.NewSSE() — headers already sent (flushed). ` +
			`Use sse.MarshalAndPatchSignals({"error": msg}) for error reporting instead`)
}

// patchElementTemplPreferred detects strings.Builder + Render manual
// pattern when PatchElementTempl should be used instead.
// Catches both `var buf strings.Builder` and `buf := strings.Builder{}`.
func patchElementTemplPreferred(m dsl.Matcher) {
	m.Match(`var $buf strings.Builder
		$comp.Render($ctx, &$buf)
		$sse.PatchElements($buf.String(), $*_)`).
		Where(m["sse"].Type.Is(`*datastar.ServerSentEventGenerator`)).
		Report(`use sse.PatchElementTempl(comp, opts...) instead of manual Render + PatchElements. ` +
			`It's more efficient (bytebufferpool) and inherits sse.Context()`).
		Suggest(`sse.PatchElementTempl($comp, $*_)`)

	m.Match(`$buf := strings.Builder{}
		$comp.Render($ctx, &$buf)
		$sse.PatchElements($buf.String(), $*_)`).
		Where(m["sse"].Type.Is(`*datastar.ServerSentEventGenerator`)).
		Report(`use sse.PatchElementTempl(comp, opts...) instead of manual Render + PatchElements`).
		Suggest(`sse.PatchElementTempl($comp, $*_)`)
}

// // isClosedNotChecked detects expensive operations without checking
// // if the SSE connection is still active.
func isClosedNotChecked(m dsl.Matcher) {
	m.Match(`$*_
		$sse := datastar.NewSSE($*_)
		$*_
		$_ = $fn($*_) // expensive call
		$*_
		$sse.MarshalAndPatchSignals($*_)
	`).
		Where(m["sse"].Type.Is(`*datastar.ServerSentEventGenerator`) &&
			m["fn"].Text.Matches(`(?i)(repo\.|db\.|api\.|llm\.|chat\.)`)).
		Report(`expensive call after NewSSE without checking sse.IsClosed() — ` +
			`client may have disconnected. Guard with: if sse.IsClosed() { return }`)
}

// // usingStringBuilderForTempl detects fmt.Sprintf with HTML tags in Go code.
// The skill already catches this via grep, but ruleguard can catch it at
// the Go AST level for function calls specifically.
func fmtSprintfHTML(m dsl.Matcher) {
	m.Match(`fmt.Sprintf($format, $*_)`).
		Where(m["format"].Text.Matches(`.*<.*>`)).
		Report(`fmt.Sprintf with HTML tag detected! Use a templ component instead. ` +
			`See: cali-coding-go-stack references`)
}

// // contextBackroundMisuseInDatastarHandler catches context.Background
// // when datastar.NewSSE is used — should use sse.Context().
func contextBackgroundInDatastarHandler(m dsl.Matcher) {
	m.Match(`$*_
		$sse := datastar.NewSSE($*_)
		$*_
		context.Background()
	`).
		Where(m["sse"].Type.Is(`*datastar.ServerSentEventGenerator`)).
		Report(`use sse.Context() instead of context.Background() in Datastar handler — ` +
			`sse.Context() is derived from r.Context() and cancels on client disconnect`)
}

// // removeElementsMissingMode detects RemoveElement used but then
// // PatchElements with remove mode — use RemoveElement helper.
func removeElementHelperAvailable(m dsl.Matcher) {
	m.Match(`$sse.PatchElements($_, datastar.WithModeRemove(), $*_)`).
		Where(m["sse"].Type.Is(`*datastar.ServerSentEventGenerator`)).
		Report(`use sse.RemoveElement(selector) or sse.RemoveElementByID(id) instead of ` +
			`PatchElements with WithModeRemove`)
}

// // consoleLogInsteadOfSDK detects console.log in ExecuteScript pattern.
func consoleLogInsteadOfSDK(m dsl.Matcher) {
	m.Match(`$sse.ExecuteScript(${script})`).
		Where(m["sse"].Type.Is(`*datastar.ServerSentEventGenerator`) &&
			m["script"].Text.Matches(`.*console\.log\b.*`)).
		Report(`use sse.ConsoleLog() or sse.ConsoleError() instead of ExecuteScript with console.log`)
}

// // redirectInsteadOfExecuteScript detects window.location in ExecuteScript.
func redirectInsteadOfExecuteScript(m dsl.Matcher) {
	m.Match(`$sse.ExecuteScript(${script})`).
		Where(m["sse"].Type.Is(`*datastar.ServerSentEventGenerator`) &&
			m["script"].Text.Matches(`.*window\.location.*`)).
		Report(`use sse.Redirect() or sse.Redirectf() instead of ExecuteScript with window.location`)
}

// // preferPostSSE detects @get used for mutations — should use @post/@put/@patch/@delete.
// // This is a frontend pattern, but the Go SDK generates the action strings.
func postSSEForMutation(m dsl.Matcher) {
	m.Match(`datastar.GetSSE($_)`).
		Where(m["_"].Text.Matches(`.*(save|create|update|delete|remove|submit|send|toggle).*`)).
		Report(`datastar.GetSSE() with mutation-like URL — GET should be idempotent. ` +
			`Use PostSSE/PutSSE/PatchSSE/DeleteSSE for state-changing operations`)
}

// ----------------------------------------------------------------
//  MISCELLANEOUS LLM-SPECIFIC MISTAKES
// ----------------------------------------------------------------

// // closingBodyInLoop detects resp.Body.Close() inside a for loop.
// This leaks connections because resp.Body is only available in the loop scope.
func closeBodyInLoop(m dsl.Matcher) {
	m.Match(`for $*_ {
		$_
		resp, err := http.$*_
		$*_
		defer resp.Body.Close()
		$*_
	}`).
		Report(`defer resp.Body.Close() inside for loop — defers accumulate until function returns. ` +
			`Close explicitly: resp.Body.Close() at end of each iteration`)
}

// // timeAfterInLoop detects time.After() inside a for loop.
// time.After creates a new timer each iteration that leaks until it fires.
func timeAfterInLoop(m dsl.Matcher) {
	m.Match(`for $*_ {
		$*_
		<-time.After($_)
		$*_
	}`).
		Report(`time.After() inside for loop leaks timers until they fire. ` +
			`Use time.NewTicker() with a single timer, not time.After() in a loop`)
}

// // sqlRowWithoutClose detects missing row.Close() after sql.Rows query.
// The rows must be closed when iteration stops early.
func sqlRowsWithoutClose(m dsl.Matcher) {
	m.Match(`$rows, err := $db.Query($*_)`).
		Where(m["rows"].Type.Is(`*sql.Rows`)).
		Report(`sql.Rows must be closed when iteration ends. Add: defer rows.Close()`)
}
