package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	ports "sipon-be/internal/modules/identity/application/ports"
	"sipon-be/internal/shared/config"
	"sipon-be/internal/shared/respond"
)

func RateLimitByIP(limiter ports.RateLimiter, cfg config.RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enabled {
			c.Next()
			return
		}

		key := "ip:" + c.ClientIP()
		result, err := limiter.Allow(c.Request.Context(), key, cfg.IPLimit, time.Duration(cfg.IPWindowSeconds)*time.Second)
		if err != nil {
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", formatInt(cfg.IPLimit))
		c.Header("X-RateLimit-Remaining", formatInt(result.Remaining))
		c.Header("X-RateLimit-Reset", formatInt(int(result.ResetAt.Unix())))

		if !result.Allowed {
			respond.AbortWithError(c, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests, please try again later")
			return
		}

		c.Next()
	}
}

func RateLimitByUser(limiter ports.RateLimiter, cfg config.RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enabled {
			c.Next()
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		uid, ok := userID.(string)
		if !ok {
			c.Next()
			return
		}

		key := "user:" + uid
		result, err := limiter.Allow(c.Request.Context(), key, cfg.UserLimit, time.Duration(cfg.UserWindowSeconds)*time.Second)
		if err != nil {
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", formatInt(cfg.UserLimit))
		c.Header("X-RateLimit-Remaining", formatInt(result.Remaining))
		c.Header("X-RateLimit-Reset", formatInt(int(result.ResetAt.Unix())))

		if !result.Allowed {
			respond.AbortWithError(c, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests, please try again later")
			return
		}

		c.Next()
	}
}

func RateLimitByAuth(limiter ports.RateLimiter, cfg config.RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enabled {
			c.Next()
			return
		}

		key := "auth:" + c.ClientIP()
		result, err := limiter.Allow(c.Request.Context(), key, cfg.AuthLimit, time.Duration(cfg.AuthWindowSeconds)*time.Second)
		if err != nil {
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", formatInt(cfg.AuthLimit))
		c.Header("X-RateLimit-Remaining", formatInt(result.Remaining))
		c.Header("X-RateLimit-Reset", formatInt(int(result.ResetAt.Unix())))

		if !result.Allowed {
			respond.AbortWithError(c, http.StatusTooManyRequests, "RATE_LIMITED", "Too many authentication attempts, please try again later")
			return
		}

		c.Next()
	}
}

func formatInt(v int) string {
	return string(appendInt(nil, v))
}

func appendInt(buf []byte, v int) []byte {
	if v < 10 {
		return append(buf, byte('0'+v))
	}
	if v < 100 {
		return append(buf, byte('0'+v/10), byte('0'+v%10))
	}
	buf = appendInt(buf, v/10)
	return append(buf, byte('0'+v%10))
}
