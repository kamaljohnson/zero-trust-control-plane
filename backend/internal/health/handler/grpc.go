package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	healthv1 "zero-trust-control-plane/backend/api/generated/health/v1"
)

// Server implements HealthService (proto server) for readiness/liveness.
// Proto: health/health.proto → internal/health/handler.
type Server struct {
	healthv1.UnimplementedHealthServiceServer
}

// NewServer returns a new Health gRPC server.
func NewServer() *Server {
	return &Server{}
}

// HealthCheck returns service health status for Kubernetes, load balancers, and CI. TODO: implement.
func (s *Server) HealthCheck(ctx context.Context, req *healthv1.HealthCheckRequest) (*healthv1.HealthCheckResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method HealthCheck not implemented")
}
