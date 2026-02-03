// Package otel provides OpenTelemetry TracerProvider, MeterProvider, and LoggerProvider
// configured with OTLP exporters for the gRPC server.
package otel

import (
	"context"
	"log"
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
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Providers holds the OpenTelemetry providers and a shutdown function.
type Providers struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *metric.MeterProvider
	LoggerProvider *sdklog.LoggerProvider
	Shutdown       func(context.Context) error
}

// NewProviders creates TracerProvider, MeterProvider, and LoggerProvider that export via OTLP to the given endpoint.
// endpoint is the OTLP gRPC endpoint (e.g. localhost:4317 or http://localhost:4317). If empty, no-op providers are returned and Shutdown is a no-op.
// serviceName is used as the service.name resource attribute (e.g. ztcp-grpc).
func NewProviders(ctx context.Context, endpoint, serviceName string) (*Providers, error) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return &Providers{
			TracerProvider: sdktrace.NewTracerProvider(),
			MeterProvider:  metric.NewMeterProvider(),
			LoggerProvider: sdklog.NewLoggerProvider(),
			Shutdown:       func(context.Context) error { return nil },
		}, nil
	}

	// Normalize endpoint: OTLP gRPC expects host:port; strip scheme for WithEndpoint.
	grpcTarget := endpoint
	if strings.HasPrefix(endpoint, "http://") {
		grpcTarget = strings.TrimPrefix(endpoint, "http://")
	} else if strings.HasPrefix(endpoint, "https://") {
		grpcTarget = strings.TrimPrefix(endpoint, "https://")
	}

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

	// Trace
	traceExp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(grpcTarget), otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
	)
	shutdownFns = append(shutdownFns, tp.Shutdown)

	// Metric
	metricExp, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint(grpcTarget), otlpmetricgrpc.WithInsecure())
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

	// Log
	logExp, err := otlploggrpc.New(ctx, otlploggrpc.WithEndpoint(grpcTarget), otlploggrpc.WithInsecure())
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
