# Testing

OmniMemory provides a conformance test suite to validate provider implementations and ensure consistent behavior.

## Conformance Test Suite

The `providertest` package contains comprehensive tests for all providers:

```go
import "github.com/plexusone/omnimemory/core/providertest"

func TestMyProvider(t *testing.T) {
    provider := createMyProvider(t)
    providertest.RunAll(t, provider)
}
```

## Test Categories

### Interface Tests

Verify the provider implements all required methods:

```go
providertest.RunInterfaceTests(t, provider)
```

Checks:

- Provider is not nil
- `Name()` returns non-empty string
- All interface methods are implemented

### CRUD Tests

Test basic create, read, update, delete operations:

```go
providertest.RunCRUDTests(t, provider)
```

Covers:

- `Add` creates memories with correct fields
- `Get` retrieves memories by ID
- `Update` modifies existing memories
- `Delete` removes memories
- `List` returns all memories for a subject

### Behavior Tests

Test isolation and validation:

```go
providertest.RunBehaviorTests(t, provider)
```

Tests:

- **Tenant Isolation**: Memories from one tenant are not visible to another
- **Subject Isolation**: Memories from one subject are not visible to another
- **Validation Errors**: Missing required fields return proper errors

### Search Tests

Test semantic search functionality:

```go
providertest.RunSearchTests(t, provider)
```

Requires an embedder for similarity calculations.

## Running Tests

### Unit Tests

```bash
go test -v ./...
```

### With Coverage

```bash
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Specific Provider

```bash
go test -v ./provider/memory/...
go test -v ./provider/postgres/...
go test -v ./provider/kvs/...
```

## Writing Provider Tests

### In-Memory Provider

```go
package memory_test

import (
    "testing"

    "github.com/plexusone/omnimemory/core"
    "github.com/plexusone/omnimemory/core/providertest"
    "github.com/plexusone/omnimemory/provider/memory"
)

func TestConformance(t *testing.T) {
    // Create a mock embedder
    embedder := &mockEmbedder{dimension: 1536}

    provider, err := memory.NewProvider(core.ProviderConfig{}, embedder)
    if err != nil {
        t.Fatal(err)
    }
    defer provider.Close()

    providertest.RunAll(t, provider)
}

type mockEmbedder struct {
    dimension int
}

func (e *mockEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
    // Return deterministic embedding based on text hash
    vec := make([]float64, e.dimension)
    for i, c := range text {
        vec[i%e.dimension] += float64(c) / 1000.0
    }
    return vec, nil
}

func (e *mockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
    result := make([][]float64, len(texts))
    for i, text := range texts {
        emb, err := e.Embed(ctx, text)
        if err != nil {
            return nil, err
        }
        result[i] = emb
    }
    return result, nil
}

func (e *mockEmbedder) Dimension() int {
    return e.dimension
}
```

### PostgreSQL Provider (Integration)

```go
package postgres_test

import (
    "os"
    "testing"

    "github.com/plexusone/omnimemory/core"
    "github.com/plexusone/omnimemory/core/providertest"
    "github.com/plexusone/omnimemory/provider/postgres"
)

func TestConformance(t *testing.T) {
    connStr := os.Getenv("TEST_DATABASE_URL")
    if connStr == "" {
        t.Skip("TEST_DATABASE_URL not set")
    }

    provider, err := postgres.NewProvider(core.ProviderConfig{
        Options: map[string]any{
            "connection_string": connStr,
        },
    }, embedder)
    if err != nil {
        t.Fatal(err)
    }
    defer provider.Close()

    providertest.RunAll(t, provider)
}
```

### External Provider (Twilio)

```go
package omnimemory_test

import (
    "os"
    "testing"

    "github.com/plexusone/omnimemory/core"
    "github.com/plexusone/omnimemory/core/providertest"
    "github.com/plexusone/omni-twilio/omnimemory"
)

func TestConformance(t *testing.T) {
    accountSID := os.Getenv("TWILIO_ACCOUNT_SID")
    authToken := os.Getenv("TWILIO_AUTH_TOKEN")

    if accountSID == "" || authToken == "" {
        t.Skip("Twilio credentials not set")
    }

    provider, err := omnimemory.NewProvider(core.ProviderConfig{
        Options: map[string]any{
            "account_sid": accountSID,
            "auth_token":  authToken,
        },
    }, nil)
    if err != nil {
        t.Fatal(err)
    }
    defer provider.Close()

    providertest.RunAll(t, provider)
}
```

## Test Isolation

Each test uses unique tenant and subject IDs to prevent interference:

```go
func TestAdd(t *testing.T) {
    ctx := core.Context{
        TenantID:  fmt.Sprintf("test-tenant-%d", time.Now().UnixNano()),
        SubjectID: fmt.Sprintf("test-subject-%d", time.Now().UnixNano()),
    }
    // Test with isolated context
}
```

## CI Integration

### GitHub Actions

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: pgvector/pgvector:pg15
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - run: go test -v ./...
        env:
          TEST_DATABASE_URL: postgres://postgres:postgres@localhost/postgres
```

## Best Practices

1. **Use conformance tests**: Validate all providers consistently
2. **Isolate tests**: Use unique IDs to prevent interference
3. **Skip when missing**: Skip integration tests without credentials
4. **Clean up**: Use `defer provider.Close()` and clean test data
5. **Test edge cases**: Empty content, long text, special characters
