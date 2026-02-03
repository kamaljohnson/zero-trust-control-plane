package telemetry

import (
	"context"
	"log"
	"time"

	telemetryv1 "zero-trust-control-plane/backend/api/generated/telemetry/v1"
)

// emitTimeout is the max time allowed for a single async emit. Used by EmitAsync and by ShutdownDrainDuration.
const emitTimeout = 5 * time.Second

// ShutdownDrainDuration is how long to wait after gRPC GracefulStop before shutting down OTel providers,
// so in-flight async telemetry emits have time to complete. Must be >= emitTimeout.
const ShutdownDrainDuration = emitTimeout

// EmitAsync runs Emit in a goroutine with a short timeout so the caller is not blocked.
// Use from request handlers for fire-and-forget, best-effort telemetry; errors are logged.
//
// emitter and event may be nil; EmitAsync returns immediately without starting a goroutine.
// The goroutine uses context.Background() with emitTimeout so request cancellation does not abort in-flight emit.
func EmitAsync(emitter EventEmitter, ctx context.Context, event *telemetryv1.TelemetryEvent) {
	if emitter == nil || event == nil {
		return
	}
	go func() {
		emitCtx, cancel := context.WithTimeout(context.Background(), emitTimeout)
		defer cancel()
		if err := emitter.Emit(emitCtx, event); err != nil {
			log.Printf("telemetry: async emit failed: %v", err)
		}
	}()
}
