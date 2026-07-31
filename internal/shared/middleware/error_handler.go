package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"sipon-be/internal/shared/httperror"
)

func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return httperror.Middleware(logger)
}
