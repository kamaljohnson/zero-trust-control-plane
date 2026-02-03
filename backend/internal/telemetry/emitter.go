package telemetry

import (
	"context"

	telemetryv1 "zero-trust-control-plane/backend/api/generated/telemetry/v1"
)

// EventEmitter emits telemetry events (e.g. to OTel Logs). Best-effort; callers log and ignore errors.
type EventEmitter interface {
	Emit(ctx context.Context, event *telemetryv1.TelemetryEvent) error
}
