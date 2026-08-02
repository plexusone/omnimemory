package openai

import (
	"testing"

	"github.com/plexusone/omnimemory/core"
)

func TestNew(t *testing.T) {
	t.Run("requires api key", func(t *testing.T) {
		if _, err := New(Config{}); err == nil {
			t.Fatal("expected error for missing api key")
		}
	})

	t.Run("defaults model and dimension", func(t *testing.T) {
		emb, err := New(Config{APIKey: "sk-test"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if emb.model != DefaultModel {
			t.Errorf("model = %q, want %q", emb.model, DefaultModel)
		}
		if emb.Dimension() != core.DefaultEmbeddingDimension {
			t.Errorf("dimension = %d, want %d", emb.Dimension(), core.DefaultEmbeddingDimension)
		}
	})

	t.Run("honors overrides", func(t *testing.T) {
		emb, err := New(Config{APIKey: "sk-test", Model: "text-embedding-3-large", Dimension: 256})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if emb.model != "text-embedding-3-large" {
			t.Errorf("model = %q, want text-embedding-3-large", emb.model)
		}
		if emb.Dimension() != 256 {
			t.Errorf("dimension = %d, want 256", emb.Dimension())
		}
	})
}

func TestInterfaceCompliance(t *testing.T) {
	emb, err := New(Config{APIKey: "sk-test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var _ core.Embedder = emb
}
