# KVS Provider

The KVS provider wraps any `kvs.ListableStore` from omnistorage-core, enabling flexible backend choices including Redis, S3, filesystem, and more.

## Features

- Pluggable key-value backends
- JSON serialization
- Prefix-based memory organization
- In-memory similarity search
- Works with any omnistorage-core store

## Installation

```bash
go get github.com/plexusone/omnistorage-core
```

```go
import _ "github.com/plexusone/omnimemory/provider/kvs"
```

## Configuration

```go
import (
    "github.com/plexusone/omnimemory"
    "github.com/plexusone/omnimemory/core"
    "github.com/plexusone/omnistorage-core/kvs"
    _ "github.com/plexusone/omnimemory/provider/kvs"
)

// Create a KVS store
store := kvs.NewMemoryStore()

client, err := omnimemory.NewClient(core.ClientConfig{
    Providers: []core.ProviderConfig{
        {
            Name: core.ProviderNameKVS,
            Options: map[string]any{
                "store":  store,
                "prefix": "memories",
            },
        },
    },
})
```

### Options

| Option | Type | Description |
|--------|------|-------------|
| `store` | `kvs.ListableStore` | Required: The underlying KVS store |
| `prefix` | `string` | Optional: Key prefix (default: "memories") |

## Key Structure

Memories are stored with keys in the format:

```
{prefix}:{tenant_id}:{subject_id}:{memory_id}
```

Example:

```
memories:acme-corp:user-123:mem_abc123
```

## Supported Backends

### In-Memory

```go
store := kvs.NewMemoryStore()
```

### Redis

```go
import "github.com/plexusone/omnistorage-core/kvs/redis"

store, err := redis.NewStore(redis.Config{
    Addr: "localhost:6379",
})
```

### S3

```go
import "github.com/plexusone/omnistorage-core/kvs/s3"

store, err := s3.NewStore(s3.Config{
    Bucket: "my-memories",
    Region: "us-west-2",
})
```

### Filesystem

```go
import "github.com/plexusone/omnistorage-core/kvs/file"

store, err := file.NewStore(file.Config{
    BasePath: "/var/data/memories",
})
```

## Storage Format

Memories are stored as JSON:

```json
{
  "id": "mem_abc123",
  "tenant_id": "acme-corp",
  "subject_id": "user-123",
  "type": "observation",
  "content": "User prefers dark mode",
  "embedding": [0.1, 0.2, ...],
  "metadata": {"source": "chat"},
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

## Search Behavior

The KVS provider performs similarity search by:

1. Listing all memories for the tenant/subject
2. Computing cosine similarity in-memory
3. Filtering by threshold
4. Sorting by score

This is efficient for small to medium datasets but may be slow for large collections.

## Limitations

- **Linear search**: O(n) similarity search
- **List performance**: Depends on underlying store
- **No indexing**: All filtering done in-memory

## Use Cases

- Serverless deployments (S3 backend)
- Redis-based caching layer
- Local file-based storage
- Custom backend integration

## Direct Provider Usage

```go
import "github.com/plexusone/omnimemory/provider/kvs"

provider, err := kvs.NewProvider(core.ProviderConfig{
    Options: map[string]any{
        "store":  store,
        "prefix": "memories",
    },
}, embedder)
if err != nil {
    log.Fatal(err)
}
defer provider.Close()

memory, err := provider.Add(ctx, &core.AddRequest{...})
```
