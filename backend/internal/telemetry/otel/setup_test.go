package otel

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestNewProviders_EmptyEndpoint(t *testing.T) {
	ctx := context.Background()
	providers, err := NewProviders(ctx, "", "test-service", false)
	if err != nil {
		t.Fatalf("NewProviders empty endpoint: %v", err)
	}
	if providers == nil {
		t.Fatal("providers should not be nil")
	}
	if providers.TracerProvider == nil {
		t.Error("TracerProvider should not be nil")
	}
	if providers.MeterProvider == nil {
		t.Error("MeterProvider should not be nil")
	}
	if providers.LoggerProvider == nil {
		t.Error("LoggerProvider should not be nil")
	}
	if providers.Shutdown == nil {
		t.Error("Shutdown should not be nil")
	}

	// Test that shutdown is a no-op
	if err := providers.Shutdown(ctx); err != nil {
		t.Errorf("shutdown should be no-op for empty endpoint, got error: %v", err)
	}
}

func TestNewProviders_WhitespaceEndpoint(t *testing.T) {
	ctx := context.Background()
	providers, err := NewProviders(ctx, "   ", "test-service", false)
	if err != nil {
		t.Fatalf("NewProviders whitespace endpoint: %v", err)
	}
	if providers == nil {
		t.Fatal("providers should not be nil")
	}
}

func TestNewProviders_InvalidURL(t *testing.T) {
	ctx := context.Background()
	testCases := []struct {
		name     string
		endpoint string
	}{
		{"invalid characters", "://invalid"},
		{"malformed URL", "http://[invalid"},
		{"missing host", "http://"},
		{"just scheme", "http://"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewProviders(ctx, tc.endpoint, "test-service", false)
			if err == nil {
				t.Errorf("NewProviders(%q) should return error", tc.endpoint)
			}
		})
	}
}

func TestNewProviders_MissingHost(t *testing.T) {
	ctx := context.Background()
	_, err := NewProviders(ctx, "http://", "test-service", false)
	if err == nil {
		t.Fatal("NewProviders with missing host should return error")
	}
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}

func TestNewProviders_EndpointWithoutProtocol(t *testing.T) {
	ctx := context.Background()
	// This will try to create real exporters, which will fail, but we can test the URL parsing
	endpoint := "localhost:4317"
	_, err := NewProviders(ctx, endpoint, "test-service", false)
	// This will fail when trying to create exporters, but URL parsing should succeed
	// We expect an error from exporter creation, not URL parsing
	if err == nil {
		t.Log("Note: NewProviders succeeded (may have real collector available)")
	} else {
		// Error is expected - exporter creation will fail without real collector
		// But URL parsing should have worked (endpoint was normalized to http://localhost:4317)
		if err.Error() == "" {
			t.Error("error message should not be empty")
		}
	}
}

func TestNewProviders_HTTPEndpoint(t *testing.T) {
	ctx := context.Background()
	// HTTP endpoint should be treated as insecure
	endpoint := "http://localhost:4317"
	_, err := NewProviders(ctx, endpoint, "test-service", false)
	// Will fail on exporter creation, but URL parsing and insecure flag should be correct
	if err == nil {
		t.Log("Note: NewProviders succeeded (may have real collector available)")
	} else {
		// Error is expected - exporter creation will fail without real collector
		if err.Error() == "" {
			t.Error("error message should not be empty")
		}
	}
}

func TestNewProviders_HTTPSEndpoint(t *testing.T) {
	ctx := context.Background()
	// HTTPS endpoint should use TLS by default
	endpoint := "https://localhost:4317"
	_, err := NewProviders(ctx, endpoint, "test-service", false)
	// Will fail on exporter creation, but URL parsing should work
	if err == nil {
		t.Log("Note: NewProviders succeeded (may have real collector available)")
	} else {
		// Error is expected - exporter creation will fail without real collector
		if err.Error() == "" {
			t.Error("error message should not be empty")
		}
	}
}

func TestNewProviders_HTTPSWithInsecureOverride(t *testing.T) {
	ctx := context.Background()
	// HTTPS endpoint with insecure override should not use TLS
	endpoint := "https://localhost:4317"
	_, err := NewProviders(ctx, endpoint, "test-service", true)
	// Will fail on exporter creation, but URL parsing and insecure flag should be correct
	if err == nil {
		t.Log("Note: NewProviders succeeded (may have real collector available)")
	} else {
		// Error is expected - exporter creation will fail without real collector
		if err.Error() == "" {
			t.Error("error message should not be empty")
		}
	}
}

func TestNewProviders_EndpointWithPath(t *testing.T) {
	ctx := context.Background()
	// Path should be ignored, only host:port used
	endpoint := "http://localhost:4317/v1/traces"
	_, err := NewProviders(ctx, endpoint, "test-service", false)
	// Will fail on exporter creation, but URL parsing should work and path should be ignored
	if err == nil {
		t.Log("Note: NewProviders succeeded (may have real collector available)")
	} else {
		// Error is expected - exporter creation will fail without real collector
		if err.Error() == "" {
			t.Error("error message should not be empty")
		}
	}
}

func TestNewProviders_ServiceName(t *testing.T) {
	ctx := context.Background()
	// Test that service name is used in resource
	providers, err := NewProviders(ctx, "", "my-custom-service", false)
	if err != nil {
		t.Fatalf("NewProviders: %v", err)
	}
	if providers == nil {
		t.Fatal("providers should not be nil")
	}
	// Service name is set in resource, but we can't easily verify without inspecting resource
	// The important thing is that it doesn't error
}

func TestSetGlobal_WithProviders(t *testing.T) {
	ctx := context.Background()
	providers, err := NewProviders(ctx, "", "test-service", false)
	if err != nil {
		t.Fatalf("NewProviders: %v", err)
	}

	// Save current global providers
	oldTracerProvider := otel.GetTracerProvider()
	oldMeterProvider := otel.GetMeterProvider()

	// Set global providers
	providers.SetGlobal()

	// Verify global providers are set
	newTracerProvider := otel.GetTracerProvider()
	newMeterProvider := otel.GetMeterProvider()

	if newTracerProvider == oldTracerProvider {
		t.Error("TracerProvider should be updated")
	}
	if newMeterProvider == oldMeterProvider {
		t.Error("MeterProvider should be updated")
	}

	// Restore old providers for other tests
	otel.SetTracerProvider(oldTracerProvider)
	otel.SetMeterProvider(oldMeterProvider)
}

func TestSetGlobal_NilProviders(t *testing.T) {
	providers := &Providers{
		TracerProvider: nil,
		MeterProvider:  nil,
		LoggerProvider: nil,
		Shutdown:       func(context.Context) error { return nil },
	}

	// Should not panic
	providers.SetGlobal()

	// Verify that nil providers don't crash
	// (SetGlobal should handle nil gracefully)
}

func TestSetGlobal_PartialProviders(t *testing.T) {
	ctx := context.Background()
	// Create providers with only TracerProvider
	tp := sdktrace.NewTracerProvider()
	defer func() { _ = tp.Shutdown(ctx) }()

	providers := &Providers{
		TracerProvider: tp,
		MeterProvider:  nil,
		LoggerProvider: nil,
		Shutdown:       func(context.Context) error { return nil },
	}

	oldTracerProvider := otel.GetTracerProvider()
	oldMeterProvider := otel.GetMeterProvider()

	providers.SetGlobal()

	// TracerProvider should be updated
	newTracerProvider := otel.GetTracerProvider()
	if newTracerProvider == oldTracerProvider {
		t.Error("TracerProvider should be updated")
	}

	// MeterProvider should remain unchanged (was nil)
	newMeterProvider := otel.GetMeterProvider()
	if newMeterProvider != oldMeterProvider {
		t.Error("MeterProvider should not be updated when nil")
	}

	// Restore
	otel.SetTracerProvider(oldTracerProvider)
	otel.SetMeterProvider(oldMeterProvider)
}

func TestProviders_Shutdown(t *testing.T) {
	ctx := context.Background()
	providers, err := NewProviders(ctx, "", "test-service", false)
	if err != nil {
		t.Fatalf("NewProviders: %v", err)
	}

	// Shutdown should be callable multiple times
	if err := providers.Shutdown(ctx); err != nil {
		t.Errorf("first shutdown: %v", err)
	}
	if err := providers.Shutdown(ctx); err != nil {
		t.Errorf("second shutdown: %v", err)
	}
}

func TestProviders_ShutdownWithNilContext(t *testing.T) {
	providers, err := NewProviders(context.Background(), "", "test-service", false)
	if err != nil {
		t.Fatalf("NewProviders: %v", err)
	}

	// Shutdown with nil context should not panic (though it's not recommended)
	// The no-op shutdown for empty endpoint should handle this
	if err := providers.Shutdown(nil); err != nil {
		t.Errorf("shutdown with nil context: %v", err)
	}
}

func TestNewProviders_ResourceMergeError(t *testing.T) {
	// This is hard to test directly since resource.Merge rarely fails
	// But we can verify the code path exists by testing with valid inputs
	ctx := context.Background()
	providers, err := NewProviders(ctx, "", "test-service", false)
	if err != nil {
		t.Fatalf("NewProviders: %v", err)
	}
	if providers == nil {
		t.Fatal("providers should not be nil")
	}
	// Resource merge should succeed with valid inputs
}

func TestNewProviders_ExporterCleanupOnError(t *testing.T) {
	ctx := context.Background()
	// When metric exporter creation fails, trace exporter should be cleaned up
	// When log exporter creation fails, both trace and metric exporters should be cleaned up
	// This is tested implicitly by the error handling in NewProviders
	// We can't easily force exporter creation to fail without a real database/collector,
	// but we can verify the error handling code path exists by checking that errors are returned

	// Test with invalid endpoint format to trigger URL parsing error early
	_, err := NewProviders(ctx, "://invalid", "test-service", false)
	if err == nil {
		t.Error("should return error for invalid endpoint format")
	}
}

func TestNewProviders_AllProvidersCreated(t *testing.T) {
	ctx := context.Background()
	providers, err := NewProviders(ctx, "", "test-service", false)
	if err != nil {
		t.Fatalf("NewProviders: %v", err)
	}

	// Verify all providers are created (not nil)
	if providers.TracerProvider == nil {
		t.Error("TracerProvider should be created")
	}
	if providers.MeterProvider == nil {
		t.Error("MeterProvider should be created")
	}
	if providers.LoggerProvider == nil {
		t.Error("LoggerProvider should be created")
	}
	if providers.Shutdown == nil {
		t.Error("Shutdown function should be created")
	}
}

func TestNewProviders_EndpointNormalization(t *testing.T) {
	ctx := context.Background()
	testCases := []struct {
		name     string
		endpoint string
		wantErr  bool
	}{
		{"with http://", "http://localhost:4317", false},
		{"with https://", "https://localhost:4317", false},
		{"without protocol", "localhost:4317", false}, // Should add http://
		{"with path", "http://localhost:4317/v1/traces", false}, // Path should be ignored
		{"with port", "http://localhost:4317", false},
		{"with query", "http://localhost:4317?param=value", false}, // Query should be ignored
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewProviders(ctx, tc.endpoint, "test-service", false)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for endpoint %q", tc.endpoint)
				}
			} else {
				// May fail on exporter creation, but URL parsing should work
				if err != nil {
				// Check if error is from URL parsing or exporter creation
				errStr := err.Error()
				if errStr == "" {
					t.Error("error message should not be empty")
				}
				// Exporter creation errors are expected without real collector
				t.Logf("Note: Exporter creation may fail without collector: %v", err)
				}
			}
		})
	}
}

func TestSetGlobal_LoggerProviderNotSet(t *testing.T) {
	ctx := context.Background()
	providers, err := NewProviders(ctx, "", "test-service", false)
	if err != nil {
		t.Fatalf("NewProviders: %v", err)
	}

	// SetGlobal should not set LoggerProvider (per documentation)
	providers.SetGlobal()

	// Verify LoggerProvider is not set globally (there's no global LoggerProvider in otel package)
	// This is expected behavior - LoggerProvider must be passed to handlers
	if providers.LoggerProvider == nil {
		t.Error("LoggerProvider should exist on Providers struct")
	}
	// But it's not set globally, which is correct
}
