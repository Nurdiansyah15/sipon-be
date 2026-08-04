package http

import (
	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/article/application/command"
	"sipon-be/internal/modules/article/application/dto"
	"sipon-be/internal/modules/article/application/query"
	"sipon-be/internal/shared/httperror"
	"sipon-be/internal/shared/middleware"
	"sipon-be/internal/shared/respond"
)

type SourceHandler struct {
	listSourcesUC            *query.ListSourcesUseCase
	createSourceUC           *command.CreateSourceUseCase
	updateSourceUC           *command.UpdateSourceUseCase
	deleteSourceUC           *command.DeleteSourceUseCase
	createSourceCategoryUC   *command.CreateSourceCategoryUseCase
	updateSourceCategoryUC   *command.UpdateSourceCategoryUseCase
	deleteSourceCategoryUC   *command.DeleteSourceCategoryUseCase
	triggerScrapeAllUC       *command.TriggerScrapeAllUseCase
}

func NewSourceHandler(
	listSourcesUC *query.ListSourcesUseCase,
	createSourceUC *command.CreateSourceUseCase,
	updateSourceUC *command.UpdateSourceUseCase,
	deleteSourceUC *command.DeleteSourceUseCase,
	createSourceCategoryUC *command.CreateSourceCategoryUseCase,
	updateSourceCategoryUC *command.UpdateSourceCategoryUseCase,
	deleteSourceCategoryUC *command.DeleteSourceCategoryUseCase,
	triggerScrapeAllUC *command.TriggerScrapeAllUseCase,
) *SourceHandler {
	return &SourceHandler{
		listSourcesUC:          listSourcesUC,
		createSourceUC:         createSourceUC,
		updateSourceUC:         updateSourceUC,
		deleteSourceUC:         deleteSourceUC,
		createSourceCategoryUC: createSourceCategoryUC,
		updateSourceCategoryUC: updateSourceCategoryUC,
		deleteSourceCategoryUC: deleteSourceCategoryUC,
		triggerScrapeAllUC:     triggerScrapeAllUC,
	}
}

func (h *SourceHandler) ListSources(c *gin.Context) {
	items, err := h.listSourcesUC.Execute(c.Request.Context())
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "daftar sumber artikel berhasil diambil", items)
}

func (h *SourceHandler) CreateSource(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.CreateSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createSourceUC.Execute(c.Request.Context(), req, userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "sumber artikel berhasil dibuat", resp)
}

func (h *SourceHandler) UpdateSource(c *gin.Context) {
	userID := middleware.GetUserID(c)
	sourceID := c.Param("source_id")
	var req dto.UpdateSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updateSourceUC.Execute(c.Request.Context(), sourceID, userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "sumber artikel berhasil diperbarui", resp)
}

func (h *SourceHandler) DeleteSource(c *gin.Context) {
	sourceID := c.Param("source_id")
	if err := h.deleteSourceUC.Execute(c.Request.Context(), sourceID); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "sumber artikel berhasil dihapus", nil)
}

func (h *SourceHandler) CreateSourceCategory(c *gin.Context) {
	sourceID := c.Param("source_id")
	var req dto.CreateSourceCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createSourceCategoryUC.Execute(c.Request.Context(), sourceID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "kategori sumber berhasil dibuat", resp)
}

func (h *SourceHandler) UpdateSourceCategory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	categoryID := c.Param("category_id")
	var req dto.UpdateSourceCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updateSourceCategoryUC.Execute(c.Request.Context(), categoryID, userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "kategori sumber berhasil diperbarui", resp)
}

func (h *SourceHandler) DeleteSourceCategory(c *gin.Context) {
	categoryID := c.Param("category_id")
	if err := h.deleteSourceCategoryUC.Execute(c.Request.Context(), categoryID); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "kategori sumber berhasil dihapus", nil)
}

func (h *SourceHandler) ScrapeAllNow(c *gin.Context) {
	sourceID := c.Param("source_id")
	resp, err := h.triggerScrapeAllUC.Execute(c.Request.Context(), sourceID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "scrape selesai", resp)
}
