package kvs

import (
	"testing"

	"github.com/plexusone/omnimemory/core"
	"github.com/plexusone/omnimemory/core/providertest"
	kvsmemory "github.com/plexusone/omnistorage-core/kvs/backend/memory"
)

func TestConformance(t *testing.T) {
	store := kvsmemory.New()
	t.Cleanup(func() { _ = store.Close() })

	p, err := NewProvider(core.ProviderConfig{
		Options: map[string]any{
			"store": store,
		},
	}, nil)
	if err != nil {
		t.Fatalf("NewProvider() error: %v", err)
	}

	providertest.RunAll(t, providertest.Config{
		Provider: p,
	})
}
