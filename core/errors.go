// Package core provides the core types and interfaces for omnimemory.
package core

import (
	"errors"
	"fmt"
)

// Common errors returned by omnimemory operations.
var (
	ErrNotFound         = errors.New("memory not found")
	ErrInvalidInput     = errors.New("invalid input")
	ErrProviderNotFound = errors.New("provider not found")
	ErrNoProviders      = errors.New("no providers configured")
	ErrEmbeddingFailed  = errors.New("embedding generation failed")
	ErrTenantRequired   = errors.New("tenant_id is required")
	ErrSubjectRequired  = errors.New("subject_id is required")
	ErrContentRequired  = errors.New("content is required")
)

// ProviderError wraps an error with provider context.
type ProviderError struct {
	Provider string
	Op       string
	Err      error
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("omnimemory: provider %s: %s: %v", e.Provider, e.Op, e.Err)
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

// NewProviderError creates a new ProviderError.
func NewProviderError(provider, op string, err error) *ProviderError {
	return &ProviderError{
		Provider: provider,
		Op:       op,
		Err:      err,
	}
}

// ValidationError represents a validation error with field context.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("omnimemory: validation error: %s: %s", e.Field, e.Message)
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
