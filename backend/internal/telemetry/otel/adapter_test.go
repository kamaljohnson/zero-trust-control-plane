package otel

import (
	"context"
	"testing"
	"time"

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

func TestEmit_ZeroTimestamp_SetsCurrentTime(t *testing.T) {
	cap := &recordCapture{}
	em := NewEventEmitterWithLogger(cap)
	event := &telemetryv1.TelemetryEvent{
		OrgId:     "org1",
		EventType: "test",
		Source:    "test",
		CreatedAt: timestamppb.New(time.Time{}), // Zero time
	}
	before := time.Now().UTC()
	if err := em.Emit(context.Background(), event); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	after := time.Now().UTC()
	rec := cap.rec
	timestamp := rec.Timestamp()
	if timestamp.IsZero() {
		t.Error("timestamp should be set when CreatedAt is zero")
	}
	if timestamp.Before(before) || timestamp.After(after) {
		t.Errorf("timestamp = %v, should be between %v and %v", timestamp, before, after)
	}
}

func TestEmit_NilCreatedAt_SetsCurrentTime(t *testing.T) {
	cap := &recordCapture{}
	em := NewEventEmitterWithLogger(cap)
	event := &telemetryv1.TelemetryEvent{
		OrgId:     "org1",
		EventType: "test",
		Source:    "test",
		CreatedAt: nil, // Nil CreatedAt
	}
	before := time.Now().UTC()
	if err := em.Emit(context.Background(), event); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	after := time.Now().UTC()
	rec := cap.rec
	timestamp := rec.Timestamp()
	if timestamp.IsZero() {
		t.Error("timestamp should be set when CreatedAt is nil")
	}
	if timestamp.Before(before) || timestamp.After(after) {
		t.Errorf("timestamp = %v, should be between %v and %v", timestamp, before, after)
	}
}

func TestEmit_PartialFields(t *testing.T) {
	cap := &recordCapture{}
	em := NewEventEmitterWithLogger(cap)
	event := &telemetryv1.TelemetryEvent{
		OrgId:     "org1",
		EventType: "test",
		// Missing UserId, DeviceId, SessionId, Source
	}
	if err := em.Emit(context.Background(), event); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	rec := cap.rec
	attrs := make(map[string]string)
	rec.WalkAttributes(func(kv otellog.KeyValue) bool {
		attrs[kv.Key] = kv.Value.AsString()
		return true
	})
	if attrs["org_id"] != "org1" {
		t.Errorf("org_id = %q, want %q", attrs["org_id"], "org1")
	}
	if attrs["event_type"] != "test" {
		t.Errorf("event_type = %q, want %q", attrs["event_type"], "test")
	}
	// Missing fields should not be in attributes
	if attrs["user_id"] != "" {
		t.Errorf("user_id should not be set, got %q", attrs["user_id"])
	}
}

func TestEmit_EmptyStringFields(t *testing.T) {
	cap := &recordCapture{}
	em := NewEventEmitterWithLogger(cap)
	event := &telemetryv1.TelemetryEvent{
		OrgId:     "",
		UserId:    "",
		DeviceId:  "",
		SessionId: "",
		EventType: "test",
		Source:    "",
		Metadata:  []byte{},
	}
	if err := em.Emit(context.Background(), event); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	rec := cap.rec
	attrs := make(map[string]string)
	rec.WalkAttributes(func(kv otellog.KeyValue) bool {
		attrs[kv.Key] = kv.Value.AsString()
		return true
	})
	// Empty string fields should not be added as attributes
	if attrs["org_id"] != "" {
		t.Errorf("org_id should not be set for empty string, got %q", attrs["org_id"])
	}
	if attrs["event_type"] != "test" {
		t.Errorf("event_type = %q, want %q", attrs["event_type"], "test")
	}
}

func TestEmit_AllFieldsPopulated(t *testing.T) {
	cap := &recordCapture{}
	em := NewEventEmitterWithLogger(cap)
	now := time.Now().UTC()
	event := &telemetryv1.TelemetryEvent{
		OrgId:     "org-1",
		UserId:    "user-1",
		DeviceId:  "device-1",
		SessionId: "session-1",
		EventType: "custom_event",
		Source:    "custom_source",
		Metadata:  []byte(`{"custom":"data"}`),
		CreatedAt: timestamppb.New(now),
	}
	if err := em.Emit(context.Background(), event); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	rec := cap.rec

	// Check timestamp
	if rec.Timestamp().Unix() != now.Unix() {
		t.Errorf("timestamp = %v, want %v", rec.Timestamp(), now)
	}

	// Check body
	if rec.Body().Empty() {
		t.Error("body should be set when metadata is non-empty")
	}
	if string(rec.Body().AsBytes()) != `{"custom":"data"}` {
		t.Errorf("body = %q, want %q", string(rec.Body().AsBytes()), `{"custom":"data"}`)
	}

	// Check all attributes
	attrs := make(map[string]string)
	rec.WalkAttributes(func(kv otellog.KeyValue) bool {
		attrs[kv.Key] = kv.Value.AsString()
		return true
	})
	wantAttrs := map[string]string{
		"org_id":     "org-1",
		"user_id":    "user-1",
		"device_id":  "device-1",
		"session_id": "session-1",
		"event_type": "custom_event",
		"source":     "custom_source",
	}
	for k, v := range wantAttrs {
		if attrs[k] != v {
			t.Errorf("attr %q = %q, want %q", k, attrs[k], v)
		}
	}
}
