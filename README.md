# OmniMemory

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Docs][docs-mkdoc-svg]][docs-mkdoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/omnimemory/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/omnimemory/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/omnimemory/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/omnimemory/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/omnimemory/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/omnimemory/actions/workflows/go-sast-codeql.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/plexusone/omnimemory
 [goreport-url]: https://goreportcard.com/report/github.com/plexusone/omnimemory
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/omnimemory
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/omnimemory
 [docs-mkdoc-svg]: https://img.shields.io/badge/Go-dev%20guide-blue.svg
 [docs-mkdoc-url]: https://plexusone.dev/omnimemory
 [viz-svg]: https://img.shields.io/badge/Go-visualizaton-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fomnimemory
 [loc-svg]: https://tokei.rs/b1/github/plexusone/omnimemory
 [repo-url]: https://github.com/plexusone/omnimemory
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/omnimemory/blob/main/LICENSE

Vendor-neutral memory abstraction layer for Go. Store and retrieve semantic memories across multiple backends with a unified API.

## Features

- **Multi-Provider Support**: PostgreSQL+pgvector, In-memory, KVS, Twilio Memory API
- **Unified API**: Same interface across all providers
- **Semantic Search**: Vector similarity search with configurable embeddings
- **Multi-Tenancy**: Built-in tenant and subject isolation
- **Memory Scopes**: User, agent, tenant, team, session, domain
- **Memory Types**: Observation, fact, preference, summary, trait, relationship
- **Fallback Support**: Automatic failover to backup providers
- **Conformance Tests**: Validate provider implementations

## Installation

```bash
go get github.com/plexusone/omnimemory
```

## Quick Start

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

    ctx := context.Background()

    // Add a memory
    memory, err := client.Add(ctx, &core.AddRequest{
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
    recalled, err := client.Recall(ctx, &core.RecallRequest{
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

## Providers

| Provider | Package | Use Case |
|----------|---------|----------|
| In-Memory | `provider/memory` | Testing, development |
| PostgreSQL | `provider/postgres` | Production with pgvector |
| KVS | `provider/kvs` | Flexible key-value backends |
| Twilio | [omni-twilio](https://github.com/plexusone/omni-twilio) | Twilio Memory API |

### PostgreSQL Setup

```go
import _ "github.com/plexusone/omnimemory/provider/postgres"

client, err := omnimemory.NewClient(core.ClientConfig{
    Providers: []core.ProviderConfig{
        {
            Name: core.ProviderNamePostgres,
            Options: map[string]any{
                "connection_string": os.Getenv("DATABASE_URL"),
            },
        },
    },
})
```

### Multi-Provider with Fallback

```go
client, err := omnimemory.NewClient(core.ClientConfig{
    Providers: []core.ProviderConfig{
        // Primary: PostgreSQL
        {Name: core.ProviderNamePostgres, Options: map[string]any{
            "connection_string": os.Getenv("DATABASE_URL"),
        }},
        // Fallback: In-memory
        {Name: core.ProviderNameMemory},
    },
})
```

## Memory Types

| Type | Description |
|------|-------------|
| `observation` | Observed behaviors or interactions |
| `fact` | Verified pieces of information |
| `preference` | User preferences |
| `summary` | Summarized information |
| `trait` | Personality traits |
| `relationship` | Relationships between entities |

## Memory Scopes

| Scope | Description |
|-------|-------------|
| `user` | Personal to one user |
| `agent` | What an agent has learned |
| `tenant` | Organization-wide shared |
| `team` | Project/group level |
| `session` | Single conversation |
| `domain` | Domain-specific (support, sales) |

## Documentation

- [Documentation Site](https://plexusone.github.io/omnimemory/)
- [Go Package Reference](https://pkg.go.dev/github.com/plexusone/omnimemory)

## Related Projects

- [omnillm-core](https://github.com/plexusone/omnillm-core) - Unified LLM SDK
- [omni-twilio](https://github.com/plexusone/omni-twilio) - Twilio adapters including Memory provider
- [omnistorage-core](https://github.com/plexusone/omnistorage-core) - Storage abstraction for KVS provider

## License

MIT License - see [LICENSE](LICENSE) for details.
