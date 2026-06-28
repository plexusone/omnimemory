package core

// Scope defines the visibility and ownership level of a memory.
type Scope string

const (
	// ScopeUser represents personal memories for one user.
	ScopeUser Scope = "user"
	// ScopeAgent represents what an agent has learned.
	ScopeAgent Scope = "agent"
	// ScopeTenant represents org-level shared memories.
	ScopeTenant Scope = "tenant"
	// ScopeTeam represents group/project level memories.
	ScopeTeam Scope = "team"
	// ScopeSession represents short-lived conversation memories.
	ScopeSession Scope = "session"
	// ScopeDomain represents domain-specific memories (support, sales, etc.).
	ScopeDomain Scope = "domain"
)

// Valid returns true if the scope is a valid scope value.
func (s Scope) Valid() bool {
	switch s {
	case ScopeUser, ScopeAgent, ScopeTenant, ScopeTeam, ScopeSession, ScopeDomain:
		return true
	default:
		return false
	}
}

// String returns the string representation of the scope.
func (s Scope) String() string {
	return string(s)
}

// MemoryType defines the category of memory content.
type MemoryType string

const (
	// MemoryTypeObservation represents an observed behavior or interaction.
	MemoryTypeObservation MemoryType = "observation"
	// MemoryTypeFact represents a verified piece of information.
	MemoryTypeFact MemoryType = "fact"
	// MemoryTypePreference represents a user preference.
	MemoryTypePreference MemoryType = "preference"
	// MemoryTypeSummary represents a summarized conversation or topic.
	MemoryTypeSummary MemoryType = "summary"
	// MemoryTypeTrait represents a personality trait or characteristic.
	MemoryTypeTrait MemoryType = "trait"
	// MemoryTypeRelationship represents a relationship between entities.
	MemoryTypeRelationship MemoryType = "relationship"
)

// Valid returns true if the memory type is a valid type value.
func (t MemoryType) Valid() bool {
	switch t {
	case MemoryTypeObservation, MemoryTypeFact, MemoryTypePreference,
		MemoryTypeSummary, MemoryTypeTrait, MemoryTypeRelationship:
		return true
	default:
		return false
	}
}

// String returns the string representation of the memory type.
func (t MemoryType) String() string {
	return string(t)
}
