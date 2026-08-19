package http

import (
	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/notification/application/command"
	"sipon-be/internal/modules/notification/application/dto"
	"sipon-be/internal/modules/notification/application/query"
	"sipon-be/internal/shared/httperror"
	"sipon-be/internal/shared/middleware"
	"sipon-be/internal/shared/respond"
)

type NotificationHandler struct {
	listInbox     *query.ListInboxUseCase
	unreadCount   *query.UnreadCountUseCase
	getPreference *query.GetPreferenceUseCase
	updatePref    *command.UpdatePreferenceUseCase
	markRead      *command.MarkNotificationReadUseCase
	markAllRead   *command.MarkAllNotificationsReadUseCase
	sendBroadcast *command.SendNotificationUseCase
}

func NewNotificationHandler(
	listInbox *query.ListInboxUseCase,
	unreadCount *query.UnreadCountUseCase,
	getPreference *query.GetPreferenceUseCase,
	updatePref *command.UpdatePreferenceUseCase,
	markRead *command.MarkNotificationReadUseCase,
	markAllRead *command.MarkAllNotificationsReadUseCase,
	sendBroadcast *command.SendNotificationUseCase,
) *NotificationHandler {
	return &NotificationHandler{
		listInbox:     listInbox,
		unreadCount:   unreadCount,
		getPreference: getPreference,
		updatePref:    updatePref,
		markRead:      markRead,
		markAllRead:   markAllRead,
		sendBroadcast: sendBroadcast,
	}
}

func (h *NotificationHandler) ListInbox(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.ListNotificationsQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listInbox.Execute(c.Request.Context(), userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "inbox notifikasi berhasil diambil", items, meta)
}

func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	resp, err := h.unreadCount.Execute(c.Request.Context(), userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "jumlah belum dibaca berhasil diambil", resp)
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	notifID := c.Param("id")
	if err := h.markRead.Execute(c.Request.Context(), notifID, userID); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "notifikasi ditandai sudah dibaca", nil)
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	resp, err := h.markAllRead.Execute(c.Request.Context(), userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "semua notifikasi ditandai sudah dibaca", resp)
}

func (h *NotificationHandler) GetPreference(c *gin.Context) {
	userID := middleware.GetUserID(c)
	resp, err := h.getPreference.Execute(c.Request.Context(), userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "preferensi notifikasi berhasil diambil", resp)
}

func (h *NotificationHandler) UpdatePreference(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.UpdatePreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updatePref.Execute(c.Request.Context(), userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "preferensi notifikasi berhasil diperbarui", resp)
}

func (h *NotificationHandler) Broadcast(c *gin.Context) {
	var req dto.BroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	if err := h.sendBroadcast.Execute(c.Request.Context(), "broadcast", "", nil, req); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "broadcast berhasil dikirim", nil)
}
