# AWS Providers

AWS-based providers are available in the [omni-aws](https://github.com/plexusone/omni-aws) package, which provides adapters for various AWS services.

## DynamoDB Provider

The DynamoDB provider stores memories in a DynamoDB table with automatic TTL support and in-memory vector search.

### Features

- Fully managed, serverless storage
- Automatic scaling (pay-per-request billing)
- Built-in TTL for memory expiration
- Multi-tenant isolation via partition keys
- Auto-create table option for development
- Custom endpoint support for DynamoDB Local

### Installation

```bash
go get github.com/plexusone/omni-aws
```

```go
import _ "github.com/plexusone/omni-aws/omnimemory/dynamodb"
```

### Configuration

#### Environment Variables

```bash
export AWS_REGION="us-east-1"
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
```

#### Client Setup

```go
import (
    "github.com/plexusone/omnimemory"
    "github.com/plexusone/omnimemory/core"
    _ "github.com/plexusone/omni-aws/omnimemory/dynamodb"
)

client, err := omnimemory.NewClient(core.ClientConfig{
    Providers: []core.ProviderConfig{
        {
            Name: core.ProviderNameAWSDynamoDB,
            Options: map[string]any{
                "table_name":   "omnimemory",
                "region":       "us-east-1",
                "create_table": true, // Auto-create for development
            },
        },
    },
})
```

#### Options

| Option | Description |
|--------|-------------|
| `table_name` | DynamoDB table name (required) |
| `region` | AWS region (optional, uses default config) |
| `endpoint` | Custom endpoint URL for DynamoDB Local (optional) |
| `create_table` | Auto-create table if not exists (default: false) |

### Table Schema

The provider uses a single-table design:

| Attribute | Type | Description |
|-----------|------|-------------|
| `pk` (Partition Key) | String | `tenant_id` |
| `sk` (Sort Key) | String | `subject_id#memory_id` |
| `expires_at` | Number | TTL attribute (Unix timestamp) |

This schema provides:

- Tenant isolation via partition key
- Efficient queries by subject within a tenant
- Automatic expiration via DynamoDB TTL

### Usage

#### Adding Memories

```go
memory, err := client.Add(ctx, &core.AddRequest{
    Context: core.Context{
        TenantID:  "tenant-123",
        SubjectID: "user-456",
    },
    Type:    core.MemoryTypeObservation,
    Content: "User prefers dark mode interfaces",
    TTL:     24 * time.Hour, // Expires in 24 hours
})
```

#### Searching Memories

```go
results, err := client.Search(ctx, &core.SearchRequest{
    Context: core.Context{
        TenantID:  "tenant-123",
        SubjectID: "user-456",
    },
    Query:     "interface preferences",
    Limit:     10,
    Threshold: 0.7,
})
```

### Local Development

Use DynamoDB Local for development:

```bash
# Start DynamoDB Local
docker run -p 8000:8000 amazon/dynamodb-local
```

```go
client, err := omnimemory.NewClient(core.ClientConfig{
    Providers: []core.ProviderConfig{
        {
            Name: core.ProviderNameAWSDynamoDB,
            Options: map[string]any{
                "table_name":   "omnimemory",
                "endpoint":     "http://localhost:8000",
                "create_table": true,
            },
        },
    },
})
```

### Semantic Search

Since DynamoDB doesn't support native vector search, the provider performs in-memory cosine similarity search:

1. Query retrieves all memories for the tenant/subject
2. Embeddings are compared using cosine similarity
3. Results are sorted by score and filtered by threshold

For production workloads requiring native vector search, consider:

- PostgreSQL with pgvector
- AWS OpenSearch with k-NN (planned)

### IAM Permissions

Required IAM permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "dynamodb:GetItem",
                "dynamodb:PutItem",
                "dynamodb:UpdateItem",
                "dynamodb:DeleteItem",
                "dynamodb:Query",
                "dynamodb:Scan"
            ],
            "Resource": "arn:aws:dynamodb:*:*:table/omnimemory"
        }
    ]
}
```

For `create_table: true`, also add:

```json
{
    "Action": [
        "dynamodb:CreateTable",
        "dynamodb:DescribeTable",
        "dynamodb:UpdateTimeToLive"
    ],
    "Resource": "arn:aws:dynamodb:*:*:table/omnimemory"
}
```

## Future Providers

Additional AWS providers are planned:

- **AWS S3**: Simple object storage for archival
- **AWS OpenSearch**: Native k-NN vector search

## Related

- [omni-aws Documentation](https://plexusone.github.io/omni-aws/)
- [DynamoDB Developer Guide](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/)
