package http

import (
	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/command"
	"sipon-be/internal/modules/kesantrian/application/dto"
	"sipon-be/internal/modules/kesantrian/application/query"
	"sipon-be/internal/shared/httperror"
	"sipon-be/internal/shared/kernel"
	"sipon-be/internal/shared/middleware"
	"sipon-be/internal/shared/respond"
)

type SantriHandler struct {
	getSantriUC     *query.GetSantriUseCase
	updateSantriUC  *command.UpdateSantriUseCase
	requestSantriUC *command.RequestSantriUseCase

	createSantriUC         *command.CreateSantriUseCase
	importSantriUC         *command.ImportSantriUseCase
	listSantriUC           *query.ListSantriUseCase
	listSantriRequestsUC   *query.ListSantriRequestsUseCase
	approveSantriRequestUC *command.ApproveSantriRequestUseCase
	rejectSantriRequestUC  *command.RejectSantriRequestUseCase

	dokumenPresignUC   *command.DokumenPresignUseCase
	dokumenConfirmUC   *command.DokumenConfirmUseCase
	dokumenListUC      *query.DokumenListUseCase
	adminDokumenListUC *query.AdminDokumenListUseCase
	dokumenAccessUC    *query.DokumenAccessUseCase
	dokumenDeleteUC    *command.DokumenDeleteUseCase
	dokumenVerifyUC    *command.DokumenVerifyUseCase
	dokumenRejectUC    *command.DokumenRejectUseCase

	createSantriFromPendaftaranUC *command.CreateSantriFromPendaftaranUseCase
	changeSantriStatusUC          *command.ChangeSantriStatusUseCase
}

func NewSantriHandler(
	getSantriUC *query.GetSantriUseCase,
	updateSantriUC *command.UpdateSantriUseCase,
	requestSantriUC *command.RequestSantriUseCase,
	createSantriUC *command.CreateSantriUseCase,
	importSantriUC *command.ImportSantriUseCase,
	listSantriUC *query.ListSantriUseCase,
	listSantriRequestsUC *query.ListSantriRequestsUseCase,
	approveSantriRequestUC *command.ApproveSantriRequestUseCase,
	rejectSantriRequestUC *command.RejectSantriRequestUseCase,
	dokumenPresignUC *command.DokumenPresignUseCase,
	dokumenConfirmUC *command.DokumenConfirmUseCase,
	dokumenListUC *query.DokumenListUseCase,
	adminDokumenListUC *query.AdminDokumenListUseCase,
	dokumenAccessUC *query.DokumenAccessUseCase,
	dokumenDeleteUC *command.DokumenDeleteUseCase,
	dokumenVerifyUC *command.DokumenVerifyUseCase,
	dokumenRejectUC *command.DokumenRejectUseCase,
	createSantriFromPendaftaranUC *command.CreateSantriFromPendaftaranUseCase,
	changeSantriStatusUC *command.ChangeSantriStatusUseCase,
) *SantriHandler {
	return &SantriHandler{
		getSantriUC:            getSantriUC,
		updateSantriUC:         updateSantriUC,
		requestSantriUC:        requestSantriUC,
		createSantriUC:         createSantriUC,
		importSantriUC:         importSantriUC,
		listSantriUC:           listSantriUC,
		listSantriRequestsUC:   listSantriRequestsUC,
		approveSantriRequestUC: approveSantriRequestUC,
		rejectSantriRequestUC:  rejectSantriRequestUC,
		dokumenPresignUC:       dokumenPresignUC,
		dokumenConfirmUC:       dokumenConfirmUC,
		dokumenListUC:          dokumenListUC,
		adminDokumenListUC:     adminDokumenListUC,
		dokumenAccessUC:        dokumenAccessUC,
		dokumenDeleteUC:        dokumenDeleteUC,
		dokumenVerifyUC:    dokumenVerifyUC,
		dokumenRejectUC:    dokumenRejectUC,
		createSantriFromPendaftaranUC: createSantriFromPendaftaranUC,
		changeSantriStatusUC:          changeSantriStatusUC,
	}
}

func (h *SantriHandler) GetSantri(c *gin.Context) {
	userID := middleware.GetUserID(c)
	resp, err := h.getSantriUC.Execute(c.Request.Context(), userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "profil santri berhasil diambil", resp)
}

func (h *SantriHandler) UpdateSantri(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.UpdateSantriRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updateSantriUC.Execute(c.Request.Context(), userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *SantriHandler) RequestSantri(c *gin.Context) {
	userID := middleware.GetUserID(c)
	resp, err := h.requestSantriUC.Execute(c.Request.Context(), userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, resp.Message, resp)
}

func (h *SantriHandler) CreateSantri(c *gin.Context) {
	var req dto.CreateSantriRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createSantriUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "santri berhasil dibuat", resp)
}

func (h *SantriHandler) ImportSantri(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		httperror.Handle(c, kernel.WrapMsg(application.ErrCodeBadRequest, "file wajib diunggah", err))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		httperror.Handle(c, kernel.Wrap(application.ErrCodeInternal, err))
		return
	}
	defer file.Close()

	rows, err := parseSantriImportExcel(file)
	if err != nil {
		httperror.Handle(c, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, err.Error(), err))
		return
	}

	resp, err := h.importSantriUC.Execute(c.Request.Context(), rows)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "import santri selesai diproses", resp)
}

func (h *SantriHandler) DownloadImportTemplate(c *gin.Context) {
	buf, err := buildSantriImportTemplate()
	if err != nil {
		httperror.Handle(c, kernel.Wrap(application.ErrCodeInternal, err))
		return
	}
	c.Header("Content-Disposition", `attachment; filename="template-import-santri.xlsx"`)
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf)
}

func (h *SantriHandler) ListSantri(c *gin.Context) {
	var req dto.ListSantriQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listSantriUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar santri berhasil diambil", items, meta)
}

func (h *SantriHandler) ListSantriRequests(c *gin.Context) {
	var req dto.ListSantriRequestsQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listSantriRequestsUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar permintaan santri berhasil diambil", items, meta)
}

func (h *SantriHandler) ApproveSantriRequest(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	requestID := c.Param("id")
	var req dto.ApproveSantriRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.approveSantriRequestUC.Execute(c.Request.Context(), reviewerID, requestID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *SantriHandler) RejectSantriRequest(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	requestID := c.Param("id")
	var req dto.RejectSantriRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.rejectSantriRequestUC.Execute(c.Request.Context(), reviewerID, requestID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *SantriHandler) DokumenPresign(c *gin.Context) {
	var req dto.DokumenPresignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.dokumenPresignUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "presign url dokumen berhasil dibuat", resp)
}

func (h *SantriHandler) DokumenConfirm(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.DokumenConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.dokumenConfirmUC.Execute(c.Request.Context(), userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "dokumen berhasil dikonfirmasi", resp)
}

func (h *SantriHandler) DokumenList(c *gin.Context) {
	userID := middleware.GetUserID(c)
	kind := c.Query("kind")
	items, err := h.dokumenListUC.Execute(c.Request.Context(), userID, kind)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "daftar dokumen berhasil diambil", items)
}

func (h *SantriHandler) AdminDokumenList(c *gin.Context) {
	santriID := c.Param("id")
	kind := c.Query("kind")
	items, err := h.adminDokumenListUC.Execute(c.Request.Context(), santriID, kind)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "daftar dokumen santri berhasil diambil", items)
}

func (h *SantriHandler) DokumenAccess(c *gin.Context) {
	userID := middleware.GetUserID(c)
	dokumenID := c.Param("id")
	resp, err := h.dokumenAccessUC.Execute(c.Request.Context(), userID, dokumenID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "akses dokumen berhasil dibuat", resp)
}

func (h *SantriHandler) DokumenDelete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	dokumenID := c.Param("id")
	resp, err := h.dokumenDeleteUC.Execute(c.Request.Context(), userID, dokumenID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *SantriHandler) DokumenVerify(c *gin.Context) {
	verifierID := middleware.GetUserID(c)
	dokumenID := c.Param("id")
	resp, err := h.dokumenVerifyUC.Execute(c.Request.Context(), verifierID, dokumenID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *SantriHandler) DokumenReject(c *gin.Context) {
	verifierID := middleware.GetUserID(c)
	dokumenID := c.Param("id")
	var req dto.RejectDokumenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.dokumenRejectUC.Execute(c.Request.Context(), verifierID, dokumenID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *SantriHandler) ChangeSantriStatus(c *gin.Context) {
	adminID := middleware.GetUserID(c)
	santriID := c.Param("id")
	var req command.ChangeSantriStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.changeSantriStatusUC.Execute(c.Request.Context(), santriID, adminID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}
