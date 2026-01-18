package tracer

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kareemhamed001/POS/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// InitTracer initializes the OpenTelemetry tracer provider with Jaeger exporter
func InitTracer(serviceName, jaegerEndpoint string) (*trace.TracerProvider, error) {
	// Create Jaeger exporter
	exporter, err := jaeger.New(
		jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(jaegerEndpoint),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create jaeger exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.DeploymentEnvironmentKey.String(getEnv("APP_ENV", "development")),
			attribute.String("service.version", getEnv("APP_VERSION", "1.0.0")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter,
			trace.WithBatchTimeout(5*time.Second),
			trace.WithMaxExportBatchSize(512),
		),
		trace.WithResource(res),
		trace.WithSampler(getSampler()),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	return tp, nil
}

// getSampler returns appropriate sampler based on environment
func getSampler() trace.Sampler {
	env := getEnv("APP_ENV", "development")
	if env == "production" {
		// Sample 20% of traces in production
		logger.Info("Using TraceIDRatioBased sampler with 20% sampling rate for production environment")
		return trace.ParentBased(trace.TraceIDRatioBased(0.2))
	}
	// Sample all traces in development
	return trace.AlwaysSample()
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Shutdown gracefully shuts down the tracer provider
func Shutdown(ctx context.Context, tp *trace.TracerProvider) error {
	if tp == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return tp.Shutdown(ctx)
}
