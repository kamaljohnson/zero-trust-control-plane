package handler

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	adminv1 "zero-trust-control-plane/backend/api/generated/admin/v1"
)

func TestNewServer(t *testing.T) {
	srv := NewServer()
	if srv == nil {
		t.Fatal("NewServer returned nil")
	}
}

func TestGetSystemStats_Unimplemented(t *testing.T) {
	srv := NewServer()
	ctx := context.Background()

	_, err := srv.GetSystemStats(ctx, &adminv1.GetSystemStatsRequest{})
	if err == nil {
		t.Fatal("expected Unimplemented error")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unimplemented)
	}
	if st.Message() != "method GetSystemStats not implemented" {
		t.Errorf("status message = %q, want %q", st.Message(), "method GetSystemStats not implemented")
	}
}
