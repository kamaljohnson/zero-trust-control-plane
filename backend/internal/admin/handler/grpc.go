package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	adminv1 "zero-trust-control-plane/backend/api/generated/admin/v1"
)

// Server implements AdminService (proto server) for system-level admin operations.
// Proto: admin/admin.proto → internal/admin/handler.
type Server struct {
	adminv1.UnimplementedAdminServiceServer
}

// NewServer returns a new Admin gRPC server.
func NewServer() *Server {
	return &Server{}
}

// GetSystemStats returns system-wide stats for platform admins.
// TODO: implement.
func (s *Server) GetSystemStats(ctx context.Context, req *adminv1.GetSystemStatsRequest) (*adminv1.GetSystemStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetSystemStats not implemented")
}
