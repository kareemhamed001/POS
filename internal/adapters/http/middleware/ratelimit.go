package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kareemhamed001/POS/internal/core/ports"
)

// RateLimit creates a rate limiting middleware
// maxRequests: maximum number of requests allowed within the window
// window: time window for rate limiting
func RateLimit(rateLimitCache ports.RateLimitCache, maxRequests int64, window time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Use IP address as the rate limit key
		ip := ctx.ClientIP()
		key := fmt.Sprintf("ip:%s:%s", ip, ctx.Request.URL.Path)

		// Increment counter
		count, err := rateLimitCache.IncrementCounter(ctx, key, window)
		if err != nil {
			// If Redis fails, allow the request (graceful degradation)
			ctx.Next()
			return
		}

		// Check if limit exceeded
		if count > maxRequests {
			ctx.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "rate limit exceeded, please try again later",
			})
			ctx.Abort()
			return
		}

		// Add rate limit headers
		ctx.Header("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
		ctx.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", maxRequests-count))
		ctx.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

		ctx.Next()
	}
}
