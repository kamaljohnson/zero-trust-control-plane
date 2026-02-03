// Package producer defines the interface for emitting telemetry events (e.g. to Kafka).
package producer

import (
	"context"

	telemetryv1 "zero-trust-control-plane/backend/api/generated/telemetry/v1"
)

// Producer emits telemetry events. Callers use it best-effort: log and ignore errors.
type Producer interface {
	// Emit sends a single telemetry event. Implementations may block briefly; call from a goroutine if needed.
	// Returns an error only on write failure; callers typically log and ignore.
	Emit(ctx context.Context, event *telemetryv1.TelemetryEvent) error
	// Close releases resources (e.g. Kafka writer). Safe to call if already closed.
	Close() error
}
