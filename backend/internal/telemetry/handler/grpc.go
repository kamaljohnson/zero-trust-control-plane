package handler

import (
	"context"
	"log"

	"google.golang.org/protobuf/types/known/timestamppb"

	telemetryv1 "zero-trust-control-plane/backend/api/generated/telemetry/v1"
	"zero-trust-control-plane/backend/internal/telemetry/producer"
)

// Server implements TelemetryService (proto server) for telemetry events.
// Proto: telemetry/telemetry.proto → internal/telemetry/handler.
type Server struct {
	telemetryv1.UnimplementedTelemetryServiceServer
	producer producer.Producer
}

// NewServer returns a new Telemetry gRPC server. producer may be nil; then Emit RPCs no-op but still return success.
func NewServer(producer producer.Producer) *Server {
	return &Server{producer: producer}
}

// EmitTelemetryEvent records a single telemetry event. Best-effort: if producer is set, emits to Kafka; always returns empty success.
func (s *Server) EmitTelemetryEvent(ctx context.Context, req *telemetryv1.EmitTelemetryEventRequest) (*telemetryv1.EmitTelemetryEventResponse, error) {
	if req == nil {
		return &telemetryv1.EmitTelemetryEventResponse{}, nil
	}
	event := requestToEvent(req)
	if s.producer != nil {
		if err := s.producer.Emit(ctx, event); err != nil {
			log.Printf("telemetry: EmitTelemetryEvent failed: %v", err)
		}
	}
	return &telemetryv1.EmitTelemetryEventResponse{}, nil
}

// BatchEmitTelemetry records multiple telemetry events. Best-effort; always returns empty success.
func (s *Server) BatchEmitTelemetry(ctx context.Context, req *telemetryv1.BatchEmitTelemetryRequest) (*telemetryv1.BatchEmitTelemetryResponse, error) {
	if req == nil || s.producer == nil {
		return &telemetryv1.BatchEmitTelemetryResponse{}, nil
	}
	for _, e := range req.Events {
		if e == nil {
			continue
		}
		event := requestToEvent(e)
		if err := s.producer.Emit(ctx, event); err != nil {
			log.Printf("telemetry: BatchEmitTelemetry emit failed: %v", err)
		}
	}
	return &telemetryv1.BatchEmitTelemetryResponse{}, nil
}

func requestToEvent(req *telemetryv1.EmitTelemetryEventRequest) *telemetryv1.TelemetryEvent {
	return &telemetryv1.TelemetryEvent{
		OrgId:     req.GetOrgId(),
		UserId:    req.GetUserId(),
		DeviceId:  req.GetDeviceId(),
		SessionId: req.GetSessionId(),
		EventType: req.GetEventType(),
		Source:    req.GetSource(),
		Metadata:  req.GetMetadata(),
		CreatedAt: timestamppb.Now(),
	}
}
