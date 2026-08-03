package http

import (
	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/psb/application/command"
	"sipon-be/internal/modules/psb/application/dto"
	"sipon-be/internal/modules/psb/application/query"
	"sipon-be/internal/shared/httperror"
	"sipon-be/internal/shared/middleware"
	"sipon-be/internal/shared/respond"
)

type PsbHandler struct {
	settingQuery   *query.SettingQueryUseCase
	getPendaftaran *query.GetPendaftaranUseCase
	listPendaftaran *query.ListPendaftaranUseCase
	listReviews    *query.ListReviewsUseCase

	upsertFormulir  *command.UpsertFormulirUseCase
	pendaftarAction *command.PendaftarActionUseCase
	adminReview     *command.AdminReviewUseCase
	generateNIS     *command.GenerateNISUseCase

	dokumenPresign  *command.DokumenPresignUseCase
	dokumenConfirm  *command.DokumenConfirmUseCase
	dokumenDelete   *command.DokumenDeleteUseCase
	dokumenVerify   *command.DokumenVerifyUseCase
	dokumenReject   *command.DokumenRejectUseCase
	dokumenList     *query.DokumenListUseCase

	manageSetting *command.ManageSettingUseCase
	purgePeriod   *command.PurgePeriodUseCase
}

func NewPsbHandler(
	settingQuery *query.SettingQueryUseCase,
	getPendaftaran *query.GetPendaftaranUseCase,
	listPendaftaran *query.ListPendaftaranUseCase,
	listReviews *query.ListReviewsUseCase,
	upsertFormulir *command.UpsertFormulirUseCase,
	pendaftarAction *command.PendaftarActionUseCase,
	adminReview *command.AdminReviewUseCase,
	generateNIS *command.GenerateNISUseCase,
	dokumenPresign *command.DokumenPresignUseCase,
	dokumenConfirm *command.DokumenConfirmUseCase,
	dokumenDelete *command.DokumenDeleteUseCase,
	dokumenVerify *command.DokumenVerifyUseCase,
	dokumenReject *command.DokumenRejectUseCase,
	dokumenList *query.DokumenListUseCase,
	manageSetting *command.ManageSettingUseCase,
	purgePeriod *command.PurgePeriodUseCase,
) *PsbHandler {
	return &PsbHandler{
		settingQuery:    settingQuery,
		getPendaftaran:  getPendaftaran,
		listPendaftaran: listPendaftaran,
		listReviews:     listReviews,
		upsertFormulir:  upsertFormulir,
		pendaftarAction: pendaftarAction,
		adminReview:     adminReview,
		generateNIS:     generateNIS,
		dokumenPresign:  dokumenPresign,
		dokumenConfirm:  dokumenConfirm,
		dokumenDelete:   dokumenDelete,
		dokumenVerify:   dokumenVerify,
		dokumenReject:   dokumenReject,
		dokumenList:     dokumenList,
		manageSetting:   manageSetting,
		purgePeriod:     purgePeriod,
	}
}

func (h *PsbHandler) GetActiveSetting(c *gin.Context) {
	resp, err := h.settingQuery.GetActive(c.Request.Context())
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "periode psb aktif berhasil diambil", resp)
}

func (h *PsbHandler) GetPendaftaran(c *gin.Context) {
	userID := middleware.GetUserID(c)
	resp, err := h.getPendaftaran.Execute(c.Request.Context(), userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "data pendaftaran berhasil diambil", resp)
}

func (h *PsbHandler) UpsertFormulir(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.UpsertFormulirRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.upsertFormulir.Execute(c.Request.Context(), userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "formulir berhasil disimpan", resp)
}

func (h *PsbHandler) SubmitPendaftaran(c *gin.Context) {
	userID := middleware.GetUserID(c)
	p, err := h.getPendaftaran.Execute(c.Request.Context(), userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	result, err := h.pendaftarAction.SubmitPendaftaran(c.Request.Context(), userID, p.PsbSettingID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, result.Message, result)
}

func (h *PsbHandler) SubmitDaftarUlang(c *gin.Context) {
	userID := middleware.GetUserID(c)
	p, err := h.getPendaftaran.Execute(c.Request.Context(), userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	result, err := h.pendaftarAction.SubmitDaftarUlang(c.Request.Context(), userID, p.PsbSettingID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, result.Message, result)
}

func (h *PsbHandler) GetRiwayat(c *gin.Context) {
	userID := middleware.GetUserID(c)
	p, err := h.getPendaftaran.Execute(c.Request.Context(), userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	items, err := h.listReviews.Execute(c.Request.Context(), p.ID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "riwayat berhasil diambil", items)
}

func (h *PsbHandler) DokumenPresign(c *gin.Context) {
	var req dto.DokumenPresignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.dokumenPresign.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "presign url berhasil dibuat", resp)
}

func (h *PsbHandler) DokumenConfirm(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.DokumenConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.dokumenConfirm.Execute(c.Request.Context(), userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "dokumen berhasil dikonfirmasi", resp)
}

func (h *PsbHandler) DokumenList(c *gin.Context) {
	userID := middleware.GetUserID(c)
	items, err := h.dokumenList.Execute(c.Request.Context(), userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "daftar dokumen berhasil diambil", items)
}

func (h *PsbHandler) DokumenDelete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	dokumenID := c.Param("id")
	resp, err := h.dokumenDelete.Execute(c.Request.Context(), userID, dokumenID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

// --- Admin handlers ---

func (h *PsbHandler) AdminListPendaftaran(c *gin.Context) {
	var req dto.ListPendaftarQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listPendaftaran.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar pendaftar berhasil diambil", items, meta)
}

func (h *PsbHandler) AdminGetPendaftaran(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.getPendaftaran.ExecuteByID(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "data pendaftar berhasil diambil", resp)
}

func (h *PsbHandler) AdminGetRiwayat(c *gin.Context) {
	pendaftarID := c.Param("id")
	items, err := h.listReviews.Execute(c.Request.Context(), pendaftarID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "riwayat berhasil diambil", items)
}

func (h *PsbHandler) AdminDokumenList(c *gin.Context) {
	pendaftarID := c.Param("id")
	items, err := h.dokumenList.ExecuteByPendaftarID(c.Request.Context(), pendaftarID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "daftar dokumen berhasil diambil", items)
}

func (h *PsbHandler) AdminDokumenVerify(c *gin.Context) {
	verifierID := middleware.GetUserID(c)
	dokumenID := c.Param("dokumenId")
	resp, err := h.dokumenVerify.Execute(c.Request.Context(), verifierID, dokumenID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *PsbHandler) AdminDokumenReject(c *gin.Context) {
	verifierID := middleware.GetUserID(c)
	dokumenID := c.Param("dokumenId")
	var req struct {
		Notes *string `json:"notes,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.dokumenReject.Execute(c.Request.Context(), verifierID, dokumenID, req.Notes)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *PsbHandler) AdminRequestRevision(c *gin.Context) {
	adminID := middleware.GetUserID(c)
	pendaftarID := c.Param("id")
	var req struct {
		Notes *string `json:"notes,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.adminReview.RequestRevision(c.Request.Context(), pendaftarID, adminID, req.Notes)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *PsbHandler) AdminReject(c *gin.Context) {
	adminID := middleware.GetUserID(c)
	pendaftarID := c.Param("id")
	var req struct {
		Notes *string `json:"notes,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.adminReview.Reject(c.Request.Context(), pendaftarID, adminID, req.Notes)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *PsbHandler) AdminAccept(c *gin.Context) {
	adminID := middleware.GetUserID(c)
	pendaftarID := c.Param("id")
	resp, err := h.adminReview.Accept(c.Request.Context(), pendaftarID, adminID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *PsbHandler) AdminMarkNotReregistered(c *gin.Context) {
	adminID := middleware.GetUserID(c)
	pendaftarID := c.Param("id")
	resp, err := h.adminReview.MarkNotReregistered(c.Request.Context(), pendaftarID, adminID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *PsbHandler) AdminRequestRevisionDaftarUlang(c *gin.Context) {
	adminID := middleware.GetUserID(c)
	pendaftarID := c.Param("id")
	var req struct {
		Notes *string `json:"notes,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.adminReview.RequestRevisionDaftarUlang(c.Request.Context(), pendaftarID, adminID, req.Notes)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *PsbHandler) AdminGenerateNIS(c *gin.Context) {
	adminID := middleware.GetUserID(c)
	pendaftarID := c.Param("id")
	resp, err := h.generateNIS.Execute(c.Request.Context(), pendaftarID, adminID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

// --- Settings admin ---

func (h *PsbHandler) ListSettings(c *gin.Context) {
	items, err := h.settingQuery.List(c.Request.Context())
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "daftar periode PSB berhasil diambil", items)
}

func (h *PsbHandler) CreateSetting(c *gin.Context) {
	var req dto.CreateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.manageSetting.Create(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "periode PSB berhasil dibuat", resp)
}

func (h *PsbHandler) UpdateSetting(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.manageSetting.Update(c.Request.Context(), id, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "periode PSB berhasil diupdate", resp)
}

func (h *PsbHandler) PurgePeriod(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.purgePeriod.Execute(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}
