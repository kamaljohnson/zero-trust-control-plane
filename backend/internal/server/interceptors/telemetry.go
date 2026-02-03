package interceptors

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	telemetryv1 "zero-trust-control-plane/backend/api/generated/telemetry/v1"
	"zero-trust-control-plane/backend/internal/telemetry/producer"
)

// grpcRequestMetadata is the JSON shape stored in TelemetryEvent.Metadata for grpc_request events.
type grpcRequestMetadata struct {
	FullMethod string `json:"full_method"`
	StatusCode string `json:"status_code"`
	DurationMs int64  `json:"duration_ms"`
	ClientIP   string `json:"client_ip"`
}

// TelemetryUnary returns a unary server interceptor that emits a telemetry event after each RPC.
// Best-effort: failures are logged and do not fail the RPC. If producer is nil, the interceptor no-ops.
// skipMethods is the set of full method names to not emit (e.g. HealthCheck).
func TelemetryUnary(p producer.Producer, skipMethods map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		if p == nil || skipMethods[info.FullMethod] {
			return resp, err
		}
		code := status.Code(err)
		meta := grpcRequestMetadata{
			FullMethod: info.FullMethod,
			StatusCode: code.String(),
			DurationMs: time.Since(start).Milliseconds(),
			ClientIP:   ClientIP(ctx),
		}
		metaJSON, _ := json.Marshal(meta)
		orgID, _ := GetOrgID(ctx)
		userID, _ := GetUserID(ctx)
		sessionID, _ := GetSessionID(ctx)
		event := &telemetryv1.TelemetryEvent{
			OrgId:     orgID,
			UserId:    userID,
			SessionId: sessionID,
			EventType: "grpc_request",
			Source:    "grpc_interceptor",
			Metadata:  metaJSON,
			CreatedAt: timestamppb.Now(),
		}
		go func() {
			emitCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if emitErr := p.Emit(emitCtx, event); emitErr != nil {
				log.Printf("telemetry: interceptor emit failed: %v", emitErr)
			}
		}()
		return resp, err
	}
}
