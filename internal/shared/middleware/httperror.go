package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HTTPError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewHTTPError(statusCode int, code, message string) *HTTPError {
	_ = statusCode
	return &HTTPError{
		Code:    code,
		Message: message,
	}
}

func AbortWithError(c *gin.Context, statusCode int, code, message string) {
	c.AbortWithStatusJSON(statusCode, gin.H{
		"error": gin.H{
			"code":       code,
			"message":    message,
			"request_id": c.GetString("request_id"),
		},
	})
}

func AbortWithValidationError(c *gin.Context, message string) {
	AbortWithError(c, http.StatusBadRequest, "VALIDATION_ERROR", message)
}

func AbortWithUnauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

func AbortWithForbidden(c *gin.Context, message string) {
	if message == "" {
		message = "Forbidden"
	}
	AbortWithError(c, http.StatusForbidden, "FORBIDDEN", message)
}

func AbortWithNotFound(c *gin.Context, message string) {
	if message == "" {
		message = "Resource not found"
	}
	AbortWithError(c, http.StatusNotFound, "NOT_FOUND", message)
}

func AbortWithConflict(c *gin.Context, message string) {
	AbortWithError(c, http.StatusConflict, "CONFLICT", message)
}

func AbortWithInternalError(c *gin.Context, message string) {
	if message == "" {
		message = "Internal server error"
	}
	AbortWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

func AbortWithTooManyRequests(c *gin.Context, message string) {
	if message == "" {
		message = "Too many requests"
	}
	AbortWithError(c, http.StatusTooManyRequests, "RATE_LIMITED", message)
}

func Errorf(format string, args ...interface{}) *HTTPError {
	return &HTTPError{
		Code:    "ERROR",
		Message: fmt.Sprintf(format, args...),
	}
}
