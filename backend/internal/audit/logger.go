package audit

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	telemetryv1 "zero-trust-control-plane/backend/api/generated/telemetry/v1"
	"zero-trust-control-plane/backend/internal/audit/domain"
	auditrepo "zero-trust-control-plane/backend/internal/audit/repository"
	"zero-trust-control-plane/backend/internal/telemetry"
)

// SentinelOrgID is the org_id used for audit events that have no org (e.g. login_failure, logout with invalid token).
const SentinelOrgID = "_system"

// IPExtractor returns the client IP from the request context (e.g. gRPC metadata or peer).
type IPExtractor func(context.Context) string

// AuditLogger writes a single audit event with explicit action/resource. Used by auth and session code paths.
// LogEvent is best-effort: failures are logged and do not affect the caller.
type AuditLogger interface {
	LogEvent(ctx context.Context, orgID, userID, action, resource, metadata string)
}

// Logger implements AuditLogger using the audit repository, optional IP extractor, and optional telemetry emitter.
type Logger struct {
	repo        auditrepo.Repository
	ipExtractor IPExtractor
	emitter     telemetry.EventEmitter
}

// NewLogger returns an AuditLogger that persists to repo and uses ipExtractor for client IP.
// ipExtractor may be nil; then IP is recorded as "unknown".
// emitter may be nil; then audit events are not sent to OTLP (e.g. Loki).
func NewLogger(repo auditrepo.Repository, ipExtractor IPExtractor, emitter telemetry.EventEmitter) *Logger {
	return &Logger{repo: repo, ipExtractor: ipExtractor, emitter: emitter}
}

// LogEvent writes one audit log entry. Best-effort: errors are logged and not returned.
func (l *Logger) LogEvent(ctx context.Context, orgID, userID, action, resource, metadata string) {
	if l.repo == nil {
		return
	}
	ip := "unknown"
	if l.ipExtractor != nil {
		ip = l.ipExtractor(ctx)
	}
	if orgID == "" {
		orgID = SentinelOrgID
	}
	entry := &domain.AuditLog{
		ID:        uuid.New().String(),
		OrgID:     orgID,
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		IP:        ip,
		Metadata:  metadata,
		CreatedAt: time.Now().UTC(),
	}
	if err := l.repo.Create(ctx, entry); err != nil {
		log.Printf("audit: failed to log event %s/%s: %v", action, resource, err)
	}
	if l.emitter != nil {
		eventType := auditActionToEventType(action)
		if eventType != "" {
			ev := &telemetryv1.TelemetryEvent{
				OrgId:     orgID,
				UserId:    userID,
				EventType: eventType,
				Source:    "audit",
				Metadata:  []byte(metadata),
				CreatedAt: timestamppb.New(entry.CreatedAt),
			}
			telemetry.EmitAsync(l.emitter, ctx, ev)
		}
	}
}

// auditActionToEventType maps audit action strings to the required telemetry event_type (for Loki/OTLP).
// Returns empty string for actions that are not emitted to telemetry.
func auditActionToEventType(action string) string {
	switch action {
	case "login_success":
		return "auth_login_success"
	case "login_failure":
		return "auth_login_failure"
	case "logout":
		return "auth_logout"
	case "session_created":
		return "session_created"
	default:
		return ""
	}
}
