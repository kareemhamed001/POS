package tracer

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kareemhamed001/POS/pkg/logger"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	metricSdk "go.opentelemetry.io/otel/sdk/metric"
)

// GenericMetricsCollector encapsulates OTEL instruments for all HTTP handlers.
type GenericMetricsCollector struct {
	requestCounter metric.Int64Counter
	errorCounter   metric.Int64Counter
	latency        metric.Float64Histogram
}

func InitPrometheusMeterProvider(namespace string) (*metricSdk.MeterProvider, http.Handler, error) {
	// Create custom prometheus registry
	reg := prom.NewRegistry()

	exporter, err := prometheus.New(
		prometheus.WithNamespace(namespace),
		prometheus.WithRegisterer(reg),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create prometheus exporter: %w", err)
	}

	provider := metricSdk.NewMeterProvider(
		metricSdk.WithReader(exporter),
	)

	otel.SetMeterProvider(provider)

	// Create handler from the registry
	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	return provider, handler, nil
}

// NewGenericMetricsCollector creates counters and histograms for HTTP handler metrics.
func NewGenericMetricsCollector(meterName string) *GenericMetricsCollector {
	meter := otel.Meter(meterName)

	requestCounter, err := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		logger.Error("failed to create http_requests_total counter", err)
	}

	errorCounter, err := meter.Int64Counter(
		"http_request_errors_total",
		metric.WithDescription("Total number of failed HTTP requests"),
	)
	if err != nil {
		logger.Error("failed to create http_request_errors_total counter", err)
	}

	latency, err := meter.Float64Histogram(
		"http_request_duration_ms",
		metric.WithDescription("HTTP request duration in milliseconds"),
	)
	if err != nil {
		logger.Error("failed to create http_request_duration_ms histogram", err)
	}

	return &GenericMetricsCollector{
		requestCounter: requestCounter,
		errorCounter:   errorCounter,
		latency:        latency,
	}
}

// Record captures request count, error count, and latency for an HTTP operation.
func (c *GenericMetricsCollector) Record(ctx context.Context, handler, operation string, statusCode int, err error, duration time.Duration) {
	if c == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("handler", handler),
		attribute.String("operation", operation),
		attribute.Int("http.status_code", statusCode),
		attribute.Bool("success", err == nil),
	}

	if c.requestCounter != nil {
		c.requestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	}

	if err != nil && c.errorCounter != nil {
		c.errorCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	}

	if c.latency != nil {
		c.latency.Record(ctx, duration.Seconds()*1000, metric.WithAttributes(attrs...))
	}
}
