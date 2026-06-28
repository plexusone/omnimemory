package core

import (
	"time"
)

// Memory is the primary storage unit for omnimemory.
type Memory struct {
	ID        string         `json:"id"`
	TenantID  string         `json:"tenant_id"`
	SubjectID string         `json:"subject_id"` // Who this memory is about
	AgentID   string         `json:"agent_id,omitempty"`
	SessionID string         `json:"session_id,omitempty"` // Session that created this memory
	Scope     Scope          `json:"scope"`
	Type      MemoryType     `json:"type"`
	Content   string         `json:"content"`
	Embedding []float64      `json:"embedding,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	ExpiresAt *time.Time     `json:"expires_at,omitempty"`
}

// Observation represents an observed behavior or interaction extracted from conversation.
type Observation struct {
	Content   string         `json:"content"`
	Source    string         `json:"source,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// Fact represents a verified piece of information extracted from text.
type Fact struct {
	Subject   string         `json:"subject"`
	Predicate string         `json:"predicate"`
	Object    string         `json:"object"`
	Source    string         `json:"source,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// Context provides multi-tenancy context for all operations.
type Context struct {
	TenantID       string `json:"tenant_id"`
	SubjectID      string `json:"subject_id"`   // Who the memory is about
	PrincipalID    string `json:"principal_id"` // Who is making the request
	AgentID        string `json:"agent_id,omitempty"`
	SessionID      string `json:"session_id,omitempty"`      // Current session identifier
	ConversationID string `json:"conversation_id,omitempty"` // Conversation within a session
	Scope          Scope  `json:"scope,omitempty"`
}

// Validate validates the context.
func (c *Context) Validate() error {
	if c.TenantID == "" {
		return ErrTenantRequired
	}
	if c.SubjectID == "" {
		return ErrSubjectRequired
	}
	return nil
}

// AddRequest is the request for adding a new memory.
type AddRequest struct {
	Context
	Type     MemoryType     `json:"type"`
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata,omitempty"`
	TTL      time.Duration  `json:"ttl,omitempty"`
}

// Validate validates the add request.
func (r *AddRequest) Validate() error {
	if err := r.Context.Validate(); err != nil {
		return err
	}
	if r.Content == "" {
		return ErrContentRequired
	}
	if !r.Type.Valid() {
		return NewValidationError("type", "invalid memory type")
	}
	return nil
}

// GetRequest is the request for getting a memory by ID.
type GetRequest struct {
	Context
	ID string `json:"id"`
}

// Validate validates the get request.
func (r *GetRequest) Validate() error {
	if err := r.Context.Validate(); err != nil {
		return err
	}
	if r.ID == "" {
		return NewValidationError("id", "id is required")
	}
	return nil
}

// UpdateRequest is the request for updating a memory.
type UpdateRequest struct {
	Context
	ID       string         `json:"id"`
	Content  string         `json:"content,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Validate validates the update request.
func (r *UpdateRequest) Validate() error {
	if err := r.Context.Validate(); err != nil {
		return err
	}
	if r.ID == "" {
		return NewValidationError("id", "id is required")
	}
	return nil
}

// DeleteRequest is the request for deleting a memory.
type DeleteRequest struct {
	Context
	ID string `json:"id"`
}

// Validate validates the delete request.
func (r *DeleteRequest) Validate() error {
	if err := r.Context.Validate(); err != nil {
		return err
	}
	if r.ID == "" {
		return NewValidationError("id", "id is required")
	}
	return nil
}

// ListRequest is the request for listing memories.
type ListRequest struct {
	Context
	Types  []MemoryType `json:"types,omitempty"`
	Scopes []Scope      `json:"scopes,omitempty"`
	Limit  int          `json:"limit,omitempty"`
	Offset int          `json:"offset,omitempty"`
}

// ListResponse is the response for listing memories.
type ListResponse struct {
	Memories   []*Memory `json:"memories"`
	TotalCount int       `json:"total_count"`
	HasMore    bool      `json:"has_more"`
}

// SearchRequest is the request for semantic search.
type SearchRequest struct {
	Context
	Query     string       `json:"query"`
	Types     []MemoryType `json:"types,omitempty"`
	Scopes    []Scope      `json:"scopes,omitempty"`
	Limit     int          `json:"limit,omitempty"`
	Threshold float64      `json:"threshold,omitempty"` // Similarity threshold
}

// Validate validates the search request.
func (r *SearchRequest) Validate() error {
	if err := r.Context.Validate(); err != nil {
		return err
	}
	if r.Query == "" {
		return NewValidationError("query", "query is required")
	}
	return nil
}

// SearchResult is a single search result with similarity score.
type SearchResult struct {
	Memory *Memory `json:"memory"`
	Score  float64 `json:"score"`
}

// SearchResponse is the response for semantic search.
type SearchResponse struct {
	Results []*SearchResult `json:"results"`
}

// RecallRequest is the request for recalling relevant memories.
type RecallRequest struct {
	Context
	Query        string       `json:"query"`
	MaxResults   int          `json:"max_results,omitempty"`
	IncludeTypes []MemoryType `json:"include_types,omitempty"`
}

// Validate validates the recall request.
func (r *RecallRequest) Validate() error {
	if err := r.Context.Validate(); err != nil {
		return err
	}
	if r.Query == "" {
		return NewValidationError("query", "query is required")
	}
	return nil
}

// RecallResponse is the response for recalling memories.
type RecallResponse struct {
	Memories []*Memory `json:"memories"`
	Summary  string    `json:"summary,omitempty"`
}
