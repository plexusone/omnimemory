// Package omnimemory provides a vendor-neutral memory abstraction layer
// for multiple memory backends (PostgreSQL+pgvector, Mem0, Graphiti, Twilio Memory).
//
// # Overview
//
// OmniMemory provides a stable API for storing and retrieving memories
// with semantic search capabilities. It supports multiple backends through
// a provider interface, allowing you to switch between implementations
// without changing application code.
//
// # Quick Start
//
//	import (
//	    "github.com/plexusone/omnimemory"
//	    "github.com/plexusone/omnimemory/core"
//	    _ "github.com/plexusone/omnimemory/provider/memory"   // Register in-memory provider
//	    _ "github.com/plexusone/omnimemory/provider/postgres" // Register PostgreSQL provider
//	)
//
//	func main() {
//	    // Create a client with in-memory provider (for testing)
//	    client, err := omnimemory.NewClient(core.ClientConfig{
//	        Providers: []core.ProviderConfig{
//	            {Name: core.ProviderNameMemory},
//	        },
//	    })
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    defer client.Close()
//
//	    // Add a memory
//	    memory, err := client.Add(ctx, &core.AddRequest{
//	        Context: core.Context{
//	            TenantID:  "tenant-1",
//	            SubjectID: "user-123",
//	        },
//	        Type:    core.MemoryTypeFact,
//	        Content: "User prefers dark mode",
//	    })
//
//	    // Search memories
//	    results, err := client.Search(ctx, &core.SearchRequest{
//	        Context: core.Context{
//	            TenantID:  "tenant-1",
//	            SubjectID: "user-123",
//	        },
//	        Query: "user interface preferences",
//	        Limit: 10,
//	    })
//	}
//
// # Memory Types
//
// OmniMemory supports different types of memories:
//   - Observation: Observed behavior or interaction
//   - Fact: Verified piece of information
//   - Preference: User preference
//   - Summary: Summarized conversation or topic
//   - Trait: Personality trait or characteristic
//   - Relationship: Relationship between entities
//
// # Memory Scopes
//
// Memories can be scoped at different levels:
//   - User: Personal to one user
//   - Agent: What an agent has learned
//   - Tenant: Org-level shared memories
//   - Team: Group/project level memories
//   - Session: Short-lived conversation memories
//   - Domain: Domain-specific memories (support, sales, etc.)
//
// # Providers
//
// Available providers:
//   - memory: In-memory provider for testing
//   - postgres: PostgreSQL+pgvector for production
//   - mem0: Mem0 adapter (future)
//   - graphiti: Graphiti temporal knowledge graph (future)
//   - twilio: Twilio Memory API (future)
package omnimemory

import (
	"github.com/plexusone/omnimemory/core"
)

// Re-export core types for convenience.
type (
	// Memory is the primary storage unit.
	Memory = core.Memory

	// MemoryType defines the category of memory content.
	MemoryType = core.MemoryType

	// Scope defines the visibility and ownership level of a memory.
	Scope = core.Scope

	// Provider is the interface that all memory providers must implement.
	Provider = core.Provider

	// Embedder generates embeddings for text.
	Embedder = core.Embedder

	// Client is a multi-provider memory client with fallback support.
	Client = core.Client

	// ClientConfig is the configuration for creating a Client.
	ClientConfig = core.ClientConfig

	// ProviderConfig is the configuration for initializing a provider.
	ProviderConfig = core.ProviderConfig

	// ProviderName identifies a provider type.
	ProviderName = core.ProviderName

	// Context provides multi-tenancy context for all operations.
	Context = core.Context

	// AddRequest is the request for adding a new memory.
	AddRequest = core.AddRequest

	// GetRequest is the request for getting a memory by ID.
	GetRequest = core.GetRequest

	// UpdateRequest is the request for updating a memory.
	UpdateRequest = core.UpdateRequest

	// DeleteRequest is the request for deleting a memory.
	DeleteRequest = core.DeleteRequest

	// ListRequest is the request for listing memories.
	ListRequest = core.ListRequest

	// ListResponse is the response for listing memories.
	ListResponse = core.ListResponse

	// SearchRequest is the request for semantic search.
	SearchRequest = core.SearchRequest

	// SearchResponse is the response for semantic search.
	SearchResponse = core.SearchResponse

	// SearchResult is a single search result with similarity score.
	SearchResult = core.SearchResult

	// RecallRequest is the request for recalling relevant memories.
	RecallRequest = core.RecallRequest

	// RecallResponse is the response for recalling memories.
	RecallResponse = core.RecallResponse

	// Observation represents an observed behavior or interaction.
	Observation = core.Observation

	// Fact represents a verified piece of information.
	Fact = core.Fact
)

// Re-export memory types.
const (
	MemoryTypeObservation  = core.MemoryTypeObservation
	MemoryTypeFact         = core.MemoryTypeFact
	MemoryTypePreference   = core.MemoryTypePreference
	MemoryTypeSummary      = core.MemoryTypeSummary
	MemoryTypeTrait        = core.MemoryTypeTrait
	MemoryTypeRelationship = core.MemoryTypeRelationship
)

// Re-export scopes.
const (
	ScopeUser    = core.ScopeUser
	ScopeAgent   = core.ScopeAgent
	ScopeTenant  = core.ScopeTenant
	ScopeTeam    = core.ScopeTeam
	ScopeSession = core.ScopeSession
	ScopeDomain  = core.ScopeDomain
)

// Re-export provider names.
const (
	ProviderNameMemory   = core.ProviderNameMemory
	ProviderNamePostgres = core.ProviderNamePostgres
	ProviderNameMem0     = core.ProviderNameMem0
	ProviderNameGraphiti = core.ProviderNameGraphiti
	ProviderNameTwilio   = core.ProviderNameTwilio
)

// Re-export priority constants.
const (
	PriorityThin  = core.PriorityThin
	PriorityThick = core.PriorityThick
)

// Re-export errors.
var (
	ErrNotFound         = core.ErrNotFound
	ErrInvalidInput     = core.ErrInvalidInput
	ErrProviderNotFound = core.ErrProviderNotFound
	ErrNoProviders      = core.ErrNoProviders
	ErrEmbeddingFailed  = core.ErrEmbeddingFailed
	ErrTenantRequired   = core.ErrTenantRequired
	ErrSubjectRequired  = core.ErrSubjectRequired
	ErrContentRequired  = core.ErrContentRequired
)

// NewClient creates a new multi-provider memory client.
func NewClient(config ClientConfig) (*Client, error) {
	return core.NewClient(config)
}

// RegisterProvider registers a provider factory with the global registry.
func RegisterProvider(name ProviderName, factory core.ProviderFactory, priority int) {
	core.RegisterProvider(name, factory, priority)
}

// GetProvider creates a provider instance from the global registry.
func GetProvider(name ProviderName, config ProviderConfig, embedder Embedder) (Provider, error) {
	return core.GetProvider(name, config, embedder)
}

// ListProviders returns all registered provider names.
func ListProviders() []ProviderName {
	return core.ListProviders()
}

// CosineSimilarity computes the cosine similarity between two vectors.
func CosineSimilarity(a, b []float64) float64 {
	return core.CosineSimilarity(a, b)
}

// EuclideanDistance computes the Euclidean distance between two vectors.
func EuclideanDistance(a, b []float64) float64 {
	return core.EuclideanDistance(a, b)
}
