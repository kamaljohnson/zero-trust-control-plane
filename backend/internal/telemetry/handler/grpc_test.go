package handler

import (
	"context"
	"testing"
	"time"

	telemetryv1 "zero-trust-control-plane/backend/api/generated/telemetry/v1"
)

// mockEmitter sends each emitted event to ch so tests can assert. Buffer ch with capacity >= maxBatchSize for batch tests.
type mockEmitter struct {
	ch chan *telemetryv1.TelemetryEvent
}

func (m *mockEmitter) Emit(ctx context.Context, event *telemetryv1.TelemetryEvent) error {
	if m.ch != nil {
		m.ch <- event
	}
	return nil
}

func TestEmitTelemetryEvent_NilRequest(t *testing.T) {
	ch := make(chan *telemetryv1.TelemetryEvent, 1)
	srv := NewServer(&mockEmitter{ch: ch})
	resp, err := srv.EmitTelemetryEvent(context.Background(), nil)
	if err != nil {
		t.Fatalf("EmitTelemetryEvent: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	select {
	case <-ch:
		t.Fatal("emitter should not be called for nil request")
	case <-time.After(100 * time.Millisecond):
		// expected: no event sent
	}
}

func TestEmitTelemetryEvent_NilEmitter(t *testing.T) {
	srv := NewServer(nil)
	req := &telemetryv1.EmitTelemetryEventRequest{OrgId: "org1", EventType: "test", Source: "unit"}
	resp, err := srv.EmitTelemetryEvent(context.Background(), req)
	if err != nil {
		t.Fatalf("EmitTelemetryEvent: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
}

func TestEmitTelemetryEvent_ValidRequest(t *testing.T) {
	ch := make(chan *telemetryv1.TelemetryEvent, 1)
	srv := NewServer(&mockEmitter{ch: ch})
	req := &telemetryv1.EmitTelemetryEventRequest{
		OrgId:     "org1",
		UserId:    "user1",
		DeviceId:  "dev1",
		SessionId: "sess1",
		EventType: "login",
		Source:    "sdk",
		Metadata:  []byte(`{"key":"value"}`),
	}
	resp, err := srv.EmitTelemetryEvent(context.Background(), req)
	if err != nil {
		t.Fatalf("EmitTelemetryEvent: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	select {
	case event := <-ch:
		if event.OrgId != "org1" || event.UserId != "user1" || event.DeviceId != "dev1" ||
			event.SessionId != "sess1" || event.EventType != "login" || event.Source != "sdk" {
			t.Errorf("event fields: org_id=%q user_id=%q device_id=%q session_id=%q event_type=%q source=%q",
				event.OrgId, event.UserId, event.DeviceId, event.SessionId, event.EventType, event.Source)
		}
		if string(event.Metadata) != `{"key":"value"}` {
			t.Errorf("metadata = %q, want %q", event.Metadata, req.Metadata)
		}
		if event.CreatedAt == nil {
			t.Error("CreatedAt should be set")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no event received on channel")
	}
}

func TestBatchEmitTelemetry_NilRequest(t *testing.T) {
	ch := make(chan *telemetryv1.TelemetryEvent, 1)
	srv := NewServer(&mockEmitter{ch: ch})
	resp, err := srv.BatchEmitTelemetry(context.Background(), nil)
	if err != nil {
		t.Fatalf("BatchEmitTelemetry: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	select {
	case <-ch:
		t.Fatal("emitter should not be called for nil request")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestBatchEmitTelemetry_NilEmitter(t *testing.T) {
	srv := NewServer(nil)
	req := &telemetryv1.BatchEmitTelemetryRequest{
		Events: []*telemetryv1.EmitTelemetryEventRequest{
			{OrgId: "org1", EventType: "e1", Source: "s"},
		},
	}
	resp, err := srv.BatchEmitTelemetry(context.Background(), req)
	if err != nil {
		t.Fatalf("BatchEmitTelemetry: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
}

func TestBatchEmitTelemetry_WithNils(t *testing.T) {
	ch := make(chan *telemetryv1.TelemetryEvent, 4)
	srv := NewServer(&mockEmitter{ch: ch})
	req := &telemetryv1.BatchEmitTelemetryRequest{
		Events: []*telemetryv1.EmitTelemetryEventRequest{
			{OrgId: "org1", EventType: "a", Source: "s"},
			nil,
			{OrgId: "org2", EventType: "b", Source: "s"},
			nil,
		},
	}
	resp, err := srv.BatchEmitTelemetry(context.Background(), req)
	if err != nil {
		t.Fatalf("BatchEmitTelemetry: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	var received int
	deadline := time.After(2 * time.Second)
	for received < 2 {
		select {
		case e := <-ch:
			received++
			if e == nil {
				t.Error("received nil event")
			}
		case <-deadline:
			if received != 2 {
				t.Errorf("received %d events, want 2", received)
			}
			return
		}
	}
}

func TestBatchEmitTelemetry_OverMaxBatchSize(t *testing.T) {
	ch := make(chan *telemetryv1.TelemetryEvent, maxBatchSize+10)
	srv := NewServer(&mockEmitter{ch: ch})
	events := make([]*telemetryv1.EmitTelemetryEventRequest, 501)
	for i := range events {
		events[i] = &telemetryv1.EmitTelemetryEventRequest{
			OrgId: "org1", EventType: "e", Source: "s",
		}
	}
	req := &telemetryv1.BatchEmitTelemetryRequest{Events: events}
	resp, err := srv.BatchEmitTelemetry(context.Background(), req)
	if err != nil {
		t.Fatalf("BatchEmitTelemetry: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	var received int
	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-ch:
			received++
			if received > maxBatchSize {
				t.Errorf("received %d events, cap is %d", received, maxBatchSize)
				return
			}
		case <-deadline:
			if received != maxBatchSize {
				t.Errorf("received %d events, want %d", received, maxBatchSize)
			}
			return
		}
	}
}
