// Package otel provides OpenTelemetry TracerProvider, MeterProvider, and LoggerProvider
// configured with OTLP exporters for the gRPC server.
package otel

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
)

// Providers holds the OpenTelemetry providers and a shutdown function.
type Providers struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *metric.MeterProvider
	LoggerProvider *sdklog.LoggerProvider
	Shutdown       func(context.Context) error
}

// NewProviders creates TracerProvider, MeterProvider, and LoggerProvider that export via OTLP to the given endpoint.
// endpoint may be a URL with optional path (e.g. http://localhost:4317 or https://collector:4317/v1/traces); path is ignored and only host:port is used for the gRPC dial.
// If empty, no-op providers are returned and Shutdown is a no-op. https endpoints use TLS unless insecureOverride is true (standard OTEL_EXPORTER_OTLP_INSECURE behavior).
func NewProviders(ctx context.Context, endpoint, serviceName string, insecureOverride bool) (*Providers, error) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return &Providers{
			TracerProvider: sdktrace.NewTracerProvider(),
			MeterProvider:  metric.NewMeterProvider(),
			LoggerProvider: sdklog.NewLoggerProvider(),
			Shutdown:       func(context.Context) error { return nil },
		}, nil
	}

	// Normalize endpoint: OTLP gRPC expects host:port; parse as URL and use Host only so paths are dropped.
	if !strings.Contains(endpoint, "://") {
		endpoint = "http://" + endpoint
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid OTLP endpoint %q: %w", endpoint, err)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("invalid OTLP endpoint %q: missing host", endpoint)
	}
	grpcTarget := u.Host
	insecure := insecureOverride || (u.Scheme != "https")

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	var shutdownFns []func(context.Context) error

	traceOpts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(grpcTarget)}
	if insecure {
		traceOpts = append(traceOpts, otlptracegrpc.WithInsecure())
	}
	traceExp, err := otlptracegrpc.New(ctx, traceOpts...)
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
	)
	shutdownFns = append(shutdownFns, tp.Shutdown)

	metricOpts := []otlpmetricgrpc.Option{otlpmetricgrpc.WithEndpoint(grpcTarget)}
	if insecure {
		metricOpts = append(metricOpts, otlpmetricgrpc.WithInsecure())
	}
	metricExp, err := otlpmetricgrpc.New(ctx, metricOpts...)
	if err != nil {
		_ = tp.Shutdown(ctx)
		return nil, err
	}
	reader := metric.NewPeriodicReader(metricExp, metric.WithInterval(10*time.Second))
	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(reader),
	)
	shutdownFns = append(shutdownFns, mp.Shutdown)

	logOpts := []otlploggrpc.Option{otlploggrpc.WithEndpoint(grpcTarget)}
	if insecure {
		logOpts = append(logOpts, otlploggrpc.WithInsecure())
	}
	logExp, err := otlploggrpc.New(ctx, logOpts...)
	if err != nil {
		_ = tp.Shutdown(ctx)
		_ = mp.Shutdown(ctx)
		return nil, err
	}
	logProcessor := sdklog.NewBatchProcessor(logExp)
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(logProcessor),
		sdklog.WithResource(res),
	)
	shutdownFns = append(shutdownFns, lp.Shutdown)

	shutdown := func(ctx context.Context) error {
		var lastErr error
		for i := len(shutdownFns) - 1; i >= 0; i-- {
			if err := shutdownFns[i](ctx); err != nil {
				log.Printf("telemetry: shutdown: %v", err)
				lastErr = err
			}
		}
		return lastErr
	}

	return &Providers{
		TracerProvider: tp,
		MeterProvider:  mp,
		LoggerProvider: lp,
		Shutdown:       shutdown,
	}, nil
}

// SetGlobal sets the global TracerProvider and MeterProvider so instrumentation (e.g. otelgrpc) uses them.
// It does not set a global LoggerProvider; pass LoggerProvider to handlers that need it.
func (p *Providers) SetGlobal() {
	if p.TracerProvider != nil {
		otel.SetTracerProvider(p.TracerProvider)
	}
	if p.MeterProvider != nil {
		otel.SetMeterProvider(p.MeterProvider)
	}
}
