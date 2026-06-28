# mem0 Provider

The mem0 provider integrates with the [mem0](https://mem0.ai) API for managed semantic memory storage. This provider is available in the [mem0-go](https://github.com/plexusone/mem0-go) package.

## Features

- Managed semantic memory service
- Automatic embedding generation
- Memory search with semantic similarity
- No infrastructure to manage
- Built-in user and agent isolation

## Installation

```bash
go get github.com/plexusone/mem0-go
```

```go
import _ "github.com/plexusone/mem0-go/omnimemory"
```

## Configuration

### Environment Variables

```bash
export MEM0_API_KEY="your-mem0-api-key"
```

### Client Setup

```go
import (
    "github.com/plexusone/omnimemory"
    "github.com/plexusone/omnimemory/core"
    _ "github.com/plexusone/mem0-go/omnimemory"
)

client, err := omnimemory.NewClient(core.ClientConfig{
    Providers: []core.ProviderConfig{
        {
            Name:   core.ProviderNameMem0,
            APIKey: os.Getenv("MEM0_API_KEY"),
        },
    },
})
```

### Options

| Option | Environment Variable | Description |
|--------|---------------------|-------------|
| `api_key` | `MEM0_API_KEY` | mem0 API key (required) |
| `base_url` | - | Custom base URL (optional, defaults to api.mem0.ai) |

## Concept Mapping

OmniMemory concepts map to mem0 API:

| OmniMemory | mem0 API |
|------------|----------|
| TenantID | app_id |
| SubjectID | user_id |
| AgentID | agent_id |
| SessionID | run_id |
| Memory | MemoryItem |
| Search/Recall | Search API |

## Usage

### Adding Memories

```go
memory, err := client.Add(ctx, &core.AddRequest{
    Context: core.Context{
        TenantID:  "my-app",
        SubjectID: "user-123",
        AgentID:   "assistant",
    },
    Type:    core.MemoryTypeObservation,
    Content: "User mentioned they prefer dark mode interfaces",
})
```

### Recalling Memories

```go
recalled, err := client.Recall(ctx, &core.RecallRequest{
    Context: core.Context{
        TenantID:  "my-app",
        SubjectID: "user-123",
    },
    Query:      "What are the user's preferences?",
    MaxResults: 5,
})

for _, mem := range recalled.Memories {
    fmt.Printf("- %s\n", mem.Content)
}
```

### Searching Memories

```go
results, err := client.Search(ctx, &core.SearchRequest{
    Context: core.Context{
        TenantID:  "my-app",
        SubjectID: "user-123",
    },
    Query:     "interface preferences",
    Limit:     10,
    Threshold: 0.7,
})

for _, r := range results.Results {
    fmt.Printf("Score: %.2f - %s\n", r.Score, r.Memory.Content)
}
```

## mem0 Dashboard

1. **Get API Key**:
   - Sign up at [mem0.ai](https://mem0.ai)
   - Navigate to Settings → API Keys
   - Create a new API key

2. **View Memories**:
   - Use the mem0 dashboard to browse stored memories
   - Filter by user_id, agent_id, or app_id

## Multi-Tenancy

mem0 supports multi-tenancy via app_id:

```go
// Tenant A
client.Add(ctx, &core.AddRequest{
    Context: core.Context{
        TenantID:  "tenant-a",  // Maps to app_id
        SubjectID: "user-123",
    },
    Content: "Memory for tenant A",
})

// Tenant B (isolated)
client.Add(ctx, &core.AddRequest{
    Context: core.Context{
        TenantID:  "tenant-b",  // Different app_id
        SubjectID: "user-123",  // Same user, different tenant
    },
    Content: "Memory for tenant B",
})
```

## Error Handling

```go
memory, err := client.Add(ctx, req)
if err != nil {
    var validationErr *core.ValidationError
    if errors.As(err, &validationErr) {
        log.Printf("Validation error: %s", validationErr.Message)
        return
    }

    var providerErr *core.ProviderError
    if errors.As(err, &providerErr) {
        log.Printf("mem0 error: %v", providerErr.Err)
        return
    }
}
```

## Related

- [mem0-go Documentation](https://github.com/plexusone/mem0-go)
- [mem0 API Documentation](https://docs.mem0.ai)
