package interceptors

import (
	"context"
	"log"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"zero-trust-control-plane/backend/internal/audit"
	"zero-trust-control-plane/backend/internal/audit/domain"
	auditrepo "zero-trust-control-plane/backend/internal/audit/repository"

	"github.com/google/uuid"
)

// AuditUnary returns a unary server interceptor that records an audit log entry after each RPC.
// skipMethods is the set of full method names to not audit (e.g. HealthCheck, optionally ListAuditLogs).
// Create is best-effort: failures are logged and do not fail the RPC. Only writes when org_id is set (authenticated context).
func AuditUnary(auditRepo auditrepo.Repository, skipMethods map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if skipMethods[info.FullMethod] {
			return resp, err
		}
		orgID, _ := GetOrgID(ctx)
		if orgID == "" {
			return resp, err
		}
		userID, _ := GetUserID(ctx)
		ar := audit.ParseFullMethod(info.FullMethod)
		ip := ClientIP(ctx)
		entry := &domain.AuditLog{
			ID:        uuid.New().String(),
			OrgID:     orgID,
			UserID:    userID,
			Action:    ar.Action,
			Resource:  ar.Resource,
			IP:        ip,
			Metadata:  "",
			CreatedAt: time.Now().UTC(),
		}
		if createErr := auditRepo.Create(ctx, entry); createErr != nil {
			log.Printf("audit: failed to create audit log: %v", createErr)
		}
		return resp, err
	}
}

// ClientIP returns the client IP from gRPC metadata (x-forwarded-for, x-real-ip) or peer, or "unknown".
func ClientIP(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-forwarded-for"); len(vals) > 0 {
			if s := strings.TrimSpace(vals[0]); s != "" {
				if i := strings.Index(s, ","); i > 0 {
					s = strings.TrimSpace(s[:i])
				}
				return s
			}
		}
		if vals := md.Get("x-real-ip"); len(vals) > 0 {
			if s := strings.TrimSpace(vals[0]); s != "" {
				return s
			}
		}
	}
	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		if host, _, err := net.SplitHostPort(p.Addr.String()); err == nil {
			return host
		}
		return p.Addr.String()
	}
	return "unknown"
}
