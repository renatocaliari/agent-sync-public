# Bleve Hybrid Search

[Bleve](https://github.com/blevesearch/bleve) v2.4+ provides native hybrid search (k-NN vector + full-text + RRF fusion) in pure Go, zero CGO.

## When to Use

- You need **hybrid search** (text + vector) and want a single engine instead of composing FTS5 + manual cosine + RRF in code.
- You want **in-memory indexing** (`NewMemOnly`) — no disk state to sync, zero split-state in backups. Rebuild from SQLite on startup.
- You need **Portuguese text analysis** (built-in `pt` analyzer with stemming).

## When NOT to Use

- Simple keyword search only → FTS5 via sqlite (`unicode61 remove_diacritics 2 prefix=3 4`).
- Large document corpus with frequent updates → Bleve in-memory rebuilds from SQLite on each restart; for >10k documents evaluate rebuild time.
- Full embedding infrastructure already in place → the manual FTS5+embedding+RRF approach (see `references/embeddings/README.md`) is equally valid.

## Setup

```go
import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
)

func initBleve() (bleve.Index, error) {
	indexMapping := bleve.NewIndexMapping()
	docMapping := bleve.NewDocumentMapping()

	// Full-text field with Portuguese analyzer
	contentMapping := bleve.NewTextFieldMapping()
	contentMapping.Analyzer = "pt"
	docMapping.AddFieldMappingsAt("content", contentMapping)

	// Vector field (384d for multilingual-e5-small)
	vectorMapping := bleve.NewVectorFieldMapping()
	vectorMapping.Dimension = 384
	docMapping.AddFieldMappingsAt("embedding", vectorMapping)

	indexMapping.AddDocumentMapping("context_block", docMapping)
	return bleve.NewMemOnly(indexMapping)
}
```

## Index from SQLite

On startup (and whenever source data changes), rebuild the in-memory index:

```go
type SearchableBlock struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Embedding []float32 `json:"embedding"`
}

func RebuildIndex(ctx context.Context, idx bleve.Index, blocks []Block) error {
	// Recreate clean index (avoids duplicates on rebuild)
	newIdx, err := initBleve()
	if err != nil {
		return err
	}

	for _, b := range blocks {
		if !b.Active || b.Content == "" || len(b.Embedding) == 0 {
			continue
		}
		item := SearchableBlock{
			ID:        b.ID,
			Content:   b.Content,
			Embedding: b.Embedding,
		}
		if err := newIdx.Index(b.ID, item); err != nil {
			log.Printf("[bleve] index error %s: %v", b.ID, err)
		}
	}

	// Swap atomically
	*idx = newIdx
	return nil
}
```

## Hybrid Search (k-NN + Text)

```go
func Search(idx bleve.Index, caseText string, vec []float32, limit int) []string {
	termQuery := bleve.NewMatchQuery(caseText)
	termQuery.SetField("content")

	searchRequest := bleve.NewSearchRequest(termQuery)
	searchRequest.Size = limit
	searchRequest.AddKNN("embedding", vec, limit, 1.0)

	results, err := idx.Search(searchRequest)
	if err != nil {
		return nil
	}

	ids := make([]string, len(results.Hits))
	for i, hit := range results.Hits {
		ids[i] = hit.ID
	}
	return ids
}
```

`AddKNN` merges k-NN vector scores with the full-text query score via Bleve's internal RRF/RSF fusion — no manual score combination needed.

## Portuguese Analyzer

Bleve ships with `"pt"` analyzer (Snowball Portuguese stemmer). For custom analyzers (asciifolding, stop words):

```go
import "github.com/blevesearch/bleve/v2/analysis/analyzer/custom"

analyzer := map[string]interface{}{
	"type":          custom.Name,
	"tokenizer":     keyword.Name,
	"token_filters": []string{"possessive_en", "lowercase", "asciifolding", "stop_en"},
}
```

## Tradeoffs

| Aspect | Bleve In-Memory | Manual FTS5 + Go Cosine |
|--------|----------------|------------------------|
| Hybrid fusion | Native (RRF built-in) | Manual k=60 RRF in code |
| Persistence | SQLite (rebuild on startup) | SQLite (always consistent) |
| Portuguese text | Built-in `pt` analyzer | sqlite unicode61 with remove_diacritics |
| Binary size | +~5MB | minimal |
| Rebuild cost | O(n) on startup | O(0) — embeddings in DB |
