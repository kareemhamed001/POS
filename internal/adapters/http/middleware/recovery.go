package middleware

import (
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kareemhamed001/POS/pkg/logger"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if recoveredValue := recover(); recoveredValue != nil {
				isBrokenPipe := false
				if errAsError, ok := recoveredValue.(error); ok {
					if netOpErr, ok := errAsError.(*net.OpError); ok {
						if syscallErr, ok := netOpErr.Err.(*os.SyscallError); ok {
							syscallMsg := strings.ToLower(syscallErr.Error())
							if strings.Contains(syscallMsg, "broken pipe") || strings.Contains(syscallMsg, "connection reset by peer") {
								isBrokenPipe = true
							}
						}
					}
				}

				logger.Errorf("Panic recovered: %v\n%s", recoveredValue, debug.Stack())
				if isBrokenPipe {
					ctx.Abort()
					return
				}
				if !ctx.Writer.Written() {
					ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
						"error": "Internal Server Error",
					})
				} else {
					ctx.Abort()
				}
			}
		}()

		ctx.Next()
	}
}
