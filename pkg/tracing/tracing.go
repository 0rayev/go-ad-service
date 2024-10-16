package tracing

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace" // stdout exporter
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// InitTracer initializes an OpenTelemetry tracer with a stdout exporter.
func InitTracer() func() {
	// Create a stdout exporter
	exp, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(), // Enables pretty-printing for easy reading
	)
	if err != nil {
		log.Fatalf("failed to create stdout exporter: %v", err)
	}

	// Create a new tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // Always sample for testing purposes
		sdktrace.WithBatcher(exp),                     // Send traces to stdout
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("ad-service"),
		)),
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
