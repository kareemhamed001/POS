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
			"/n status_code", statusCode,
			"/n latency", latency,
			"/n client_ip", clientIP,
			"/n method", method,
			"/n path", path,
		)
	}
}
