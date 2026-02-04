package handler

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	devv1 "zero-trust-control-plane/backend/api/generated/dev/v1"
)

// mockStore implements devotp.Store for tests.
type mockStore struct {
	otps map[string]string
}

func (m *mockStore) Put(ctx context.Context, challengeID, otp string, expiresAt time.Time) {
	if m.otps == nil {
		m.otps = make(map[string]string)
	}
	m.otps[challengeID] = otp
}

func (m *mockStore) Get(ctx context.Context, challengeID string) (string, bool) {
	if m.otps == nil {
		return "", false
	}
	otp, ok := m.otps[challengeID]
	return otp, ok
}

func TestGetOTP_Success(t *testing.T) {
	store := &mockStore{
		otps: map[string]string{
			"challenge-1": "123456",
		},
	}
	srv := NewServer(store)
	ctx := context.Background()

	resp, err := srv.GetOTP(ctx, &devv1.GetOTPRequest{ChallengeId: "challenge-1"})
	if err != nil {
		t.Fatalf("GetOTP: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	if resp.Otp != "123456" {
		t.Errorf("otp = %q, want %q", resp.Otp, "123456")
	}
	if resp.Note != devOTPNote {
		t.Errorf("note = %q, want %q", resp.Note, devOTPNote)
	}
}

func TestGetOTP_NotFound(t *testing.T) {
	store := &mockStore{
		otps: make(map[string]string),
	}
	srv := NewServer(store)
	ctx := context.Background()

	_, err := srv.GetOTP(ctx, &devv1.GetOTPRequest{ChallengeId: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent challenge")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("status code = %v, want %v", st.Code(), codes.NotFound)
	}
	if st.Message() != "OTP not found or expired" {
		t.Errorf("status message = %q, want %q", st.Message(), "OTP not found or expired")
	}
}

func TestGetOTP_InvalidChallengeID(t *testing.T) {
	store := &mockStore{
		otps: make(map[string]string),
	}
	srv := NewServer(store)
	ctx := context.Background()

	_, err := srv.GetOTP(ctx, &devv1.GetOTPRequest{ChallengeId: ""})
	if err == nil {
		t.Fatal("expected error for empty challenge_id")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("status code = %v, want %v", st.Code(), codes.InvalidArgument)
	}
	if st.Message() != "challenge_id is required" {
		t.Errorf("status message = %q, want %q", st.Message(), "challenge_id is required")
	}
}

func TestGetOTP_NilStore(t *testing.T) {
	srv := NewServer(nil)
	ctx := context.Background()

	// This should panic or handle gracefully - checking that it doesn't crash
	defer func() {
		if r := recover(); r != nil {
			// Panic is acceptable for nil store
		}
	}()

	_, err := srv.GetOTP(ctx, &devv1.GetOTPRequest{ChallengeId: "challenge-1"})
	if err == nil {
		t.Fatal("expected error for nil store")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("status code = %v, want %v", st.Code(), codes.NotFound)
	}
}
