# Quick Start

This guide walks you through your first memory operations with OmniMemory.

## Setup

First, create a client with a provider:

```go
package main

import (
    "context"
    "log"

    "github.com/plexusone/omnimemory"
    "github.com/plexusone/omnimemory/core"
    _ "github.com/plexusone/omnimemory/provider/memory"
)

func main() {
    client, err := omnimemory.NewClient(core.ClientConfig{
        Providers: []core.ProviderConfig{
            {Name: core.ProviderNameMemory},
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()
    // ... operations below
}
```

## Add Memories

Store observations, facts, or preferences:

```go
// Define the context (who this memory belongs to)
memCtx := core.Context{
    TenantID:  "acme-corp",
    SubjectID: "user-123",
}

// Add an observation
memory, err := client.Add(ctx, &core.AddRequest{
    Context: memCtx,
    Type:    core.MemoryTypeObservation,
    Content: "User asked about pricing plans",
})

// Add a preference
_, err = client.Add(ctx, &core.AddRequest{
    Context: memCtx,
    Type:    core.MemoryTypePreference,
    Content: "User prefers email communication over phone",
})

// Add a fact
_, err = client.Add(ctx, &core.AddRequest{
    Context: memCtx,
    Type:    core.MemoryTypeFact,
    Content: "User is on the Enterprise plan since 2024",
})
```

## Search Memories

Find relevant memories using semantic search:

```go
results, err := client.Search(ctx, &core.SearchRequest{
    Context:   memCtx,
    Query:     "communication preferences",
    Limit:     5,
    Threshold: 0.7, // Minimum similarity score
})

for _, result := range results.Results {
    fmt.Printf("Score: %.2f - %s\n", result.Score, result.Memory.Content)
}
```

## Recall Memories

Get contextually relevant memories with optional summarization:

```go
recalled, err := client.Recall(ctx, &core.RecallRequest{
    Context:    memCtx,
    Query:      "What do we know about this user?",
    MaxResults: 10,
})

fmt.Printf("Summary: %s\n", recalled.Summary)
for _, mem := range recalled.Memories {
    fmt.Printf("- [%s] %s\n", mem.Type, mem.Content)
}
```

## List Memories

Retrieve all memories for a subject:

```go
list, err := client.List(ctx, &core.ListRequest{
    Context: memCtx,
    Limit:   100,
})

fmt.Printf("Total memories: %d\n", list.TotalCount)
for _, mem := range list.Memories {
    fmt.Printf("- %s: %s\n", mem.ID, mem.Content)
}
```

## Update Memories

Modify existing memories:

```go
updated, err := client.Update(ctx, &core.UpdateRequest{
    Context: memCtx,
    ID:      memory.ID,
    Content: "User asked about Enterprise pricing plans",
})
```

## Delete Memories

Remove memories when no longer needed:

```go
err = client.Delete(ctx, &core.DeleteRequest{
    Context: memCtx,
    ID:      memory.ID,
})
```

## Memory Types

OmniMemory supports several memory types:

| Type | Use Case |
|------|----------|
| `observation` | Observed behaviors or interactions |
| `fact` | Verified pieces of information |
| `preference` | User preferences |
| `summary` | Summarized information |
| `trait` | Personality traits |
| `relationship` | Relationships between entities |

## Next Steps

- [Configuration](configuration.md) - Advanced configuration options
- [Providers](../providers/index.md) - Choose and configure providers
- [Semantic Search](../features/search.md) - Deep dive into search capabilities
