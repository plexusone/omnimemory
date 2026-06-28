package core

import (
	"testing"
)

func TestScope_Valid(t *testing.T) {
	tests := []struct {
		scope Scope
		valid bool
	}{
		{ScopeUser, true},
		{ScopeAgent, true},
		{ScopeTenant, true},
		{ScopeTeam, true},
		{ScopeSession, true},
		{ScopeDomain, true},
		{Scope("invalid"), false},
		{Scope(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.scope), func(t *testing.T) {
			if got := tt.scope.Valid(); got != tt.valid {
				t.Errorf("Scope(%q).Valid() = %v, want %v", tt.scope, got, tt.valid)
			}
		})
	}
}

func TestMemoryType_Valid(t *testing.T) {
	tests := []struct {
		memType MemoryType
		valid   bool
	}{
		{MemoryTypeObservation, true},
		{MemoryTypeFact, true},
		{MemoryTypePreference, true},
		{MemoryTypeSummary, true},
		{MemoryTypeTrait, true},
		{MemoryTypeRelationship, true},
		{MemoryType("invalid"), false},
		{MemoryType(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.memType), func(t *testing.T) {
			if got := tt.memType.Valid(); got != tt.valid {
				t.Errorf("MemoryType(%q).Valid() = %v, want %v", tt.memType, got, tt.valid)
			}
		})
	}
}

func TestContext_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ctx     Context
		wantErr error
	}{
		{
			name: "valid context",
			ctx: Context{
				TenantID:  "tenant-1",
				SubjectID: "user-123",
			},
			wantErr: nil,
		},
		{
			name: "missing tenant_id",
			ctx: Context{
				SubjectID: "user-123",
			},
			wantErr: ErrTenantRequired,
		},
		{
			name: "missing subject_id",
			ctx: Context{
				TenantID: "tenant-1",
			},
			wantErr: ErrSubjectRequired,
		},
		{
			name:    "missing both",
			ctx:     Context{},
			wantErr: ErrTenantRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ctx.Validate()
			if err != tt.wantErr {
				t.Errorf("Context.Validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     AddRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: AddRequest{
				Context: Context{
					TenantID:  "tenant-1",
					SubjectID: "user-123",
				},
				Type:    MemoryTypeFact,
				Content: "Test content",
			},
			wantErr: false,
		},
		{
			name: "missing tenant_id",
			req: AddRequest{
				Context: Context{
					SubjectID: "user-123",
				},
				Type:    MemoryTypeFact,
				Content: "Test content",
			},
			wantErr: true,
		},
		{
			name: "missing content",
			req: AddRequest{
				Context: Context{
					TenantID:  "tenant-1",
					SubjectID: "user-123",
				},
				Type: MemoryTypeFact,
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			req: AddRequest{
				Context: Context{
					TenantID:  "tenant-1",
					SubjectID: "user-123",
				},
				Type:    MemoryType("invalid"),
				Content: "Test content",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("AddRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSearchRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     SearchRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: SearchRequest{
				Context: Context{
					TenantID:  "tenant-1",
					SubjectID: "user-123",
				},
				Query: "test query",
			},
			wantErr: false,
		},
		{
			name: "missing query",
			req: SearchRequest{
				Context: Context{
					TenantID:  "tenant-1",
					SubjectID: "user-123",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
