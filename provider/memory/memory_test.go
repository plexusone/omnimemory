package memory

import (
	"context"
	"testing"
	"time"

	"github.com/plexusone/omnimemory/core"
)

// mockEmbedder returns deterministic embeddings for testing.
type mockEmbedder struct {
	dimension int
}

func (e *mockEmbedder) Embed(_ context.Context, text string) ([]float64, error) {
	// Generate a simple deterministic embedding based on text length
	embedding := make([]float64, e.dimension)
	for i := range embedding {
		embedding[i] = float64(len(text)+i) / float64(e.dimension)
	}
	return embedding, nil
}

func (e *mockEmbedder) EmbedBatch(_ context.Context, texts []string) ([][]float64, error) {
	embeddings := make([][]float64, len(texts))
	for i, text := range texts {
		emb, err := e.Embed(context.Background(), text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = emb
	}
	return embeddings, nil
}

func (e *mockEmbedder) Dimension() int {
	return e.dimension
}

func newTestProvider(t *testing.T) *Provider {
	t.Helper()
	embedder := &mockEmbedder{dimension: 8}
	provider, err := NewProvider(core.ProviderConfig{}, embedder)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}
	return provider.(*Provider)
}

func TestProvider_Name(t *testing.T) {
	p := newTestProvider(t)
	if p.Name() != "memory" {
		t.Errorf("expected name 'memory', got %q", p.Name())
	}
}

func TestProvider_Add(t *testing.T) {
	p := newTestProvider(t)
	ctx := context.Background()

	req := &core.AddRequest{
		Context: core.Context{
			TenantID:  "tenant-1",
			SubjectID: "user-123",
		},
		Type:    core.MemoryTypeFact,
		Content: "User prefers dark mode",
	}

	mem, err := p.Add(ctx, req)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if mem.ID == "" {
		t.Error("expected non-empty ID")
	}
	if mem.TenantID != "tenant-1" {
		t.Errorf("expected TenantID 'tenant-1', got %q", mem.TenantID)
	}
	if mem.SubjectID != "user-123" {
		t.Errorf("expected SubjectID 'user-123', got %q", mem.SubjectID)
	}
	if mem.Type != core.MemoryTypeFact {
		t.Errorf("expected Type 'fact', got %q", mem.Type)
	}
	if mem.Content != "User prefers dark mode" {
		t.Errorf("unexpected Content: %q", mem.Content)
	}
	if len(mem.Embedding) != 8 {
		t.Errorf("expected embedding dimension 8, got %d", len(mem.Embedding))
	}
}

func TestProvider_AddWithTTL(t *testing.T) {
	p := newTestProvider(t)
	ctx := context.Background()

	req := &core.AddRequest{
		Context: core.Context{
			TenantID:  "tenant-1",
			SubjectID: "user-123",
		},
		Type:    core.MemoryTypeFact,
		Content: "Temporary memory",
		TTL:     time.Hour,
	}

	mem, err := p.Add(ctx, req)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if mem.ExpiresAt == nil {
		t.Error("expected ExpiresAt to be set")
	}

	expectedExpiry := time.Now().Add(time.Hour)
	if mem.ExpiresAt.Before(expectedExpiry.Add(-time.Minute)) ||
		mem.ExpiresAt.After(expectedExpiry.Add(time.Minute)) {
		t.Errorf("ExpiresAt outside expected range: %v", mem.ExpiresAt)
	}
}

func TestProvider_Get(t *testing.T) {
	p := newTestProvider(t)
	ctx := context.Background()

	// Add a memory first
	addReq := &core.AddRequest{
		Context: core.Context{
			TenantID:  "tenant-1",
			SubjectID: "user-123",
		},
		Type:    core.MemoryTypeFact,
		Content: "Test memory",
	}

	added, err := p.Add(ctx, addReq)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Get the memory
	getReq := &core.GetRequest{
		Context: core.Context{
			TenantID:  "tenant-1",
			SubjectID: "user-123",
		},
		ID: added.ID,
	}

	mem, err := p.Get(ctx, getReq)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if mem.ID != added.ID {
		t.Errorf("expected ID %q, got %q", added.ID, mem.ID)
	}
	if mem.Content != "Test memory" {
		t.Errorf("expected Content 'Test memory', got %q", mem.Content)
	}
}

func TestProvider_GetNotFound(t *testing.T) {
	p := newTestProvider(t)
	ctx := context.Background()

	getReq := &core.GetRequest{
		Context: core.Context{
			TenantID:  "tenant-1",
			SubjectID: "user-123",
		},
		ID: "non-existent-id",
	}

	_, err := p.Get(ctx, getReq)
	if err != core.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestProvider_GetTenantIsolation(t *testing.T) {
	p := newTestProvider(t)
	ctx := context.Background()

	// Add a memory for tenant-1
	addReq := &core.AddRequest{
		Context: core.Context{
			TenantID:  "tenant-1",
			SubjectID: "user-123",
		},
		Type:    core.MemoryTypeFact,
		Content: "Tenant 1 memory",
	}

	added, err := p.Add(ctx, addReq)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Try to get from tenant-2
	getReq := &core.GetRequest{
		Context: core.Context{
			TenantID:  "tenant-2",
			SubjectID: "user-123",
		},
		ID: added.ID,
	}

	_, err = p.Get(ctx, getReq)
	if err != core.ErrNotFound {
		t.Errorf("expected ErrNotFound for different tenant, got %v", err)
	}
}

func TestProvider_Update(t *testing.T) {
	p := newTestProvider(t)
	ctx := context.Background()

	// Add a memory first
	addReq := &core.AddRequest{
		Context: core.Context{
			TenantID:  "tenant-1",
			SubjectID: "user-123",
		},
		Type:    core.MemoryTypeFact,
		Content: "Original content",
	}

	added, err := p.Add(ctx, addReq)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Update the memory
	updateReq := &core.UpdateRequest{
		Context: core.Context{
			TenantID:  "tenant-1",
			SubjectID: "user-123",
		},
		ID:      added.ID,
		Content: "Updated content",
		Metadata: map[string]any{
			"updated": true,
		},
	}

	updated, err := p.Update(ctx, updateReq)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.Content != "Updated content" {
		t.Errorf("expected Content 'Updated content', got %q", updated.Content)
	}
	if updated.Metadata["updated"] != true {
		t.Errorf("expected metadata 'updated' to be true")
	}
	// UpdatedAt should be at or after original (may be equal if test runs fast)
	if updated.UpdatedAt.Before(added.UpdatedAt) {
		t.Error("expected UpdatedAt to be at or after original")
	}
}

func TestProvider_Delete(t *testing.T) {
	p := newTestProvider(t)
	ctx := context.Background()

	// Add a memory first
	addReq := &core.AddRequest{
		Context: core.Context{
			TenantID:  "tenant-1",
			SubjectID: "user-123",
		},
		Type:    core.MemoryTypeFact,
		Content: "To be deleted",
	}

	added, err := p.Add(ctx, addReq)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Delete the memory
	deleteReq := &core.DeleteRequest{
		Context: core.Context{
			TenantID:  "tenant-1",
			SubjectID: "user-123",
		},
		ID: added.ID,
	}

	err = p.Delete(ctx, deleteReq)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	getReq := &core.GetRequest{
		Context: core.Context{
			TenantID:  "tenant-1",
			SubjectID: "user-123",
		},
		ID: added.ID,
	}

	_, err = p.Get(ctx, getReq)
	if err != core.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestProvider_List(t *testing.T) {
	p := newTestProvider(t)
	ctx := context.Background()

	// Add multiple memories
	for i := 0; i < 5; i++ {
		addReq := &core.AddRequest{
			Context: core.Context{
				TenantID:  "tenant-1",
				SubjectID: "user-123",
				Scope:     core.ScopeUser,
			},
			Type:    core.MemoryTypeFact,
			Content: "Memory content",
		}
		if _, err := p.Add(ctx, addReq); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// List memories
	listReq := &core.ListRequest{
		Context: core.Context{
			TenantID:  "tenant-1",
			SubjectID: "user-123",
		},
		Limit: 3,
	}

	resp, err := p.List(ctx, listReq)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(resp.Memories) != 3 {
		t.Errorf("expected 3 memories, got %d", len(resp.Memories))
	}
	if resp.TotalCount != 5 {
		t.Errorf("expected TotalCount 5, got %d", resp.TotalCount)
	}
	if !resp.HasMore {
		t.Error("expected HasMore to be true")
	}
}

func TestProvider_ListWithTypeFilter(t *testing.T) {
	p := newTestProvider(t)
	ctx := context.Background()

	// Add memories of different types
	types := []core.MemoryType{core.MemoryTypeFact, core.MemoryTypePreference, core.MemoryTypeFact}
	for _, typ := range types {
		addReq := &core.AddRequest{
			Context: core.Context{
				TenantID:  "tenant-1",
				SubjectID: "user-123",
			},
			Type:    typ,
			Content: "Memory content",
		}
		if _, err := p.Add(ctx, addReq); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// List only facts
	listReq := &core.ListRequest{
		Context: core.Context{
			TenantID:  "tenant-1",
			SubjectID: "user-123",
		},
		Types: []core.MemoryType{core.MemoryTypeFact},
	}

	resp, err := p.List(ctx, listReq)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(resp.Memories) != 2 {
		t.Errorf("expected 2 fact memories, got %d", len(resp.Memories))
	}
	for _, m := range resp.Memories {
		if m.Type != core.MemoryTypeFact {
			t.Errorf("expected type 'fact', got %q", m.Type)
		}
	}
}

func TestProvider_Search(t *testing.T) {
	p := newTestProvider(t)
	ctx := context.Background()

	// Add some memories with different content
	contents := []string{
		"User prefers dark mode interface",
		"User likes coffee in the morning",
		"User is a software engineer",
	}

	for _, content := range contents {
		addReq := &core.AddRequest{
			Context: core.Context{
				TenantID:  "tenant-1",
				SubjectID: "user-123",
			},
			Type:    core.MemoryTypeFact,
			Content: content,
		}
		if _, err := p.Add(ctx, addReq); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// Search
	searchReq := &core.SearchRequest{
		Context: core.Context{
			TenantID:  "tenant-1",
			SubjectID: "user-123",
		},
		Query: "dark mode preferences",
		Limit: 10,
	}

	resp, err := p.Search(ctx, searchReq)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(resp.Results) == 0 {
		t.Error("expected at least one search result")
	}

	// Results should be sorted by score
	for i := 1; i < len(resp.Results); i++ {
		if resp.Results[i].Score > resp.Results[i-1].Score {
			t.Error("results not sorted by score descending")
		}
	}
}

func TestProvider_Recall(t *testing.T) {
	p := newTestProvider(t)
	ctx := context.Background()

	// Add some memories
	contents := []string{
		"User prefers dark mode",
		"User likes TypeScript",
	}

	for _, content := range contents {
		addReq := &core.AddRequest{
			Context: core.Context{
				TenantID:  "tenant-1",
				SubjectID: "user-123",
			},
			Type:    core.MemoryTypeFact,
			Content: content,
		}
		if _, err := p.Add(ctx, addReq); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// Recall
	recallReq := &core.RecallRequest{
		Context: core.Context{
			TenantID:  "tenant-1",
			SubjectID: "user-123",
		},
		Query:      "programming preferences",
		MaxResults: 5,
	}

	resp, err := p.Recall(ctx, recallReq)
	if err != nil {
		t.Fatalf("Recall failed: %v", err)
	}

	if len(resp.Memories) == 0 {
		t.Error("expected at least one memory in recall response")
	}
}

func TestProvider_ValidationErrors(t *testing.T) {
	p := newTestProvider(t)
	ctx := context.Background()

	// Missing tenant_id
	addReq := &core.AddRequest{
		Context: core.Context{
			SubjectID: "user-123",
		},
		Type:    core.MemoryTypeFact,
		Content: "Test",
	}

	_, err := p.Add(ctx, addReq)
	if err != core.ErrTenantRequired {
		t.Errorf("expected ErrTenantRequired, got %v", err)
	}

	// Missing subject_id
	addReq = &core.AddRequest{
		Context: core.Context{
			TenantID: "tenant-1",
		},
		Type:    core.MemoryTypeFact,
		Content: "Test",
	}

	_, err = p.Add(ctx, addReq)
	if err != core.ErrSubjectRequired {
		t.Errorf("expected ErrSubjectRequired, got %v", err)
	}

	// Missing content
	addReq = &core.AddRequest{
		Context: core.Context{
			TenantID:  "tenant-1",
			SubjectID: "user-123",
		},
		Type: core.MemoryTypeFact,
	}

	_, err = p.Add(ctx, addReq)
	if err != core.ErrContentRequired {
		t.Errorf("expected ErrContentRequired, got %v", err)
	}
}

func TestProvider_Close(t *testing.T) {
	p := newTestProvider(t)
	ctx := context.Background()

	// Add a memory
	addReq := &core.AddRequest{
		Context: core.Context{
			TenantID:  "tenant-1",
			SubjectID: "user-123",
		},
		Type:    core.MemoryTypeFact,
		Content: "Test",
	}

	if _, err := p.Add(ctx, addReq); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Close should clear memories
	if err := p.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// List should return empty
	listReq := &core.ListRequest{
		Context: core.Context{
			TenantID:  "tenant-1",
			SubjectID: "user-123",
		},
	}

	resp, err := p.List(ctx, listReq)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(resp.Memories) != 0 {
		t.Errorf("expected 0 memories after close, got %d", len(resp.Memories))
	}
}
