package http

import (
	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/dokumen_aset/application/command"
	"sipon-be/internal/modules/dokumen_aset/application/dto"
	"sipon-be/internal/modules/dokumen_aset/application/query"
	"sipon-be/internal/shared/httperror"
	"sipon-be/internal/shared/middleware"
	"sipon-be/internal/shared/respond"
)

type DokumenAsetHandler struct {
	presignUC  *command.CreateDokumenAsetPresignUseCase
	confirmUC  *command.CreateDokumenAsetConfirmUseCase
	updateUC   *command.UpdateDokumenAsetUseCase
	deleteUC   *command.DeleteDokumenAsetUseCase
	listUC     *query.ListDokumenAsetUseCase
	getUC      *query.GetDokumenAsetUseCase
	downloadUC *query.DownloadDokumenAsetUseCase
}

func NewDokumenAsetHandler(
	presignUC *command.CreateDokumenAsetPresignUseCase,
	confirmUC *command.CreateDokumenAsetConfirmUseCase,
	updateUC *command.UpdateDokumenAsetUseCase,
	deleteUC *command.DeleteDokumenAsetUseCase,
	listUC *query.ListDokumenAsetUseCase,
	getUC *query.GetDokumenAsetUseCase,
	downloadUC *query.DownloadDokumenAsetUseCase,
) *DokumenAsetHandler {
	return &DokumenAsetHandler{
		presignUC:  presignUC,
		confirmUC:  confirmUC,
		updateUC:   updateUC,
		deleteUC:   deleteUC,
		listUC:     listUC,
		getUC:      getUC,
		downloadUC: downloadUC,
	}
}

func (h *DokumenAsetHandler) Presign(c *gin.Context) {
	var req dto.DokumenAsetPresignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.presignUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "presign url dokumen berhasil dibuat", resp)
}

func (h *DokumenAsetHandler) Confirm(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.DokumenAsetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.confirmUC.Execute(c.Request.Context(), userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "dokumen aset berhasil dibuat", resp)
}

func (h *DokumenAsetHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.DokumenAsetUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updateUC.Execute(c.Request.Context(), id, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *DokumenAsetHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.deleteUC.Execute(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *DokumenAsetHandler) List(c *gin.Context) {
	var req dto.DokumenAsetListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}

	isAuthenticated := middleware.GetUserID(c) != ""

	items, meta, err := h.listUC.Execute(c.Request.Context(), isAuthenticated, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar dokumen aset berhasil diambil", items, meta)
}

func (h *DokumenAsetHandler) Get(c *gin.Context) {
	id := c.Param("id")

	isAuthenticated := middleware.GetUserID(c) != ""

	detail, err := h.getUC.Execute(c.Request.Context(), id, nil, isAuthenticated)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "dokumen aset berhasil diambil", detail)
}

func (h *DokumenAsetHandler) Download(c *gin.Context) {
	id := c.Param("id")

	isAuthenticated := middleware.GetUserID(c) != ""

	resp, err := h.downloadUC.Execute(c.Request.Context(), id, isAuthenticated)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "akses download dokumen berhasil dibuat", resp)
}
