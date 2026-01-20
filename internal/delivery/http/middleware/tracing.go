package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/kareemhamed001/POS/http"

func TracingMiddleware(serviceName string) gin.HandlerFunc {
	tracer := otel.Tracer(tracerName)

	return func(c *gin.Context) {
		// Extract trace context from incoming request headers
		ctx := otel.GetTextMapPropagator().Extract(
			c.Request.Context(),
			propagation.HeaderCarrier(c.Request.Header),
		)

		// Generate span name from HTTP method and route
		spanName := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
		if c.FullPath() == "" {
			spanName = fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)
		}

		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPMethodKey.String(c.Request.Method),
				semconv.HTTPURLKey.String(c.Request.URL.String()),
				semconv.HTTPTargetKey.String(c.Request.URL.Path),
				semconv.HTTPSchemeKey.String(c.Request.URL.Scheme),
				semconv.HTTPUserAgentKey.String(c.Request.UserAgent()),
				semconv.HTTPClientIPKey.String(c.ClientIP()),
				semconv.NetHostNameKey.String(c.Request.Host),
				attribute.String("service.name", serviceName),
			),
		)
		defer span.End()
		defer func() {
			statusCode := c.Writer.Status()
			span.SetAttributes(semconv.HTTPStatusCodeKey.Int(statusCode))

			if statusCode >= 400 {
				span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", statusCode))
			} else {
				span.SetStatus(codes.Ok, "")
			}

			if len(c.Errors) > 0 {
				span.RecordError(c.Errors.Last())
				span.SetAttributes(attribute.String("error.message", c.Errors.String()))
			}
		}()

		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
