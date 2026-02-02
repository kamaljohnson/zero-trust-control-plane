package handler

import (
	"context"
	"log"

	healthv1 "zero-trust-control-plane/backend/api/generated/health/v1"
)

// Pinger is used by the health handler to check dependency connectivity (e.g. database).
// *sql.DB satisfies this interface via PingContext. If nil, HealthCheck returns SERVING without probing.
type Pinger interface {
	PingContext(context.Context) error
}

// PolicyChecker is used by the health handler to verify the in-process policy engine (e.g. OPA Rego).
// If nil, HealthCheck does not run a policy check.
type PolicyChecker interface {
	HealthCheck(context.Context) error
}

// Server implements HealthService (proto server) for readiness.
// When pinger or policyChecker is set, HealthCheck returns SERVING only if all configured checks succeed; otherwise NOT_SERVING (no gRPC error).
// Proto: health/health.proto → internal/health/handler.
type Server struct {
	healthv1.UnimplementedHealthServiceServer
	pinger        Pinger
	policyChecker PolicyChecker
}

// NewServer returns a new Health gRPC server. Pass nil pinger or policyChecker when not configured (that check is skipped).
func NewServer(pinger Pinger, policyChecker PolicyChecker) *Server {
	return &Server{pinger: pinger, policyChecker: policyChecker}
}

// HealthCheck returns readiness status for Kubernetes, load balancers, and CI.
// Runs pinger (if set), then policyChecker (if set). Returns SERVING only when all configured checks pass;
// on any failure logs and returns NOT_SERVING without a gRPC error so probes receive a successful RPC with status NOT_SERVING.
func (s *Server) HealthCheck(ctx context.Context, req *healthv1.HealthCheckRequest) (*healthv1.HealthCheckResponse, error) {
	if s.pinger != nil {
		if err := s.pinger.PingContext(ctx); err != nil {
			log.Printf("health: database ping failed: %v", err)
			return &healthv1.HealthCheckResponse{Status: healthv1.ServingStatus_SERVING_STATUS_NOT_SERVING}, nil
		}
	}
	if s.policyChecker != nil {
		if err := s.policyChecker.HealthCheck(ctx); err != nil {
			log.Printf("health: policy check failed: %v", err)
			return &healthv1.HealthCheckResponse{Status: healthv1.ServingStatus_SERVING_STATUS_NOT_SERVING}, nil
		}
	}
	return &healthv1.HealthCheckResponse{Status: healthv1.ServingStatus_SERVING_STATUS_SERVING}, nil
}
