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

		defer func() {
			logger.Info("HTTP Request",
				"status_code", statusCode,
				"latency", latency,
				"client_ip", clientIP,
				"method", method,
				"path", path,
			)
		}()

	}

}
