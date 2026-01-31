package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	telemetryv1 "zero-trust-control-plane/backend/api/generated/telemetry/v1"
)

// Server implements TelemetryService (proto server) for telemetry events.
// Proto: telemetry/telemetry.proto → internal/telemetry/handler.
type Server struct {
	telemetryv1.UnimplementedTelemetryServiceServer
}

// NewServer returns a new Telemetry gRPC server.
func NewServer() *Server {
	return &Server{}
}

// EmitTelemetryEvent emits a single telemetry event. TODO: implement.
func (s *Server) EmitTelemetryEvent(ctx context.Context, req *telemetryv1.EmitTelemetryEventRequest) (*telemetryv1.EmitTelemetryEventResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method EmitTelemetryEvent not implemented")
}

// BatchEmitTelemetry emits multiple telemetry events. TODO: implement.
func (s *Server) BatchEmitTelemetry(ctx context.Context, req *telemetryv1.BatchEmitTelemetryRequest) (*telemetryv1.BatchEmitTelemetryResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method BatchEmitTelemetry not implemented")
}
