package http

import (
	"github.com/gin-gonic/gin"

	"sipon-be/internal/shared/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, h *NotificationHandler, jwtAuth, principalLoad gin.HandlerFunc) {
	grp := router.Group("/api/v1/web/notifications")
	grp.Use(jwtAuth, principalLoad)
	{
		grp.GET("/inbox", h.ListInbox)
		grp.GET("/unread-count", h.UnreadCount)
		grp.POST("/:id/read", h.MarkRead)
		grp.POST("/read-all", h.MarkAllRead)
		grp.GET("/preferences", h.GetPreference)
		grp.PUT("/preferences", h.UpdatePreference)
		grp.POST("/devices", h.RegisterDevice)
		grp.DELETE("/devices", h.UnregisterDevice)
		grp.GET("/devices", h.ListDevices)
	}

	admin := router.Group("/api/v1/web/notification/admin")
	admin.Use(jwtAuth, principalLoad, middleware.RequirePermission("manage_notification"))
	{
		admin.POST("/broadcast", h.Broadcast)
	}
}
