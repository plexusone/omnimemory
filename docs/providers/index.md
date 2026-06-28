# Providers Overview

OmniMemory uses a provider architecture that allows pluggable storage backends. Each provider implements the same `Provider` interface, ensuring consistent behavior across different storage solutions.

## Available Providers

| Provider | Package | Use Case |
|----------|---------|----------|
| [In-Memory](memory.md) | `provider/memory` | Testing, development, ephemeral storage |
| [PostgreSQL](postgres.md) | `provider/postgres` | Production with pgvector |
| [KVS](kvs.md) | `provider/kvs` | Flexible key-value backends |
| [AWS DynamoDB](aws.md) | External: `omni-aws` | Serverless, auto-scaling |
| [mem0](mem0.md) | External: `mem0-go` | Managed memory service |
| [Twilio](twilio.md) | External: `omni-twilio` | Twilio Memory API |

## Provider Interface

All providers implement the `core.Provider` interface:

```go
type Provider interface {
    // Core CRUD operations
    Add(ctx context.Context, req *AddRequest) (*Memory, error)
    Get(ctx context.Context, req *GetRequest) (*Memory, error)
    Update(ctx context.Context, req *UpdateRequest) (*Memory, error)
    Delete(ctx context.Context, req *DeleteRequest) error
    List(ctx context.Context, req *ListRequest) (*ListResponse, error)

    // Semantic operations
    Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error)
    Recall(ctx context.Context, req *RecallRequest) (*RecallResponse, error)

    // Lifecycle
    Close() error
    Name() string
}
```

## Provider Registration

Providers register themselves via `init()`:

```go
import (
    "github.com/plexusone/omnimemory/core"
    _ "github.com/plexusone/omnimemory/provider/memory"  // Registers "memory"
    _ "github.com/plexusone/omnimemory/provider/postgres" // Registers "postgres"
)

// List registered providers
providers := core.ListProviders()
// ["memory", "postgres"]
```

## Choosing a Provider

### Development & Testing

Use the **In-Memory** provider for fast iteration:

```go
{Name: core.ProviderNameMemory}
```

### Production

Use **PostgreSQL** with pgvector for production workloads:

```go
{Name: core.ProviderNamePostgres, Options: map[string]any{
    "connection_string": os.Getenv("DATABASE_URL"),
}}
```

### External Services

Use **AWS DynamoDB** for serverless, auto-scaling storage:

```go
{Name: core.ProviderNameAWSDynamoDB, Options: map[string]any{
    "table_name": "omnimemory",
    "region":     "us-east-1",
}}
```

Use **mem0** for managed memory with automatic embeddings:

```go
{Name: core.ProviderNameMem0, APIKey: os.Getenv("MEM0_API_KEY")}
```

Use **Twilio Memory** for managed semantic memory:

```go
{Name: core.ProviderNameTwilio, Options: map[string]any{
    "account_sid": os.Getenv("TWILIO_ACCOUNT_SID"),
    "auth_token":  os.Getenv("TWILIO_AUTH_TOKEN"),
}}
```

## Multi-Provider Setup

Configure multiple providers for fallback:

```go
client, err := omnimemory.NewClient(core.ClientConfig{
    Providers: []core.ProviderConfig{
        // Primary: PostgreSQL for production
        {Name: core.ProviderNamePostgres, Options: map[string]any{
            "connection_string": os.Getenv("DATABASE_URL"),
        }},
        // Fallback: In-memory if PostgreSQL fails
        {Name: core.ProviderNameMemory},
    },
})
```

## Custom Providers

Implement `core.Provider` for custom backends:

```go
type MyProvider struct {
    // ...
}

func (p *MyProvider) Name() string {
    return "my-provider"
}

func (p *MyProvider) Add(ctx context.Context, req *core.AddRequest) (*core.Memory, error) {
    // Implementation
}

// ... implement remaining methods

func init() {
    core.RegisterProvider("my-provider", func(cfg core.ProviderConfig) (core.Provider, error) {
        return NewMyProvider(cfg)
    }, core.PriorityThin)
}
```

## Conformance Testing

Validate custom providers with the conformance test suite:

```go
import "github.com/plexusone/omnimemory/core/providertest"

func TestMyProvider(t *testing.T) {
    provider := createMyProvider(t)
    providertest.RunAll(t, provider)
}
```

See [Testing](../testing.md) for details.
