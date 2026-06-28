package core

import (
	"context"
)

// Provider is the interface that all memory providers must implement.
type Provider interface {
	// Add adds a new memory.
	Add(ctx context.Context, req *AddRequest) (*Memory, error)

	// Get retrieves a memory by ID.
	Get(ctx context.Context, req *GetRequest) (*Memory, error)

	// Update updates an existing memory.
	Update(ctx context.Context, req *UpdateRequest) (*Memory, error)

	// Delete deletes a memory by ID.
	Delete(ctx context.Context, req *DeleteRequest) error

	// List lists memories with optional filters.
	List(ctx context.Context, req *ListRequest) (*ListResponse, error)

	// Search performs semantic search on memories.
	Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error)

	// Recall retrieves relevant memories for a given query.
	Recall(ctx context.Context, req *RecallRequest) (*RecallResponse, error)

	// Close closes the provider and releases resources.
	Close() error

	// Name returns the provider name.
	Name() string
}

// ExtractorProvider extends Provider with extraction capabilities.
type ExtractorProvider interface {
	Provider

	// ExtractObservations extracts observations from a conversation.
	ExtractObservations(ctx context.Context, conversation string) ([]Observation, error)

	// ExtractFacts extracts facts from text.
	ExtractFacts(ctx context.Context, text string) ([]Fact, error)
}

// ProviderConfig is the configuration for initializing a provider.
type ProviderConfig struct {
	Name     ProviderName   `json:"name"`
	DSN      string         `json:"dsn,omitempty"`
	APIKey   string         `json:"api_key,omitempty"`
	Endpoint string         `json:"endpoint,omitempty"`
	Options  map[string]any `json:"options,omitempty"`
}

// ProviderName identifies a provider type.
type ProviderName string

const (
	ProviderNameMemory   ProviderName = "memory"
	ProviderNamePostgres ProviderName = "postgres"
	ProviderNameKVS      ProviderName = "kvs"
	ProviderNameMem0     ProviderName = "mem0"
	ProviderNameGraphiti ProviderName = "graphiti"
	ProviderNameTwilio   ProviderName = "twilio"
)

// Valid returns true if the provider name is valid.
func (n ProviderName) Valid() bool {
	switch n {
	case ProviderNameMemory, ProviderNamePostgres, ProviderNameKVS,
		ProviderNameMem0, ProviderNameGraphiti, ProviderNameTwilio:
		return true
	default:
		return false
	}
}

// String returns the string representation of the provider name.
func (n ProviderName) String() string {
	return string(n)
}
