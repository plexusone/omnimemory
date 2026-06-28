// Package postgres provides a PostgreSQL+pgvector Provider implementation via Ent.
package postgres

import (
	"context"
	"database/sql"
	"sort"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	"github.com/pgvector/pgvector-go"
	"github.com/plexusone/omnimemory/core"
	"github.com/plexusone/omnimemory/ent"
	"github.com/plexusone/omnimemory/ent/memory"
)

func init() {
	core.RegisterProvider(core.ProviderNamePostgres, NewProvider, core.PriorityThick)
}

// Provider wraps Ent client for PostgreSQL+pgvector.
type Provider struct {
	client   *ent.Client
	db       *sql.DB
	embedder core.Embedder
	config   core.ProviderConfig
}

// NewProvider creates a new PostgreSQL Provider.
func NewProvider(config core.ProviderConfig, embedder core.Embedder) (core.Provider, error) {
	if config.DSN == "" {
		return nil, core.NewValidationError("dsn", "DSN is required for postgres provider")
	}

	db, err := sql.Open("pgx", config.DSN)
	if err != nil {
		return nil, core.NewProviderError("postgres", "Open", err)
	}

	driver := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(driver))

	return &Provider{
		client:   client,
		db:       db,
		embedder: embedder,
		config:   config,
	}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return core.ProviderNamePostgres.String()
}

// Close closes the provider.
func (p *Provider) Close() error {
	if err := p.client.Close(); err != nil {
		return err
	}
	return p.db.Close()
}

// Add adds a new memory.
func (p *Provider) Add(ctx context.Context, req *core.AddRequest) (*core.Memory, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	id := uuid.New().String()

	// Generate embedding
	var embedding []float32
	if p.embedder != nil {
		emb, err := p.embedder.Embed(ctx, req.Content)
		if err != nil {
			return nil, core.NewProviderError(p.Name(), "Add", err)
		}
		embedding = float32Slice(emb)
	}

	// Build create query
	create := p.client.Memory.Create().
		SetID(id).
		SetTenantID(req.TenantID).
		SetSubjectID(req.SubjectID).
		SetScope(req.Scope.String()).
		SetType(req.Type.String()).
		SetContent(req.Content).
		SetCreatedAt(now).
		SetUpdatedAt(now)

	if req.AgentID != "" {
		create.SetAgentID(req.AgentID)
	}

	if req.SessionID != "" {
		create.SetSessionID(req.SessionID)
	}

	if len(embedding) > 0 {
		create.SetEmbedding(embedding)
	}

	if req.Metadata != nil {
		create.SetMetadata(req.Metadata)
	}

	if req.TTL > 0 {
		expiresAt := now.Add(req.TTL)
		create.SetExpiresAt(expiresAt)
	}

	entMem, err := create.Save(ctx)
	if err != nil {
		return nil, core.NewProviderError(p.Name(), "Add", err)
	}

	return entToCore(entMem), nil
}

// Get retrieves a memory by ID.
func (p *Provider) Get(ctx context.Context, req *core.GetRequest) (*core.Memory, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	entMem, err := p.client.Memory.Query().
		Where(
			memory.ID(req.ID),
			memory.TenantID(req.TenantID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, core.ErrNotFound
		}
		return nil, core.NewProviderError(p.Name(), "Get", err)
	}

	// Check expiration
	if entMem.ExpiresAt != nil && time.Now().After(*entMem.ExpiresAt) {
		return nil, core.ErrNotFound
	}

	return entToCore(entMem), nil
}

// Update updates an existing memory.
func (p *Provider) Update(ctx context.Context, req *core.UpdateRequest) (*core.Memory, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// First check if memory exists and belongs to tenant
	existing, err := p.Get(ctx, &core.GetRequest{
		Context: req.Context,
		ID:      req.ID,
	})
	if err != nil {
		return nil, err
	}

	update := p.client.Memory.UpdateOneID(req.ID).
		SetUpdatedAt(time.Now())

	if req.Content != "" {
		update.SetContent(req.Content)

		// Regenerate embedding
		if p.embedder != nil {
			emb, err := p.embedder.Embed(ctx, req.Content)
			if err != nil {
				return nil, core.NewProviderError(p.Name(), "Update", err)
			}
			update.SetEmbedding(float32Slice(emb))
		}
	}

	if req.Metadata != nil {
		merged := existing.Metadata
		if merged == nil {
			merged = make(map[string]any)
		}
		for k, v := range req.Metadata {
			merged[k] = v
		}
		update.SetMetadata(merged)
	}

	entMem, err := update.Save(ctx)
	if err != nil {
		return nil, core.NewProviderError(p.Name(), "Update", err)
	}

	return entToCore(entMem), nil
}

// Delete deletes a memory by ID.
func (p *Provider) Delete(ctx context.Context, req *core.DeleteRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	// Check tenant isolation
	_, err := p.client.Memory.Query().
		Where(
			memory.ID(req.ID),
			memory.TenantID(req.TenantID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return core.ErrNotFound
		}
		return core.NewProviderError(p.Name(), "Delete", err)
	}

	err = p.client.Memory.DeleteOneID(req.ID).Exec(ctx)
	if err != nil {
		return core.NewProviderError(p.Name(), "Delete", err)
	}

	return nil
}

// List lists memories with optional filters.
func (p *Provider) List(ctx context.Context, req *core.ListRequest) (*core.ListResponse, error) {
	query := p.client.Memory.Query().
		Where(memory.TenantID(req.TenantID))

	if req.SubjectID != "" {
		query.Where(memory.SubjectID(req.SubjectID))
	}

	// Filter by types
	if len(req.Types) > 0 {
		types := make([]string, len(req.Types))
		for i, t := range req.Types {
			types[i] = t.String()
		}
		query.Where(memory.TypeIn(types...))
	}

	// Filter by scopes
	if len(req.Scopes) > 0 {
		scopes := make([]string, len(req.Scopes))
		for i, s := range req.Scopes {
			scopes[i] = s.String()
		}
		query.Where(memory.ScopeIn(scopes...))
	}

	// Get total count
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, core.NewProviderError(p.Name(), "List", err)
	}

	// Apply pagination
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}

	query.Order(ent.Desc(memory.FieldCreatedAt)).
		Offset(req.Offset).
		Limit(limit + 1) // Get one extra to check hasMore

	entMems, err := query.All(ctx)
	if err != nil {
		return nil, core.NewProviderError(p.Name(), "List", err)
	}

	hasMore := len(entMems) > limit
	if hasMore {
		entMems = entMems[:limit]
	}

	memories := make([]*core.Memory, len(entMems))
	for i, em := range entMems {
		memories[i] = entToCore(em)
	}

	return &core.ListResponse{
		Memories:   memories,
		TotalCount: total,
		HasMore:    hasMore,
	}, nil
}

// Search performs semantic search on memories using pgvector.
func (p *Provider) Search(ctx context.Context, req *core.SearchRequest) (*core.SearchResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Generate embedding for query
	var queryEmbedding []float64
	if p.embedder != nil {
		emb, err := p.embedder.Embed(ctx, req.Query)
		if err != nil {
			return nil, core.NewProviderError(p.Name(), "Search", err)
		}
		queryEmbedding = emb
	}

	if len(queryEmbedding) == 0 {
		// Fall back to listing if no embedding
		listResp, err := p.List(ctx, &core.ListRequest{
			Context: req.Context,
			Types:   req.Types,
			Scopes:  req.Scopes,
			Limit:   req.Limit,
		})
		if err != nil {
			return nil, err
		}

		results := make([]*core.SearchResult, len(listResp.Memories))
		for i, m := range listResp.Memories {
			results[i] = &core.SearchResult{Memory: m, Score: 0}
		}
		return &core.SearchResponse{Results: results}, nil
	}

	// Use raw SQL for vector similarity search
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	queryVec := pgvector.NewVector(float32Slice(queryEmbedding))

	// Build query with vector similarity (cosine distance)
	query := `
		SELECT id, tenant_id, subject_id, agent_id, session_id, scope, type, content,
		       embedding, metadata, created_at, updated_at, expires_at,
		       1 - (embedding <=> $1::vector) as score
		FROM memories
		WHERE tenant_id = $2
		  AND subject_id = $3
		  AND embedding IS NOT NULL
		ORDER BY embedding <=> $1::vector
		LIMIT $4
	`

	rows, err := p.db.QueryContext(ctx, query, queryVec, req.TenantID, req.SubjectID, limit)
	if err != nil {
		return nil, core.NewProviderError(p.Name(), "Search", err)
	}
	defer func() { _ = rows.Close() }()

	var results []*core.SearchResult
	for rows.Next() {
		var (
			id, tenantID, subjectID, scope, memType, content string
			agentID, sessionID                               sql.NullString
			embeddingBytes                                   []byte
			metadataBytes                                    []byte
			createdAt, updatedAt                             time.Time
			expiresAt                                        sql.NullTime
			score                                            float64
		)

		err := rows.Scan(&id, &tenantID, &subjectID, &agentID, &sessionID, &scope, &memType,
			&content, &embeddingBytes, &metadataBytes, &createdAt, &updatedAt, &expiresAt, &score)
		if err != nil {
			continue
		}

		// Apply threshold
		if req.Threshold > 0 && score < req.Threshold {
			continue
		}

		mem := &core.Memory{
			ID:        id,
			TenantID:  tenantID,
			SubjectID: subjectID,
			Scope:     core.Scope(scope),
			Type:      core.MemoryType(memType),
			Content:   content,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}

		if agentID.Valid {
			mem.AgentID = agentID.String
		}

		if sessionID.Valid {
			mem.SessionID = sessionID.String
		}

		if expiresAt.Valid {
			mem.ExpiresAt = &expiresAt.Time
		}

		results = append(results, &core.SearchResult{
			Memory: mem,
			Score:  score,
		})
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return &core.SearchResponse{Results: results}, nil
}

// Recall retrieves relevant memories for a given query.
func (p *Provider) Recall(ctx context.Context, req *core.RecallRequest) (*core.RecallResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	searchReq := &core.SearchRequest{
		Context: req.Context,
		Query:   req.Query,
		Types:   req.IncludeTypes,
		Limit:   req.MaxResults,
	}

	searchResp, err := p.Search(ctx, searchReq)
	if err != nil {
		return nil, err
	}

	memories := make([]*core.Memory, len(searchResp.Results))
	for i, r := range searchResp.Results {
		memories[i] = r.Memory
	}

	return &core.RecallResponse{
		Memories: memories,
	}, nil
}

// entToCore converts an Ent Memory to a core Memory.
func entToCore(em *ent.Memory) *core.Memory {
	m := &core.Memory{
		ID:        em.ID,
		TenantID:  em.TenantID,
		SubjectID: em.SubjectID,
		AgentID:   em.AgentID,
		SessionID: em.SessionID,
		Scope:     core.Scope(em.Scope),
		Type:      core.MemoryType(em.Type),
		Content:   em.Content,
		Metadata:  em.Metadata,
		CreatedAt: em.CreatedAt,
		UpdatedAt: em.UpdatedAt,
		ExpiresAt: em.ExpiresAt,
	}

	if len(em.Embedding) > 0 {
		m.Embedding = float64Slice(em.Embedding)
	}

	return m
}

// float32Slice converts []float64 to []float32.
func float32Slice(f64 []float64) []float32 {
	f32 := make([]float32, len(f64))
	for i, v := range f64 {
		f32[i] = float32(v)
	}
	return f32
}

// float64Slice converts []float32 to []float64.
func float64Slice(f32 []float32) []float64 {
	f64 := make([]float64, len(f32))
	for i, v := range f32 {
		f64[i] = float64(v)
	}
	return f64
}
