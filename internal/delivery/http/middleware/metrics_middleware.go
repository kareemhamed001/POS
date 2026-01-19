package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kareemhamed001/POS/pkg/tracer"
)

// MetricsMiddleware records HTTP metrics for all requests.
func MetricsMiddleware(metricsCollector *tracer.GenericMetricsCollector) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		// Continue request
		ctx.Next()

		// Extract handler name from route
		handlerName := extractHandlerName(ctx.Request.URL.Path)
		operation := extractOperationName(ctx.Request.Method, ctx.Request.URL.Path)

		// Record metrics
		metricsCollector.Record(
			ctx.Request.Context(),
			handlerName,
			operation,
			ctx.Writer.Status(),
			nil,
			time.Since(start),
		)
	}
}

func extractHandlerName(path string) string {
	// Extract handler name from API path
	// e.g., "/api/v1/users" -> "user"
	// e.g., "/api/v1/products" -> "product"
	parts := strings.Split(path, "/")
	if len(parts) >= 4 {
		resource := parts[3]
		if resource != "" {
			// Remove trailing 's' for singular form
			if strings.HasSuffix(resource, "s") {
				return strings.TrimSuffix(resource, "s")
			}
			return resource
		}
	}
	return "unknown"
}

func extractOperationName(method, path string) string {
	// Extract operation name from method and path
	// e.g., "GET /api/v1/users" -> "list_users"
	// e.g., "POST /api/v1/users" -> "create_user"
	parts := strings.Split(path, "/")
	if len(parts) >= 4 {
		resource := parts[3]

		switch method {
		case "GET":
			// Check if it's a detail endpoint (has ID)
			if len(parts) > 4 && parts[4] != "" {
				return "get_" + strings.TrimSuffix(resource, "s")
			}
			return "list_" + resource
		case "POST":
			return "create_" + strings.TrimSuffix(resource, "s")
		case "PUT":
			return "update_" + strings.TrimSuffix(resource, "s")
		case "DELETE":
			return "delete_" + strings.TrimSuffix(resource, "s")
		case "PATCH":
			return "patch_" + strings.TrimSuffix(resource, "s")
		}
	}

	return strings.ToLower(method)
}
