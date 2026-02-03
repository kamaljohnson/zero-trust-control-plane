package handler

import (
	"context"
	"log"

	"google.golang.org/protobuf/types/known/timestamppb"

	telemetryv1 "zero-trust-control-plane/backend/api/generated/telemetry/v1"
	"zero-trust-control-plane/backend/internal/telemetry"
)

// maxBatchSize is the maximum number of events processed per BatchEmitTelemetry request; excess are dropped.
const maxBatchSize = 500

// Server implements TelemetryService (proto server) for telemetry events.
// Proto: telemetry/telemetry.proto → internal/telemetry/handler.
type Server struct {
	telemetryv1.UnimplementedTelemetryServiceServer
	emitter EventEmitter
}

// NewServer returns a new Telemetry gRPC server. emitter may be nil; then Emit RPCs no-op but still return success.
func NewServer(emitter EventEmitter) *Server {
	return &Server{emitter: emitter}
}

// EmitTelemetryEvent records a single telemetry event. Best-effort: if emitter is set, emits via OTel Logs; always returns empty success.
func (s *Server) EmitTelemetryEvent(ctx context.Context, req *telemetryv1.EmitTelemetryEventRequest) (*telemetryv1.EmitTelemetryEventResponse, error) {
	if req == nil {
		return &telemetryv1.EmitTelemetryEventResponse{}, nil
	}
	event := requestToEvent(req)
	telemetry.EmitAsync(s.emitter, ctx, event)
	return &telemetryv1.EmitTelemetryEventResponse{}, nil
}

// BatchEmitTelemetry records multiple telemetry events. Best-effort; always returns empty success.
// At most maxBatchSize events are processed per request; excess are dropped and a log line is written.
func (s *Server) BatchEmitTelemetry(ctx context.Context, req *telemetryv1.BatchEmitTelemetryRequest) (*telemetryv1.BatchEmitTelemetryResponse, error) {
	// Guard nil request first, consistent with EmitTelemetryEvent; emitter may be nil (EmitAsync no-ops).
	if req == nil {
		return &telemetryv1.BatchEmitTelemetryResponse{}, nil
	}
	events := req.Events
	if len(events) > maxBatchSize {
		log.Printf("telemetry: BatchEmitTelemetry truncated to %d events (received %d)", maxBatchSize, len(req.Events))
		events = events[:maxBatchSize]
	}
	for _, e := range events {
		if e == nil {
			continue
		}
		telemetry.EmitAsync(s.emitter, ctx, requestToEvent(e))
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
