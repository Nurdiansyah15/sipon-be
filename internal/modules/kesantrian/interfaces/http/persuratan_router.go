package http

import (
	"github.com/gin-gonic/gin"

	"sipon-be/internal/shared/middleware"
)

func RegisterPersuratanRoutes(admin *gin.RouterGroup, h *PersuratanHandler) {
	persuratan := admin.Group("/persuratan")
	persuratan.Use(middleware.RequirePermission("manage_persuratan"))
	{
		tipeSurat := persuratan.Group("/tipe-surat")
		{
			tipeSurat.GET("", h.ListTipeSurat)
			tipeSurat.GET("/:id", h.GetTipeSurat)
			tipeSurat.POST("", h.CreateTipeSurat)
			tipeSurat.PUT("/:id", h.UpdateTipeSurat)
			tipeSurat.DELETE("/:id", h.DeleteTipeSurat)
		}

		surat := persuratan.Group("/surat")
		{
			surat.GET("", h.ListSurat)
			surat.POST("", h.CreateSurat)
			surat.GET("/:id", h.GetSurat)
			surat.DELETE("/:id", h.DeleteSurat)
			surat.POST("/:id/dokumen", h.AddSuratDokumen)
			surat.DELETE("/:id/dokumen/:dokumenAsetId", h.RemoveSuratDokumen)
			surat.GET("/:id/dokumen/:dokumenAsetId/download", h.GetSuratDownload)
		}
	}
}
