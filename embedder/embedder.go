// Package embedder constructs core.Embedder implementations from configuration.
//
// Embeddings are produced by omnillm-core's embedding providers; this package
// adapts an omnillm-core EmbeddingProvider to omnimemory's core.Embedder
// interface. Any provider omnillm-core registers (openai today, more over time)
// is selectable via core.EmbedderConfig.Provider, so omnimemory does not
// maintain its own per-vendor embedding integrations.
package embedder

import (
	"context"
	"errors"
	"fmt"
	"strings"

	llmcore "github.com/plexusone/omnillm-core"
	"github.com/plexusone/omnillm-core/provider"

	"github.com/plexusone/omnimemory/core"
)

// DefaultModel is used when core.EmbedderConfig.Model is empty.
const DefaultModel = "text-embedding-3-small"

// NewFromConfig constructs a core.Embedder backed by an omnillm-core embedding
// provider selected by cfg.Provider (e.g. "openai"). cfg.Endpoint overrides the
// provider base URL when set, and cfg.Dimension, when > 0, requests a specific
// output dimension (only supported by some models).
func NewFromConfig(cfg core.EmbedderConfig) (core.Embedder, error) {
	name := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if name == "" {
		return nil, errors.New("embedder: provider is required")
	}

	p, err := llmcore.GetEmbeddingProvider(
		llmcore.ProviderName(name),
		llmcore.ProviderConfig{
			Provider: llmcore.ProviderName(name),
			APIKey:   cfg.APIKey,
			BaseURL:  cfg.Endpoint,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("embedder: %w", err)
	}

	model := cfg.Model
	if model == "" {
		model = DefaultModel
	}

	// Report a dimension for schema/validation callers. Only forward an explicit
	// dimension to the provider; sending the default would break models that do
	// not support the dimensions parameter (e.g. text-embedding-ada-002).
	reported := cfg.Dimension
	if reported <= 0 {
		reported = core.DefaultEmbeddingDimension
	}

	return &omnillmEmbedder{
		provider:      p,
		model:         model,
		dimension:     reported,
		sendDimension: cfg.Dimension > 0,
	}, nil
}

// omnillmEmbedder adapts an omnillm-core EmbeddingProvider to core.Embedder.
type omnillmEmbedder struct {
	provider      provider.EmbeddingProvider
	model         string
	dimension     int
	sendDimension bool
}

var _ core.Embedder = (*omnillmEmbedder)(nil)

// Embed generates an embedding for a single text.
func (e *omnillmEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
	embeddings, err := e.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, errors.New("embedder: no embedding returned")
	}
	return embeddings[0], nil
}

// EmbedBatch generates embeddings for multiple texts, returned in input order.
func (e *omnillmEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	req := &provider.EmbeddingRequest{Model: e.model, Input: texts}
	if e.sendDimension {
		d := e.dimension
		req.Dimensions = &d
	}

	resp, err := e.provider.CreateEmbedding(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("embedder: create embeddings: %w", err)
	}
	if len(resp.Data) != len(texts) {
		return nil, fmt.Errorf("embedder: expected %d embeddings, got %d", len(texts), len(resp.Data))
	}

	// Providers return data with an explicit index; sort back into input order.
	embeddings := make([][]float64, len(texts))
	for i := range resp.Data {
		d := resp.Data[i]
		if d.Index < 0 || d.Index >= len(embeddings) {
			return nil, fmt.Errorf("embedder: embedding index %d out of range", d.Index)
		}
		embeddings[d.Index] = d.Embedding
	}
	return embeddings, nil
}

// Dimension returns the embedding dimension.
func (e *omnillmEmbedder) Dimension() int {
	return e.dimension
}
