package http

import (
	"github.com/gin-gonic/gin"

	"sipon-be/internal/shared/middleware"
)

// RegisterRoutes mendaftarkan route modul fingerprint. Route sandbox hanya
// didaftarkan ketika sandboxEnabled true — di production endpoint ini tidak
// ada sama sekali, bukan cuma disembunyikan di permission.
func RegisterRoutes(router *gin.RouterGroup, h *FingerprintHandler, jwtAuth, principalLoad gin.HandlerFunc, sandboxEnabled bool) {
	fp := router.Group("/api/v1/web/fingerprint")

	admin := fp.Group("/")
	admin.Use(jwtAuth, principalLoad, middleware.RequirePermission("manage_akademik"))
	{
		admin.GET("/scans", h.ListScans)
	}

	if sandboxEnabled {
		sandbox := fp.Group("/sandbox")
		sandbox.Use(jwtAuth, principalLoad, middleware.RequirePermission("manage_akademik"))
		{
			sandbox.POST("/scan", h.SimulateScan)
		}
	}
}
