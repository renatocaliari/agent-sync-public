# Embeddings: Local ONNX with Hugot + pt-BR SBD + Sliding Window + Relevance Stack

Production-ready embedding pipeline for Go projects. Three orthogonal pieces:

1. **Embedding infra** — `multilingual-e5-small` (384d) via Hugot (pure Go, no CGO).
2. **Retrieval helpers** — FTS5-integratable semantic search over messages.
3. **Relevance strategy stack** — composable strategies (token overlap → bi-encoder → cross-encoder rerank) with circuit-breaker fallbacks.

Zero external deps for text processing. Requires Hugot for ONNX inference backend.

## Files (in `features/<name>/embeddings/`)

| File | Purpose |
|------|---------|
| `embedder.go` | `ONNXEmbedder` + `FakeEmbedder` — lazy init, Hugot session, download, inference, long-form sampling |
| `cross_encoder.go` | `CrossEncoder` interface + `ONNXCrossEncoder` (BERT ms-marco-MiniLM-L-6-v2) + `FakeCrossEncoder` for tests |
| `text.go` | `SplitSentencesPT` (SBD), `ChunkSentences` (sliding window), `meanPoolVectors` |
| `vector.go` | `CosineSimilarity`, `Float32ToBytes`/`BytesToFloat32`, `TopKResults` |
| `search.go` | `EmbeddingMessageSearcher` — FTS5-integratable semantic search |

## Embedding Pipeline

```
[text.go] SplitSentencesPT → [text.go] ChunkSentences(maxBytes, overlap)
  → [embedder.go] Hugot embed each window
  → [text.go] meanPoolVectors → single []float32 (384d)
```

`Embed(text)` returns one vector regardless of input length. Caller never sees windowing.

## Key Design Decisions

- **SBD:** rule-based (abbreviation list + uppercase heuristic), pt-BR specific. Uses
  `FindAllStringSubmatchIndex` + position split (Go's RE2 has no lookahead).
  Same approach as prose, sentencizer, sentencex, PySBD.
- **Overlap:** 1 sentence between windows prevents context loss at boundaries.
- **Max window size:** MaxSeqLen * 4 bytes (≈512 tokens for XLMRoberta pt-BR).
- **Mean-pooling:** dilutes very specific terms but hybrid search (FTS5) covers exact matches.
- **Serialized:** mutex guards Hugot pipeline (not goroutine-safe). Fine for single-worker.
- **Pure Go:** gomlx backend — no CGO, no system libs. Trade: slower than ORT for large batches.

## Long-Form Article Support (>50KB / ~50 pages)

`ONNXEmbedder.EmbedBatch` caps window count at 16 for very long texts and samples evenly via stride:

```go
// from embedder.go
const longTextWindowCount = 16

if len(chunks) > longTextWindowCount {
    stride := len(chunks) / longTextWindowCount
    sampled := make([]string, 0, longTextWindowCount)
    for j := 0; j < len(chunks); j += stride {
        sampled = append(sampled, chunks[j])
        if len(sampled) >= longTextWindowCount { break }
    }
    // Always include the last chunk so the document's end is represented.
    if sampled[len(sampled)-1] != chunks[len(chunks)-1] {
        sampled[len(sampled)-1] = chunks[len(chunks)-1]
    }
    chunks = sampled
}
```

Empirical results (pure-Go backend, multilingual-e5-small, 384d):

| Text size | Sliding window (old) | 16-window cap (new) | Speedup |
|-----------|----------------------|---------------------|---------|
| 158KB (50 pages) | 169s | 27s | 6.3x |
| 2 × 150KB | ~340s | 53s | ~6.4x |

Texts <50KB use full sliding window (no quality loss). For higher throughput on
large documents, switch to ORT CGO backend (5–10x faster than pure-Go backend).

---

## Relevance Strategy Stack

When you have a corpus of user-authored texts (context blocks, source documents, KB articles)
that should influence LLM prompts, compose three strategies in priority order. Each tier
degrades gracefully to the next on failure — **prompts never fail**. Proven pattern.

```
┌──────────────────────────────────────────────────────────────────┐
│ CrossEncoderRerankRelevancer (best quality — composite)          │
│   ├─ FirstStage: SemanticEmbedRelevancer (bi-encoder cosine)     │
│   │   └─ breaker open → TokenOverlapRelevancer (bag-of-words)    │
│   └─ Rerank: BERT ms-marco-MiniLM-L-6-v2 (cross-encoder)          │
│       └─ breaker open / error → FirstStage scores                 │
└──────────────────────────────────────────────────────────────────┘
                                    ↓ (if both embedders nil)
                    TokenOverlapRelevancer (zero cost, always safe)
```

### Strategy Selection (in service.Initialize)

```go
if crossEncoder != nil → composite (CrossEncoderRerankRelevancer)
elif embedder != nil  → semantic only (SemanticEmbedRelevancer)
else                 → token overlap (TokenOverlapRelevancer, default)
```

Opt-out flags (env, read at init only):
- `ENABLE_SEMANTIC_BLOCK_RELEVANCE=false` — disable semantic + cross-encoder, keep token overlap
- `ENABLE_CROSS_ENCODER_RERANK=false` — disable rerank, keep semantic only

### Phase 1 — Save (async, pre-compute block embeddings)

```
User edits context block in UI
  → PB saves record (with Embedding=nil for new blocks)
  → PB hook OnRecordAfterUpdateSuccess("collection")
    enqueues "BLK:<id>" into the same embedding worker channel as messages
  → worker dispatches by prefix:
      bare ID         → embedMessage()
      "BLK:<id>"      → embedContextBlock()  // prefix discriminator
  → embedder.Embed(content) → repo.SetEmbedding(id, Float32ToBytes(vec))
```

- **No schema migration**: store the vector as `[]byte json:"embedding,omitempty"`
  on the existing entity. Old rows deserialize with nil → automatic fallback.
- **Reuse the existing worker**: don't create a parallel queue. Discriminator prefix
  (`BLK:`, `DOC:`, `KB:`) reuses panic recovery, retry/backoff, and backpressure.
- **Backfill on startup**: scan for entities with empty `Embedding`, enqueue them.

### Phase 2 — Prompt (hot path, sync, batched)

```
Handler builds prompt variables:
  "fontesContexto": relevanceStrategy.Select(blocks, promptType, caseText)
  → filterByPromptType(blocks, promptType)         // hard filter (cheap)
  → selectRelevantBlocks(filtered, caseText, N, r) // sort + top N
    └→ r.ScoreBatch(toScore, caseText)             // ONE batch call, not N per block
      │   (strategy-specific — see below)
  → render as string, inject via {{fontesContexto}}
```

**`ScoreBatch` is the hot-path API.** It amortizes the embedder call across N blocks. Latency profile:

| Strategy | Per-prompt cost | Quality |
|----------|-----------------|---------|
| Token overlap | ~µs (no embed) | Lexical only |
| Bi-encoder (semantic cosine) | ~50–100ms (1 embed) | Semantic (recall-focused) |
| Cross-encoder rerank | ~50–100ms + ~30ms × K pairs | Best (cross-attention) |

### Phase 3 — Failsafe (4 fallback points per strategy)

Every strategy in the stack implements the same fallback contract. Prompts **never fail** because of embedding. 4 fallback points:

1. empty `caseText`
2. `embedder == nil` (model didn't load)
3. breaker open (3 consecutive failures → 5min open)
4. embedder error → record failure + fallback

### Relevancer interface

```go
// All strategies implement this. Name() identifies the strategy for observability
// (composite strategies name their sub-strategies, e.g. "cross-encoder-rerank(semantic-embed)").
type Relevancer interface {
    Score(block, caseText string) float64
    ScoreBatch(blocks []Entity, caseText string) []float64  // hot path: amortize cost
    Name() string
}
```

### Implementation stack

```go
// Tier 1 (default — zero cost): bag-of-words on Title + first 200 bytes of Content.
// Best fallback. Safe because no external dependencies.
type TokenOverlapRelevancer struct{}
func (TokenOverlapRelevancer) Score(block Entity, caseText string) float64 {
    return calculateRelevance(snippet(block), block.Title, caseText)
    // snippet = Title + " " + first 200 bytes of Content
    // Title-substring match adds +0.3 bonus
}
func (TokenOverlapRelevancer) Name() string { return "token-overlap" }

// Tier 2 (bi-encoder): cosine similarity over pre-computed block embeddings.
// Best recall. Async pre-compute keeps hot path at ~1 embed call.
type SemanticEmbedRelevancer struct {
    embedder embeddings.Embedder
    cache    *BlockEmbeddingsCache        // in-memory map[id][]float32, RWMutex
    fallback Relevancer                   // TokenOverlapRelevancer
    breaker  *Breaker                     // 3-strike, 5min open
}
func (s *SemanticEmbedRelevancer) ScoreBatch(blocks, caseText string) []float64 {
    if caseText == "" || s.embedder == nil        { /* all fallback */ }
    if !s.breaker.Allow()                          { /* all fallback */ }
    caseVec, err := s.embedder.Embed(caseText)     // 1 call
    if err != nil { s.breaker.Record(true); /* all fallback */ }
    s.breaker.Record(false)
    for each block: cosine(blockVec, caseVec)      // ~µs each, pure math
}

// Tier 3 (cross-encoder composite): best quality via BERT cross-attention rerank.
// Reranks only top-K (e.g. 10) candidates from Stage 1 bi-encoder.
// Latency: stage 1 (~100ms) + stage 2 (~300ms for 10 pairs sequential, or ~50ms batched).
type CrossEncoderRerankRelevancer struct {
    FirstStage Relevancer              // typically SemanticEmbedRelevancer
    Rerank     embeddings.CrossEncoder // BERT ms-marco via embeddings.ONNXCrossEncoder
    FirstK     int                     // candidate set size (default 10)
    Breaker    *Breaker                // shared or own (3-strike, 5min open)
    Fallback   Relevancer              // when rerank fails: FirstStage (not raw overlap)
}
func (c *CrossEncoderRerankRelevancer) ScoreBatch(blocks, caseText string) []float64 {
    firstScores := c.FirstStage.ScoreBatch(blocks, caseText)
    if caseText == "" || !c.Breaker.Allow() { return firstScores }
    // Sort by first-stage score, take top FirstK
    topN := min(len(blocks), c.FirstK)
    documents := make([]string, topN) // block.Content for each top-N
    rerankScores, err := c.Rerank.ScoreBatch(ctx, caseText, documents)
    if err != nil { c.Breaker.Record(true); return firstScores }
    c.Breaker.Record(false)
    // Replace top-N scores with rerank scores; rest keep first-stage.
    return mergeScores(firstScores, topIdx, rerankScores)
}
```

### Cross-encoder module (cross_encoder.go)

```go
// Public interface — used by CrossEncoderRerankRelevancer.
type CrossEncoder interface {
    Score(ctx, query, document string) (float32, error)        // sigmoid score in [0, 1]
    ScoreBatch(ctx, query string, documents []string) ([]float32, error)  // faster: 1 ONNX forward pass
    Close() error
}

// Hugot-based ONNX impl. Lazy model download. Mutex-guarded session (not goroutine-safe).
// Runs on the same gomlx pure-Go backend as the bi-encoder (no CGO).
type ONNXCrossEncoder struct { /* ... */ }

func NewONNXCrossEncoder(opts CrossEncoderOpts) *ONNXCrossEncoder
// Default: "cross-encoder/ms-marco-MiniLM-L-6-v2", BatchSize=32, cacheDir=XDG_CACHE_HOME

// For tests — deterministic scores, no model load.
type FakeCrossEncoder struct {
    DefaultScore float32
    ScoreFn      func(query, document string) float32
    FailScore    bool  // forces error on every call (for breaker tests)
}
```

**Why cross-encoder over LLM judge?** Cross-encoder is the gold-standard for retrieval
reranking: same inference infra (ONNX), ~30ms per pair vs 200-2000ms for LLM, no
hallucination (returns numeric score not text), no API cost. Use LLM judge only when
you need reasoning over cross-block context or have labeled training data.

### Circuit breaker

```go
type breaker struct {
    mu           sync.Mutex
    failures     int
    openUntil    time.Time
    threshold    int           // 3
    openDuration time.Duration // 5 * time.Minute
    now          func() time.Time  // injected for tests
}

func (b *breaker) Allow() bool     { return !b.now().Before(b.openUntil) }
func (b *breaker) Record(failure bool) {
    if failure {
        b.failures++
        if b.failures >= b.threshold {
            b.openUntil = b.now().Add(b.openDuration)
            b.failures = 0
        }
        return
    }
    if !b.openUntil.IsZero() { log.Printf("circuit breaker closed") }
    b.failures = 0
    b.openUntil = time.Time{}
}
```

- No goroutines, no timers — caller invokes `Allow()` before each call, `Record()` after.
- For deterministic tests: add `newBreakerWithClock(threshold, openDuration, clock func() time.Time)`
  so tests can advance time without `time.Sleep`.
- Logs at state transitions: `"circuit breaker opened for 5m (3 consecutive embedder failures)"` / `"circuit breaker closed (embedder recovered)"`.

### In-memory block embeddings cache

```go
type BlockEmbeddingsCache struct {
    mu   sync.RWMutex
    data map[string][]float32
}

func (c *BlockEmbeddingsCache) Get(id string) ([]float32, bool) { /* O(1) under RLock */ }
func (c *BlockEmbeddingsCache) Load(repo BlockEmbeddingRepo) error {
    raw, _ := repo.LoadEmbeddings()                    // map[id][]byte
    decoded := decodeRaw(raw)                          // bytes → float32
    c.mu.Lock(); c.data = decoded; c.mu.Unlock()       // atomic swap
    return nil
}
```

- Loads once at startup (`NarrativeService.Initialize`).
- Refreshed on every save (PB hook calls `svc.RefreshBlockEmbeddings()` after enqueuing new embeddings).
- Hot path: RLock only, zero mutation.

### Storage pattern (no schema migration)

```go
// db/types.go
type ContextBlock struct {
    // ... existing fields (Title, Content, Active, PromptTypes, etc.) ...
    Embedding []byte `json:"embedding,omitempty"`   // 384 floats = 1536 bytes
}

// db/pb_repos.go — returns raw bytes (NOT []float32) to avoid circular import
func (r *PBSettingsRepo) LoadBlockEmbeddings() (map[string][]byte, error)
func (r *PBSettingsRepo) SetBlockEmbedding(id string, vec []byte) error
```

The `db` package cannot import `features/.../embeddings` (circular). Repo returns `[]byte`;
conversion `[]byte → []float32` via `embeddings.BytesToFloat32` happens in the cache layer
(the only place that imports both packages).

### Validation (integration tests, real model)

```go
//go:build integration

// TestIntegration_SemanticMatch_PTBR: "infantojuvenil" vs block about "crianças"
//   b-criancas=0.9114  b-luto=0.8638  ✓ semantic match wins
//   token-overlap score for the same pair = 0  (negative control)

// TestIntegration_EmbedBatch_LongArticle_50Pages: 158KB article embeds in <30s
//   old: 169s (75 windows)   new: 27s (16 windows, stride sampling)   6.3x speedup
```

Always include **negative control** tests that prove the lower tier returns 0 where the
higher tier succeeds — confirms each tier actually solves the problem it claims.

---

## When to use this stack

- User-authored reference texts (context blocks, KB articles, custom prompts)
- Corpus is small enough to fit in RAM (< 100K blocks → < 600MB cache)
- Query path is hot (multiple relevance queries per request)
- Token-overlap fails on synonymy ("ansiedade" ≠ "preocupação", "luto" ≠ "perda recente")
- Documents range from short (1KB) to long-form (50 pages, 150KB)

## When NOT to use

- Real-time generated content (no pre-compute opportunity)
- Cor > 1M blocks (RAM becomes a problem; consider vector index)
- Already have a vector DB (Pinecone, Weaviate, etc.) — use that, don't reinvent
- Documents are uniformly > 1MB (embedding takes >1 minute; chunk at ingest time)

## See also

- `references/datastore/pocketbase.md` — PB hook patterns used by Phase 1
- `references/llm-streaming.md` — SSE patterns for streaming LLM responses
- `cali-coding-go-standards` — circuit breaker pattern (general), error handling
