package otel

import (
	"context"
	"time"

	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"

	telemetryv1 "zero-trust-control-plane/backend/api/generated/telemetry/v1"
	"zero-trust-control-plane/backend/internal/telemetry"
)

// loggerEmit is the subset of the OTel Logger interface used to emit records. Used so tests can inject a mock.
type loggerEmit interface {
	Emit(context.Context, otellog.Record)
}

// NewEventEmitter returns an EventEmitter that sends events as OTel log records via the given LoggerProvider.
// If provider is nil, returns a no-op emitter.
func NewEventEmitter(provider *sdklog.LoggerProvider) telemetry.EventEmitter {
	if provider == nil {
		return noopEmitter{}
	}
	return &otelEmitter{logger: provider.Logger("ztcp.telemetry")}
}

// NewEventEmitterWithLogger returns an EventEmitter that uses the given logger (for tests).
func NewEventEmitterWithLogger(l loggerEmit) telemetry.EventEmitter {
	return &otelEmitter{logger: l}
}

type noopEmitter struct{}

func (noopEmitter) Emit(context.Context, *telemetryv1.TelemetryEvent) error { return nil }

type otelEmitter struct {
	logger loggerEmit
}

// Emit converts the telemetry event to an OTel log record and emits it. Best-effort; errors are logged.
func (e *otelEmitter) Emit(ctx context.Context, event *telemetryv1.TelemetryEvent) error {
	if event == nil {
		return nil
	}
	rec := otellog.Record{}
	if event.CreatedAt != nil {
		if t := event.CreatedAt.AsTime(); !t.IsZero() {
			rec.SetTimestamp(t)
		}
	}
	// Body is set only when metadata is non-empty; when empty the record has no body. In Loki/LogQL query by attributes (org_id, event_type, source), not by body.
	if len(event.Metadata) > 0 {
		rec.SetBody(otellog.BytesValue(event.Metadata))
	}
	if event.OrgId != "" {
		rec.AddAttributes(otellog.String("org_id", event.OrgId))
	}
	if event.UserId != "" {
		rec.AddAttributes(otellog.String("user_id", event.UserId))
	}
	if event.DeviceId != "" {
		rec.AddAttributes(otellog.String("device_id", event.DeviceId))
	}
	if event.SessionId != "" {
		rec.AddAttributes(otellog.String("session_id", event.SessionId))
	}
	if event.EventType != "" {
		rec.AddAttributes(otellog.String("event_type", event.EventType))
	}
	if event.Source != "" {
		rec.AddAttributes(otellog.String("source", event.Source))
	}
	if rec.Timestamp().IsZero() {
		rec.SetTimestamp(time.Now().UTC())
	}
	e.logger.Emit(ctx, rec)
	return nil
}
