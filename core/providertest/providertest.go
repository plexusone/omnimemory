// Package providertest provides conformance tests for memory provider implementations.
//
// Provider implementations can use this package to verify they correctly implement
// the core.Provider interface with consistent behavior.
//
// Basic usage:
//
//	func TestConformance(t *testing.T) {
//	    store := kvsmemory.New()
//	    p, _ := kvs.NewProvider(core.ProviderConfig{
//	        Options: map[string]any{"store": store},
//	    }, nil)
//
//	    providertest.RunAll(t, providertest.Config{
//	        Provider: p,
//	    })
//	}
//
// For providers requiring API credentials (like Twilio):
//
//	func TestConformance(t *testing.T) {
//	    if os.Getenv("TWILIO_ACCOUNT_SID") == "" {
//	        t.Skip("TWILIO_ACCOUNT_SID not set")
//	    }
//
//	    p, _ := NewProvider(core.ProviderConfig{}, nil)
//
//	    providertest.RunAll(t, providertest.Config{
//	        Provider:        p,
//	        SkipIntegration: false,
//	        TenantID:        "your-store-id",
//	        SubjectID:       "your-profile-id",
//	    })
//	}
package providertest

import (
	"context"
	"testing"
	"time"

	"github.com/plexusone/omnimemory/core"
)

// Config configures the memory provider conformance test suite.
type Config struct {
	// Provider is the memory provider implementation to test.
	Provider core.Provider

	// Embedder is an optional embedder for testing semantic search.
	// If nil, search tests will verify behavior without embeddings.
	Embedder core.Embedder

	// SkipIntegration skips tests that require real API calls.
	// Set to true for local-only providers (memory, kvs, postgres).
	SkipIntegration bool

	// TenantID is the tenant ID to use for tests.
	// Defaults to "test-tenant" if empty.
	TenantID string

	// SubjectID is the subject ID to use for tests.
	// Defaults to "test-subject" if empty.
	SubjectID string

	// Timeout for individual test operations.
	// Defaults to 30 seconds if zero.
	Timeout time.Duration
}

// withDefaults returns a copy of Config with default values applied.
func (c Config) withDefaults() Config {
	if c.TenantID == "" {
		c.TenantID = "test-tenant"
	}
	if c.SubjectID == "" {
		c.SubjectID = "test-subject"
	}
	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second
	}
	return c
}

// testContext returns a context for the test.
func (c Config) testContext() core.Context {
	return core.Context{
		TenantID:  c.TenantID,
		SubjectID: c.SubjectID,
	}
}

// RunAll runs all conformance tests for a memory provider.
func RunAll(t *testing.T, cfg Config) {
	t.Helper()
	cfg = cfg.withDefaults()

	// Interface tests (always run)
	t.Run("Interface", func(t *testing.T) {
		RunInterfaceTests(t, cfg)
	})

	// CRUD tests (always run)
	t.Run("CRUD", func(t *testing.T) {
		RunCRUDTests(t, cfg)
	})

	// Behavior tests (always run)
	t.Run("Behavior", func(t *testing.T) {
		RunBehaviorTests(t, cfg)
	})

	// Search tests (may require embedder)
	t.Run("Search", func(t *testing.T) {
		RunSearchTests(t, cfg)
	})
}

// RunInterfaceTests runs only interface compliance tests.
// These tests verify the provider correctly implements the interface contract.
func RunInterfaceTests(t *testing.T, cfg Config) {
	t.Helper()
	cfg = cfg.withDefaults()

	t.Run("Name", func(t *testing.T) { testName(t, cfg) })
	t.Run("Close", func(t *testing.T) { testClose(t, cfg) })
}

// RunCRUDTests runs CRUD operation tests.
func RunCRUDTests(t *testing.T, cfg Config) {
	t.Helper()
	cfg = cfg.withDefaults()

	t.Run("Add", func(t *testing.T) { testAdd(t, cfg) })
	t.Run("Add_WithTTL", func(t *testing.T) { testAddWithTTL(t, cfg) })
	t.Run("Add_WithMetadata", func(t *testing.T) { testAddWithMetadata(t, cfg) })
	t.Run("Get", func(t *testing.T) { testGet(t, cfg) })
	t.Run("Get_NotFound", func(t *testing.T) { testGetNotFound(t, cfg) })
	t.Run("Update", func(t *testing.T) { testUpdate(t, cfg) })
	t.Run("Update_NotFound", func(t *testing.T) { testUpdateNotFound(t, cfg) })
	t.Run("Delete", func(t *testing.T) { testDelete(t, cfg) })
	t.Run("List", func(t *testing.T) { testList(t, cfg) })
	t.Run("List_WithTypeFilter", func(t *testing.T) { testListWithTypeFilter(t, cfg) })
	t.Run("List_WithScopeFilter", func(t *testing.T) { testListWithScopeFilter(t, cfg) })
}

// RunBehaviorTests runs behavioral contract tests.
func RunBehaviorTests(t *testing.T, cfg Config) {
	t.Helper()
	cfg = cfg.withDefaults()

	t.Run("TenantIsolation", func(t *testing.T) { testTenantIsolation(t, cfg) })
	t.Run("SubjectIsolation", func(t *testing.T) { testSubjectIsolation(t, cfg) })
	t.Run("ValidationErrors", func(t *testing.T) { testValidationErrors(t, cfg) })
}

// RunSearchTests runs search and recall tests.
func RunSearchTests(t *testing.T, cfg Config) {
	t.Helper()
	cfg = cfg.withDefaults()

	t.Run("Search", func(t *testing.T) { testSearch(t, cfg) })
	t.Run("Recall", func(t *testing.T) { testRecall(t, cfg) })
}

// Interface Tests

func testName(t *testing.T, cfg Config) {
	t.Helper()
	name := cfg.Provider.Name()
	if name == "" {
		t.Error("Name() returned empty string")
	}
	// Verify name is lowercase, alphanumeric with hyphens/underscores
	for _, r := range name {
		isLowerAlpha := r >= 'a' && r <= 'z'
		isDigit := r >= '0' && r <= '9'
		isSpecial := r == '-' || r == '_'
		if !isLowerAlpha && !isDigit && !isSpecial {
			t.Errorf("Name() contains invalid character %q; should be lowercase alphanumeric with hyphens/underscores", r)
		}
	}
	t.Logf("Provider name: %s", name)
}

func testClose(t *testing.T, cfg Config) {
	t.Helper()
	// We don't actually close the provider since it may be reused
	// This test just verifies the method exists
	_ = cfg.Provider
}

// CRUD Tests

func testAdd(t *testing.T, cfg Config) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	memory, err := cfg.Provider.Add(ctx, &core.AddRequest{
		Context: cfg.testContext(),
		Type:    core.MemoryTypeObservation,
		Content: "Test memory content for conformance test",
	})

	if err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	if memory.ID == "" {
		t.Error("Add() returned memory with empty ID")
	}
	if memory.TenantID != cfg.TenantID {
		t.Errorf("Add() TenantID = %q, want %q", memory.TenantID, cfg.TenantID)
	}
	if memory.SubjectID != cfg.SubjectID {
		t.Errorf("Add() SubjectID = %q, want %q", memory.SubjectID, cfg.SubjectID)
	}
	if memory.Content != "Test memory content for conformance test" {
		t.Errorf("Add() Content mismatch")
	}
	if memory.Type != core.MemoryTypeObservation {
		t.Errorf("Add() Type = %q, want %q", memory.Type, core.MemoryTypeObservation)
	}
	if memory.CreatedAt.IsZero() {
		t.Error("Add() CreatedAt is zero")
	}

	t.Logf("Added memory ID: %s", memory.ID)

	// Cleanup
	_ = cfg.Provider.Delete(ctx, &core.DeleteRequest{
		Context: cfg.testContext(),
		ID:      memory.ID,
	})
}

func testAddWithTTL(t *testing.T, cfg Config) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	memory, err := cfg.Provider.Add(ctx, &core.AddRequest{
		Context: cfg.testContext(),
		Type:    core.MemoryTypeObservation,
		Content: "Memory with TTL",
		TTL:     time.Hour,
	})

	if err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	if memory.ExpiresAt == nil {
		t.Error("Add() with TTL should set ExpiresAt")
	} else if time.Until(*memory.ExpiresAt) < 59*time.Minute {
		t.Error("Add() ExpiresAt should be approximately 1 hour in the future")
	}

	// Cleanup
	_ = cfg.Provider.Delete(ctx, &core.DeleteRequest{
		Context: cfg.testContext(),
		ID:      memory.ID,
	})
}

func testAddWithMetadata(t *testing.T, cfg Config) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	metadata := map[string]any{
		"source":     "test",
		"importance": 5,
	}

	memory, err := cfg.Provider.Add(ctx, &core.AddRequest{
		Context:  cfg.testContext(),
		Type:     core.MemoryTypeObservation,
		Content:  "Memory with metadata",
		Metadata: metadata,
	})

	if err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	if memory.Metadata == nil {
		t.Error("Add() with metadata should preserve metadata")
	}

	// Cleanup
	_ = cfg.Provider.Delete(ctx, &core.DeleteRequest{
		Context: cfg.testContext(),
		ID:      memory.ID,
	})
}

func testGet(t *testing.T, cfg Config) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Add a memory first
	added, err := cfg.Provider.Add(ctx, &core.AddRequest{
		Context: cfg.testContext(),
		Type:    core.MemoryTypeObservation,
		Content: "Memory to retrieve",
	})
	if err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	// Get it back
	retrieved, err := cfg.Provider.Get(ctx, &core.GetRequest{
		Context: cfg.testContext(),
		ID:      added.ID,
	})
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	if retrieved.ID != added.ID {
		t.Errorf("Get() ID = %q, want %q", retrieved.ID, added.ID)
	}
	if retrieved.Content != added.Content {
		t.Errorf("Get() Content = %q, want %q", retrieved.Content, added.Content)
	}

	// Cleanup
	_ = cfg.Provider.Delete(ctx, &core.DeleteRequest{
		Context: cfg.testContext(),
		ID:      added.ID,
	})
}

func testGetNotFound(t *testing.T, cfg Config) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	_, err := cfg.Provider.Get(ctx, &core.GetRequest{
		Context: cfg.testContext(),
		ID:      "nonexistent-memory-id",
	})

	if err != core.ErrNotFound {
		t.Errorf("Get() for nonexistent ID should return ErrNotFound, got: %v", err)
	}
}

func testUpdate(t *testing.T, cfg Config) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Add a memory first
	added, _ := cfg.Provider.Add(ctx, &core.AddRequest{
		Context: cfg.testContext(),
		Type:    core.MemoryTypeObservation,
		Content: "Original content",
	})

	// Update it
	updated, err := cfg.Provider.Update(ctx, &core.UpdateRequest{
		Context: cfg.testContext(),
		ID:      added.ID,
		Content: "Updated content",
	})

	if err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	if updated.Content != "Updated content" {
		t.Errorf("Update() Content = %q, want %q", updated.Content, "Updated content")
	}

	// Verify persistence
	retrieved, _ := cfg.Provider.Get(ctx, &core.GetRequest{
		Context: cfg.testContext(),
		ID:      added.ID,
	})
	if retrieved.Content != "Updated content" {
		t.Errorf("Get() after Update() Content = %q, want %q", retrieved.Content, "Updated content")
	}

	// Cleanup
	_ = cfg.Provider.Delete(ctx, &core.DeleteRequest{
		Context: cfg.testContext(),
		ID:      added.ID,
	})
}

func testUpdateNotFound(t *testing.T, cfg Config) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	_, err := cfg.Provider.Update(ctx, &core.UpdateRequest{
		Context: cfg.testContext(),
		ID:      "nonexistent-memory-id",
		Content: "New content",
	})

	if err != core.ErrNotFound {
		t.Errorf("Update() for nonexistent ID should return ErrNotFound, got: %v", err)
	}
}

func testDelete(t *testing.T, cfg Config) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Add a memory first
	added, _ := cfg.Provider.Add(ctx, &core.AddRequest{
		Context: cfg.testContext(),
		Type:    core.MemoryTypeObservation,
		Content: "Memory to delete",
	})

	// Delete it
	err := cfg.Provider.Delete(ctx, &core.DeleteRequest{
		Context: cfg.testContext(),
		ID:      added.ID,
	})

	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	// Verify deletion
	_, err = cfg.Provider.Get(ctx, &core.GetRequest{
		Context: cfg.testContext(),
		ID:      added.ID,
	})
	if err != core.ErrNotFound {
		t.Errorf("Get() after Delete() should return ErrNotFound, got: %v", err)
	}
}

func testList(t *testing.T, cfg Config) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Add some memories
	var ids []string
	for i := 0; i < 3; i++ {
		m, _ := cfg.Provider.Add(ctx, &core.AddRequest{
			Context: cfg.testContext(),
			Type:    core.MemoryTypeObservation,
			Content: "List test memory",
		})
		ids = append(ids, m.ID)
	}

	// List them
	resp, err := cfg.Provider.List(ctx, &core.ListRequest{
		Context: cfg.testContext(),
	})

	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(resp.Memories) < 3 {
		t.Errorf("List() returned %d memories, want at least 3", len(resp.Memories))
	}

	// Cleanup
	for _, id := range ids {
		_ = cfg.Provider.Delete(ctx, &core.DeleteRequest{
			Context: cfg.testContext(),
			ID:      id,
		})
	}
}

func testListWithTypeFilter(t *testing.T, cfg Config) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Add memories of different types
	obs, _ := cfg.Provider.Add(ctx, &core.AddRequest{
		Context: cfg.testContext(),
		Type:    core.MemoryTypeObservation,
		Content: "Observation",
	})
	fact, _ := cfg.Provider.Add(ctx, &core.AddRequest{
		Context: cfg.testContext(),
		Type:    core.MemoryTypeFact,
		Content: "Fact",
	})

	// Filter by type
	resp, err := cfg.Provider.List(ctx, &core.ListRequest{
		Context: cfg.testContext(),
		Types:   []core.MemoryType{core.MemoryTypeFact},
	})

	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	for _, m := range resp.Memories {
		if m.Type != core.MemoryTypeFact {
			t.Errorf("List() with type filter returned memory with type %q", m.Type)
		}
	}

	// Cleanup
	_ = cfg.Provider.Delete(ctx, &core.DeleteRequest{Context: cfg.testContext(), ID: obs.ID})
	_ = cfg.Provider.Delete(ctx, &core.DeleteRequest{Context: cfg.testContext(), ID: fact.ID})
}

func testListWithScopeFilter(t *testing.T, cfg Config) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Add memories with different scopes
	ctxUser := cfg.testContext()
	ctxUser.Scope = core.ScopeUser

	ctxAgent := cfg.testContext()
	ctxAgent.Scope = core.ScopeAgent

	user, _ := cfg.Provider.Add(ctx, &core.AddRequest{
		Context: ctxUser,
		Type:    core.MemoryTypeObservation,
		Content: "User scoped",
	})
	agent, _ := cfg.Provider.Add(ctx, &core.AddRequest{
		Context: ctxAgent,
		Type:    core.MemoryTypeObservation,
		Content: "Agent scoped",
	})

	// Filter by scope
	resp, err := cfg.Provider.List(ctx, &core.ListRequest{
		Context: cfg.testContext(),
		Scopes:  []core.Scope{core.ScopeUser},
	})

	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	for _, m := range resp.Memories {
		if m.Scope != core.ScopeUser {
			t.Errorf("List() with scope filter returned memory with scope %q", m.Scope)
		}
	}

	// Cleanup
	_ = cfg.Provider.Delete(ctx, &core.DeleteRequest{Context: cfg.testContext(), ID: user.ID})
	_ = cfg.Provider.Delete(ctx, &core.DeleteRequest{Context: cfg.testContext(), ID: agent.ID})
}

// Behavior Tests

func testTenantIsolation(t *testing.T, cfg Config) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Add memory to tenant1
	tenant1Ctx := cfg.testContext()
	tenant1Ctx.TenantID = "tenant1"

	added, _ := cfg.Provider.Add(ctx, &core.AddRequest{
		Context: tenant1Ctx,
		Type:    core.MemoryTypeObservation,
		Content: "Tenant 1 memory",
	})

	// Try to get from tenant2
	tenant2Ctx := cfg.testContext()
	tenant2Ctx.TenantID = "tenant2"

	_, err := cfg.Provider.Get(ctx, &core.GetRequest{
		Context: tenant2Ctx,
		ID:      added.ID,
	})

	if err != core.ErrNotFound {
		t.Errorf("Get() across tenants should return ErrNotFound, got: %v", err)
	}

	// Cleanup
	_ = cfg.Provider.Delete(ctx, &core.DeleteRequest{Context: tenant1Ctx, ID: added.ID})
}

func testSubjectIsolation(t *testing.T, cfg Config) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Add memory to subject1
	subject1Ctx := cfg.testContext()
	subject1Ctx.SubjectID = "subject1"

	added, _ := cfg.Provider.Add(ctx, &core.AddRequest{
		Context: subject1Ctx,
		Type:    core.MemoryTypeObservation,
		Content: "Subject 1 memory",
	})

	// Try to get from subject2
	subject2Ctx := cfg.testContext()
	subject2Ctx.SubjectID = "subject2"

	_, err := cfg.Provider.Get(ctx, &core.GetRequest{
		Context: subject2Ctx,
		ID:      added.ID,
	})

	if err != core.ErrNotFound {
		t.Errorf("Get() across subjects should return ErrNotFound, got: %v", err)
	}

	// Cleanup
	_ = cfg.Provider.Delete(ctx, &core.DeleteRequest{Context: subject1Ctx, ID: added.ID})
}

func testValidationErrors(t *testing.T, cfg Config) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	tests := []struct {
		name string
		req  *core.AddRequest
	}{
		{
			name: "missing_tenant_id",
			req: &core.AddRequest{
				Context: core.Context{SubjectID: "subject"},
				Type:    core.MemoryTypeObservation,
				Content: "Content",
			},
		},
		{
			name: "missing_subject_id",
			req: &core.AddRequest{
				Context: core.Context{TenantID: "tenant"},
				Type:    core.MemoryTypeObservation,
				Content: "Content",
			},
		},
		{
			name: "missing_content",
			req: &core.AddRequest{
				Context: cfg.testContext(),
				Type:    core.MemoryTypeObservation,
			},
		},
		{
			name: "invalid_type",
			req: &core.AddRequest{
				Context: cfg.testContext(),
				Type:    core.MemoryType("invalid"),
				Content: "Content",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := cfg.Provider.Add(ctx, tc.req)
			if err == nil {
				t.Errorf("Add() should return validation error for %s", tc.name)
			}
		})
	}
}

// Search Tests

func testSearch(t *testing.T, cfg Config) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Add some memories
	m1, _ := cfg.Provider.Add(ctx, &core.AddRequest{
		Context: cfg.testContext(),
		Type:    core.MemoryTypeObservation,
		Content: "The user likes programming in Go",
	})
	m2, _ := cfg.Provider.Add(ctx, &core.AddRequest{
		Context: cfg.testContext(),
		Type:    core.MemoryTypeFact,
		Content: "The weather today is sunny",
	})

	// Search (without embedder, this tests the basic flow)
	resp, err := cfg.Provider.Search(ctx, &core.SearchRequest{
		Context: cfg.testContext(),
		Query:   "programming",
	})

	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}

	// Without embedder, search should still return results (may not be ranked by relevance)
	t.Logf("Search returned %d results", len(resp.Results))

	// Cleanup
	_ = cfg.Provider.Delete(ctx, &core.DeleteRequest{Context: cfg.testContext(), ID: m1.ID})
	_ = cfg.Provider.Delete(ctx, &core.DeleteRequest{Context: cfg.testContext(), ID: m2.ID})
}

func testRecall(t *testing.T, cfg Config) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Add some memories
	m1, _ := cfg.Provider.Add(ctx, &core.AddRequest{
		Context: cfg.testContext(),
		Type:    core.MemoryTypeObservation,
		Content: "User prefers dark mode interfaces",
	})

	// Recall
	resp, err := cfg.Provider.Recall(ctx, &core.RecallRequest{
		Context:    cfg.testContext(),
		Query:      "interface preferences",
		MaxResults: 5,
	})

	if err != nil {
		t.Fatalf("Recall() error: %v", err)
	}

	t.Logf("Recall returned %d memories", len(resp.Memories))

	// Cleanup
	_ = cfg.Provider.Delete(ctx, &core.DeleteRequest{Context: cfg.testContext(), ID: m1.ID})
}
