// Package kvs provides a KVS-backed Provider implementation using omnistorage-core.
//
// This provider stores memories as JSON blobs in a key-value store, supporting
// SQLite, Redis, and in-memory backends via omnistorage-core/kvs.
//
// # Key Format
//
// Memories are stored with keys in the format:
//
//	{prefix}:{tenant_id}:{subject_id}:{memory_id}
//
// This allows efficient listing by tenant and subject.
//
// # Semantic Search
//
// Since KVS backends don't support vector search natively, this provider
// performs in-memory similarity search by loading matching memories and
// computing cosine similarity on their embeddings.
//
// # Usage
//
//	import (
//	    "github.com/plexusone/omnimemory"
//	    "github.com/plexusone/omnimemory/core"
//	    "github.com/plexusone/omnimemory/provider/kvs"
//	    kvsmemory "github.com/plexusone/omnistorage-core/kvs/backend/memory"
//	)
//
//	func main() {
//	    // Create a KVS backend
//	    store := kvsmemory.New()
//
//	    // Create the provider
//	    provider, err := kvs.NewProvider(core.ProviderConfig{
//	        Options: map[string]any{
//	            "store": store,
//	            "key_prefix": "memories",
//	        },
//	    }, embedder)
//	}
package kvs

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/plexusone/omnimemory/core"
	"github.com/plexusone/omnistorage-core/kvs"
)

func init() {
	core.RegisterProvider(core.ProviderNameKVS, NewProvider, core.PriorityThin)
}

// Provider stores memories in a KVS backend.
type Provider struct {
	store     kvs.ListableStore
	embedder  core.Embedder
	keyPrefix string
	ownsStore bool // Whether we own the store and should close it
}

// NewProvider creates a new KVS Provider.
//
// The provider accepts the following options:
//
//   - store: A kvs.ListableStore instance (required)
//   - key_prefix: Prefix for all keys (default: "memories")
func NewProvider(config core.ProviderConfig, embedder core.Embedder) (core.Provider, error) {
	storeVal, ok := config.Options["store"]
	if !ok {
		return nil, core.NewValidationError("store", "KVS store is required")
	}

	store, ok := storeVal.(kvs.ListableStore)
	if !ok {
		return nil, core.NewValidationError("store", "store must implement kvs.ListableStore")
	}

	keyPrefix := "memories"
	if prefix, ok := config.Options["key_prefix"].(string); ok && prefix != "" {
		keyPrefix = prefix
	}

	return &Provider{
		store:     store,
		embedder:  embedder,
		keyPrefix: keyPrefix,
		ownsStore: false,
	}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return core.ProviderNameKVS.String()
}

// Close closes the provider.
func (p *Provider) Close() error {
	if p.ownsStore {
		return p.store.Close()
	}
	return nil
}

// memoryKey builds a key for a memory.
func (p *Provider) memoryKey(tenantID, subjectID, id string) string {
	return p.keyPrefix + ":" + tenantID + ":" + subjectID + ":" + id
}

// listPrefix returns the prefix for listing memories by tenant and optionally subject.
func (p *Provider) listPrefix(tenantID, subjectID string) string {
	if subjectID != "" {
		return p.keyPrefix + ":" + tenantID + ":" + subjectID + ":"
	}
	return p.keyPrefix + ":" + tenantID + ":"
}

// Add adds a new memory.
func (p *Provider) Add(ctx context.Context, req *core.AddRequest) (*core.Memory, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	id := uuid.New().String()

	memory := &core.Memory{
		ID:        id,
		TenantID:  req.TenantID,
		SubjectID: req.SubjectID,
		AgentID:   req.AgentID,
		SessionID: req.SessionID,
		Scope:     req.Scope,
		Type:      req.Type,
		Content:   req.Content,
		Metadata:  req.Metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Set expiration if TTL is provided
	if req.TTL > 0 {
		expiresAt := now.Add(req.TTL)
		memory.ExpiresAt = &expiresAt
	}

	// Generate embedding if embedder is available
	if p.embedder != nil {
		embedding, err := p.embedder.Embed(ctx, req.Content)
		if err != nil {
			return nil, core.NewProviderError(p.Name(), "Add", err)
		}
		memory.Embedding = embedding
	}

	// Serialize to JSON
	data, err := json.Marshal(memory)
	if err != nil {
		return nil, core.NewProviderError(p.Name(), "Add", err)
	}

	// Store in KVS
	key := p.memoryKey(req.TenantID, req.SubjectID, id)
	if err := p.store.Set(ctx, key, data, req.TTL); err != nil {
		return nil, core.NewProviderError(p.Name(), "Add", err)
	}

	return memory, nil
}

// Get retrieves a memory by ID.
func (p *Provider) Get(ctx context.Context, req *core.GetRequest) (*core.Memory, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	key := p.memoryKey(req.TenantID, req.SubjectID, req.ID)
	data, err := p.store.Get(ctx, key)
	if err != nil {
		if errors.Is(err, kvs.ErrNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, core.NewProviderError(p.Name(), "Get", err)
	}

	var memory core.Memory
	if err := json.Unmarshal(data, &memory); err != nil {
		return nil, core.NewProviderError(p.Name(), "Get", err)
	}

	// Check expiration (KVS should handle TTL, but double-check)
	if memory.ExpiresAt != nil && time.Now().After(*memory.ExpiresAt) {
		return nil, core.ErrNotFound
	}

	return &memory, nil
}

// Update updates an existing memory.
func (p *Provider) Update(ctx context.Context, req *core.UpdateRequest) (*core.Memory, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Get existing memory
	existing, err := p.Get(ctx, &core.GetRequest{
		Context: req.Context,
		ID:      req.ID,
	})
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Content != "" {
		existing.Content = req.Content

		// Regenerate embedding if content changed
		if p.embedder != nil {
			embedding, err := p.embedder.Embed(ctx, req.Content)
			if err != nil {
				return nil, core.NewProviderError(p.Name(), "Update", err)
			}
			existing.Embedding = embedding
		}
	}

	if req.Metadata != nil {
		if existing.Metadata == nil {
			existing.Metadata = make(map[string]any)
		}
		for k, v := range req.Metadata {
			existing.Metadata[k] = v
		}
	}

	existing.UpdatedAt = time.Now()

	// Serialize and store
	data, err := json.Marshal(existing)
	if err != nil {
		return nil, core.NewProviderError(p.Name(), "Update", err)
	}

	key := p.memoryKey(req.TenantID, req.SubjectID, req.ID)

	// Calculate remaining TTL if applicable
	var ttl time.Duration
	if existing.ExpiresAt != nil {
		ttl = time.Until(*existing.ExpiresAt)
		if ttl < 0 {
			ttl = 0
		}
	}

	if err := p.store.Set(ctx, key, data, ttl); err != nil {
		return nil, core.NewProviderError(p.Name(), "Update", err)
	}

	return existing, nil
}

// Delete deletes a memory by ID.
func (p *Provider) Delete(ctx context.Context, req *core.DeleteRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	key := p.memoryKey(req.TenantID, req.SubjectID, req.ID)
	if err := p.store.Delete(ctx, key); err != nil {
		return core.NewProviderError(p.Name(), "Delete", err)
	}

	return nil
}

// List lists memories with optional filters.
func (p *Provider) List(ctx context.Context, req *core.ListRequest) (*core.ListResponse, error) {
	if req.TenantID == "" {
		return nil, core.ErrTenantRequired
	}

	// List all keys matching the prefix
	prefix := p.listPrefix(req.TenantID, req.SubjectID)
	keys, err := p.store.List(ctx, prefix)
	if err != nil {
		return nil, core.NewProviderError(p.Name(), "List", err)
	}

	var memories []*core.Memory
	now := time.Now()

	for _, key := range keys {
		data, err := p.store.Get(ctx, key)
		if err != nil {
			if errors.Is(err, kvs.ErrNotFound) {
				continue // Key was deleted between List and Get
			}
			return nil, core.NewProviderError(p.Name(), "List", err)
		}

		var memory core.Memory
		if err := json.Unmarshal(data, &memory); err != nil {
			continue // Skip malformed entries
		}

		// Check expiration
		if memory.ExpiresAt != nil && now.After(*memory.ExpiresAt) {
			continue
		}

		// Filter by types
		if len(req.Types) > 0 && !containsType(req.Types, memory.Type) {
			continue
		}

		// Filter by scopes
		if len(req.Scopes) > 0 && !containsScope(req.Scopes, memory.Scope) {
			continue
		}

		memories = append(memories, &memory)
	}

	// Sort by created_at descending
	sort.Slice(memories, func(i, j int) bool {
		return memories[i].CreatedAt.After(memories[j].CreatedAt)
	})

	totalCount := len(memories)

	// Apply pagination
	if req.Offset > 0 && req.Offset < len(memories) {
		memories = memories[req.Offset:]
	} else if req.Offset >= len(memories) {
		memories = nil
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}

	hasMore := len(memories) > limit
	if len(memories) > limit {
		memories = memories[:limit]
	}

	return &core.ListResponse{
		Memories:   memories,
		TotalCount: totalCount,
		HasMore:    hasMore,
	}, nil
}

// Search performs semantic search on memories.
func (p *Provider) Search(ctx context.Context, req *core.SearchRequest) (*core.SearchResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Generate embedding for query
	var queryEmbedding []float64
	if p.embedder != nil {
		embedding, err := p.embedder.Embed(ctx, req.Query)
		if err != nil {
			return nil, core.NewProviderError(p.Name(), "Search", err)
		}
		queryEmbedding = embedding
	}

	// List all memories for this tenant/subject
	listResp, err := p.List(ctx, &core.ListRequest{
		Context: req.Context,
		Types:   req.Types,
		Scopes:  req.Scopes,
		Limit:   0, // Get all for scoring
	})
	if err != nil {
		return nil, err
	}

	type scoredMemory struct {
		memory *core.Memory
		score  float64
	}

	var scored []scoredMemory

	for _, memory := range listResp.Memories {
		// Calculate similarity score
		var score float64
		if len(queryEmbedding) > 0 && len(memory.Embedding) > 0 {
			score = core.CosineSimilarity(queryEmbedding, memory.Embedding)
		}

		// Apply threshold
		if req.Threshold > 0 && score < req.Threshold {
			continue
		}

		scored = append(scored, scoredMemory{memory: memory, score: score})
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Apply limit
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	if len(scored) > limit {
		scored = scored[:limit]
	}

	results := make([]*core.SearchResult, len(scored))
	for i, sm := range scored {
		results[i] = &core.SearchResult{
			Memory: sm.memory,
			Score:  sm.score,
		}
	}

	return &core.SearchResponse{
		Results: results,
	}, nil
}

// Recall retrieves relevant memories for a given query.
func (p *Provider) Recall(ctx context.Context, req *core.RecallRequest) (*core.RecallResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Use Search under the hood
	searchReq := &core.SearchRequest{
		Context: req.Context,
		Query:   req.Query,
		Types:   req.IncludeTypes,
		Limit:   req.MaxResults,
	}

	searchResp, err := p.Search(ctx, searchReq)
	if err != nil {
		return nil, err
	}

	memories := make([]*core.Memory, len(searchResp.Results))
	for i, r := range searchResp.Results {
		memories[i] = r.Memory
	}

	return &core.RecallResponse{
		Memories: memories,
	}, nil
}

// containsType checks if a slice contains a memory type.
func containsType(types []core.MemoryType, t core.MemoryType) bool {
	for _, mt := range types {
		if mt == t {
			return true
		}
	}
	return false
}

// containsScope checks if a slice contains a scope.
func containsScope(scopes []core.Scope, s core.Scope) bool {
	for _, sc := range scopes {
		if sc == s {
			return true
		}
	}
	return false
}
