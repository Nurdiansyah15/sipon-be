package http

import (
	"github.com/gin-gonic/gin"

	"sipon-be/internal/shared/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, h *DokumenAsetHandler, jwtAuth, principalLoad gin.HandlerFunc) {
	da := router.Group("/api/v1/web/dokumen-aset")
	{
		da.GET("", h.List)
		da.GET("/:id", optionalAuth(jwtAuth), h.Get)
		da.GET("/:id/download", optionalAuth(jwtAuth), h.Download)

		admin := da.Group("/admin")
		admin.Use(jwtAuth, principalLoad, middleware.RequirePermission("manage_dokumen"))
		{
			admin.GET("", h.List)
			admin.POST("", h.Presign)
			admin.POST("/confirm", h.Confirm)
			admin.PUT("/:id", h.Update)
			admin.DELETE("/:id", h.Delete)
		}
	}
}

func optionalAuth(jwtAuth gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}
		jwtAuth(c)
	}
}
