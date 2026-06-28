# Twilio Memory Provider

The Twilio Memory provider integrates with Twilio's Memory API for managed semantic memory storage. This provider is available in the [omni-twilio](https://github.com/plexusone/omni-twilio) package.

## Features

- Managed semantic memory service
- Automatic embedding generation
- Recall API with summarization
- No infrastructure to manage
- Built-in multi-tenancy via Stores/Profiles

## Installation

```bash
go get github.com/plexusone/omni-twilio
```

```go
import _ "github.com/plexusone/omni-twilio/omnimemory"
```

## Configuration

### Environment Variables

```bash
export TWILIO_ACCOUNT_SID="ACxxxxxxxx"
export TWILIO_AUTH_TOKEN="your-auth-token"
```

### Client Setup

```go
import (
    "github.com/plexusone/omnimemory"
    "github.com/plexusone/omnimemory/core"
    _ "github.com/plexusone/omni-twilio/omnimemory"
)

client, err := omnimemory.NewClient(core.ClientConfig{
    Providers: []core.ProviderConfig{
        {
            Name: core.ProviderNameTwilio,
            Options: map[string]any{
                "account_sid": os.Getenv("TWILIO_ACCOUNT_SID"),
                "auth_token":  os.Getenv("TWILIO_AUTH_TOKEN"),
            },
        },
    },
})
```

### Options

| Option | Environment Variable | Description |
|--------|---------------------|-------------|
| `account_sid` | `TWILIO_ACCOUNT_SID` | Twilio Account SID (required) |
| `auth_token` | `TWILIO_AUTH_TOKEN` | Twilio Auth Token (required) |

## Concept Mapping

OmniMemory concepts map to Twilio Memory API:

| OmniMemory | Twilio Memory API |
|------------|-------------------|
| TenantID | Store ID |
| SubjectID | Profile ID |
| Memory | Observation |
| Search/Recall | Recall API |

## Usage

### Adding Memories

```go
memory, err := client.Add(ctx, &core.AddRequest{
    Context: core.Context{
        TenantID:  "your-store-id",   // Twilio Store ID
        SubjectID: "your-profile-id", // Twilio Profile ID
    },
    Type:    core.MemoryTypeObservation,
    Content: "User mentioned they prefer dark mode interfaces",
})
```

### Recalling Memories

```go
recalled, err := client.Recall(ctx, &core.RecallRequest{
    Context: core.Context{
        TenantID:  "your-store-id",
        SubjectID: "your-profile-id",
    },
    Query:      "What are the user's preferences?",
    MaxResults: 5,
})

fmt.Printf("Summary: %s\n", recalled.Summary)
for _, mem := range recalled.Memories {
    fmt.Printf("- %s\n", mem.Content)
}
```

### Searching Memories

```go
results, err := client.Search(ctx, &core.SearchRequest{
    Context: core.Context{
        TenantID:  "your-store-id",
        SubjectID: "your-profile-id",
    },
    Query:     "interface preferences",
    Limit:     10,
    Threshold: 0.7,
})
```

## Twilio Console Setup

1. **Create a Memory Store**:
   - Go to Twilio Console -> Memory -> Stores
   - Create a new store and note the Store ID

2. **Create a Profile**:
   - Within your store, create a profile
   - Note the Profile ID

3. **Set Credentials**:
   - Use your Account SID and Auth Token from Twilio Console

## Testing

Run conformance tests:

```bash
export TWILIO_ACCOUNT_SID="ACxxxxxxxx"
export TWILIO_AUTH_TOKEN="your-token"
export TWILIO_MEMORY_STORE_ID="your-store-id"
export TWILIO_MEMORY_PROFILE_ID="your-profile-id"

go test -v github.com/plexusone/omni-twilio/omnimemory/...
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
        log.Printf("Twilio error: %v", providerErr.Err)
        return
    }
}
```

## Related

- [omni-twilio Documentation](https://plexusone.github.io/omni-twilio/)
- [Twilio Memory API Documentation](https://www.twilio.com/docs/memory)
