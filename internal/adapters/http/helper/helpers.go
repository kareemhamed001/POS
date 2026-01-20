package helper

import (
	"github.com/gin-gonic/gin"
)

func WriteAPIResponse(ctx *gin.Context, data interface{}, message string, statusCode int) map[string]interface{} {
	if statusCode >= 400 {
		WriteError(ctx, statusCode, "ERROR", message)
		return nil
	}
	WriteSuccess(ctx, statusCode, data, nil)
	return nil
}
