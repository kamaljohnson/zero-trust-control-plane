package audit

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"

	"zero-trust-control-plane/backend/internal/audit/domain"
	auditrepo "zero-trust-control-plane/backend/internal/audit/repository"
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

// Logger implements AuditLogger using the audit repository and an optional IP extractor.
type Logger struct {
	repo        auditrepo.Repository
	ipExtractor IPExtractor
}

// NewLogger returns an AuditLogger that persists to repo and uses ipExtractor for client IP.
// ipExtractor may be nil; then IP is recorded as "unknown".
func NewLogger(repo auditrepo.Repository, ipExtractor IPExtractor) *Logger {
	return &Logger{repo: repo, ipExtractor: ipExtractor}
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
}
