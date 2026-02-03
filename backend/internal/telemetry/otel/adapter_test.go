package otel

import (
	"context"
	"testing"

	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"

	"google.golang.org/protobuf/types/known/timestamppb"

	telemetryv1 "zero-trust-control-plane/backend/api/generated/telemetry/v1"
)

func TestNewEventEmitter_NilProvider_ReturnsNoop(t *testing.T) {
	em := NewEventEmitter(nil)
	if em == nil {
		t.Fatal("NewEventEmitter(nil) returned nil")
	}
	if err := em.Emit(context.Background(), nil); err != nil {
		t.Errorf("noop Emit(ctx, nil): %v", err)
	}
	if err := em.Emit(context.Background(), &telemetryv1.TelemetryEvent{OrgId: "org1"}); err != nil {
		t.Errorf("noop Emit(ctx, event): %v", err)
	}
}

func TestEmit_NilEvent_ReturnsNil(t *testing.T) {
	provider := sdklog.NewLoggerProvider()
	defer func() { _ = provider.Shutdown(context.Background()) }()
	em := NewEventEmitter(provider)
	if err := em.Emit(context.Background(), nil); err != nil {
		t.Errorf("Emit(ctx, nil): %v", err)
	}
}

// recordCapture stores the last Record passed to Emit for assertion.
type recordCapture struct {
	rec otellog.Record
}

func (r *recordCapture) Emit(ctx context.Context, rec otellog.Record) {
	r.rec = rec
}

func TestEmit_AttributeAndBodyMapping(t *testing.T) {
	cap := &recordCapture{}
	em := NewEventEmitterWithLogger(cap)
	event := &telemetryv1.TelemetryEvent{
		OrgId:     "org1",
		UserId:    "user1",
		DeviceId:  "dev1",
		SessionId: "sess1",
		EventType: "login",
		Source:    "sdk",
		Metadata:  []byte(`{"key":"value"}`),
		CreatedAt: timestamppb.Now(),
	}
	if err := em.Emit(context.Background(), event); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	rec := cap.rec

	// Body
	if rec.Body().Empty() {
		t.Error("body should be set when metadata is non-empty")
	}
	if got := rec.Body().AsBytes(); string(got) != `{"key":"value"}` {
		t.Errorf("body = %q, want %q", got, event.Metadata)
	}

	// Attributes
	attrs := make(map[string]string)
	rec.WalkAttributes(func(kv otellog.KeyValue) bool {
		attrs[kv.Key] = kv.Value.AsString()
		return true
	})
	want := map[string]string{
		"org_id": "org1", "user_id": "user1", "device_id": "dev1",
		"session_id": "sess1", "event_type": "login", "source": "sdk",
	}
	for k, v := range want {
		if attrs[k] != v {
			t.Errorf("attr %q = %q, want %q", k, attrs[k], v)
		}
	}
}

func TestEmit_EmptyMetadata_NoBodySet(t *testing.T) {
	cap := &recordCapture{}
	em := NewEventEmitterWithLogger(cap)
	event := &telemetryv1.TelemetryEvent{
		OrgId:     "org1",
		EventType: "ping",
		Source:    "test",
		Metadata:  nil,
	}
	if err := em.Emit(context.Background(), event); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	rec := cap.rec
	if !rec.Body().Empty() {
		t.Error("body should be empty when metadata is nil")
	}
	attrs := make(map[string]string)
	rec.WalkAttributes(func(kv otellog.KeyValue) bool {
		attrs[kv.Key] = kv.Value.AsString()
		return true
	})
	if attrs["org_id"] != "org1" || attrs["event_type"] != "ping" || attrs["source"] != "test" {
		t.Errorf("attributes = %v", attrs)
	}
}
