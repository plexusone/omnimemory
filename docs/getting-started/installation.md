# Installation

## Requirements

- Go 1.21 or later
- PostgreSQL 15+ with pgvector extension (for PostgreSQL provider)

## Install Package

```bash
go get github.com/plexusone/omnimemory
```

## Provider-Specific Dependencies

### In-Memory Provider

No additional dependencies required. Ideal for testing and development.

```go
import _ "github.com/plexusone/omnimemory/provider/memory"
```

### PostgreSQL Provider

Requires PostgreSQL with the pgvector extension installed.

```bash
# Install pgvector in PostgreSQL
CREATE EXTENSION IF NOT EXISTS vector;
```

```go
import _ "github.com/plexusone/omnimemory/provider/postgres"
```

### KVS Provider

Requires omnistorage-core for the underlying key-value store.

```bash
go get github.com/plexusone/omnistorage-core
```

```go
import _ "github.com/plexusone/omnimemory/provider/kvs"
```

### Twilio Memory Provider

Available as a separate module in omni-twilio.

```bash
go get github.com/plexusone/omni-twilio
```

```go
import _ "github.com/plexusone/omni-twilio/omnimemory"
```

## Embedding Generation

OmniMemory requires an embedder for semantic search. The default uses omnillm-core:

```bash
go get github.com/plexusone/omnillm-core
```

Configure your embedding provider:

```bash
export OPENAI_API_KEY="your-api-key"
```

## Verify Installation

```go
package main

import (
    "fmt"

    "github.com/plexusone/omnimemory/core"
)

func main() {
    providers := core.ListProviders()
    fmt.Printf("Available providers: %v\n", providers)
}
```
