package handler

import (
	"context"
	"errors"
	"testing"

	healthv1 "zero-trust-control-plane/backend/api/generated/health/v1"
)

// mockPinger implements Pinger for tests.
type mockPinger struct {
	pingErr error
}

func (m *mockPinger) PingContext(context.Context) error {
	return m.pingErr
}

// mockPolicyChecker implements PolicyChecker for tests.
type mockPolicyChecker struct {
	healthErr error
}

func (m *mockPolicyChecker) HealthCheck(context.Context) error {
	return m.healthErr
}

func TestHealthCheck_NilPinger(t *testing.T) {
	srv := NewServer(nil, nil)
	resp, err := srv.HealthCheck(context.Background(), &healthv1.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
	if resp.GetStatus() != healthv1.ServingStatus_SERVING_STATUS_SERVING {
		t.Errorf("status = %v, want SERVING", resp.GetStatus())
	}
}

func TestHealthCheck_PingerSuccess(t *testing.T) {
	srv := NewServer(&mockPinger{}, nil)
	resp, err := srv.HealthCheck(context.Background(), &healthv1.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
	if resp.GetStatus() != healthv1.ServingStatus_SERVING_STATUS_SERVING {
		t.Errorf("status = %v, want SERVING", resp.GetStatus())
	}
}

func TestHealthCheck_PingerFailure(t *testing.T) {
	srv := NewServer(&mockPinger{pingErr: errors.New("connection refused")}, nil)
	resp, err := srv.HealthCheck(context.Background(), &healthv1.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("HealthCheck must not return gRPC error on ping failure: %v", err)
	}
	if resp.GetStatus() != healthv1.ServingStatus_SERVING_STATUS_NOT_SERVING {
		t.Errorf("status = %v, want NOT_SERVING", resp.GetStatus())
	}
}

func TestHealthCheck_PolicyCheckerSuccess(t *testing.T) {
	srv := NewServer(nil, &mockPolicyChecker{})
	resp, err := srv.HealthCheck(context.Background(), &healthv1.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
	if resp.GetStatus() != healthv1.ServingStatus_SERVING_STATUS_SERVING {
		t.Errorf("status = %v, want SERVING", resp.GetStatus())
	}
}

func TestHealthCheck_PolicyCheckerFailure(t *testing.T) {
	srv := NewServer(nil, &mockPolicyChecker{healthErr: errors.New("rego compile failed")})
	resp, err := srv.HealthCheck(context.Background(), &healthv1.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("HealthCheck must not return gRPC error on policy check failure: %v", err)
	}
	if resp.GetStatus() != healthv1.ServingStatus_SERVING_STATUS_NOT_SERVING {
		t.Errorf("status = %v, want NOT_SERVING", resp.GetStatus())
	}
}

func TestHealthCheck_BothChecksPolicyFails(t *testing.T) {
	srv := NewServer(&mockPinger{}, &mockPolicyChecker{healthErr: errors.New("policy error")})
	resp, err := srv.HealthCheck(context.Background(), &healthv1.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
	if resp.GetStatus() != healthv1.ServingStatus_SERVING_STATUS_NOT_SERVING {
		t.Errorf("status = %v, want NOT_SERVING", resp.GetStatus())
	}
}
