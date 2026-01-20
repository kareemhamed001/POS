# Distributed Tracing Implementation Guide

## Overview

This project now includes comprehensive distributed tracing using **OpenTelemetry** and **Jaeger** to monitor request flows, identify performance bottlenecks, and debug issues in production.

## What's Been Added

### 1. **Tracer Package** (`pkg/tracer/tracer.go`)

- Initializes OpenTelemetry tracer provider
- Configures Jaeger exporter
- Supports environment-based sampling (100% in dev, 20% in production)
- Graceful shutdown handling

### 2. **HTTP Tracing Middleware** (`internal/delivery/http/middleware/tracing.go`)

- Automatically traces all HTTP requests
- Extracts trace context from incoming headers (W3C Trace Context)
- Captures HTTP metadata (method, path, status code, client IP, user agent)
- Records errors and sets span status based on response

### 3. **Product Handler Instrumentation** (`internal/delivery/http/handler/product_handler.go`)

- Each handler method creates a span
- Nested spans for validation, file processing, and business logic
- Attributes include product details (ID, name, price, discount type)
- Error recording with stack traces

### 4. **Product Usecase Instrumentation** (`internal/usecase/product_usecase.go`)

- Separate spans for database operations, cache operations
- Cache hit/miss tracking
- Performance metrics for each operation
- Product count and metadata attributes

## Architecture

```
HTTP Request
    │
    ├─ TracingMiddleware (automatic)
    │   └─ Span: "POST /api/products"
    │
    ├─ ProductHandler.CreateProduct
    │   ├─ Span: "ProductHandler.ValidateProduct"
    │   ├─ Span: "ProductHandler.ProcessProductData"
    │   │   └─ File upload tracing
    │   └─ Calls ProductUsecase
    │
    └─ ProductUsecase.CreateProduct
        ├─ Span: "Database.CreateProduct"
        └─ Span: "Cache.InvalidateProductList"
```

## Setup Instructions

### 1. Start Jaeger (Optional but recommended for visualization)

```bash
# Start Jaeger using docker-compose
docker-compose -f docker-compose.jaeger.yml up -d

# Verify Jaeger is running
curl http://localhost:16686
```

**Jaeger UI**: http://localhost:16686

### 2. Configure Environment Variables

Add to your `.env` file:

```bash
# Jaeger endpoint (default works with docker-compose setup)
JAEGER_ENDPOINT=http://jaeger:14268/api/traces

# For local development (if Jaeger is on host machine)
# JAEGER_ENDPOINT=http://localhost:14268/api/traces

# Environment (affects sampling rate)
ENV=development  # or production
```

### 3. Rebuild and Run Your Application

```bash
# Rebuild with new dependencies
docker-compose up app -d --build

# Or for local development
go run cmd/api/main.go
```

### 4. Generate Some Traffic

```bash
# Create a product
curl -X POST http://localhost:8080/api/products \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "name=Test Product" \
  -F "description=Test Description" \
  -F "price=99.99" \
  -F "quantity=10" \
  -F "image=@/path/to/image.jpg"

# List products
curl http://localhost:8080/api/products

# Get a specific product
curl http://localhost:8080/api/products/1

# Update product
curl -X PUT http://localhost:8080/api/products/1 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "price=89.99"

# Delete product
curl -X DELETE http://localhost:8080/api/products/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 5. View Traces in Jaeger UI

1. Open http://localhost:16686
2. Select "pos-api" from the Service dropdown
3. Click "Find Traces"
4. Click on any trace to see the detailed timeline

## What You'll See in Jaeger

### Trace Timeline Example

```
Trace: Create Product (Total: 142ms)
├─ POST /api/products (142ms)
│  ├─ ProductHandler.CreateProduct (138ms)
│  │  ├─ ProductHandler.ValidateProduct (2ms)
│  │  ├─ ProductHandler.ProcessProductData (45ms)
│  │  │  └─ File Upload to S3 (43ms)
│  │  └─ ProductUsecase.CreateProduct (88ms)
│  │     ├─ Database.CreateProduct (85ms)
│  │     └─ Cache.InvalidateProductList (3ms)
│  └─ JSON Serialization (2ms)
```

### Span Attributes

Each span includes contextual information:

```json
{
  "product.id": 123,
  "product.name": "Wireless Mouse",
  "product.price": 29.99,
  "product.discount_type": "percentage",
  "products.count": 45,
  "cache.hit": false,
  "image.updated": true,
  "http.method": "POST",
  "http.url": "/api/products",
  "http.status_code": 201,
  "service.name": "pos-api"
}
```

## Performance Insights

### What to Look For

1. **Slow Operations**
   - Database queries taking too long
   - File uploads causing delays
   - Cache operations failing

2. **Cache Effectiveness**
   - `cache.hit: true` vs `cache.hit: false` ratio
   - Time saved by cache hits

3. **Error Patterns**
   - Which operations fail most frequently
   - Error propagation through layers

4. **Request Distribution**
   - Which endpoints are most used
   - Peak traffic patterns

### Example Queries in Jaeger

```
# Find slow product creation requests
service=pos-api operation="ProductHandler.CreateProduct" minDuration=500ms

# Find all cache misses
service=pos-api tags={"cache.hit":"false"}

# Find errors
service=pos-api tags={"error":"true"}

# Find requests for specific product
service=pos-api tags={"product.id":"123"}
```

## Sampling Strategy

### Development

- **Sample Rate**: 100% (all traces)
- **Purpose**: Full visibility during development

### Production

- **Sample Rate**: 20% (1 in 5 traces)
- **Purpose**: Reduce overhead while maintaining statistical significance

To change sampling rate, modify `pkg/tracer/tracer.go`:

```go
func getSampler() trace.Sampler {
    env := getEnv("ENV", "development")
    if env == "production" {
        // Change 0.2 to your desired rate (0.0 to 1.0)
        return trace.ParentBased(trace.TraceIDRatioBased(0.2))
    }
    return trace.AlwaysSample()
}
```

## Integration with Existing Tools

### Logging Correlation

Traces include trace IDs that can be correlated with logs:

```go
// In your handler/usecase, extract trace ID
span := trace.SpanFromContext(ctx)
traceID := span.SpanContext().TraceID().String()

// Add to logs
logger.Info("Processing request", "trace_id", traceID, "product_id", id)
```

### Prometheus Metrics

Combine with metrics for complete observability:

```
# Request duration from traces -> Alert if P95 > 500ms
# Error rate from traces -> Alert if > 1%
# Cache hit ratio from trace attributes -> Alert if < 80%
```

## Production Deployment

### Docker Compose Update

Add Jaeger to your main `docker-compose.yml`:

```yaml
services:
  app:
    environment:
      - JAEGER_ENDPOINT=http://jaeger:14268/api/traces
    depends_on:
      - jaeger

  jaeger:
    image: jaegertracing/all-in-one:1.51
    ports:
      - "16686:16686"
      - "14268:14268"
```

### Kubernetes Deployment

For production Kubernetes clusters, consider:

1. **Jaeger Operator**: Manage Jaeger deployment
2. **Elasticsearch Backend**: Store traces long-term
3. **Kafka for Buffering**: Handle high trace volume

Example Jaeger deployment:

```yaml
apiVersion: jaegertracing.io/v1
kind: Jaeger
metadata:
  name: pos-jaeger
spec:
  strategy: production
  storage:
    type: elasticsearch
    options:
      es:
        server-urls: http://elasticsearch:9200
```

## Extending Tracing

### Add Tracing to Other Handlers

Follow the same pattern as ProductHandler:

```go
type OrderHandler struct {
    orderUsecase usecase.OrderUsecaseInterface
    tracer       trace.Tracer
}

func NewOrderHandler(...) *OrderHandler {
    return &OrderHandler{
        // ... other fields
        tracer: otel.Tracer("order-handler"),
    }
}

func (h *OrderHandler) CreateOrder(ctx *gin.Context) {
    reqCtx, span := h.tracer.Start(ctx.Request.Context(), "OrderHandler.CreateOrder")
    defer span.End()

    span.SetAttributes(
        attribute.Int("order.items_count", len(items)),
        attribute.Float64("order.total", total),
    )

    // ... handler logic
}
```

### Add Database Tracing (GORM)

Install GORM OpenTelemetry plugin:

```bash
go get go.opentelemetry.io/contrib/instrumentation/gorm.io/gorm/otelgorm
```

In `internal/db/db.go`:

```go
import "go.opentelemetry.io/contrib/instrumentation/gorm.io/gorm/otelgorm"

func InitializeDB(...) (*gorm.DB, error) {
    // ... existing code

    // Add OpenTelemetry plugin
    if err := db.Use(otelgorm.NewPlugin()); err != nil {
        return nil, err
    }

    return db, nil
}
```

### Add Redis Tracing

Install Redis OpenTelemetry hook:

```bash
go get github.com/go-redis/redis/extra/redisotel/v9
```

In `pkg/redis/client.go`:

```go
import "github.com/go-redis/redis/extra/redisotel/v9"

func NewClient(cfg *config.Config) (*Client, error) {
    rdb := redis.NewClient(&redis.Options{...})

    // Enable tracing
    if err := redisotel.InstrumentTracing(rdb); err != nil {
        return nil, err
    }

    return &Client{Client: rdb}, nil
}
```

## Troubleshooting

### Traces Not Appearing in Jaeger

1. **Check Jaeger is running**:

   ```bash
   docker ps | grep jaeger
   ```

2. **Verify endpoint**:

   ```bash
   curl http://localhost:14268/api/traces
   ```

3. **Check application logs**:

   ```bash
   docker logs pos_app | grep -i tracer
   ```

4. **Network connectivity** (if using Docker):
   ```bash
   # Ensure app and Jaeger are on same network
   docker network inspect pos_network
   ```

### High Memory Usage

If tracing causes memory issues:

1. Reduce sampling rate in production
2. Adjust batch size in `pkg/tracer/tracer.go`:
   ```go
   trace.WithBatcher(exporter,
       trace.WithMaxExportBatchSize(256), // Reduce from 512
   )
   ```

### Performance Impact

- **Development**: ~2-5% overhead
- **Production (20% sampling)**: <1% overhead

## Resources

- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/instrumentation/go/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [W3C Trace Context](https://www.w3.org/TR/trace-context/)
- [Project Tracing Guide](TRACING_OBSERVABILITY.md)

## Summary

You now have:

✅ **Automatic HTTP request tracing**  
✅ **Detailed product operation tracing**  
✅ **Cache and database operation visibility**  
✅ **Error tracking and performance monitoring**  
✅ **Production-ready sampling strategy**  
✅ **Jaeger UI for visualization**

Start making requests and watch your traces appear in real-time at http://localhost:16686!
