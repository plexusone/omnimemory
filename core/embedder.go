package core

import (
	"context"
)

// Embedder generates embeddings for text.
type Embedder interface {
	// Embed generates an embedding for a single text.
	Embed(ctx context.Context, text string) ([]float64, error)

	// EmbedBatch generates embeddings for multiple texts.
	EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)

	// Dimension returns the embedding dimension.
	Dimension() int
}

// EmbedderConfig is the configuration for creating an Embedder.
type EmbedderConfig struct {
	// Provider is the embedding provider (openai, anthropic, cohere, etc.)
	Provider string `json:"provider"`

	// APIKey is the API key for the provider.
	APIKey string `json:"api_key"`

	// Model is the embedding model to use.
	Model string `json:"model"`

	// Dimension is the embedding dimension (if configurable).
	Dimension int `json:"dimension,omitempty"`

	// Endpoint is the custom API endpoint (optional).
	Endpoint string `json:"endpoint,omitempty"`
}

// DefaultEmbeddingDimension is the default dimension for OpenAI text-embedding-3-small.
const DefaultEmbeddingDimension = 1536

// CosineSimilarity computes the cosine similarity between two vectors.
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (sqrt(normA) * sqrt(normB))
}

// sqrt computes the square root without importing math package.
func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 100; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// EuclideanDistance computes the Euclidean distance between two vectors.
func EuclideanDistance(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var sum float64
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}

	return sqrt(sum)
}
