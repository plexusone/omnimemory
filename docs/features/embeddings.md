# Embeddings

OmniMemory uses vector embeddings for semantic search. This guide covers embedding configuration and best practices.

## Embedder Interface

```go
type Embedder interface {
    Embed(ctx context.Context, text string) ([]float64, error)
    EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)
    Dimension() int
}
```

## OmniLLM Integration

The recommended embedder uses omnillm-core:

```go
import "github.com/plexusone/omnimemory/core"

embedder, err := core.NewOmniLLMEmbedder(core.EmbedderConfig{
    Provider: "openai",
    APIKey:   os.Getenv("OPENAI_API_KEY"),
    Model:    "text-embedding-3-small",
})
if err != nil {
    log.Fatal(err)
}
```

### Configuration Options

| Field | Type | Description |
|-------|------|-------------|
| `Provider` | `string` | LLM provider name |
| `APIKey` | `string` | API key |
| `Model` | `string` | Embedding model |
| `Dimension` | `int` | Vector dimension (optional) |
| `BaseURL` | `string` | Custom API endpoint (optional) |

## Supported Models

### OpenAI

| Model | Dimensions | Notes |
|-------|------------|-------|
| `text-embedding-3-small` | 1536 | Recommended for most use cases |
| `text-embedding-3-large` | 3072 | Higher quality, more expensive |
| `text-embedding-ada-002` | 1536 | Legacy model |

```go
embedder, _ := core.NewOmniLLMEmbedder(core.EmbedderConfig{
    Provider: "openai",
    Model:    "text-embedding-3-small",
    APIKey:   os.Getenv("OPENAI_API_KEY"),
})
```

### Anthropic

Anthropic doesn't provide embedding models directly. Use OpenAI or a local model.

### Local Models (Ollama)

```go
embedder, _ := core.NewOmniLLMEmbedder(core.EmbedderConfig{
    Provider: "ollama",
    Model:    "nomic-embed-text",
    BaseURL:  "http://localhost:11434",
})
```

## Using with Providers

### PostgreSQL

```go
import "github.com/plexusone/omnimemory/provider/postgres"

provider, err := postgres.NewProvider(core.ProviderConfig{
    Options: map[string]any{
        "connection_string": os.Getenv("DATABASE_URL"),
    },
}, embedder)
```

### In-Memory

```go
import "github.com/plexusone/omnimemory/provider/memory"

provider, err := memory.NewProvider(core.ProviderConfig{}, embedder)
```

### KVS

```go
import "github.com/plexusone/omnimemory/provider/kvs"

provider, err := kvs.NewProvider(core.ProviderConfig{
    Options: map[string]any{
        "store": store,
    },
}, embedder)
```

## Custom Embedder

Implement the `Embedder` interface for custom embedding solutions:

```go
type MyEmbedder struct {
    dimension int
}

func (e *MyEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
    // Call your embedding API
    return myEmbeddingAPI(text)
}

func (e *MyEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
    embeddings := make([][]float64, len(texts))
    for i, text := range texts {
        emb, err := e.Embed(ctx, text)
        if err != nil {
            return nil, err
        }
        embeddings[i] = emb
    }
    return embeddings, nil
}

func (e *MyEmbedder) Dimension() int {
    return e.dimension
}
```

## Dimension Matching

Embedding dimensions must match across:

1. **Embedder output**: The dimension from your embedding model
2. **Database schema**: The vector column dimension
3. **Stored embeddings**: Previously stored memory embeddings

```go
// Check dimension matches
if embedder.Dimension() != 1536 {
    log.Fatal("Embedder dimension must match database schema")
}
```

## Performance Optimization

### Batch Embedding

```go
// Embed multiple texts at once
texts := []string{"text1", "text2", "text3"}
embeddings, err := embedder.EmbedBatch(ctx, texts)
```

### Caching

Consider caching embeddings for frequently used queries:

```go
type CachedEmbedder struct {
    inner Embedder
    cache map[string][]float64
    mu    sync.RWMutex
}

func (e *CachedEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
    e.mu.RLock()
    if emb, ok := e.cache[text]; ok {
        e.mu.RUnlock()
        return emb, nil
    }
    e.mu.RUnlock()

    emb, err := e.inner.Embed(ctx, text)
    if err != nil {
        return nil, err
    }

    e.mu.Lock()
    e.cache[text] = emb
    e.mu.Unlock()

    return emb, nil
}
```

## Best Practices

1. **Use consistent models**: Don't mix embedding models
2. **Normalize text**: Clean input before embedding
3. **Handle errors**: Embedding APIs can fail
4. **Monitor costs**: Track embedding API usage
5. **Consider local models**: For high volume or privacy
