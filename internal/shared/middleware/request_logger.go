package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		requestID, _ := c.Get("request_id")
		rid, _ := requestID.(string)

		attrs := []slog.Attr{
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("duration", duration),
			slog.String("request_id", rid),
			slog.String("client_ip", c.ClientIP()),
		}

		switch {
		case status >= 500:
			logger.LogAttrs(c.Request.Context(), slog.LevelError, "request completed", attrs...)
		case status >= 400:
			logger.LogAttrs(c.Request.Context(), slog.LevelWarn, "request completed", attrs...)
		default:
			logger.LogAttrs(c.Request.Context(), slog.LevelInfo, "request completed", attrs...)
		}
	}
}
