// Package memory provides an in-memory Provider implementation for testing.
package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/plexusone/omnimemory/core"
)

func init() {
	core.RegisterProvider(core.ProviderNameMemory, NewProvider, core.PriorityThin)
}

// Provider stores memories in-memory.
type Provider struct {
	memories map[string]*core.Memory
	mu       sync.RWMutex
	embedder core.Embedder
}

// NewProvider creates a new in-memory Provider.
func NewProvider(_ core.ProviderConfig, embedder core.Embedder) (core.Provider, error) {
	return &Provider{
		memories: make(map[string]*core.Memory),
		embedder: embedder,
	}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return core.ProviderNameMemory.String()
}

// Close closes the provider.
func (p *Provider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.memories = make(map[string]*core.Memory)
	return nil
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

	p.mu.Lock()
	p.memories[id] = memory
	p.mu.Unlock()

	return memory, nil
}

// Get retrieves a memory by ID.
func (p *Provider) Get(_ context.Context, req *core.GetRequest) (*core.Memory, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	p.mu.RLock()
	memory, ok := p.memories[req.ID]
	p.mu.RUnlock()

	if !ok {
		return nil, core.ErrNotFound
	}

	// Check tenant and subject isolation
	if memory.TenantID != req.TenantID || memory.SubjectID != req.SubjectID {
		return nil, core.ErrNotFound
	}

	// Check expiration
	if memory.ExpiresAt != nil && time.Now().After(*memory.ExpiresAt) {
		return nil, core.ErrNotFound
	}

	return memory, nil
}

// Update updates an existing memory.
func (p *Provider) Update(ctx context.Context, req *core.UpdateRequest) (*core.Memory, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	memory, ok := p.memories[req.ID]
	if !ok {
		return nil, core.ErrNotFound
	}

	// Check tenant and subject isolation
	if memory.TenantID != req.TenantID || memory.SubjectID != req.SubjectID {
		return nil, core.ErrNotFound
	}

	// Update fields
	if req.Content != "" {
		memory.Content = req.Content

		// Regenerate embedding if content changed
		if p.embedder != nil {
			embedding, err := p.embedder.Embed(ctx, req.Content)
			if err != nil {
				return nil, core.NewProviderError(p.Name(), "Update", err)
			}
			memory.Embedding = embedding
		}
	}

	if req.Metadata != nil {
		if memory.Metadata == nil {
			memory.Metadata = make(map[string]any)
		}
		for k, v := range req.Metadata {
			memory.Metadata[k] = v
		}
	}

	memory.UpdatedAt = time.Now()

	return memory, nil
}

// Delete deletes a memory by ID.
func (p *Provider) Delete(_ context.Context, req *core.DeleteRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	memory, ok := p.memories[req.ID]
	if !ok {
		return core.ErrNotFound
	}

	// Check tenant and subject isolation
	if memory.TenantID != req.TenantID || memory.SubjectID != req.SubjectID {
		return core.ErrNotFound
	}

	delete(p.memories, req.ID)
	return nil
}

// List lists memories with optional filters.
func (p *Provider) List(_ context.Context, req *core.ListRequest) (*core.ListResponse, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var results []*core.Memory
	now := time.Now()

	for _, memory := range p.memories {
		// Check tenant isolation
		if memory.TenantID != req.TenantID {
			continue
		}

		// Check subject if specified
		if req.SubjectID != "" && memory.SubjectID != req.SubjectID {
			continue
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

		results = append(results, memory)
	}

	// Sort by created_at descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})

	totalCount := len(results)

	// Apply pagination
	if req.Offset > 0 && req.Offset < len(results) {
		results = results[req.Offset:]
	} else if req.Offset >= len(results) {
		results = nil
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}

	hasMore := len(results) > limit
	if len(results) > limit {
		results = results[:limit]
	}

	return &core.ListResponse{
		Memories:   results,
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

	p.mu.RLock()
	defer p.mu.RUnlock()

	type scoredMemory struct {
		memory *core.Memory
		score  float64
	}

	var scored []scoredMemory
	now := time.Now()

	for _, memory := range p.memories {
		// Check tenant isolation
		if memory.TenantID != req.TenantID {
			continue
		}

		// Check subject if specified
		if req.SubjectID != "" && memory.SubjectID != req.SubjectID {
			continue
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
