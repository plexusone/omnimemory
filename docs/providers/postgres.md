# PostgreSQL Provider

The PostgreSQL provider uses PostgreSQL with the pgvector extension for production-grade semantic memory storage.

## Features

- Persistent storage
- Native vector similarity search via pgvector
- HNSW indexing for fast queries
- Full SQL query capabilities
- ACID transactions
- Horizontal scaling via read replicas

## Prerequisites

PostgreSQL 15+ with pgvector extension:

```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

## Installation

```go
import _ "github.com/plexusone/omnimemory/provider/postgres"
```

## Configuration

```go
import (
    "github.com/plexusone/omnimemory"
    "github.com/plexusone/omnimemory/core"
    _ "github.com/plexusone/omnimemory/provider/postgres"
)

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

### Options

| Option | Environment Variable | Description |
|--------|---------------------|-------------|
| `connection_string` | `DATABASE_URL` | PostgreSQL connection string |

### Connection String Format

```
postgres://user:password@host:port/database?sslmode=require
```

## Schema

The provider uses Ent ORM with the following schema:

```sql
CREATE TABLE memories (
    id VARCHAR PRIMARY KEY,
    tenant_id VARCHAR NOT NULL,
    subject_id VARCHAR NOT NULL,
    agent_id VARCHAR,
    session_id VARCHAR,
    scope VARCHAR NOT NULL,
    type VARCHAR NOT NULL,
    content TEXT NOT NULL,
    embedding vector(1536),
    metadata JSONB,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP
);

CREATE INDEX idx_memories_tenant ON memories(tenant_id);
CREATE INDEX idx_memories_tenant_subject ON memories(tenant_id, subject_id);
CREATE INDEX idx_memories_tenant_scope ON memories(tenant_id, scope);
CREATE INDEX idx_memories_embedding ON memories USING hnsw (embedding vector_l2_ops);
```

## Migrations

Run migrations automatically:

```go
import "github.com/plexusone/omnimemory/provider/postgres"

provider, err := postgres.NewProvider(config, embedder)
if err != nil {
    log.Fatal(err)
}

// Migrations run automatically on provider creation
```

Or manually with Ent:

```go
import "github.com/plexusone/omnimemory/ent"

client, err := ent.Open("postgres", connectionString)
if err != nil {
    log.Fatal(err)
}

if err := client.Schema.Create(context.Background()); err != nil {
    log.Fatal(err)
}
```

## Vector Search

The provider uses pgvector's L2 distance for similarity:

```go
results, err := client.Search(ctx, &core.SearchRequest{
    Context: core.Context{
        TenantID:  "tenant-1",
        SubjectID: "user-1",
    },
    Query:     "user preferences",
    Limit:     10,
    Threshold: 0.7, // Minimum similarity
})
```

### HNSW Index

The HNSW index provides approximate nearest neighbor search with configurable parameters:

```sql
-- Tune for your workload
SET hnsw.ef_search = 100; -- Higher = more accurate, slower
```

## Performance Tips

1. **Connection pooling**: Use pgbouncer or built-in pool
2. **Index maintenance**: Regularly VACUUM and ANALYZE
3. **Embedding dimension**: Smaller dimensions (384) are faster
4. **Batch operations**: Use transactions for bulk inserts

## Direct Provider Usage

```go
import "github.com/plexusone/omnimemory/provider/postgres"

provider, err := postgres.NewProvider(core.ProviderConfig{
    Options: map[string]any{
        "connection_string": os.Getenv("DATABASE_URL"),
    },
}, embedder)
if err != nil {
    log.Fatal(err)
}
defer provider.Close()

memory, err := provider.Add(ctx, &core.AddRequest{...})
```
