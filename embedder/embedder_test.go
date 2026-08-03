package embedder

import (
	"testing"

	"github.com/plexusone/omnimemory/core"
)

func TestNewFromConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     core.EmbedderConfig
		wantErr bool
	}{
		{
			name:    "openai with api key",
			cfg:     core.EmbedderConfig{Provider: "openai", APIKey: "sk-test"},
			wantErr: false,
		},
		{
			name:    "openai case-insensitive",
			cfg:     core.EmbedderConfig{Provider: "OpenAI", APIKey: "sk-test"},
			wantErr: false,
		},
		{
			name:    "openai missing api key",
			cfg:     core.EmbedderConfig{Provider: "openai"},
			wantErr: true,
		},
		{
			name:    "empty provider",
			cfg:     core.EmbedderConfig{},
			wantErr: true,
		},
		{
			name:    "unknown provider",
			cfg:     core.EmbedderConfig{Provider: "cohere", APIKey: "x"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emb, err := NewFromConfig(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if emb == nil {
				t.Fatal("expected non-nil embedder")
			}
		})
	}
}

func TestNewFromConfigDimension(t *testing.T) {
	t.Run("defaults dimension when unset", func(t *testing.T) {
		emb, err := NewFromConfig(core.EmbedderConfig{Provider: "openai", APIKey: "sk-test"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if emb.Dimension() != core.DefaultEmbeddingDimension {
			t.Errorf("Dimension() = %d, want %d", emb.Dimension(), core.DefaultEmbeddingDimension)
		}
	})

	t.Run("honors explicit dimension", func(t *testing.T) {
		emb, err := NewFromConfig(core.EmbedderConfig{Provider: "openai", APIKey: "sk-test", Dimension: 256})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if emb.Dimension() != 256 {
			t.Errorf("Dimension() = %d, want 256", emb.Dimension())
		}
	})
}
