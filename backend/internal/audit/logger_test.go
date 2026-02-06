package audit

import (
	"context"
	"errors"
	"testing"

	"zero-trust-control-plane/backend/internal/audit/domain"
)

// mockAuditRepo implements audit repository interface for tests.
type mockAuditRepo struct {
	entries []*domain.AuditLog
	createErr error
}

func (m *mockAuditRepo) GetByID(ctx context.Context, id string) (*domain.AuditLog, error) {
	return nil, nil
}

func (m *mockAuditRepo) Create(ctx context.Context, entry *domain.AuditLog) error {
	if m.createErr != nil {
		return m.createErr
	}
	if m.entries == nil {
		m.entries = make([]*domain.AuditLog, 0)
	}
	m.entries = append(m.entries, entry)
	return nil
}

func (m *mockAuditRepo) ListByOrg(ctx context.Context, orgID string, limit, offset int32) ([]*domain.AuditLog, error) {
	return nil, nil
}

func (m *mockAuditRepo) ListByOrgFiltered(ctx context.Context, orgID string, limit, offset int32, userID, action, resource *string) ([]*domain.AuditLog, error) {
	return nil, nil
}


func TestLogger_LogEvent_Success(t *testing.T) {
	repo := &mockAuditRepo{}
	ipExtractor := func(ctx context.Context) string {
		return "192.168.1.1"
	}
	logger := NewLogger(repo, ipExtractor)
	ctx := context.Background()

	logger.LogEvent(ctx, "org-1", "user-1", "test_action", "test_resource", "metadata")

	if len(repo.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(repo.entries))
	}
	entry := repo.entries[0]
	if entry.OrgID != "org-1" {
		t.Errorf("org_id = %q, want %q", entry.OrgID, "org-1")
	}
	if entry.UserID != "user-1" {
		t.Errorf("user_id = %q, want %q", entry.UserID, "user-1")
	}
	if entry.Action != "test_action" {
		t.Errorf("action = %q, want %q", entry.Action, "test_action")
	}
	if entry.Resource != "test_resource" {
		t.Errorf("resource = %q, want %q", entry.Resource, "test_resource")
	}
	if entry.IP != "192.168.1.1" {
		t.Errorf("ip = %q, want %q", entry.IP, "192.168.1.1")
	}
	if entry.Metadata != "metadata" {
		t.Errorf("metadata = %q, want %q", entry.Metadata, "metadata")
	}
	if entry.ID == "" {
		t.Error("entry ID should be set")
	}
	if entry.CreatedAt.IsZero() {
		t.Error("entry CreatedAt should be set")
	}
}

func TestLogger_LogEvent_UsesIPExtractor(t *testing.T) {
	repo := &mockAuditRepo{}
	ipExtractor := func(ctx context.Context) string {
		return "10.0.0.1"
	}
	logger := NewLogger(repo, ipExtractor)
	ctx := context.Background()

	logger.LogEvent(ctx, "org-1", "user-1", "action", "resource", "")

	if len(repo.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(repo.entries))
	}
	if repo.entries[0].IP != "10.0.0.1" {
		t.Errorf("ip = %q, want %q", repo.entries[0].IP, "10.0.0.1")
	}
}

func TestLogger_LogEvent_NilIPExtractor(t *testing.T) {
	repo := &mockAuditRepo{}
	logger := NewLogger(repo, nil)
	ctx := context.Background()

	logger.LogEvent(ctx, "org-1", "user-1", "action", "resource", "")

	if len(repo.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(repo.entries))
	}
	if repo.entries[0].IP != "unknown" {
		t.Errorf("ip = %q, want %q", repo.entries[0].IP, "unknown")
	}
}

func TestLogger_LogEvent_SentinelOrgID(t *testing.T) {
	repo := &mockAuditRepo{}
	logger := NewLogger(repo, nil)
	ctx := context.Background()

	logger.LogEvent(ctx, "", "user-1", "action", "resource", "")

	if len(repo.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(repo.entries))
	}
	if repo.entries[0].OrgID != SentinelOrgID {
		t.Errorf("org_id = %q, want %q", repo.entries[0].OrgID, SentinelOrgID)
	}
}


func TestLogger_LogEvent_RepositoryError(t *testing.T) {
	repo := &mockAuditRepo{
		createErr: errors.New("database error"),
	}
	logger := NewLogger(repo, nil)
	ctx := context.Background()

	// Should not panic or return error - best-effort logging
	logger.LogEvent(ctx, "org-1", "user-1", "action", "resource", "")
}

func TestLogger_LogEvent_NilRepo(t *testing.T) {
	logger := NewLogger(nil, nil)
	ctx := context.Background()

	// Should not panic - no-op when repo is nil
	logger.LogEvent(ctx, "org-1", "user-1", "action", "resource", "")
}

