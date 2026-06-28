# Semantic Search

OmniMemory provides semantic search capabilities using vector embeddings, allowing you to find relevant memories based on meaning rather than exact text matches.

## How It Works

1. **Embedding Generation**: Text is converted to a vector embedding
2. **Similarity Calculation**: Query embedding is compared to stored memory embeddings
3. **Ranking**: Results are sorted by similarity score
4. **Filtering**: Results are filtered by threshold and limit

## Search API

```go
results, err := client.Search(ctx, &core.SearchRequest{
    Context: core.Context{
        TenantID:  "tenant-1",
        SubjectID: "user-1",
    },
    Query:     "user interface preferences",
    Types:     []core.MemoryType{core.MemoryTypePreference},
    Scopes:    []core.Scope{core.ScopeUser},
    Limit:     10,
    Threshold: 0.7,
})
```

### Parameters

| Field | Type | Description |
|-------|------|-------------|
| `Context` | `Context` | Required: Tenant and subject context |
| `Query` | `string` | Required: Search query text |
| `Types` | `[]MemoryType` | Optional: Filter by memory types |
| `Scopes` | `[]Scope` | Optional: Filter by scopes |
| `Limit` | `int` | Optional: Max results (default: 10) |
| `Threshold` | `float64` | Optional: Min similarity (0.0-1.0) |

### Response

```go
type SearchResponse struct {
    Results []SearchResult
}

type SearchResult struct {
    Memory *Memory
    Score  float64  // Similarity score (0.0-1.0)
}
```

## Example

```go
results, err := client.Search(ctx, &core.SearchRequest{
    Context: core.Context{
        TenantID:  "acme-corp",
        SubjectID: "user-123",
    },
    Query: "communication preferences",
    Limit: 5,
})
if err != nil {
    log.Fatal(err)
}

for _, result := range results.Results {
    fmt.Printf("Score: %.2f | %s: %s\n",
        result.Score,
        result.Memory.Type,
        result.Memory.Content,
    )
}
```

Output:

```
Score: 0.92 | preference: User prefers email over phone calls
Score: 0.85 | observation: User responded quickly to Slack messages
Score: 0.78 | fact: User timezone is PST
```

## Recall API

The `Recall` API provides contextual memory retrieval with optional summarization:

```go
recalled, err := client.Recall(ctx, &core.RecallRequest{
    Context: core.Context{
        TenantID:  "acme-corp",
        SubjectID: "user-123",
    },
    Query:        "What do we know about this user's preferences?",
    MaxResults:   10,
    IncludeTypes: []core.MemoryType{core.MemoryTypePreference},
})

fmt.Printf("Summary: %s\n", recalled.Summary)
for _, mem := range recalled.Memories {
    fmt.Printf("- %s\n", mem.Content)
}
```

### Recall vs Search

| Feature | Search | Recall |
|---------|--------|--------|
| Returns scores | Yes | No |
| Summarization | No | Yes (provider-dependent) |
| Use case | Finding specific memories | Contextual understanding |

## Similarity Threshold

The threshold controls result quality:

| Threshold | Behavior |
|-----------|----------|
| `0.9+` | Very similar, near-exact matches |
| `0.7-0.9` | Related content, good for most cases |
| `0.5-0.7` | Loosely related content |
| `<0.5` | May include irrelevant results |

## Filtering

### By Memory Type

```go
results, err := client.Search(ctx, &core.SearchRequest{
    Context: memCtx,
    Query:   "user preferences",
    Types: []core.MemoryType{
        core.MemoryTypePreference,
        core.MemoryTypeFact,
    },
})
```

### By Scope

```go
results, err := client.Search(ctx, &core.SearchRequest{
    Context: memCtx,
    Query:   "team knowledge",
    Scopes: []core.Scope{
        core.ScopeTeam,
        core.ScopeTenant,
    },
})
```

## Provider Differences

| Provider | Search Implementation |
|----------|----------------------|
| PostgreSQL | Native pgvector similarity |
| In-Memory | Brute-force cosine similarity |
| KVS | In-memory after listing |
| Twilio | Twilio Recall API |

## Performance Tips

1. **Use appropriate limits**: Start with small limits and increase as needed
2. **Filter early**: Use `Types` and `Scopes` to reduce search space
3. **Choose the right provider**: PostgreSQL with HNSW for large datasets
4. **Cache embeddings**: Query embedding generation adds latency
