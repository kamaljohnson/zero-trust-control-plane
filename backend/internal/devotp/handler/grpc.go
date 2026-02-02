// Package handler implements the dev-only gRPC DevService (e.g. GetOTP).
package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	devv1 "zero-trust-control-plane/backend/api/generated/dev/v1"
	"zero-trust-control-plane/backend/internal/devotp"
)

const devOTPNote = "DEV MODE ONLY"

// Server implements DevService. Only registered when dev OTP is enabled and not production.
type Server struct {
	devv1.UnimplementedDevServiceServer
	store devotp.Store
}

// NewServer returns a DevService server that reads OTP from the given store.
func NewServer(store devotp.Store) *Server {
	return &Server{store: store}
}

// GetOTP returns the plain OTP for the given challenge_id from the dev store. Returns NotFound if missing or expired.
func (s *Server) GetOTP(ctx context.Context, req *devv1.GetOTPRequest) (*devv1.GetOTPResponse, error) {
	challengeID := req.GetChallengeId()
	if challengeID == "" {
		return nil, status.Error(codes.InvalidArgument, "challenge_id is required")
	}
	otp, ok := s.store.Get(ctx, challengeID)
	if !ok {
		return nil, status.Error(codes.NotFound, "OTP not found or expired")
	}
	return &devv1.GetOTPResponse{
		Otp:  otp,
		Note: devOTPNote,
	}, nil
}
