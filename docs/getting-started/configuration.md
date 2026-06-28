# Configuration

## Client Configuration

The `ClientConfig` struct controls client behavior:

```go
client, err := omnimemory.NewClient(core.ClientConfig{
    Providers: []core.ProviderConfig{
        {Name: core.ProviderNamePostgres, Options: map[string]any{
            "connection_string": os.Getenv("DATABASE_URL"),
        }},
        {Name: core.ProviderNameMemory}, // Fallback
    },
    DefaultTTL: 24 * time.Hour,
    Logger:     slog.Default(),
})
```

### Options

| Field | Type | Description |
|-------|------|-------------|
| `Providers` | `[]ProviderConfig` | Ordered list of providers (first is primary) |
| `DefaultTTL` | `time.Duration` | Default memory expiration time |
| `Logger` | `*slog.Logger` | Logger for operations |

## Provider Configuration

Each provider accepts specific options:

### In-Memory

```go
core.ProviderConfig{
    Name: core.ProviderNameMemory,
    // No options required
}
```

### PostgreSQL

```go
core.ProviderConfig{
    Name: core.ProviderNamePostgres,
    Options: map[string]any{
        "connection_string": "postgres://user:pass@localhost/db",
    },
}
```

### KVS

```go
import "github.com/plexusone/omnistorage-core/kvs"

store := kvs.NewMemoryStore() // Or any ListableStore

core.ProviderConfig{
    Name: core.ProviderNameKVS,
    Options: map[string]any{
        "store":  store,
        "prefix": "memories",
    },
}
```

### Twilio Memory

```go
core.ProviderConfig{
    Name: core.ProviderNameTwilio,
    Options: map[string]any{
        "account_sid": os.Getenv("TWILIO_ACCOUNT_SID"),
        "auth_token":  os.Getenv("TWILIO_AUTH_TOKEN"),
    },
}
```

## Embedder Configuration

Configure the embedding model for semantic search:

```go
import "github.com/plexusone/omnimemory/core"

embedder, err := core.NewOmniLLMEmbedder(core.EmbedderConfig{
    Provider: "openai",
    APIKey:   os.Getenv("OPENAI_API_KEY"),
    Model:    "text-embedding-3-small",
})

// Pass to provider
provider, err := postgres.NewProvider(config, embedder)
```

### Embedder Options

| Field | Type | Description |
|-------|------|-------------|
| `Provider` | `string` | LLM provider (openai, anthropic, etc.) |
| `APIKey` | `string` | API key for the provider |
| `Model` | `string` | Embedding model name |
| `Dimension` | `int` | Vector dimension (default: 1536) |

## Environment Variables

Common environment variables:

```bash
# PostgreSQL
export DATABASE_URL="postgres://user:pass@localhost/omnimemory"

# Twilio
export TWILIO_ACCOUNT_SID="ACxxxxxxxx"
export TWILIO_AUTH_TOKEN="your-auth-token"

# Embeddings
export OPENAI_API_KEY="sk-..."
```

## Logging

Configure structured logging:

```go
import "log/slog"

logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

client, err := omnimemory.NewClient(core.ClientConfig{
    Logger: logger,
    // ...
})
```

## Context Fields

The `Context` struct identifies the memory owner:

```go
ctx := core.Context{
    TenantID:       "acme-corp",     // Required: organization/workspace
    SubjectID:      "user-123",       // Required: who the memory is about
    PrincipalID:    "admin-456",      // Optional: who is making the request
    AgentID:        "support-bot",    // Optional: which agent is storing
    SessionID:      "sess-789",       // Optional: conversation session
    ConversationID: "conv-abc",       // Optional: conversation thread
    Scope:          core.ScopeUser,   // Optional: memory scope
}
```

## Memory Scopes

Control visibility with scopes:

| Scope | Description |
|-------|-------------|
| `user` | Personal to one user |
| `agent` | What an agent has learned |
| `tenant` | Organization-level shared |
| `team` | Group/project level |
| `session` | Short-lived conversation |
| `domain` | Domain-specific (support, sales) |
