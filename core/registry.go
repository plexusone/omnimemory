package core

import (
	"sort"
	"sync"
)

// Priority levels for provider registration.
const (
	PriorityThin  = 0  // Lightweight implementations (in-memory, mock)
	PriorityThick = 10 // Full SDK implementations (PostgreSQL, external services)
)

// ProviderFactory is a function that creates a new Provider from config.
type ProviderFactory func(config ProviderConfig, embedder Embedder) (Provider, error)

// registeredProvider holds provider registration info.
type registeredProvider struct {
	name     ProviderName
	factory  ProviderFactory
	priority int
}

// Registry manages provider registrations.
type Registry struct {
	mu        sync.RWMutex
	providers map[ProviderName]*registeredProvider
}

// globalRegistry is the default registry instance.
var globalRegistry = &Registry{
	providers: make(map[ProviderName]*registeredProvider),
}

// RegisterProvider registers a provider factory with the global registry.
func RegisterProvider(name ProviderName, factory ProviderFactory, priority int) {
	globalRegistry.Register(name, factory, priority)
}

// GetProvider creates a provider instance from the global registry.
func GetProvider(name ProviderName, config ProviderConfig, embedder Embedder) (Provider, error) {
	return globalRegistry.Get(name, config, embedder)
}

// ListProviders returns all registered provider names from the global registry.
func ListProviders() []ProviderName {
	return globalRegistry.List()
}

// Register registers a provider factory with this registry.
func (r *Registry) Register(name ProviderName, factory ProviderFactory, priority int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.providers[name] = &registeredProvider{
		name:     name,
		factory:  factory,
		priority: priority,
	}
}

// Get creates a provider instance from this registry.
func (r *Registry) Get(name ProviderName, config ProviderConfig, embedder Embedder) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reg, ok := r.providers[name]
	if !ok {
		return nil, ErrProviderNotFound
	}

	config.Name = name
	return reg.factory(config, embedder)
}

// List returns all registered provider names sorted by priority (highest first).
func (r *Registry) List() []ProviderName {
	r.mu.RLock()
	defer r.mu.RUnlock()

	regs := make([]*registeredProvider, 0, len(r.providers))
	for _, reg := range r.providers {
		regs = append(regs, reg)
	}

	sort.Slice(regs, func(i, j int) bool {
		return regs[i].priority > regs[j].priority
	})

	names := make([]ProviderName, len(regs))
	for i, reg := range regs {
		names[i] = reg.name
	}
	return names
}

// Has returns true if a provider is registered.
func (r *Registry) Has(name ProviderName) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.providers[name]
	return ok
}

// Unregister removes a provider from the registry.
func (r *Registry) Unregister(name ProviderName) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.providers, name)
}
