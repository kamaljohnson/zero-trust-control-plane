package audit

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	telemetryv1 "zero-trust-control-plane/backend/api/generated/telemetry/v1"
	"zero-trust-control-plane/backend/internal/audit/domain"
	"zero-trust-control-plane/backend/internal/telemetry"
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

// mockEventEmitter implements telemetry.EventEmitter for tests.
type mockEventEmitter struct {
	mu     sync.Mutex
	events []*telemetryv1.TelemetryEvent
	emitErr error
}

func (m *mockEventEmitter) Emit(ctx context.Context, event *telemetryv1.TelemetryEvent) error {
	if m.emitErr != nil {
		return m.emitErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.events == nil {
		m.events = make([]*telemetryv1.TelemetryEvent, 0)
	}
	m.events = append(m.events, event)
	return nil
}

func (m *mockEventEmitter) getEvents() []*telemetryv1.TelemetryEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.events
}

// Ensure mockEventEmitter implements telemetry.EventEmitter
var _ telemetry.EventEmitter = (*mockEventEmitter)(nil)

func TestLogger_LogEvent_Success(t *testing.T) {
	repo := &mockAuditRepo{}
	ipExtractor := func(ctx context.Context) string {
		return "192.168.1.1"
	}
	logger := NewLogger(repo, ipExtractor, nil)
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
	logger := NewLogger(repo, ipExtractor, nil)
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
	logger := NewLogger(repo, nil, nil)
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
	logger := NewLogger(repo, nil, nil)
	ctx := context.Background()

	logger.LogEvent(ctx, "", "user-1", "action", "resource", "")

	if len(repo.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(repo.entries))
	}
	if repo.entries[0].OrgID != SentinelOrgID {
		t.Errorf("org_id = %q, want %q", repo.entries[0].OrgID, SentinelOrgID)
	}
}

func TestLogger_LogEvent_EmitsToTelemetry(t *testing.T) {
	repo := &mockAuditRepo{}
	emitter := &mockEventEmitter{}
	logger := NewLogger(repo, nil, emitter)
	ctx := context.Background()

	logger.LogEvent(ctx, "org-1", "user-1", "login_success", "auth", "")

	// EmitAsync is async, so wait a bit for it to complete
	time.Sleep(100 * time.Millisecond)

	events := emitter.getEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 telemetry event, got %d", len(events))
	}
	event := events[0]
	if event.OrgId != "org-1" {
		t.Errorf("event org_id = %q, want %q", event.OrgId, "org-1")
	}
	if event.UserId != "user-1" {
		t.Errorf("event user_id = %q, want %q", event.UserId, "user-1")
	}
	if event.EventType != "auth_login_success" {
		t.Errorf("event type = %q, want %q", event.EventType, "auth_login_success")
	}
	if event.Source != "audit" {
		t.Errorf("event source = %q, want %q", event.Source, "audit")
	}
}

func TestLogger_LogEvent_RepositoryError(t *testing.T) {
	repo := &mockAuditRepo{
		createErr: errors.New("database error"),
	}
	logger := NewLogger(repo, nil, nil)
	ctx := context.Background()

	// Should not panic or return error - best-effort logging
	logger.LogEvent(ctx, "org-1", "user-1", "action", "resource", "")
}

func TestLogger_LogEvent_NilRepo(t *testing.T) {
	logger := NewLogger(nil, nil, nil)
	ctx := context.Background()

	// Should not panic - no-op when repo is nil
	logger.LogEvent(ctx, "org-1", "user-1", "action", "resource", "")
}

func TestAuditActionToEventType_LoginSuccess(t *testing.T) {
	eventType := auditActionToEventType("login_success")
	if eventType != "auth_login_success" {
		t.Errorf("event type = %q, want %q", eventType, "auth_login_success")
	}
}

func TestAuditActionToEventType_LoginFailure(t *testing.T) {
	eventType := auditActionToEventType("login_failure")
	if eventType != "auth_login_failure" {
		t.Errorf("event type = %q, want %q", eventType, "auth_login_failure")
	}
}

func TestAuditActionToEventType_Logout(t *testing.T) {
	eventType := auditActionToEventType("logout")
	if eventType != "auth_logout" {
		t.Errorf("event type = %q, want %q", eventType, "auth_logout")
	}
}

func TestAuditActionToEventType_SessionCreated(t *testing.T) {
	eventType := auditActionToEventType("session_created")
	if eventType != "session_created" {
		t.Errorf("event type = %q, want %q", eventType, "session_created")
	}
}

func TestAuditActionToEventType_UnknownAction(t *testing.T) {
	eventType := auditActionToEventType("unknown_action")
	if eventType != "" {
		t.Errorf("event type = %q, want empty string", eventType)
	}
}

func TestLogger_LogEvent_TelemetryError(t *testing.T) {
	repo := &mockAuditRepo{}
	emitter := &mockEventEmitter{
		emitErr: errors.New("telemetry error"),
	}
	logger := NewLogger(repo, nil, emitter)
	ctx := context.Background()

	// Should not panic - telemetry errors are logged but don't fail audit logging
	logger.LogEvent(ctx, "org-1", "user-1", "login_success", "auth", "")

	// Audit log should still be created
	if len(repo.entries) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(repo.entries))
	}
}

func TestLogger_LogEvent_NonMappedAction(t *testing.T) {
	repo := &mockAuditRepo{}
	emitter := &mockEventEmitter{}
	logger := NewLogger(repo, nil, emitter)
	ctx := context.Background()

	logger.LogEvent(ctx, "org-1", "user-1", "custom_action", "resource", "")

	// Audit log should be created
	if len(repo.entries) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(repo.entries))
	}
	// But telemetry event should not be emitted for non-mapped actions
	if len(emitter.events) != 0 {
		t.Errorf("expected 0 telemetry events, got %d", len(emitter.events))
	}
}
