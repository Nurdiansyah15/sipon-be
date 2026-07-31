package httperror

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

func Middleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.ErrorContext(c.Request.Context(), "panic recovered",
					slog.Any("recovered", rec),
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
				)
				Handle(c, internalError("internal server error"))
			}
		}()

		c.Next()

		if len(c.Errors) > 0 && !c.Writer.Written() {
			Handle(c, c.Errors.Last().Err)
		}
	}
}
