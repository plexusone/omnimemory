// Package openai provides an OpenAI-backed implementation of core.Embedder.
//
// It wraps the official openai-go SDK's embeddings endpoint and produces
// float64 vectors suitable for the semantic-search paths in omnimemory
// providers (e.g. the kvs provider's in-memory cosine similarity).
//
// # Usage
//
//	emb, err := openai.New(openai.Config{
//	    APIKey: os.Getenv("OPENAI_API_KEY"),
//	    Model:  "text-embedding-3-small",
//	})
//	if err != nil {
//	    // handle error
//	}
//
//	client, err := core.NewClient(core.ClientConfig{
//	    Providers: []core.ProviderConfig{{Name: core.ProviderNameKVS, Options: opts}},
//	    Embedder:  emb,
//	})
package openai

import (
	"context"
	"errors"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"github.com/plexusone/omnimemory/core"
)

// DefaultModel is the embedding model used when Config.Model is empty.
const DefaultModel = openai.EmbeddingModelTextEmbedding3Small

// Config configures the OpenAI embedder.
type Config struct {
	// APIKey is the OpenAI API key. Required.
	APIKey string

	// Model is the embedding model (default: text-embedding-3-small).
	Model string

	// Dimension optionally overrides the output dimension. Only supported by
	// text-embedding-3 and later models. When 0, the model default is used.
	Dimension int

	// BaseURL optionally overrides the API endpoint (for proxies or
	// OpenAI-compatible services). When empty, the SDK default is used.
	BaseURL string
}

// Embedder implements core.Embedder using the OpenAI embeddings API.
type Embedder struct {
	client    openai.Client
	model     openai.EmbeddingModel
	dimension int
}

// Verify interface compliance.
var _ core.Embedder = (*Embedder)(nil)

// New creates an OpenAI embedder from the given configuration.
func New(cfg Config) (*Embedder, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("openai embedder: APIKey is required")
	}

	model := cfg.Model
	if model == "" {
		model = DefaultModel
	}

	opts := []option.RequestOption{option.WithAPIKey(cfg.APIKey)}
	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}

	dimension := cfg.Dimension
	if dimension <= 0 {
		dimension = core.DefaultEmbeddingDimension
	}

	return &Embedder{
		client:    openai.NewClient(opts...),
		model:     model,
		dimension: dimension,
	}, nil
}

// Embed generates an embedding for a single text.
func (e *Embedder) Embed(ctx context.Context, text string) ([]float64, error) {
	embeddings, err := e.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, errors.New("openai embedder: no embedding returned")
	}
	return embeddings[0], nil
}

// EmbedBatch generates embeddings for multiple texts.
//
// The OpenAI response is ordered by an explicit index; results are sorted back
// into input order before returning so callers can rely on positional mapping.
func (e *Embedder) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	params := openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{OfArrayOfStrings: texts},
		Model: e.model,
	}
	if e.dimension > 0 {
		params.Dimensions = openai.Int(int64(e.dimension))
	}

	resp, err := e.client.Embeddings.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("openai embedder: create embeddings: %w", err)
	}
	if len(resp.Data) != len(texts) {
		return nil, fmt.Errorf("openai embedder: expected %d embeddings, got %d", len(texts), len(resp.Data))
	}

	embeddings := make([][]float64, len(texts))
	for i := range resp.Data {
		d := resp.Data[i]
		idx := int(d.Index)
		if idx < 0 || idx >= len(embeddings) {
			return nil, fmt.Errorf("openai embedder: embedding index %d out of range", idx)
		}
		embeddings[idx] = d.Embedding
	}
	return embeddings, nil
}

// Dimension returns the embedding dimension.
func (e *Embedder) Dimension() int {
	return e.dimension
}
