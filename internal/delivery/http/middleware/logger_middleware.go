package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kareemhamed001/POS/pkg/logger"
)

func LoggerMiddleware(logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		latency := time.Since(startTime)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		logger.Info("HTTP Request",
		zap.Int(
			"status_code", statusCode
		),
		zap.Duration(
			"latency", latency,
		),
		zap.String(
			"client_ip", clientIP,
		),
		zap.String(
			"method", method,
		),
		zap.String(
			"path", path,
		),
	)
}
}
