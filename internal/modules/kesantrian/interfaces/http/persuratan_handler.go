package http

import (
	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/kesantrian/application/command"
	"sipon-be/internal/modules/kesantrian/application/dto"
	"sipon-be/internal/modules/kesantrian/application/query"
	"sipon-be/internal/shared/httperror"
	"sipon-be/internal/shared/middleware"
	"sipon-be/internal/shared/respond"
)

type PersuratanHandler struct {
	createTipeSuratUC *command.CreateTipeSuratUseCase
	updateTipeSuratUC *command.UpdateTipeSuratUseCase
	deleteTipeSuratUC *command.DeleteTipeSuratUseCase
	listTipeSuratUC   *query.ListTipeSuratUseCase
	getTipeSuratUC    *query.GetTipeSuratUseCase

	createSuratUC        *command.CreateSuratUseCase
	deleteSuratUC        *command.DeleteSuratUseCase
	addSuratDokumenUC    *command.AddSuratDokumenUseCase
	removeSuratDokumenUC *command.RemoveSuratDokumenUseCase
	listSuratUC          *query.ListSuratUseCase
	getSuratUC           *query.GetSuratUseCase
	getSuratDownloadUC   *query.GetSuratDownloadUseCase
}

func NewPersuratanHandler(
	createTipeSuratUC *command.CreateTipeSuratUseCase,
	updateTipeSuratUC *command.UpdateTipeSuratUseCase,
	deleteTipeSuratUC *command.DeleteTipeSuratUseCase,
	listTipeSuratUC *query.ListTipeSuratUseCase,
	getTipeSuratUC *query.GetTipeSuratUseCase,
	createSuratUC *command.CreateSuratUseCase,
	deleteSuratUC *command.DeleteSuratUseCase,
	addSuratDokumenUC *command.AddSuratDokumenUseCase,
	removeSuratDokumenUC *command.RemoveSuratDokumenUseCase,
	listSuratUC *query.ListSuratUseCase,
	getSuratUC *query.GetSuratUseCase,
	getSuratDownloadUC *query.GetSuratDownloadUseCase,
) *PersuratanHandler {
	return &PersuratanHandler{
		createTipeSuratUC:    createTipeSuratUC,
		updateTipeSuratUC:    updateTipeSuratUC,
		deleteTipeSuratUC:    deleteTipeSuratUC,
		listTipeSuratUC:      listTipeSuratUC,
		getTipeSuratUC:       getTipeSuratUC,
		createSuratUC:        createSuratUC,
		deleteSuratUC:        deleteSuratUC,
		addSuratDokumenUC:    addSuratDokumenUC,
		removeSuratDokumenUC: removeSuratDokumenUC,
		listSuratUC:          listSuratUC,
		getSuratUC:           getSuratUC,
		getSuratDownloadUC:   getSuratDownloadUC,
	}
}

func (h *PersuratanHandler) ListTipeSurat(c *gin.Context) {
	var q dto.ListTipeSuratQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, err := h.listTipeSuratUC.Execute(c.Request.Context(), q)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "daftar tipe surat berhasil diambil", items)
}

func (h *PersuratanHandler) GetTipeSurat(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.getTipeSuratUC.Execute(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "tipe surat berhasil diambil", resp)
}

func (h *PersuratanHandler) CreateTipeSurat(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.CreateTipeSuratRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createTipeSuratUC.Execute(c.Request.Context(), &userID, req.Nama, req.Kode)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "tipe surat berhasil dibuat", resp)
}

func (h *PersuratanHandler) UpdateTipeSurat(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateTipeSuratRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	if err := h.updateTipeSuratUC.Execute(c.Request.Context(), id, req.Nama, req.Kode); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "tipe surat berhasil diperbarui", nil)
}

func (h *PersuratanHandler) DeleteTipeSurat(c *gin.Context) {
	id := c.Param("id")
	if err := h.deleteTipeSuratUC.Execute(c.Request.Context(), id); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "tipe surat berhasil dihapus", nil)
}

func (h *PersuratanHandler) ListSurat(c *gin.Context) {
	var q dto.ListSuratQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listSuratUC.Execute(c.Request.Context(), q)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar surat berhasil diambil", items, meta)
}

func (h *PersuratanHandler) GetSurat(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.getSuratUC.Execute(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "surat berhasil diambil", resp)
}

func (h *PersuratanHandler) CreateSurat(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.CreateSuratRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createSuratUC.Execute(c.Request.Context(), userID, req.TipeSuratID, req.Keterangan, req.Tanggal, req.DokumenAsetIDs)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	detail, err := h.getSuratUC.Execute(c.Request.Context(), resp.ID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "surat berhasil dibuat", detail)
}

func (h *PersuratanHandler) DeleteSurat(c *gin.Context) {
	id := c.Param("id")
	if err := h.deleteSuratUC.Execute(c.Request.Context(), id); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "surat berhasil dihapus", nil)
}

func (h *PersuratanHandler) AddSuratDokumen(c *gin.Context) {
	suratID := c.Param("id")
	var req dto.AddDokumenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	if err := h.addSuratDokumenUC.Execute(c.Request.Context(), suratID, req.DokumenAsetID); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "dokumen berhasil ditautkan", dto.TautDokumenResponse{
		SuratID:       suratID,
		DokumenAsetID: req.DokumenAsetID,
	})
}

func (h *PersuratanHandler) RemoveSuratDokumen(c *gin.Context) {
	suratID := c.Param("id")
	dokumenAsetID := c.Param("dokumenAsetId")
	if err := h.removeSuratDokumenUC.Execute(c.Request.Context(), suratID, dokumenAsetID); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "dokumen berhasil dilepas", nil)
}

func (h *PersuratanHandler) GetSuratDownload(c *gin.Context) {
	suratID := c.Param("id")
	dokumenAsetID := c.Param("dokumenAsetId")
	resp, err := h.getSuratDownloadUC.Execute(c.Request.Context(), suratID, dokumenAsetID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "url download berhasil dibuat", resp)
}
