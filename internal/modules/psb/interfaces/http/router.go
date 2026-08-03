package http

import (
	"github.com/gin-gonic/gin"

	"sipon-be/internal/shared/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, h *PsbHandler, jwtAuth, principalLoad gin.HandlerFunc) {
	psb := router.Group("/api/v1/web/psb")
	psb.Use(jwtAuth, principalLoad)
	{
		psb.GET("/setting/active", h.GetActiveSetting)

		psb.GET("/pendaftaran", h.GetPendaftaran)
		psb.PUT("/pendaftaran", h.UpsertFormulir)
		psb.POST("/pendaftaran/submit", h.SubmitPendaftaran)
		psb.GET("/pendaftaran/riwayat", h.GetRiwayat)

		psb.POST("/dokumen/presign", h.DokumenPresign)
		psb.POST("/dokumen/confirm", h.DokumenConfirm)
		psb.GET("/dokumen", h.DokumenList)
		psb.DELETE("/dokumen/:id", h.DokumenDelete)

		psb.POST("/daftar-ulang/submit", h.SubmitDaftarUlang)

		admin := psb.Group("/admin")
		admin.Use(middleware.RequirePermission("manage_psb"))
		{
			admin.GET("/pendaftaran", h.AdminListPendaftaran)
			admin.GET("/pendaftaran/:id", h.AdminGetPendaftaran)
			admin.GET("/pendaftaran/:id/riwayat", h.AdminGetRiwayat)
			admin.GET("/pendaftaran/:id/dokumen", h.AdminDokumenList)
			admin.POST("/pendaftaran/:id/dokumen/:dokumenId/verify", h.AdminDokumenVerify)
			admin.POST("/pendaftaran/:id/dokumen/:dokumenId/reject", h.AdminDokumenReject)
			admin.POST("/pendaftaran/:id/request-revision", h.AdminRequestRevision)
			admin.POST("/pendaftaran/:id/reject", h.AdminReject)
			admin.POST("/pendaftaran/:id/accept", h.AdminAccept)
			admin.POST("/pendaftaran/:id/mark-not-reregistered", h.AdminMarkNotReregistered)
			admin.POST("/pendaftaran/:id/request-revision-daftar-ulang", h.AdminRequestRevisionDaftarUlang)
			admin.POST("/pendaftaran/:id/generate-nis", h.AdminGenerateNIS)

			settings := admin.Group("/settings")
			settings.Use(middleware.RequirePermission("manage_psb_settings"))
			{
				settings.GET("", h.ListSettings)
				settings.POST("", h.CreateSetting)
				settings.PUT("/:id", h.UpdateSetting)
				settings.POST("/:id/purge", h.PurgePeriod)
			}
		}
	}
}
