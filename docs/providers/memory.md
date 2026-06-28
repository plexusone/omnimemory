# In-Memory Provider

The in-memory provider stores memories in a Go map. Ideal for testing, development, and ephemeral use cases.

## Features

- Zero configuration
- Fast operations
- Brute-force cosine similarity search
- Thread-safe with mutex protection
- No persistence (data lost on restart)

## Installation

```go
import _ "github.com/plexusone/omnimemory/provider/memory"
```

## Configuration

```go
import (
    "github.com/plexusone/omnimemory"
    "github.com/plexusone/omnimemory/core"
    _ "github.com/plexusone/omnimemory/provider/memory"
)

client, err := omnimemory.NewClient(core.ClientConfig{
    Providers: []core.ProviderConfig{
        {Name: core.ProviderNameMemory},
    },
})
```

No options are required.

## Usage

```go
ctx := context.Background()

// Add a memory
memory, err := client.Add(ctx, &core.AddRequest{
    Context: core.Context{
        TenantID:  "tenant-1",
        SubjectID: "user-1",
    },
    Type:    core.MemoryTypeObservation,
    Content: "User prefers dark mode",
})

// Search memories
results, err := client.Search(ctx, &core.SearchRequest{
    Context: core.Context{
        TenantID:  "tenant-1",
        SubjectID: "user-1",
    },
    Query: "user preferences",
    Limit: 10,
})
```

## Embeddings

The in-memory provider requires an embedder for semantic search:

```go
embedder, _ := core.NewOmniLLMEmbedder(core.EmbedderConfig{
    Provider: "openai",
    APIKey:   os.Getenv("OPENAI_API_KEY"),
    Model:    "text-embedding-3-small",
})

provider, err := memory.NewProvider(core.ProviderConfig{}, embedder)
```

If no embedder is provided, search operations will return an error.

## Limitations

- **No persistence**: Data is lost when the process exits
- **Memory usage**: All data stored in RAM
- **Not distributed**: Single-process only
- **Linear search**: O(n) similarity search

## Use Cases

- Unit tests
- Integration tests
- Local development
- Prototyping
- Ephemeral conversations

## Direct Provider Usage

```go
import "github.com/plexusone/omnimemory/provider/memory"

provider, err := memory.NewProvider(core.ProviderConfig{}, embedder)
if err != nil {
    log.Fatal(err)
}
defer provider.Close()

// Use provider directly
memory, err := provider.Add(ctx, &core.AddRequest{...})
```
