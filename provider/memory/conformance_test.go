package memory

import (
	"testing"

	"github.com/plexusone/omnimemory/core"
	"github.com/plexusone/omnimemory/core/providertest"
)

func TestConformance(t *testing.T) {
	p, err := NewProvider(core.ProviderConfig{}, nil)
	if err != nil {
		t.Fatalf("NewProvider() error: %v", err)
	}
	defer func() { _ = p.Close() }()

	providertest.RunAll(t, providertest.Config{
		Provider: p,
	})
}
