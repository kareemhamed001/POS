# Distributed Tracing & Observability - Senior Engineering Guide

> **Target Audience**: Senior Backend Engineers implementing production-grade observability in Go microservices

This guide demonstrates enterprise-level tracing, logging, and observability patterns using real examples from this POS system. We'll cover advanced topics including distributed tracing, structured logging, metrics collection, and correlation strategies.

## Table of Contents

1. [Observability Pillars](#observability-pillars)
2. [Structured Logging with Zap](#structured-logging-with-zap)
3. [Distributed Tracing Architecture](#distributed-tracing-architecture)
4. [Context Propagation Patterns](#context-propagation-patterns)
5. [Request Correlation & Tracking](#request-correlation--tracking)
6. [Performance Monitoring](#performance-monitoring)
7. [Error Tracking & Aggregation](#error-tracking--aggregation)
8. [Production Best Practices](#production-best-practices)

---

## Observability Pillars

### The Three Pillars

1. **Logs**: Discrete events with timestamps (what happened)
2. **Metrics**: Numerical measurements over time (how much/how many)
3. **Traces**: Request flow across services (where time was spent)

### Current Implementation Status

```
✅ Structured Logging (Zap)
✅ HTTP Request Logging
✅ Cache Operation Logging
⚠️  Distributed Tracing (Manual - needs OpenTelemetry)
⚠️  Metrics Collection (needs Prometheus)
❌ Error Aggregation (needs Sentry/Rollbar)
```

---

## Structured Logging with Zap

### Architecture Overview

Our logging implementation uses **Uber's Zap** library with a singleton pattern for global access.

**File**: [`pkg/logger/logger.go`](pkg/logger/logger.go)

```go
type Logger struct {
    *zap.SugaredLogger
}

var (
    globalLogger *Logger
    once         sync.Once
)

func InitGlobal(env string) {
    once.Do(func() {
        if env == "production" {
            base, _ = zap.NewProduction()  // JSON output, sampled
        } else {
            base, _ = zap.NewDevelopment() // Console output, verbose
        }
        globalLogger = &Logger{base.Sugar()}
    })
}
```

### Key Design Decisions

| Decision                     | Rationale                                                       |
| ---------------------------- | --------------------------------------------------------------- |
| **Singleton Pattern**        | Ensures consistent logger configuration across application      |
| **SugaredLogger**            | More ergonomic API for most use cases (slight performance cost) |
| **Environment-based Config** | Different logging strategies for dev vs prod                    |
| **Sync.Once**                | Thread-safe initialization, prevents double-init                |

### Log Levels in Production

```go
// Debug - Disabled in production (sampled at 1:100)
logger.Debug("Cache hit for key", "key", productID)

// Info - Normal operation events
logger.Info("Product created", "product_id", product.ID, "user_id", userID)

// Warn - Degraded state, non-critical failures
logger.Warn("Cache write failed, continuing", "error", err, "operation", "product_cache")

// Error - Operation failures that need attention
logger.Error("Database connection lost", "error", err, "retry_count", retries)

// Fatal - Unrecoverable errors (calls os.Exit(1))
logger.Fatal("Failed to load config", "error", err)
```

### Structured Logging Best Practices

#### ❌ Bad: String Concatenation

```go
logger.Info("User " + userID + " created product " + productID)
// Issues:
// - No machine-readable fields
// - String concatenation overhead
// - Can't filter/search by user_id or product_id
```

#### ✅ Good: Structured Fields

```go
logger.Info("Product created",
    "user_id", userID,
    "product_id", productID,
    "price", product.Price,
    "discount_type", product.DiscountType,
)
// Benefits:
// - Indexable fields in log aggregators
// - Type-safe values
// - Easy filtering: user_id="123"
```

### Real Example: Cache Operations

**File**: [`internal/usecase/product_usecase.go`](internal/usecase/product_usecase.go#L47-L60)

```go
func (u *ProductUsecase) GetProductByID(ctx context.Context, id uint) (*entity.Product, error) {
    // Try cache first
    product, err := u.cache.Get(ctx, id)
    if err == nil {
        logger.Debug("Product cache hit", "product_id", id)
        return product, nil
    }

    logger.Debug("Product cache miss, fetching from DB", "product_id", id)
    product, err = u.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Async cache write
    if err := u.cache.Set(ctx, product); err != nil {
        logger.Warnf("Failed to cache product: %v", err)
    }

    return product, nil
}
```

**Key Takeaways**:

- Cache misses are **Debug** level (normal operation)
- Cache write failures are **Warn** (degraded, not critical)
- Product ID included for correlation

---

## Distributed Tracing Architecture

### What is Distributed Tracing?

Distributed tracing tracks a single request as it flows through multiple services, layers, and infrastructure components.

```
User Request → HTTP Handler → Use Case → Repository → Database
                   ↓              ↓           ↓
                 Trace 1       Trace 2     Trace 3
                   └─────────────┴───────────┴────────→ Timeline
```

### Implementation with OpenTelemetry (Recommended)

#### 1. Add Dependencies

```bash
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/sdk/trace
go get go.opentelemetry.io/otel/exporters/jaeger
go get go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin
```

#### 2. Initialize Tracer

**File**: `pkg/tracer/tracer.go` (create this)

```go
package tracer

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func InitTracer(serviceName, jaegerEndpoint string) (*trace.TracerProvider, error) {
    exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
        jaeger.WithEndpoint(jaegerEndpoint),
    ))
    if err != nil {
        return nil, err
    }

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String(serviceName),
            semconv.DeploymentEnvironmentKey.String(os.Getenv("ENV")),
        )),
    )

    otel.SetTracerProvider(tp)
    return tp, nil
}
```

#### 3. Instrument HTTP Layer

**File**: [`cmd/api/main.go`](cmd/api/main.go)

```go
import "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

func main() {
    // Initialize tracer
    tp, err := tracer.InitTracer("pos-api", "http://jaeger:14268/api/traces")
    if err != nil {
        log.Fatal(err)
    }
    defer tp.Shutdown(context.Background())

    router := gin.Default()

    // Add OpenTelemetry middleware (before other middleware)
    router.Use(otelgin.Middleware("pos-api"))

    // Your routes...
}
```

#### 4. Add Custom Spans in Use Cases

**Example**: Tracing product creation with cache operations

```go
package usecase

import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("product-usecase")

func (u *ProductUsecase) CreateProduct(ctx context.Context, product *entity.Product) error {
    // Start a span
    ctx, span := tracer.Start(ctx, "CreateProduct")
    defer span.End()

    // Add attributes
    span.SetAttributes(
        attribute.String("product.name", product.Name),
        attribute.Float64("product.price", float64(product.Price)),
        attribute.String("product.discount_type", string(product.DiscountType)),
    )

    // Database operation (child span)
    if err := u.repo.Create(ctx, product); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "database insert failed")
        return err
    }

    // Cache invalidation (sibling span)
    u.invalidateCacheAsync(ctx)

    span.SetStatus(codes.Ok, "product created successfully")
    return nil
}

func (u *ProductUsecase) invalidateCacheAsync(ctx context.Context) {
    _, span := tracer.Start(ctx, "InvalidateProductCache")
    defer span.End()

    if err := u.cache.Invalidate(ctx, "products:list"); err != nil {
        span.RecordError(err)
        logger.Warnf("Failed to invalidate cache: %v", err)
    }
}
```

#### 5. Database Instrumentation (GORM)

```go
import "go.opentelemetry.io/contrib/instrumentation/gorm.io/gorm/otelgorm"

db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
if err := db.Use(otelgorm.NewPlugin()); err != nil {
    panic(err)
}
```

#### 6. Redis Instrumentation

```go
import "github.com/go-redis/redis/extra/redisotel/v9"

rdb := redis.NewClient(&redis.Options{...})
if err := redisotel.InstrumentTracing(rdb); err != nil {
    panic(err)
}
```

### Trace Visualization in Jaeger

```
Trace ID: 7f8a9b2c3d4e5f6a
Total Duration: 142ms

┌─ HTTP POST /api/products (142ms) ──────────────────────┐
│  ├─ CreateProduct (138ms)                              │
│  │  ├─ ValidateProduct (2ms)                           │
│  │  ├─ UploadImage (45ms)                              │
│  │  │  └─ S3Upload (43ms)                              │
│  │  ├─ Database.Insert (85ms)                          │
│  │  │  └─ SQL: INSERT INTO products (82ms)             │
│  │  └─ InvalidateProductCache (4ms)                    │
│  │     └─ Redis.Del (3ms)                              │
│  └─ JSON Serialization (2ms)                           │
└────────────────────────────────────────────────────────┘
```

---

## Context Propagation Patterns

### The Problem

How do we pass request-scoped data (user ID, request ID, trace context) through all layers without polluting function signatures?

### Solution: Context.Context

**Current Implementation**: [`internal/delivery/http/handler/product_handler.go`](internal/delivery/http/handler/product_handler.go#L48-L68)

```go
func (h *ProductHandler) CreateProduct(ctx *gin.Context) {
    var productRequest request.CreateProductRequest

    // Bind request
    if err := ctx.ShouldBind(&productRequest); err != nil {
        helper.WriteAPIResponse(ctx, nil, "invalid request body", http.StatusBadRequest)
        return
    }

    // Convert to entity (passes ctx.Request.Context())
    product, err := productRequest.ToProduct(ctx.Request.Context(), h.fileStorage)
    if err != nil {
        helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusBadRequest)
        return
    }

    // Use case layer (passes context down)
    if err := h.productUsecase.CreateProduct(ctx.Request.Context(), product); err != nil {
        helper.WriteAPIResponse(ctx, nil, err.Error(), http.StatusInternalServerError)
        return
    }
}
```

### Context Values for Request Metadata

#### Adding Request ID to Context

**File**: `internal/delivery/http/middleware/request_id.go` (create this)

```go
package middleware

import (
    "context"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

type contextKey string

const (
    RequestIDKey contextKey = "request_id"
    UserIDKey    contextKey = "user_id"
    TraceIDKey   contextKey = "trace_id"
)

func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        // Add to Gin context
        c.Set(string(RequestIDKey), requestID)

        // Add to request context for propagation
        ctx := context.WithValue(c.Request.Context(), RequestIDKey, requestID)
        c.Request = c.Request.WithContext(ctx)

        // Add to response headers
        c.Header("X-Request-ID", requestID)

        c.Next()
    }
}

// Helper to extract from context
func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(RequestIDKey).(string); ok {
        return id
    }
    return ""
}
```

#### Enhanced Logger with Context

```go
func (l *Logger) WithContext(ctx context.Context) *Logger {
    fields := []interface{}{}

    if reqID := GetRequestID(ctx); reqID != "" {
        fields = append(fields, "request_id", reqID)
    }

    if userID := GetUserID(ctx); userID != "" {
        fields = append(fields, "user_id", userID)
    }

    if traceID := trace.SpanFromContext(ctx).SpanContext().TraceID(); traceID.IsValid() {
        fields = append(fields, "trace_id", traceID.String())
    }

    return &Logger{l.With(fields...)}
}

// Usage in use cases
func (u *ProductUsecase) CreateProduct(ctx context.Context, product *entity.Product) error {
    log := logger.Get().WithContext(ctx)
    log.Info("Creating product", "name", product.Name)
    // All subsequent logs will include request_id, user_id, trace_id
}
```

---

## Request Correlation & Tracking

### Current Middleware Stack

**File**: [`internal/delivery/http/middleware/logger_middleware.go`](internal/delivery/http/middleware/logger_middleware.go)

```go
func LoggerMiddleware(logger *logger.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        startTime := time.Now()
        c.Next()

        logger.Info("HTTP Request",
            "status_code", c.Writer.Status(),
            "latency", time.Since(startTime),
            "client_ip", c.ClientIP(),
            "method", c.Request.Method,
            "path", c.Request.URL.Path,
        )
    }
}
```

### Enhanced Production Version

```go
func LoggerMiddleware(logger *logger.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery

        c.Next()

        latency := time.Since(start)
        statusCode := c.Writer.Status()

        // Log level based on status code
        logFunc := logger.Info
        if statusCode >= 500 {
            logFunc = logger.Error
        } else if statusCode >= 400 {
            logFunc = logger.Warn
        }

        fields := []interface{}{
            "status", statusCode,
            "method", c.Request.Method,
            "path", path,
            "query", query,
            "ip", c.ClientIP(),
            "user_agent", c.Request.UserAgent(),
            "latency_ms", latency.Milliseconds(),
            "request_id", c.GetString("request_id"),
        }

        // Add user context if authenticated
        if userID, exists := c.Get("user_id"); exists {
            fields = append(fields, "user_id", userID)
        }

        // Add error if present
        if len(c.Errors) > 0 {
            fields = append(fields, "errors", c.Errors.String())
        }

        logFunc("HTTP request completed", fields...)
    }
}
```

### Correlation Across Services

If you had multiple services (e.g., POS API + Inventory Service + Payment Service):

```go
// Propagate trace context via HTTP headers
func (c *HTTPClient) DoRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
    // Extract span context
    span := trace.SpanFromContext(ctx)

    // Inject into HTTP headers (W3C Trace Context standard)
    otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

    // Also add custom headers
    req.Header.Set("X-Request-ID", GetRequestID(ctx))
    req.Header.Set("X-User-ID", GetUserID(ctx))

    return c.client.Do(req)
}
```

---

## Performance Monitoring

### Critical Metrics to Track

1. **Request Rate**: Requests per second (RPS)
2. **Latency Percentiles**: P50, P95, P99
3. **Error Rate**: 4xx and 5xx responses
4. **Database Query Time**: Query duration distribution
5. **Cache Hit Ratio**: Cache effectiveness

### Prometheus Integration

#### 1. Add Prometheus Client

```bash
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
```

#### 2. Define Metrics

**File**: `pkg/metrics/metrics.go` (create this)

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // HTTP Metrics
    HTTPRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request latency in seconds",
            Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"method", "path", "status"},
    )

    HTTPRequestTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    // Database Metrics
    DBQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "db_query_duration_seconds",
            Help:    "Database query latency",
            Buckets: prometheus.ExponentialBuckets(0.001, 2, 12),
        },
        []string{"operation", "table"},
    )

    // Cache Metrics
    CacheHitTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cache_hits_total",
            Help: "Total cache hits",
        },
        []string{"cache_name"},
    )

    CacheMissTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cache_misses_total",
            Help: "Total cache misses",
        },
        []string{"cache_name"},
    )
)
```

#### 3. Instrument Code

**HTTP Middleware**:

```go
func PrometheusMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        c.Next()

        duration := time.Since(start).Seconds()
        status := strconv.Itoa(c.Writer.Status())
        path := c.FullPath() // Template path, not actual (avoids high cardinality)

        metrics.HTTPRequestDuration.WithLabelValues(
            c.Request.Method,
            path,
            status,
        ).Observe(duration)

        metrics.HTTPRequestTotal.WithLabelValues(
            c.Request.Method,
            path,
            status,
        ).Inc()
    }
}
```

**Cache Layer** (update existing):

```go
func (c *ProductCache) Get(ctx context.Context, id uint) (*entity.Product, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start).Seconds()
        metrics.DBQueryDuration.WithLabelValues("get", "product_cache").Observe(duration)
    }()

    result, err := c.redis.Get(ctx, c.key(id)).Result()
    if err == redis.Nil {
        metrics.CacheMissTotal.WithLabelValues("product").Inc()
        return nil, ErrCacheMiss
    }
    if err != nil {
        return nil, err
    }

    metrics.CacheHitTotal.WithLabelValues("product").Inc()

    var product entity.Product
    if err := json.Unmarshal([]byte(result), &product); err != nil {
        return nil, err
    }

    return &product, nil
}
```

#### 4. Expose Metrics Endpoint

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

func main() {
    router := gin.Default()

    // Metrics endpoint (should be on separate port in production)
    router.GET("/metrics", gin.WrapH(promhttp.Handler()))

    // ... rest of setup
}
```

### Grafana Dashboard Example

```yaml
# PromQL Queries for Dashboard

# Request Rate
rate(http_requests_total[5m])

# P95 Latency
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Error Rate
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))

# Cache Hit Ratio
sum(rate(cache_hits_total[5m])) / (sum(rate(cache_hits_total[5m])) + sum(rate(cache_misses_total[5m])))

# Database Query P99
histogram_quantile(0.99, rate(db_query_duration_seconds_bucket[5m]))
```

---

## Error Tracking & Aggregation

### Sentry Integration

#### 1. Install Sentry SDK

```bash
go get github.com/getsentry/sentry-go
go get github.com/getsentry/sentry-go/gin
```

#### 2. Initialize Sentry

```go
import (
    "github.com/getsentry/sentry-go"
    sentrygin "github.com/getsentry/sentry-go/gin"
)

func main() {
    if err := sentry.Init(sentry.ClientOptions{
        Dsn:              os.Getenv("SENTRY_DSN"),
        Environment:      os.Getenv("ENV"),
        Release:          os.Getenv("APP_VERSION"),
        TracesSampleRate: 0.2, // 20% of transactions
        AttachStacktrace: true,
    }); err != nil {
        log.Fatal(err)
    }
    defer sentry.Flush(2 * time.Second)

    router := gin.Default()
    router.Use(sentrygin.New(sentrygin.Options{
        Repanic: true,
    }))
}
```

#### 3. Capture Custom Errors

```go
func (u *ProductUsecase) CreateProduct(ctx context.Context, product *entity.Product) error {
    if err := u.repo.Create(ctx, product); err != nil {
        // Capture error with context
        sentry.WithScope(func(scope *sentry.Scope) {
            scope.SetContext("product", map[string]interface{}{
                "name":  product.Name,
                "price": product.Price,
            })
            scope.SetUser(sentry.User{
                ID: GetUserID(ctx),
            })
            scope.SetTag("operation", "product_create")
            sentry.CaptureException(err)
        })

        return err
    }
    return nil
}
```

---

## Production Best Practices

### 1. Log Sampling in High-Traffic Scenarios

```go
// Sample Debug logs in production (1 in 100)
if env == "production" {
    cfg := zap.NewProductionConfig()
    cfg.Sampling = &zap.SamplingConfig{
        Initial:    100,
        Thereafter: 100,
    }
    base, _ = cfg.Build()
}
```

### 2. Avoid High Cardinality in Metrics

```go
// ❌ BAD: Using actual user ID (unbounded cardinality)
metrics.UserRequests.WithLabelValues(userID).Inc()

// ✅ GOOD: Using user tier/role
metrics.UserRequests.WithLabelValues(user.Tier).Inc()
```

### 3. Structured Error Messages

```go
// ❌ BAD
return fmt.Errorf("failed to create product")

// ✅ GOOD
return fmt.Errorf("failed to create product: %w (user_id=%s, product_name=%s)",
    err, userID, product.Name)
```

### 4. Context Timeouts

```go
func (u *ProductUsecase) CreateProduct(ctx context.Context, product *entity.Product) error {
    // Set timeout for external operations
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    if err := u.repo.Create(ctx, product); err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            logger.Error("Database timeout during product creation",
                "timeout", "5s",
                "product_name", product.Name,
            )
        }
        return err
    }
    return nil
}
```

### 5. Panic Recovery with Reporting

```go
func RecoveryMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                // Log the panic
                logger.Error("Panic recovered",
                    "error", err,
                    "stack", string(debug.Stack()),
                    "path", c.Request.URL.Path,
                )

                // Report to Sentry
                sentry.CurrentHub().Recover(err)
                sentry.Flush(2 * time.Second)

                // Return error response
                c.JSON(500, gin.H{"error": "Internal server error"})
                c.Abort()
            }
        }()
        c.Next()
    }
}
```

### 6. Health Check with Dependencies

```go
func (h *HealthHandler) Check(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
    defer cancel()

    health := map[string]string{
        "status": "ok",
    }

    // Check database
    if err := h.db.PingContext(ctx); err != nil {
        health["database"] = "unhealthy"
        health["status"] = "degraded"
        logger.Error("Health check: database unreachable", "error", err)
    } else {
        health["database"] = "healthy"
    }

    // Check Redis
    if err := h.redis.Ping(ctx).Err(); err != nil {
        health["cache"] = "unhealthy"
        health["status"] = "degraded"
    } else {
        health["cache"] = "healthy"
    }

    statusCode := 200
    if health["status"] != "ok" {
        statusCode = 503
    }

    c.JSON(statusCode, health)
}
```

---

## Advanced: Custom Trace Exporter

For scenarios where you need custom trace storage or analysis:

```go
type CustomSpanExporter struct {
    logger *logger.Logger
    db     *gorm.DB
}

type StoredSpan struct {
    TraceID    string
    SpanID     string
    ParentID   string
    Name       string
    StartTime  time.Time
    EndTime    time.Time
    Duration   int64
    Attributes datatypes.JSON
    Status     string
}

func (e *CustomSpanExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
    for _, span := range spans {
        storedSpan := &StoredSpan{
            TraceID:   span.SpanContext().TraceID().String(),
            SpanID:    span.SpanContext().SpanID().String(),
            Name:      span.Name(),
            StartTime: span.StartTime(),
            EndTime:   span.EndTime(),
            Duration:  span.EndTime().Sub(span.StartTime()).Milliseconds(),
            Status:    span.Status().Code.String(),
        }

        // Store in database for custom analytics
        if err := e.db.Create(storedSpan).Error; err != nil {
            e.logger.Error("Failed to store span", "error", err)
        }
    }
    return nil
}
```

---

## Checklist for Production Readiness

- [ ] **Structured logging** with request IDs in all layers
- [ ] **OpenTelemetry** instrumentation (HTTP, DB, Cache, External APIs)
- [ ] **Prometheus metrics** for SLIs (latency, error rate, throughput)
- [ ] **Sentry** for error aggregation and alerting
- [ ] **Context propagation** throughout application
- [ ] **Log sampling** configured for production
- [ ] **Health checks** for all dependencies
- [ ] **Graceful shutdown** with trace/log flushing
- [ ] **Dashboards** in Grafana for key metrics
- [ ] **Alerts** configured for critical thresholds

---

## References & Further Reading

- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/instrumentation/go/)
- [Uber's Guide to Logging](https://github.com/uber-go/zap/blob/master/FAQ.md)
- [Google SRE Book - Monitoring Distributed Systems](https://sre.google/sre-book/monitoring-distributed-systems/)
- [The Three Pillars of Observability](https://www.oreilly.com/library/view/distributed-systems-observability/9781492033431/ch04.html)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/naming/)

---

**Last Updated**: January 2026  
**Maintainer**: Backend Engineering Team  
**Related Docs**: [REDIS.md](REDIS.md), [README.md](README.md)
