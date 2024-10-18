package tracing

import (
	"ad_service/internal/config"
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// InitTracer initializes an OpenTelemetry tracer with a Jaeger exporter.
func InitTracer(cfg config.TracingConfig) func() {

	// Set up headers for the HTTP client
	headers := map[string]string{
		"content-type": "application/json",
	}

	// Create OTLP exporter using HTTP and the Jaeger endpoint
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(cfg.JaegerEndpoint),
		otlptracehttp.WithHeaders(headers),
		otlptracehttp.WithInsecure(), // Disable TLS
	)

	// Initialize the trace exporter
	exp, err := otlptrace.New(context.Background(), client)
	if err != nil {
		log.Fatalf("failed to create Jaeger exporter: %v", err)
	}

	// Create and configure a new tracer provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(
			exp,
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
			trace.WithBatchTimeout(trace.DefaultScheduleDelay*time.Millisecond),
		),
		trace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("ad-service"),
			),
		),
	)

	// Set the global tracer provider
	otel.SetTracerProvider(tp)

	// Return a shutdown function for graceful shutdown
	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("failed to shutdown tracer provider: %v", err)
		}
	}
}
