# Architecture

OmniMemory follows a modular architecture with clear separation between the core API, provider implementations, and storage backends.

## Package Structure

```
omnimemory/
├── omnimemory.go          # Package entry point, re-exports
├── core/
│   ├── provider.go        # Provider interface
│   ├── types.go           # Memory, Context, Request/Response types
│   ├── scope.go           # Memory scopes
│   ├── registry.go        # Provider registry
│   ├── client.go          # Multi-provider client
│   ├── embedder.go        # Embedding interface
│   ├── errors.go          # Error types
│   └── providertest/      # Conformance test suite
├── ent/
│   ├── schema/            # Ent schema definitions
│   └── ...                # Generated Ent code
└── provider/
    ├── memory/            # In-memory provider
    ├── postgres/          # PostgreSQL+pgvector provider
    ├── kvs/               # KVS provider
    ├── mem0/              # Mem0 provider (stub)
    ├── graphiti/          # Graphiti provider (stub)
    └── twilio/            # Twilio provider (stub)
```

## Core Components

### Provider Interface

The `Provider` interface defines all memory operations:

```go
type Provider interface {
    // CRUD
    Add(ctx context.Context, req *AddRequest) (*Memory, error)
    Get(ctx context.Context, req *GetRequest) (*Memory, error)
    Update(ctx context.Context, req *UpdateRequest) (*Memory, error)
    Delete(ctx context.Context, req *DeleteRequest) error
    List(ctx context.Context, req *ListRequest) (*ListResponse, error)

    // Semantic
    Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error)
    Recall(ctx context.Context, req *RecallRequest) (*RecallResponse, error)

    // Lifecycle
    Close() error
    Name() string
}
```

### Registry

Providers register themselves via `init()`:

```go
func init() {
    core.RegisterProvider(
        core.ProviderNameMemory,
        func(cfg core.ProviderConfig, emb core.Embedder) (core.Provider, error) {
            return NewProvider(cfg, emb)
        },
        core.PriorityThin,
    )
}
```

### Client

The `Client` orchestrates multiple providers with fallback:

```
┌─────────────────────────────────────────────────────────┐
│                        Client                           │
├─────────────────────────────────────────────────────────┤
│  providers: map[ProviderName]Provider                   │
│  primary:   ProviderName                                │
│  fallbacks: []ProviderName                              │
│  logger:    *slog.Logger                                │
└─────────────────────────────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        ▼                 ▼                 ▼
   ┌─────────┐       ┌─────────┐       ┌─────────┐
   │PostgreSQL│       │ Memory  │       │  KVS    │
   │ Provider │       │ Provider│       │ Provider│
   └─────────┘       └─────────┘       └─────────┘
```

## Data Model

### Memory Entity

```go
type Memory struct {
    ID          string            // Unique identifier
    TenantID    string            // Organization/workspace
    SubjectID   string            // Who this memory is about
    AgentID     string            // Which agent stored it
    SessionID   string            // Conversation session
    Scope       Scope             // Visibility scope
    Type        MemoryType        // observation, fact, etc.
    Content     string            // Memory content
    Embedding   []float64         // Vector embedding
    Metadata    map[string]any    // Custom metadata
    CreatedAt   time.Time
    UpdatedAt   time.Time
    ExpiresAt   *time.Time        // Optional TTL
}
```

### Context

Every operation requires context for isolation:

```go
type Context struct {
    TenantID       string  // Required
    SubjectID      string  // Required
    PrincipalID    string  // Who is making the request
    AgentID        string  // Which agent
    SessionID      string  // Conversation session
    ConversationID string  // Conversation thread
    Scope          Scope   // Memory visibility
}
```

## Request Flow

```
┌──────────┐     ┌────────┐     ┌──────────┐     ┌─────────┐
│  Client  │────▶│ Client │────▶│ Provider │────▶│ Storage │
│   Code   │     │        │     │          │     │         │
└──────────┘     └────────┘     └──────────┘     └─────────┘
                     │               │
                     │    ┌──────────┴──────────┐
                     │    │      Embedder       │
                     │    │  (for Add/Search)   │
                     │    └─────────────────────┘
                     │
              ┌──────┴──────┐
              │   Fallback  │
              │  Provider   │
              └─────────────┘
```

## Storage Layer

### PostgreSQL

Uses Ent ORM with pgvector:

```
┌─────────────────────────────────────────┐
│              PostgreSQL                  │
├─────────────────────────────────────────┤
│  memories table                          │
│  ├── id (VARCHAR, PK)                   │
│  ├── tenant_id (VARCHAR, indexed)       │
│  ├── subject_id (VARCHAR, indexed)      │
│  ├── embedding (vector(1536), HNSW)     │
│  └── ...                                │
└─────────────────────────────────────────┘
```

### KVS

Stores JSON documents with prefix-based keys:

```
┌─────────────────────────────────────────┐
│            KVS Backend                   │
├─────────────────────────────────────────┤
│  Key: memories:tenant-1:user-1:mem-abc  │
│  Value: {"id":"mem-abc","content":...}  │
└─────────────────────────────────────────┘
```

## Extension Points

### Custom Providers

Implement `Provider` interface for new backends:

```go
type MyProvider struct{}

func (p *MyProvider) Add(...) (*Memory, error) { ... }
// ... implement all methods

func init() {
    core.RegisterProvider("my-provider", NewMyProvider, 0)
}
```

### Custom Embedders

Implement `Embedder` interface for custom embedding:

```go
type MyEmbedder struct{}

func (e *MyEmbedder) Embed(...) ([]float64, error) { ... }
func (e *MyEmbedder) EmbedBatch(...) ([][]float64, error) { ... }
func (e *MyEmbedder) Dimension() int { ... }
```

## Design Principles

1. **Interface-first**: Clean interfaces enable multiple implementations
2. **Multi-tenancy**: Built-in tenant and subject isolation
3. **Fallback-ready**: Multiple providers with automatic failover
4. **Embedding-agnostic**: Pluggable embedding providers
5. **Test-friendly**: Conformance tests validate implementations
