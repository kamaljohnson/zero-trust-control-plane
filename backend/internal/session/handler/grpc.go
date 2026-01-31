package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sessionv1 "zero-trust-control-plane/backend/api/generated/session/v1"
)

// Server implements SessionService (proto server) for session lifecycle.
// Proto: session/session.proto → internal/session/handler.
type Server struct {
	sessionv1.UnimplementedSessionServiceServer
}

// NewServer returns a new Session gRPC server.
func NewServer() *Server {
	return &Server{}
}

// RevokeSession revokes a session. TODO: implement.
func (s *Server) RevokeSession(ctx context.Context, req *sessionv1.RevokeSessionRequest) (*sessionv1.RevokeSessionResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method RevokeSession not implemented")
}

// ListSessions returns a paginated list of sessions. TODO: implement.
func (s *Server) ListSessions(ctx context.Context, req *sessionv1.ListSessionsRequest) (*sessionv1.ListSessionsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListSessions not implemented")
}

// GetSession returns a session by ID. TODO: implement.
func (s *Server) GetSession(ctx context.Context, req *sessionv1.GetSessionRequest) (*sessionv1.GetSessionResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetSession not implemented")
}
