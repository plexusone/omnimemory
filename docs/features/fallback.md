# Fallback & Recovery

OmniMemory supports multiple providers with automatic fallback when the primary provider fails.

## Multi-Provider Configuration

Configure providers in priority order:

```go
client, err := omnimemory.NewClient(core.ClientConfig{
    Providers: []core.ProviderConfig{
        // Primary: PostgreSQL for production
        {
            Name: core.ProviderNamePostgres,
            Options: map[string]any{
                "connection_string": os.Getenv("DATABASE_URL"),
            },
        },
        // Fallback: In-memory if PostgreSQL fails
        {Name: core.ProviderNameMemory},
    },
})
```

## Fallback Behavior

When an operation fails on the primary provider:

1. Error is logged
2. Next provider in the list is tried
3. Process repeats until success or all providers exhausted
4. Final error is returned if all fail

```go
// If PostgreSQL is down, automatically uses in-memory
memory, err := client.Add(ctx, &core.AddRequest{
    Context: memCtx,
    Content: "User preference",
})
// Operations continue without manual intervention
```

## Provider Priorities

Providers have built-in priorities:

| Priority | Type | Description |
|----------|------|-------------|
| `PriorityThick` (10) | SDK-based | Full SDK implementations |
| `PriorityThin` (0) | HTTP-based | Lightweight implementations |

Higher priority providers are preferred when multiple implementations exist for the same backend.

## Error Handling

The client tracks errors and attempts fallback:

```go
memory, err := client.Add(ctx, req)
if err != nil {
    // All providers failed
    var providerErr *core.ProviderError
    if errors.As(err, &providerErr) {
        log.Printf("Provider %s failed: %v", providerErr.Provider, providerErr.Err)
    }
}
```

## Context Cancellation

Fallback respects context cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// If timeout occurs, stops trying fallback providers
memory, err := client.Add(ctx, req)
```

## Use Cases

### Development to Production

```go
providers := []core.ProviderConfig{
    {Name: core.ProviderNameMemory}, // Works without database
}

if os.Getenv("DATABASE_URL") != "" {
    // Add PostgreSQL as primary in production
    providers = append([]core.ProviderConfig{
        {Name: core.ProviderNamePostgres, Options: map[string]any{
            "connection_string": os.Getenv("DATABASE_URL"),
        }},
    }, providers...)
}

client, _ := omnimemory.NewClient(core.ClientConfig{
    Providers: providers,
})
```

### High Availability

```go
client, _ := omnimemory.NewClient(core.ClientConfig{
    Providers: []core.ProviderConfig{
        // Primary database
        {Name: core.ProviderNamePostgres, Options: map[string]any{
            "connection_string": os.Getenv("PRIMARY_DB"),
        }},
        // Replica database
        {Name: core.ProviderNamePostgres, Options: map[string]any{
            "connection_string": os.Getenv("REPLICA_DB"),
        }},
        // External service
        {Name: core.ProviderNameTwilio, Options: map[string]any{
            "account_sid": os.Getenv("TWILIO_ACCOUNT_SID"),
            "auth_token":  os.Getenv("TWILIO_AUTH_TOKEN"),
        }},
    },
})
```

### Testing with Real Data

```go
client, _ := omnimemory.NewClient(core.ClientConfig{
    Providers: []core.ProviderConfig{
        // Test against real provider
        {Name: core.ProviderNamePostgres, Options: map[string]any{
            "connection_string": os.Getenv("TEST_DATABASE_URL"),
        }},
        // Fall back to in-memory for CI
        {Name: core.ProviderNameMemory},
    },
})
```

## Logging

Enable debug logging to see fallback behavior:

```go
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

client, _ := omnimemory.NewClient(core.ClientConfig{
    Logger:    logger,
    Providers: providers,
})
```

Output:

```
level=WARN msg="primary provider failed, trying fallbacks" provider=postgres op=Add error="connection refused"
level=DEBUG msg="fallback provider succeeded" provider=memory op=Add
```

## Checking Active Provider

Query which providers are active:

```go
activeProviders := client.Providers()
fmt.Printf("Active providers: %v\n", activeProviders)
```

## Best Practices

1. **Order matters**: Put most reliable provider first
2. **Include a fallback**: Always have an in-memory fallback for resilience
3. **Monitor failures**: Log and alert on fallback usage
4. **Test failover**: Regularly test that fallback works
5. **Consider data consistency**: Fallback providers may not have the same data
