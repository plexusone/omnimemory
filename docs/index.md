# OmniMemory

**Vendor-neutral memory abstraction layer for Go**

OmniMemory provides a unified interface for storing and retrieving semantic memories across multiple backends. It supports multi-tenancy, semantic search via embeddings, and automatic provider fallback.

## Features

- **Multi-Provider Support**: PostgreSQL+pgvector, In-memory, KVS, Twilio Memory API
- **Unified API**: Same interface across all providers
- **Semantic Search**: Vector similarity search with configurable embeddings
- **Multi-Tenancy**: Built-in tenant and subject isolation
- **Memory Scopes**: User, agent, tenant, team, session, domain
- **Memory Types**: Observation, fact, preference, summary, trait, relationship
- **Fallback Support**: Automatic failover to backup providers
- **Conformance Tests**: Validate provider implementations

## Quick Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/plexusone/omnimemory"
    "github.com/plexusone/omnimemory/core"
    _ "github.com/plexusone/omnimemory/provider/memory" // Register provider
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

    // Add a memory
    memory, err := client.Add(context.Background(), &core.AddRequest{
        Context: core.Context{
            TenantID:  "tenant-123",
            SubjectID: "user-456",
        },
        Type:    core.MemoryTypeObservation,
        Content: "User prefers dark mode interfaces",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Added memory: %s\n", memory.ID)

    // Recall memories
    recalled, err := client.Recall(context.Background(), &core.RecallRequest{
        Context: core.Context{
            TenantID:  "tenant-123",
            SubjectID: "user-456",
        },
        Query: "user interface preferences",
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, mem := range recalled.Memories {
        fmt.Printf("- %s\n", mem.Content)
    }
}
```

## Supported Providers

| Provider | Backend | Features |
|----------|---------|----------|
| [In-Memory](providers/memory.md) | Go map | Testing, development |
| [PostgreSQL](providers/postgres.md) | PostgreSQL + pgvector | Production, vector search |
| [KVS](providers/kvs.md) | omnistorage-core | Flexible key-value backends |
| [Twilio](providers/twilio.md) | Twilio Memory API | External via omni-twilio |

## Concept Mapping

OmniMemory uses consistent terminology that maps to provider-specific concepts:

| OmniMemory | Description | Twilio Memory |
|------------|-------------|---------------|
| TenantID | Organization/workspace | Store ID |
| SubjectID | User/entity the memory is about | Profile ID |
| Memory | Stored semantic information | Observation |
| Search | Vector similarity query | - |
| Recall | Contextual memory retrieval | Recall API |

## Next Steps

- [Installation](getting-started/installation.md) - Get OmniMemory set up
- [Quick Start](getting-started/quickstart.md) - Your first memory operations
- [Providers](providers/index.md) - Configure specific providers
- [Features](features/search.md) - Explore semantic search
