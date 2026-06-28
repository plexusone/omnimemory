# Multi-Tenancy

OmniMemory is designed for multi-tenant applications with built-in isolation between tenants and subjects.

## Context Structure

Every operation requires a `Context` that identifies ownership:

```go
type Context struct {
    TenantID       string  // Required: Organization/workspace
    SubjectID      string  // Required: Who the memory is about
    PrincipalID    string  // Optional: Who is making the request
    AgentID        string  // Optional: Which agent is acting
    SessionID      string  // Optional: Conversation session
    ConversationID string  // Optional: Conversation thread
    Scope          Scope   // Optional: Memory visibility
}
```

## Tenant Isolation

Memories are strictly isolated by `TenantID`:

```go
// Tenant A's memories
client.Add(ctx, &core.AddRequest{
    Context: core.Context{
        TenantID:  "tenant-a",
        SubjectID: "user-1",
    },
    Content: "Secret information for Tenant A",
})

// Tenant B cannot access Tenant A's memories
results, _ := client.Search(ctx, &core.SearchRequest{
    Context: core.Context{
        TenantID:  "tenant-b",  // Different tenant
        SubjectID: "user-1",    // Same subject ID
    },
    Query: "secret",
})
// results.Results is empty - no cross-tenant leakage
```

## Subject Isolation

Within a tenant, memories are isolated by `SubjectID`:

```go
// Add memory for user-1
client.Add(ctx, &core.AddRequest{
    Context: core.Context{
        TenantID:  "acme-corp",
        SubjectID: "user-1",
    },
    Content: "User 1's private preference",
})

// User-2 cannot access user-1's memories
results, _ := client.Search(ctx, &core.SearchRequest{
    Context: core.Context{
        TenantID:  "acme-corp",
        SubjectID: "user-2",  // Different subject
    },
    Query: "private",
})
// results.Results is empty
```

## Memory Scopes

Scopes provide visibility control within a tenant:

| Scope | Visibility |
|-------|------------|
| `user` | Personal to one user |
| `agent` | What an agent has learned |
| `tenant` | Organization-wide shared |
| `team` | Project/group level |
| `session` | Single conversation |
| `domain` | Domain-specific (support, sales) |

### Example: Shared Knowledge

```go
// Add tenant-wide knowledge
client.Add(ctx, &core.AddRequest{
    Context: core.Context{
        TenantID:  "acme-corp",
        SubjectID: "shared",
        Scope:     core.ScopeTenant,
    },
    Type:    core.MemoryTypeFact,
    Content: "Company policy: No meetings on Fridays",
})

// Any user can search tenant-scope memories
results, _ := client.Search(ctx, &core.SearchRequest{
    Context: core.Context{
        TenantID:  "acme-corp",
        SubjectID: "shared",
    },
    Query:  "meeting policy",
    Scopes: []core.Scope{core.ScopeTenant},
})
```

## Agent Memory

Track what agents have learned:

```go
// Store agent observations
client.Add(ctx, &core.AddRequest{
    Context: core.Context{
        TenantID:  "acme-corp",
        SubjectID: "user-123",
        AgentID:   "support-bot",
        Scope:     core.ScopeAgent,
    },
    Type:    core.MemoryTypeObservation,
    Content: "User prefers technical explanations",
})

// Query agent's knowledge about a user
results, _ := client.Search(ctx, &core.SearchRequest{
    Context: core.Context{
        TenantID:  "acme-corp",
        SubjectID: "user-123",
        AgentID:   "support-bot",
    },
    Query:  "communication style",
    Scopes: []core.Scope{core.ScopeAgent},
})
```

## Session Memory

Short-lived conversation context:

```go
sessionCtx := core.Context{
    TenantID:  "acme-corp",
    SubjectID: "user-123",
    SessionID: "sess-abc",
    Scope:     core.ScopeSession,
}

// Add session-specific memory
client.Add(ctx, &core.AddRequest{
    Context: sessionCtx,
    Type:    core.MemoryTypeObservation,
    Content: "User mentioned they're in a hurry",
    TTL:     30 * time.Minute, // Auto-expire
})
```

## Access Control Patterns

### Principal Validation

```go
func validateAccess(ctx core.Context, principal User) error {
    // Check tenant membership
    if !principal.BelongsToTenant(ctx.TenantID) {
        return errors.New("access denied: wrong tenant")
    }

    // Check subject access
    if ctx.SubjectID != principal.ID && !principal.IsAdmin() {
        return errors.New("access denied: cannot access other users")
    }

    return nil
}
```

### Audit Logging

```go
// Use PrincipalID to track who accessed what
client.Search(ctx, &core.SearchRequest{
    Context: core.Context{
        TenantID:    "acme-corp",
        SubjectID:   "user-123",
        PrincipalID: "admin-456", // Who is searching
    },
    Query: "sensitive information",
})
```

## Provider Mapping

Different providers use different terminology:

| OmniMemory | PostgreSQL | Twilio | Description |
|------------|------------|--------|-------------|
| TenantID | tenant_id column | Store ID | Organization |
| SubjectID | subject_id column | Profile ID | User/entity |
| SessionID | session_id column | - | Conversation |

## Best Practices

1. **Always set TenantID and SubjectID**: Never leave them empty
2. **Use meaningful IDs**: UUIDs or prefixed identifiers
3. **Validate at the edge**: Check access before calling OmniMemory
4. **Log access**: Use PrincipalID for audit trails
5. **Scope appropriately**: Use the narrowest scope that fits
