package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	auditv1 "zero-trust-control-plane/backend/api/generated/audit/v1"
)

// Server implements AuditService (proto server) for audit logs.
// Proto: audit/audit.proto → internal/audit/handler.
type Server struct {
	auditv1.UnimplementedAuditServiceServer
}

// NewServer returns a new Audit gRPC server.
func NewServer() *Server {
	return &Server{}
}

// ListAuditLogs returns a paginated list of audit logs. TODO: implement.
func (s *Server) ListAuditLogs(ctx context.Context, req *auditv1.ListAuditLogsRequest) (*auditv1.ListAuditLogsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListAuditLogs not implemented")
}
