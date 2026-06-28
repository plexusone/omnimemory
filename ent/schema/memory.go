package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Memory holds the schema definition for the Memory entity.
type Memory struct {
	ent.Schema
}

// Fields of the Memory.
func (Memory) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			Immutable().
			NotEmpty().
			Comment("Unique identifier for the memory"),

		field.String("tenant_id").
			NotEmpty().
			Comment("Tenant ID for multi-tenancy"),

		field.String("subject_id").
			NotEmpty().
			Comment("Who this memory is about"),

		field.String("agent_id").
			Optional().
			Comment("Agent that created or owns this memory"),

		field.String("session_id").
			Optional().
			Comment("Session that created this memory"),

		field.String("scope").
			NotEmpty().
			Comment("Memory scope: user, agent, tenant, team, session, domain"),

		field.String("type").
			NotEmpty().
			Comment("Memory type: observation, fact, preference, summary, trait, relationship"),

		field.Text("content").
			NotEmpty().
			Comment("The memory content"),

		// Store embedding as JSON array of float64
		// PostgreSQL schema can be customized via migration to use vector type
		field.JSON("embedding", []float32{}).
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "vector(1536)",
			}).
			Comment("Vector embedding for semantic search"),

		field.JSON("metadata", map[string]any{}).
			Optional().
			Comment("Additional metadata"),

		field.Time("created_at").
			Immutable().
			Default(time.Now).
			Comment("When the memory was created"),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("When the memory was last updated"),

		field.Time("expires_at").
			Optional().
			Nillable().
			Comment("When the memory expires (nil = never)"),
	}
}

// Edges of the Memory.
func (Memory) Edges() []ent.Edge {
	return nil
}

// Indexes of the Memory.
func (Memory) Indexes() []ent.Index {
	return []ent.Index{
		// Fast lookup by tenant
		index.Fields("tenant_id"),

		// Fast lookup by tenant + subject
		index.Fields("tenant_id", "subject_id"),

		// Fast lookup by tenant + scope
		index.Fields("tenant_id", "scope"),

		// Fast lookup by tenant + type
		index.Fields("tenant_id", "type"),

		// Composite index for common query patterns
		index.Fields("tenant_id", "subject_id", "scope"),

		// HNSW index for vector similarity search
		index.Fields("embedding").
			Annotations(
				entsql.IndexType("hnsw"),
				entsql.OpClass("vector_l2_ops"),
			),
	}
}
