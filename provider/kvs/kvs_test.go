package kvs

import (
	"context"
	"testing"
	"time"

	"github.com/plexusone/omnimemory/core"
	kvsmemory "github.com/plexusone/omnistorage-core/kvs/backend/memory"
)

func newTestProvider(t *testing.T) *Provider {
	t.Helper()
	store := kvsmemory.New()
	t.Cleanup(func() { _ = store.Close() })

	provider, err := NewProvider(core.ProviderConfig{
		Options: map[string]any{
			"store": store,
		},
	}, nil)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}
	return provider.(*Provider)
}

func TestProvider_Name(t *testing.T) {
	provider := newTestProvider(t)
	if provider.Name() != "kvs" {
		t.Errorf("expected name 'kvs', got %q", provider.Name())
	}
}

func TestProvider_Add(t *testing.T) {
	provider := newTestProvider(t)
	ctx := context.Background()

	memory, err := provider.Add(ctx, &core.AddRequest{
		Context: core.Context{
			TenantID:  "tenant1",
			SubjectID: "subject1",
		},
		Type:    core.MemoryTypeObservation,
		Content: "Test memory content",
	})

	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if memory.ID == "" {
		t.Error("expected memory ID to be set")
	}
	if memory.TenantID != "tenant1" {
		t.Errorf("expected tenant ID 'tenant1', got %q", memory.TenantID)
	}
	if memory.SubjectID != "subject1" {
		t.Errorf("expected subject ID 'subject1', got %q", memory.SubjectID)
	}
	if memory.Content != "Test memory content" {
		t.Errorf("expected content 'Test memory content', got %q", memory.Content)
	}
	if memory.Type != core.MemoryTypeObservation {
		t.Errorf("expected type 'observation', got %q", memory.Type)
	}
}

func TestProvider_AddWithTTL(t *testing.T) {
	provider := newTestProvider(t)
	ctx := context.Background()

	memory, err := provider.Add(ctx, &core.AddRequest{
		Context: core.Context{
			TenantID:  "tenant1",
			SubjectID: "subject1",
		},
		Type:    core.MemoryTypeObservation,
		Content: "Test memory with TTL",
		TTL:     time.Hour,
	})

	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if memory.ExpiresAt == nil {
		t.Error("expected ExpiresAt to be set")
	}
	if time.Until(*memory.ExpiresAt) < 59*time.Minute {
		t.Error("expected ExpiresAt to be about 1 hour in the future")
	}
}

func TestProvider_Get(t *testing.T) {
	provider := newTestProvider(t)
	ctx := context.Background()

	// Add a memory
	added, err := provider.Add(ctx, &core.AddRequest{
		Context: core.Context{
			TenantID:  "tenant1",
			SubjectID: "subject1",
		},
		Type:    core.MemoryTypeObservation,
		Content: "Test memory content",
	})
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Get the memory
	retrieved, err := provider.Get(ctx, &core.GetRequest{
		Context: core.Context{
			TenantID:  "tenant1",
			SubjectID: "subject1",
		},
		ID: added.ID,
	})
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.ID != added.ID {
		t.Errorf("expected ID %q, got %q", added.ID, retrieved.ID)
	}
	if retrieved.Content != added.Content {
		t.Errorf("expected content %q, got %q", added.Content, retrieved.Content)
	}
}

func TestProvider_GetNotFound(t *testing.T) {
	provider := newTestProvider(t)
	ctx := context.Background()

	_, err := provider.Get(ctx, &core.GetRequest{
		Context: core.Context{
			TenantID:  "tenant1",
			SubjectID: "subject1",
		},
		ID: "nonexistent",
	})

	if err != core.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestProvider_GetTenantIsolation(t *testing.T) {
	provider := newTestProvider(t)
	ctx := context.Background()

	// Add a memory to tenant1
	added, _ := provider.Add(ctx, &core.AddRequest{
		Context: core.Context{
			TenantID:  "tenant1",
			SubjectID: "subject1",
		},
		Type:    core.MemoryTypeObservation,
		Content: "Tenant 1 memory",
	})

	// Try to get from tenant2
	_, err := provider.Get(ctx, &core.GetRequest{
		Context: core.Context{
			TenantID:  "tenant2",
			SubjectID: "subject1",
		},
		ID: added.ID,
	})

	if err != core.ErrNotFound {
		t.Errorf("expected ErrNotFound for cross-tenant access, got %v", err)
	}
}

func TestProvider_Update(t *testing.T) {
	provider := newTestProvider(t)
	ctx := context.Background()

	// Add a memory
	added, _ := provider.Add(ctx, &core.AddRequest{
		Context: core.Context{
			TenantID:  "tenant1",
			SubjectID: "subject1",
		},
		Type:    core.MemoryTypeObservation,
		Content: "Original content",
	})

	// Update the memory
	updated, err := provider.Update(ctx, &core.UpdateRequest{
		Context: core.Context{
			TenantID:  "tenant1",
			SubjectID: "subject1",
		},
		ID:      added.ID,
		Content: "Updated content",
	})

	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.Content != "Updated content" {
		t.Errorf("expected content 'Updated content', got %q", updated.Content)
	}

	// Verify persistence
	retrieved, _ := provider.Get(ctx, &core.GetRequest{
		Context: core.Context{
			TenantID:  "tenant1",
			SubjectID: "subject1",
		},
		ID: added.ID,
	})
	if retrieved.Content != "Updated content" {
		t.Errorf("expected persisted content 'Updated content', got %q", retrieved.Content)
	}
}

func TestProvider_Delete(t *testing.T) {
	provider := newTestProvider(t)
	ctx := context.Background()

	// Add a memory
	added, _ := provider.Add(ctx, &core.AddRequest{
		Context: core.Context{
			TenantID:  "tenant1",
			SubjectID: "subject1",
		},
		Type:    core.MemoryTypeObservation,
		Content: "Test memory",
	})

	// Delete the memory
	err := provider.Delete(ctx, &core.DeleteRequest{
		Context: core.Context{
			TenantID:  "tenant1",
			SubjectID: "subject1",
		},
		ID: added.ID,
	})

	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = provider.Get(ctx, &core.GetRequest{
		Context: core.Context{
			TenantID:  "tenant1",
			SubjectID: "subject1",
		},
		ID: added.ID,
	})
	if err != core.ErrNotFound {
		t.Errorf("expected ErrNotFound after deletion, got %v", err)
	}
}

func TestProvider_List(t *testing.T) {
	provider := newTestProvider(t)
	ctx := context.Background()

	// Add multiple memories
	for i := 0; i < 5; i++ {
		_, err := provider.Add(ctx, &core.AddRequest{
			Context: core.Context{
				TenantID:  "tenant1",
				SubjectID: "subject1",
			},
			Type:    core.MemoryTypeObservation,
			Content: "Memory content",
		})
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// Add memory to different subject
	_, _ = provider.Add(ctx, &core.AddRequest{
		Context: core.Context{
			TenantID:  "tenant1",
			SubjectID: "subject2",
		},
		Type:    core.MemoryTypeObservation,
		Content: "Different subject",
	})

	// List for subject1
	resp, err := provider.List(ctx, &core.ListRequest{
		Context: core.Context{
			TenantID:  "tenant1",
			SubjectID: "subject1",
		},
	})

	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(resp.Memories) != 5 {
		t.Errorf("expected 5 memories, got %d", len(resp.Memories))
	}
}

func TestProvider_ListWithTypeFilter(t *testing.T) {
	provider := newTestProvider(t)
	ctx := context.Background()

	// Add memories of different types
	_, _ = provider.Add(ctx, &core.AddRequest{
		Context: core.Context{
			TenantID:  "tenant1",
			SubjectID: "subject1",
		},
		Type:    core.MemoryTypeObservation,
		Content: "Observation",
	})
	_, _ = provider.Add(ctx, &core.AddRequest{
		Context: core.Context{
			TenantID:  "tenant1",
			SubjectID: "subject1",
		},
		Type:    core.MemoryTypeFact,
		Content: "Fact",
	})
	_, _ = provider.Add(ctx, &core.AddRequest{
		Context: core.Context{
			TenantID:  "tenant1",
			SubjectID: "subject1",
		},
		Type:    core.MemoryTypePreference,
		Content: "Preference",
	})

	// Filter by type
	resp, err := provider.List(ctx, &core.ListRequest{
		Context: core.Context{
			TenantID:  "tenant1",
			SubjectID: "subject1",
		},
		Types: []core.MemoryType{core.MemoryTypeFact},
	})

	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(resp.Memories) != 1 {
		t.Errorf("expected 1 memory, got %d", len(resp.Memories))
	}
	if resp.Memories[0].Type != core.MemoryTypeFact {
		t.Errorf("expected type 'fact', got %q", resp.Memories[0].Type)
	}
}

func TestProvider_ValidationErrors(t *testing.T) {
	provider := newTestProvider(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		request *core.AddRequest
	}{
		{
			name: "missing_tenant_id",
			request: &core.AddRequest{
				Context: core.Context{SubjectID: "subject1"},
				Type:    core.MemoryTypeObservation,
				Content: "Content",
			},
		},
		{
			name: "missing_subject_id",
			request: &core.AddRequest{
				Context: core.Context{TenantID: "tenant1"},
				Type:    core.MemoryTypeObservation,
				Content: "Content",
			},
		},
		{
			name: "missing_content",
			request: &core.AddRequest{
				Context: core.Context{
					TenantID:  "tenant1",
					SubjectID: "subject1",
				},
				Type: core.MemoryTypeObservation,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := provider.Add(ctx, tc.request)
			if err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

func TestProvider_Close(t *testing.T) {
	provider := newTestProvider(t)

	if err := provider.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestNewProvider_MissingStore(t *testing.T) {
	_, err := NewProvider(core.ProviderConfig{}, nil)
	if err == nil {
		t.Error("expected error for missing store")
	}

	validationErr, ok := err.(*core.ValidationError)
	if !ok {
		t.Errorf("expected ValidationError, got %T", err)
	}
	if validationErr.Field != "store" {
		t.Errorf("expected field 'store', got %q", validationErr.Field)
	}
}

func TestNewProvider_CustomKeyPrefix(t *testing.T) {
	store := kvsmemory.New()
	defer func() { _ = store.Close() }()

	provider, err := NewProvider(core.ProviderConfig{
		Options: map[string]any{
			"store":      store,
			"key_prefix": "custom_prefix",
		},
	}, nil)

	if err != nil {
		t.Fatalf("NewProvider failed: %v", err)
	}

	kvsProv := provider.(*Provider)
	if kvsProv.keyPrefix != "custom_prefix" {
		t.Errorf("expected key prefix 'custom_prefix', got %q", kvsProv.keyPrefix)
	}
}
