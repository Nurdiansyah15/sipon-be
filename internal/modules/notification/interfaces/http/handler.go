package http

import (
	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/notification/application/command"
	"sipon-be/internal/modules/notification/application/dto"
	"sipon-be/internal/modules/notification/application/query"
	"sipon-be/internal/modules/notification/domain/device/constant"
	"sipon-be/internal/shared/httperror"
	"sipon-be/internal/shared/middleware"
	"sipon-be/internal/shared/respond"
)

type NotificationHandler struct {
	listInbox        *query.ListInboxUseCase
	unreadCount      *query.UnreadCountUseCase
	getPreference    *query.GetPreferenceUseCase
	updatePref       *command.UpdatePreferenceUseCase
	markRead         *command.MarkNotificationReadUseCase
	markAllRead      *command.MarkAllNotificationsReadUseCase
	sendBroadcast    *command.SendNotificationUseCase
	registerDevice   *command.RegisterDeviceUseCase
	unregisterDevice *command.UnregisterDeviceUseCase
	listDevices      *command.ListDevicesUseCase
}

func NewNotificationHandler(
	listInbox *query.ListInboxUseCase,
	unreadCount *query.UnreadCountUseCase,
	getPreference *query.GetPreferenceUseCase,
	updatePref *command.UpdatePreferenceUseCase,
	markRead *command.MarkNotificationReadUseCase,
	markAllRead *command.MarkAllNotificationsReadUseCase,
	sendBroadcast *command.SendNotificationUseCase,
	registerDevice *command.RegisterDeviceUseCase,
	unregisterDevice *command.UnregisterDeviceUseCase,
	listDevices *command.ListDevicesUseCase,
) *NotificationHandler {
	return &NotificationHandler{
		listInbox:        listInbox,
		unreadCount:      unreadCount,
		getPreference:    getPreference,
		updatePref:       updatePref,
		markRead:         markRead,
		markAllRead:      markAllRead,
		sendBroadcast:    sendBroadcast,
		registerDevice:   registerDevice,
		unregisterDevice: unregisterDevice,
		listDevices:      listDevices,
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
	if err := h.sendBroadcast.Execute(c.Request.Context(), req); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "broadcast berhasil dikirim", nil)
}

func (h *NotificationHandler) RegisterDevice(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}

	input := command.RegisterDeviceInput{
		UserID:        userID,
		Platform:      parsePlatform(req.Platform),
		PushProvider:  parsePushProvider(req.PushProvider),
		ProviderToken: req.ProviderToken,
	}
	if req.DeviceID != "" {
		input.DeviceID = &req.DeviceID
	}
	if req.DeviceName != "" {
		input.DeviceName = &req.DeviceName
	}
	if req.DeviceModel != "" {
		input.DeviceModel = &req.DeviceModel
	}
	if req.OSVersion != "" {
		input.OSVersion = &req.OSVersion
	}
	if req.AppVersion != "" {
		input.AppVersion = &req.AppVersion
	}
	if req.Timezone != "" {
		input.Timezone = &req.Timezone
	}

	dr, err := h.registerDevice.Execute(c.Request.Context(), input)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "device berhasil didaftarkan", dto.DeviceResponse{
		ID:           dr.ID,
		Platform:     string(dr.Platform),
		PushProvider: string(dr.PushProvider),
		Active:       dr.Active,
		LastSeenAt:   dr.LastSeenAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *NotificationHandler) UnregisterDevice(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.UnregisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	if err := h.unregisterDevice.Execute(c.Request.Context(), userID, req.ProviderToken); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "device berhasil dinonaktifkan", nil)
}

func (h *NotificationHandler) ListDevices(c *gin.Context) {
	userID := middleware.GetUserID(c)
	devices, err := h.listDevices.Execute(c.Request.Context(), userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	result := make([]dto.DeviceResponse, 0, len(devices))
	for _, dr := range devices {
		result = append(result, dto.DeviceResponse{
			ID:           dr.ID,
			Platform:     string(dr.Platform),
			PushProvider: string(dr.PushProvider),
			DeviceName:   derefStr(dr.DeviceName),
			DeviceModel:  derefStr(dr.DeviceModel),
			OSVersion:    derefStr(dr.OSVersion),
			AppVersion:   derefStr(dr.AppVersion),
			Active:       dr.Active,
			LastSeenAt:   dr.LastSeenAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	respond.OK(c, "daftar device berhasil diambil", result)
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func parsePlatform(s string) constant.Platform {
	switch constant.Platform(s) {
	case constant.PlatformAndroid, constant.PlatformIOS, constant.PlatformWeb:
		return constant.Platform(s)
	default:
		return constant.PlatformAndroid
	}
}

func parsePushProvider(s string) constant.PushProvider {
	switch constant.PushProvider(s) {
	case constant.PushProviderFCM, constant.PushProviderAPNS, constant.PushProviderHuawei, constant.PushProviderWebPush:
		return constant.PushProvider(s)
	default:
		return constant.PushProviderFCM
	}
}
