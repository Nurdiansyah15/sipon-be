package http

import (
	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/feedback/application/command"
	"sipon-be/internal/modules/feedback/application/dto"
	"sipon-be/internal/modules/feedback/application/query"
	"sipon-be/internal/shared/httperror"
	"sipon-be/internal/shared/middleware"
	"sipon-be/internal/shared/respond"
)

const feedbackPermission = "manage_feedback"

type FeedbackHandler struct {
	listFeedbacks   *query.ListFeedbacksUseCase
	getFeedback     *query.GetFeedbackUseCase
	listComments    *query.ListCommentsUseCase
	listAttachments *query.ListAttachmentsUseCase

	createFeedback   *command.CreateFeedbackUseCase
	updateFeedback   *command.UpdateFeedbackUseCase
	deleteFeedback   *command.DeleteFeedbackUseCase
	moderateFeedback *command.ModerateFeedbackUseCase
	createComment    *command.CreateCommentUseCase
	updateComment    *command.UpdateCommentUseCase
	deleteComment    *command.DeleteCommentUseCase
	moderateComment  *command.ModerateCommentUseCase
	toggleLike       *command.ToggleLikeUseCase
	attachment       *command.AttachmentUseCase
}

func NewFeedbackHandler(
	listFeedbacks *query.ListFeedbacksUseCase,
	getFeedback *query.GetFeedbackUseCase,
	listComments *query.ListCommentsUseCase,
	listAttachments *query.ListAttachmentsUseCase,
	createFeedback *command.CreateFeedbackUseCase,
	updateFeedback *command.UpdateFeedbackUseCase,
	deleteFeedback *command.DeleteFeedbackUseCase,
	moderateFeedback *command.ModerateFeedbackUseCase,
	createComment *command.CreateCommentUseCase,
	updateComment *command.UpdateCommentUseCase,
	deleteComment *command.DeleteCommentUseCase,
	moderateComment *command.ModerateCommentUseCase,
	toggleLike *command.ToggleLikeUseCase,
	attachment *command.AttachmentUseCase,
) *FeedbackHandler {
	return &FeedbackHandler{
		listFeedbacks:    listFeedbacks,
		getFeedback:      getFeedback,
		listComments:     listComments,
		listAttachments:  listAttachments,
		createFeedback:   createFeedback,
		updateFeedback:   updateFeedback,
		deleteFeedback:   deleteFeedback,
		moderateFeedback: moderateFeedback,
		createComment:    createComment,
		updateComment:    updateComment,
		deleteComment:    deleteComment,
		moderateComment:  moderateComment,
		toggleLike:       toggleLike,
		attachment:       attachment,
	}
}

// --- Self-service ---

func (h *FeedbackHandler) ListFeedbacks(c *gin.Context) {
	var req dto.ListFeedbackQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listFeedbacks.Execute(c.Request.Context(), query.ListFeedbacksParams{
		Query:        req,
		ViewerUserID: middleware.GetUserID(c),
	})
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar feedback berhasil diambil", items, meta)
}

func (h *FeedbackHandler) ListMyFeedbacks(c *gin.Context) {
	var req dto.ListFeedbackQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listFeedbacks.Execute(c.Request.Context(), query.ListFeedbacksParams{
		Query:           req,
		ViewerUserID:    middleware.GetUserID(c),
		IncludeTakedown: true,
		OnlyMine:        true,
	})
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar feedback milik saya berhasil diambil", items, meta)
}

func (h *FeedbackHandler) GetFeedback(c *gin.Context) {
	userID := middleware.GetUserID(c)
	resp, err := h.getFeedback.Execute(c.Request.Context(), query.GetFeedbackParams{
		FeedbackID:   c.Param("id"),
		ViewerUserID: userID,
		IsModerator:  h.isModerator(c),
	})
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "detail feedback berhasil diambil", resp)
}

func (h *FeedbackHandler) CreateFeedback(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.CreateFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createFeedback.Execute(c.Request.Context(), userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "feedback berhasil dibuat", resp)
}

func (h *FeedbackHandler) UpdateFeedback(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.UpdateFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updateFeedback.Execute(c.Request.Context(), userID, c.Param("id"), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *FeedbackHandler) DeleteFeedback(c *gin.Context) {
	userID := middleware.GetUserID(c)
	resp, err := h.deleteFeedback.Execute(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *FeedbackHandler) ListComments(c *gin.Context) {
	var req struct {
		Page  int `form:"page"`
		Limit int `form:"limit"`
	}
	_ = c.ShouldBindQuery(&req)
	items, meta, err := h.listComments.Execute(c.Request.Context(), query.ListCommentsParams{
		FeedbackID:   c.Param("id"),
		ViewerUserID: middleware.GetUserID(c),
		Page:         req.Page,
		Limit:        req.Limit,
	})
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar komentar berhasil diambil", items, meta)
}

func (h *FeedbackHandler) CreateComment(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createComment.Execute(c.Request.Context(), userID, c.Param("id"), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, resp.Message, resp)
}

func (h *FeedbackHandler) UpdateComment(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updateComment.Execute(c.Request.Context(), userID, c.Param("commentId"), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *FeedbackHandler) DeleteComment(c *gin.Context) {
	userID := middleware.GetUserID(c)
	resp, err := h.deleteComment.Execute(c.Request.Context(), userID, c.Param("commentId"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *FeedbackHandler) ToggleLikeFeedback(c *gin.Context) {
	h.toggleLikeTarget(c, dto.ToggleLikeRequest{TargetType: "feedback", TargetID: c.Param("id")})
}

func (h *FeedbackHandler) ToggleLikeComment(c *gin.Context) {
	h.toggleLikeTarget(c, dto.ToggleLikeRequest{TargetType: "comment", TargetID: c.Param("commentId")})
}

func (h *FeedbackHandler) toggleLikeTarget(c *gin.Context, req dto.ToggleLikeRequest) {
	userID := middleware.GetUserID(c)
	resp, err := h.toggleLike.Execute(c.Request.Context(), userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "like berhasil diubah", resp)
}

func (h *FeedbackHandler) AttachmentPresign(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.AttachmentPresignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.attachment.Presign(c.Request.Context(), userID, c.Param("id"), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "presign url attachment berhasil dibuat", resp)
}

func (h *FeedbackHandler) AttachmentConfirm(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.AttachmentConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.attachment.Confirm(c.Request.Context(), userID, c.Param("id"), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "attachment berhasil dikonfirmasi", resp)
}

func (h *FeedbackHandler) ListAttachments(c *gin.Context) {
	userID := middleware.GetUserID(c)
	items, err := h.listAttachments.Execute(c.Request.Context(), query.ListAttachmentsParams{
		FeedbackID:   c.Param("id"),
		ViewerUserID: userID,
		IsModerator:  h.isModerator(c),
	})
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "daftar attachment berhasil diambil", items)
}

func (h *FeedbackHandler) DeleteAttachment(c *gin.Context) {
	userID := middleware.GetUserID(c)
	resp, err := h.attachment.Delete(c.Request.Context(), userID, c.Param("id"), c.Param("attachmentId"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

// --- Admin moderation ---

func (h *FeedbackHandler) AdminListFeedbacks(c *gin.Context) {
	var req dto.ListFeedbackQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listFeedbacks.Execute(c.Request.Context(), query.ListFeedbacksParams{
		Query:           req,
		ViewerUserID:    middleware.GetUserID(c),
		IncludeTakedown: true,
	})
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar feedback berhasil diambil", items, meta)
}

func (h *FeedbackHandler) AdminGetFeedback(c *gin.Context) {
	resp, err := h.getFeedback.Execute(c.Request.Context(), query.GetFeedbackParams{
		FeedbackID:   c.Param("id"),
		ViewerUserID: middleware.GetUserID(c),
		IsModerator:  true,
	})
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "detail feedback berhasil diambil", resp)
}

func (h *FeedbackHandler) AdminListComments(c *gin.Context) {
	var req struct {
		Page  int `form:"page"`
		Limit int `form:"limit"`
	}
	_ = c.ShouldBindQuery(&req)
	items, meta, err := h.listComments.Execute(c.Request.Context(), query.ListCommentsParams{
		FeedbackID:      c.Param("id"),
		ViewerUserID:    middleware.GetUserID(c),
		IncludeTakedown: true,
		Page:            req.Page,
		Limit:           req.Limit,
	})
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar komentar berhasil diambil", items, meta)
}

func (h *FeedbackHandler) AdminTakedownFeedback(c *gin.Context) {
	adminID := middleware.GetUserID(c)
	var req command.ModerateFeedbackRequest
	_ = c.ShouldBindJSON(&req)
	resp, err := h.moderateFeedback.Takedown(c.Request.Context(), adminID, c.Param("id"), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *FeedbackHandler) AdminRestoreFeedback(c *gin.Context) {
	resp, err := h.moderateFeedback.Restore(c.Request.Context(), c.Param("id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *FeedbackHandler) AdminTakedownComment(c *gin.Context) {
	adminID := middleware.GetUserID(c)
	var req command.ModerateCommentRequest
	_ = c.ShouldBindJSON(&req)
	resp, err := h.moderateComment.Takedown(c.Request.Context(), adminID, c.Param("commentId"), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *FeedbackHandler) AdminRestoreComment(c *gin.Context) {
	resp, err := h.moderateComment.Restore(c.Request.Context(), c.Param("commentId"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, resp)
}

func (h *FeedbackHandler) isModerator(c *gin.Context) bool {
	p := middleware.GetPrincipal(c)
	if p == nil {
		return false
	}
	for _, perm := range p.Permissions {
		if perm == feedbackPermission {
			return true
		}
	}
	return false
}
