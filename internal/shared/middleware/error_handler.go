package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			logger.Error("unhandled error in request pipeline",
				"error", err.Err,
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
			)

			if !c.Writer.Written() {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "An unexpected error occurred",
					},
				})
			}
		}
	}
}
