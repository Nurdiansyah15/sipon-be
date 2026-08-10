package http

import (
	"github.com/gin-gonic/gin"

	"sipon-be/internal/shared/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, h *FeedbackHandler, jwtAuth, principalLoad gin.HandlerFunc) {
	fb := router.Group("/api/v1/web/feedbacks")
	fb.Use(jwtAuth, principalLoad)
	{
		fb.GET("", h.ListFeedbacks)
		fb.GET("/my", h.ListMyFeedbacks)
		fb.GET("/:id", h.GetFeedback)
		fb.POST("", h.CreateFeedback)
		fb.PUT("/:id", h.UpdateFeedback)
		fb.DELETE("/:id", h.DeleteFeedback)

		fb.GET("/:id/comments", h.ListComments)
		fb.POST("/:id/comments", h.CreateComment)
		fb.POST("/:id/like", h.ToggleLikeFeedback)

		fb.POST("/:id/attachments/presign", h.AttachmentPresign)
		fb.POST("/:id/attachments/confirm", h.AttachmentConfirm)
		fb.GET("/:id/attachments", h.ListAttachments)
		fb.DELETE("/:id/attachments/:attachmentId", h.DeleteAttachment)
	}

	comments := router.Group("/api/v1/web/comments")
	comments.Use(jwtAuth, principalLoad)
	{
		comments.PUT("/:commentId", h.UpdateComment)
		comments.DELETE("/:commentId", h.DeleteComment)
		comments.POST("/:commentId/like", h.ToggleLikeComment)
	}

	admin := router.Group("/api/v1/web/feedback/admin")
	admin.Use(jwtAuth, principalLoad, middleware.RequirePermission("manage_feedback"))
	{
		admin.GET("/feedbacks", h.AdminListFeedbacks)
		admin.GET("/feedbacks/:id", h.AdminGetFeedback)
		admin.GET("/feedbacks/:id/comments", h.AdminListComments)
		admin.POST("/feedbacks/:id/takedown", h.AdminTakedownFeedback)
		admin.POST("/feedbacks/:id/restore", h.AdminRestoreFeedback)
		admin.POST("/comments/:commentId/takedown", h.AdminTakedownComment)
		admin.POST("/comments/:commentId/restore", h.AdminRestoreComment)
	}
}
