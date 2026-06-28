package core

import (
	"context"
	"log/slog"
	"time"
)

// ClientConfig is the configuration for creating a Client.
type ClientConfig struct {
	// Providers is the list of provider configurations.
	// The first provider is the primary, others are fallbacks.
	Providers []ProviderConfig `json:"providers"`

	// DefaultTTL is the default time-to-live for memories.
	DefaultTTL time.Duration `json:"default_ttl,omitempty"`

	// Embedder is the embedding generator.
	Embedder Embedder `json:"-"`

	// Logger is the logger for the client.
	Logger *slog.Logger `json:"-"`
}

// ObservabilityHook is called for provider operations.
type ObservabilityHook func(provider string, op string, duration time.Duration, err error)

// Client is a multi-provider memory client with fallback support.
type Client struct {
	providers map[ProviderName]Provider
	primary   ProviderName
	fallbacks []ProviderName
	embedder  Embedder
	config    ClientConfig
	hook      ObservabilityHook
	logger    *slog.Logger
}

// NewClient creates a new Client with the given configuration.
func NewClient(config ClientConfig) (*Client, error) {
	if len(config.Providers) == 0 {
		return nil, ErrNoProviders
	}

	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	embedder := config.Embedder
	if embedder == nil {
		embedder = &noopEmbedder{}
	}

	c := &Client{
		providers: make(map[ProviderName]Provider),
		embedder:  embedder,
		config:    config,
		logger:    logger,
	}

	// Initialize providers
	for i, pc := range config.Providers {
		provider, err := GetProvider(pc.Name, pc, embedder)
		if err != nil {
			logger.Warn("failed to initialize provider",
				"provider", pc.Name,
				"error", err)
			continue
		}

		c.providers[pc.Name] = provider

		if i == 0 {
			c.primary = pc.Name
		} else {
			c.fallbacks = append(c.fallbacks, pc.Name)
		}
	}

	if c.primary == "" {
		return nil, ErrNoProviders
	}

	return c, nil
}

// SetHook sets the observability hook.
func (c *Client) SetHook(hook ObservabilityHook) {
	c.hook = hook
}

// Add adds a new memory.
func (c *Client) Add(ctx context.Context, req *AddRequest) (*Memory, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	return c.withFallback(ctx, "Add", func(p Provider) (*Memory, error) {
		return p.Add(ctx, req)
	})
}

// Get retrieves a memory by ID.
func (c *Client) Get(ctx context.Context, req *GetRequest) (*Memory, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	return c.withFallback(ctx, "Get", func(p Provider) (*Memory, error) {
		return p.Get(ctx, req)
	})
}

// Update updates an existing memory.
func (c *Client) Update(ctx context.Context, req *UpdateRequest) (*Memory, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	return c.withFallback(ctx, "Update", func(p Provider) (*Memory, error) {
		return p.Update(ctx, req)
	})
}

// Delete deletes a memory by ID.
func (c *Client) Delete(ctx context.Context, req *DeleteRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	_, err := c.withFallback(ctx, "Delete", func(p Provider) (*Memory, error) {
		return nil, p.Delete(ctx, req)
	})
	return err
}

// List lists memories with optional filters.
func (c *Client) List(ctx context.Context, req *ListRequest) (*ListResponse, error) {
	provider := c.providers[c.primary]
	if provider == nil {
		return nil, ErrNoProviders
	}

	start := time.Now()
	resp, err := provider.List(ctx, req)
	c.recordOp(c.primary.String(), "List", time.Since(start), err)
	return resp, err
}

// Search performs semantic search on memories.
func (c *Client) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	provider := c.providers[c.primary]
	if provider == nil {
		return nil, ErrNoProviders
	}

	start := time.Now()
	resp, err := provider.Search(ctx, req)
	c.recordOp(c.primary.String(), "Search", time.Since(start), err)
	return resp, err
}

// Recall retrieves relevant memories for a given query.
func (c *Client) Recall(ctx context.Context, req *RecallRequest) (*RecallResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	provider := c.providers[c.primary]
	if provider == nil {
		return nil, ErrNoProviders
	}

	start := time.Now()
	resp, err := provider.Recall(ctx, req)
	c.recordOp(c.primary.String(), "Recall", time.Since(start), err)
	return resp, err
}

// Close closes all providers.
func (c *Client) Close() error {
	var lastErr error
	for name, provider := range c.providers {
		if err := provider.Close(); err != nil {
			c.logger.Error("failed to close provider",
				"provider", name,
				"error", err)
			lastErr = err
		}
	}
	return lastErr
}

// Primary returns the primary provider name.
func (c *Client) Primary() ProviderName {
	return c.primary
}

// Providers returns all provider names.
func (c *Client) Providers() []ProviderName {
	names := make([]ProviderName, 0, len(c.providers))
	for name := range c.providers {
		names = append(names, name)
	}
	return names
}

// withFallback executes an operation with fallback to other providers.
func (c *Client) withFallback(ctx context.Context, op string, fn func(Provider) (*Memory, error)) (*Memory, error) {
	// Try primary first
	if provider := c.providers[c.primary]; provider != nil {
		start := time.Now()
		result, err := fn(provider)
		c.recordOp(c.primary.String(), op, time.Since(start), err)
		if err == nil {
			return result, nil
		}
		c.logger.Warn("primary provider failed, trying fallbacks",
			"provider", c.primary,
			"op", op,
			"error", err)
	}

	// Try fallbacks
	for _, name := range c.fallbacks {
		provider := c.providers[name]
		if provider == nil {
			continue
		}

		start := time.Now()
		result, err := fn(provider)
		c.recordOp(name.String(), op, time.Since(start), err)
		if err == nil {
			return result, nil
		}
		c.logger.Warn("fallback provider failed",
			"provider", name,
			"op", op,
			"error", err)
	}

	return nil, NewProviderError(c.primary.String(), op, ErrNoProviders)
}

// recordOp records operation metrics via the hook.
func (c *Client) recordOp(provider, op string, duration time.Duration, err error) {
	if c.hook != nil {
		c.hook(provider, op, duration, err)
	}
}

// noopEmbedder is a no-op embedder that returns empty embeddings.
type noopEmbedder struct{}

func (e *noopEmbedder) Embed(_ context.Context, _ string) ([]float64, error) {
	return nil, nil
}

func (e *noopEmbedder) EmbedBatch(_ context.Context, _ []string) ([][]float64, error) {
	return nil, nil
}

func (e *noopEmbedder) Dimension() int {
	return 0
}
