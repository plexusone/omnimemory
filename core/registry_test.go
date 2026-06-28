package core

import (
	"context"
	"testing"
)

// mockProvider is a test implementation of Provider.
type mockProvider struct {
	name string
}

func (p *mockProvider) Name() string { return p.name }
func (p *mockProvider) Close() error { return nil }

func (p *mockProvider) Add(_ context.Context, _ *AddRequest) (*Memory, error) {
	return nil, nil
}

func (p *mockProvider) Get(_ context.Context, _ *GetRequest) (*Memory, error) {
	return nil, nil
}

func (p *mockProvider) Update(_ context.Context, _ *UpdateRequest) (*Memory, error) {
	return nil, nil
}

func (p *mockProvider) Delete(_ context.Context, _ *DeleteRequest) error {
	return nil
}

func (p *mockProvider) List(_ context.Context, _ *ListRequest) (*ListResponse, error) {
	return nil, nil
}

func (p *mockProvider) Search(_ context.Context, _ *SearchRequest) (*SearchResponse, error) {
	return nil, nil
}

func (p *mockProvider) Recall(_ context.Context, _ *RecallRequest) (*RecallResponse, error) {
	return nil, nil
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := &Registry{
		providers: make(map[ProviderName]*registeredProvider),
	}

	factory := func(_ ProviderConfig, _ Embedder) (Provider, error) {
		return &mockProvider{name: "test"}, nil
	}

	r.Register("test", factory, PriorityThin)

	if !r.Has("test") {
		t.Error("expected registry to have 'test' provider")
	}

	provider, err := r.Get("test", ProviderConfig{}, nil)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if provider.Name() != "test" {
		t.Errorf("expected provider name 'test', got %q", provider.Name())
	}
}

func TestRegistry_GetNotFound(t *testing.T) {
	r := &Registry{
		providers: make(map[ProviderName]*registeredProvider),
	}

	_, err := r.Get("nonexistent", ProviderConfig{}, nil)
	if err != ErrProviderNotFound {
		t.Errorf("expected ErrProviderNotFound, got %v", err)
	}
}

func TestRegistry_List(t *testing.T) {
	r := &Registry{
		providers: make(map[ProviderName]*registeredProvider),
	}

	factory := func(_ ProviderConfig, _ Embedder) (Provider, error) {
		return &mockProvider{name: "mock"}, nil
	}

	r.Register("thin", factory, PriorityThin)
	r.Register("thick", factory, PriorityThick)

	list := r.List()

	if len(list) != 2 {
		t.Errorf("expected 2 providers, got %d", len(list))
	}

	// Should be sorted by priority descending (thick first)
	if list[0] != "thick" {
		t.Errorf("expected first provider to be 'thick', got %q", list[0])
	}
	if list[1] != "thin" {
		t.Errorf("expected second provider to be 'thin', got %q", list[1])
	}
}

func TestRegistry_Unregister(t *testing.T) {
	r := &Registry{
		providers: make(map[ProviderName]*registeredProvider),
	}

	factory := func(_ ProviderConfig, _ Embedder) (Provider, error) {
		return &mockProvider{name: "test"}, nil
	}

	r.Register("test", factory, PriorityThin)

	if !r.Has("test") {
		t.Error("expected registry to have 'test' provider")
	}

	r.Unregister("test")

	if r.Has("test") {
		t.Error("expected registry to not have 'test' provider after unregister")
	}
}

func TestProviderName_Valid(t *testing.T) {
	tests := []struct {
		name  ProviderName
		valid bool
	}{
		{ProviderNameMemory, true},
		{ProviderNamePostgres, true},
		{ProviderNameKVS, true},
		{ProviderNameMem0, true},
		{ProviderNameGraphiti, true},
		{ProviderNameTwilio, true},
		{ProviderName("invalid"), false},
		{ProviderName(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.name), func(t *testing.T) {
			if got := tt.name.Valid(); got != tt.valid {
				t.Errorf("ProviderName(%q).Valid() = %v, want %v", tt.name, got, tt.valid)
			}
		})
	}
}
