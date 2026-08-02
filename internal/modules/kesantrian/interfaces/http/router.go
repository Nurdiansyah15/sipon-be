package http

import (
	"github.com/gin-gonic/gin"

	"sipon-be/internal/shared/middleware"
)

// RegisterRoutes takes the two already-built auth/principal middleware
// funcs from identity (via *identity.Module's AuthMiddleware/
// PrincipalMiddleware — NOT identity.Contract) instead of owning a
// MiddlewareBuilder of its own: kesantrian has no TokenGenerator/
// SessionStore/PrincipalBuilder to wrap, so a builder struct here would
// just carry 2 fields for no behavioral gain. See
// docs/architecture/module-boundaries.md and the kesantrian port plan §0.6.
func RegisterRoutes(router *gin.RouterGroup, h *SantriHandler, jwtAuth, principalLoad gin.HandlerFunc) {
	santri := router.Group("/api/v1/web/santri")
	santri.Use(jwtAuth, principalLoad)
	{
		santri.GET("/profile", h.GetSantri)
		santri.PUT("/profile", h.UpdateSantri)
		santri.POST("/request", h.RequestSantri)

		santri.POST("/dokumen/presign", h.DokumenPresign)
		santri.POST("/dokumen/confirm", h.DokumenConfirm)
		santri.GET("/dokumen", h.DokumenList)
		santri.GET("/dokumen/:id/access", h.DokumenAccess)
		santri.DELETE("/dokumen/:id", h.DokumenDelete)

		admin := santri.Group("/admin")
		admin.Use(middleware.RequirePermission("manage_santri"))
		{
			admin.GET("", h.ListSantri)
			admin.POST("", h.CreateSantri)
			admin.POST("/import", h.ImportSantri)
			admin.GET("/import/template", h.DownloadImportTemplate)
			admin.GET("/requests", h.ListSantriRequests)
			admin.POST("/requests/:id/approve", h.ApproveSantriRequest)
			admin.POST("/requests/:id/reject", h.RejectSantriRequest)
			admin.GET("/:id/dokumen", h.AdminDokumenList)
			admin.POST("/verify/:id", h.DokumenVerify)
			admin.POST("/reject/:id", h.DokumenReject)
		}
	}
}
