package helper

import "github.com/gin-gonic/gin"

// APIResponse standardizes successful responses.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// ErrorBody holds structured error details.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteSuccess writes a success response.
func WriteSuccess(ctx *gin.Context, status int, data interface{}, meta interface{}) {
	ctx.JSON(status, APIResponse{Success: true, Data: data, Meta: meta})
}

// WriteError writes an error response.
func WriteError(ctx *gin.Context, status int, code string, message string) {
	ctx.JSON(status, APIResponse{Success: false, Error: &ErrorBody{Code: code, Message: message}})
}
