// Package embedder provides a factory for constructing core.Embedder
// implementations from configuration.
//
// It decouples callers from concrete embedder packages: pass a
// core.EmbedderConfig with a Provider name and the factory returns the matching
// implementation. New providers are added here as they are implemented.
package embedder

import (
	"fmt"
	"strings"

	"github.com/plexusone/omnimemory/core"
	openaiembedder "github.com/plexusone/omnimemory/embedder/openai"
)

// Provider names understood by NewFromConfig.
const (
	// ProviderOpenAI selects the OpenAI embeddings implementation.
	ProviderOpenAI = "openai"
)

// NewFromConfig constructs a core.Embedder from the given configuration,
// dispatching on cfg.Provider. Returns an error for unknown providers.
func NewFromConfig(cfg core.EmbedderConfig) (core.Embedder, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case ProviderOpenAI:
		return openaiembedder.New(openaiembedder.Config{
			APIKey:    cfg.APIKey,
			Model:     cfg.Model,
			Dimension: cfg.Dimension,
			BaseURL:   cfg.Endpoint,
		})
	case "":
		return nil, fmt.Errorf("embedder: provider is required")
	default:
		return nil, fmt.Errorf("embedder: unknown provider %q", cfg.Provider)
	}
}
