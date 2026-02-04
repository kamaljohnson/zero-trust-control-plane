package telemetry

import (
	"context"
	"sync"
	"testing"
	"time"

	telemetryv1 "zero-trust-control-plane/backend/api/generated/telemetry/v1"
)

// mockEventEmitter implements EventEmitter for tests.
type mockEventEmitter struct {
	mu     sync.Mutex
	events []*telemetryv1.TelemetryEvent
	emitErr error
	delay   time.Duration
}

func (m *mockEventEmitter) Emit(ctx context.Context, event *telemetryv1.TelemetryEvent) error {
	if m.delay > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(m.delay):
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.events == nil {
		m.events = make([]*telemetryv1.TelemetryEvent, 0)
	}
	m.events = append(m.events, event)
	return m.emitErr
}

func (m *mockEventEmitter) getEvents() []*telemetryv1.TelemetryEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.events
}

func TestEmitAsync_NilEmitter(t *testing.T) {
	ctx := context.Background()
	event := &telemetryv1.TelemetryEvent{
		OrgId:     "org-1",
		EventType: "test",
	}

	// Should not panic
	EmitAsync(nil, ctx, event)
}

func TestEmitAsync_NilEvent(t *testing.T) {
	emitter := &mockEventEmitter{}
	ctx := context.Background()

	// Should not panic
	EmitAsync(emitter, ctx, nil)

	// Give goroutine time to run (if it starts)
	time.Sleep(10 * time.Millisecond)

	// Should not have emitted anything
	events := emitter.getEvents()
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestEmitAsync_SuccessfulEmit(t *testing.T) {
	emitter := &mockEventEmitter{}
	ctx := context.Background()
	event := &telemetryv1.TelemetryEvent{
		OrgId:     "org-1",
		UserId:    "user-1",
		EventType: "test_event",
		Source:    "test",
	}

	EmitAsync(emitter, ctx, event)

	// Wait for goroutine to complete
	time.Sleep(100 * time.Millisecond)

	events := emitter.getEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].OrgId != "org-1" {
		t.Errorf("event org_id = %q, want %q", events[0].OrgId, "org-1")
	}
	if events[0].UserId != "user-1" {
		t.Errorf("event user_id = %q, want %q", events[0].UserId, "user-1")
	}
	if events[0].EventType != "test_event" {
		t.Errorf("event type = %q, want %q", events[0].EventType, "test_event")
	}
}

func TestEmitAsync_UsesBackgroundContext(t *testing.T) {
	emitter := &mockEventEmitter{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel the request context immediately

	event := &telemetryv1.TelemetryEvent{
		OrgId:     "org-1",
		EventType: "test",
	}

	// Should still emit even though request context is cancelled
	EmitAsync(emitter, ctx, event)

	// Wait for goroutine to complete
	time.Sleep(100 * time.Millisecond)

	events := emitter.getEvents()
	if len(events) != 1 {
		t.Errorf("expected 1 event (context.Background used), got %d", len(events))
	}
}

func TestEmitAsync_Timeout(t *testing.T) {
	emitter := &mockEventEmitter{
		delay: emitTimeout + 100*time.Millisecond, // Longer than timeout
	}
	ctx := context.Background()
	event := &telemetryv1.TelemetryEvent{
		OrgId:     "org-1",
		EventType: "test",
	}

	EmitAsync(emitter, ctx, event)

	// Wait for timeout
	time.Sleep(emitTimeout + 200*time.Millisecond)

	// Event might not be emitted due to timeout, but should not panic
	// The error is logged but doesn't affect the caller
}

func TestEmitAsync_ErrorHandling(t *testing.T) {
	emitter := &mockEventEmitter{
		emitErr: context.DeadlineExceeded,
	}
	ctx := context.Background()
	event := &telemetryv1.TelemetryEvent{
		OrgId:     "org-1",
		EventType: "test",
	}

	// Should not panic on error
	EmitAsync(emitter, ctx, event)

	// Wait for goroutine to complete
	time.Sleep(100 * time.Millisecond)

	// Error is logged but doesn't affect the caller
	// Event might still be recorded (implementation detail)
}

func TestEmitAsync_MultipleEvents(t *testing.T) {
	emitter := &mockEventEmitter{}
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		event := &telemetryv1.TelemetryEvent{
			OrgId:     "org-1",
			EventType: "test",
		}
		EmitAsync(emitter, ctx, event)
	}

	// Wait for all goroutines to complete
	time.Sleep(200 * time.Millisecond)

	events := emitter.getEvents()
	if len(events) != 5 {
		t.Errorf("expected 5 events, got %d", len(events))
	}
}

func TestEmitAsync_ConcurrentAccess(t *testing.T) {
	emitter := &mockEventEmitter{}
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			event := &telemetryv1.TelemetryEvent{
				OrgId:     "org-1",
				EventType: "test",
			}
			EmitAsync(emitter, ctx, event)
		}(i)
	}

	wg.Wait()
	// Wait for all async emits to complete
	time.Sleep(200 * time.Millisecond)

	events := emitter.getEvents()
	if len(events) != 10 {
		t.Errorf("expected 10 events, got %d", len(events))
	}
}
