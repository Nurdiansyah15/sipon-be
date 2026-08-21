package notification

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/gin-gonic/gin"

	notifApp "sipon-be/internal/modules/notification/application"
	"sipon-be/internal/modules/notification/application/command"
	"sipon-be/internal/modules/notification/application/ports"
	"sipon-be/internal/modules/notification/application/query"
	notifMQ "sipon-be/internal/modules/notification/interfaces/mq"

	messaging "sipon-be/internal/modules/messaging"
	"sipon-be/internal/modules/notification/infrastructure/external"
	"sipon-be/internal/modules/notification/infrastructure/persistence"
	notifHTTP "sipon-be/internal/modules/notification/interfaces/http"
	"sipon-be/internal/shared/config"
)

type Module struct {
	handler       *notifHTTP.NotificationHandler
	sendBroadcast *command.SendNotificationUseCase
	mqDeps        notifMQ.Dependencies
	jwtAuth       gin.HandlerFunc
	principalLoad gin.HandlerFunc
}

func NewModule(
	db *sql.DB,
	cfg *config.Config,
	jwtAuth gin.HandlerFunc,
	principalLoad gin.HandlerFunc,
) *Module {
	notifRepo := persistence.NewPostgresNotificationRepository(db)
	deliveryRepo := persistence.NewPostgresDeliveryAttemptRepository(db)
	prefRepo := persistence.NewPostgresPreferenceRepository(db)
	deviceRepo := persistence.NewPostgresDeviceRegistrationRepository(db)

	pushSender, err := external.NewFCMPushSender(context.Background(), &cfg.FCM)
	if err != nil {
		slog.Warn("gagal init FCM push sender, push dinonaktifkan", slog.Any("error", err))
		pushSender = &external.NoopPushSender{}
	}

	dispatcher := notifApp.NewDispatcher(notifRepo, deliveryRepo, prefRepo, deviceRepo, pushSender, slog.Default())

	listInbox := query.NewListInboxUseCase(deliveryRepo)
	unreadCount := query.NewUnreadCountUseCase(deliveryRepo)
	getPreference := query.NewGetPreferenceUseCase(prefRepo)
	updatePref := command.NewUpdatePreferenceUseCase(prefRepo)
	markRead := command.NewMarkNotificationReadUseCase(deliveryRepo)
	markAllRead := command.NewMarkAllNotificationsReadUseCase(deliveryRepo)
	sendBroadcast := command.NewSendNotificationUseCase()

	registerDevice := command.NewRegisterDeviceUseCase(deviceRepo)
	unregisterDevice := command.NewUnregisterDeviceUseCase(deviceRepo)
	listDevices := command.NewListDevicesUseCase(deviceRepo)

	handler := notifHTTP.NewNotificationHandler(
		listInbox, unreadCount, getPreference, updatePref,
		markRead, markAllRead, sendBroadcast,
		registerDevice, unregisterDevice, listDevices,
	)

	mqDeps := notifMQ.Dependencies{Dispatcher: dispatcher}

	return &Module{
		handler:       handler,
		sendBroadcast: sendBroadcast,
		mqDeps:        mqDeps,
		jwtAuth:       jwtAuth,
		principalLoad: principalLoad,
	}
}

func (m *Module) SetUserProvider(p notifMQ.UserProvider) {
	m.mqDeps.UserProvider = p
}

func (m *Module) SetOutboxWriter(w ports.OutboxWriter) {
	m.sendBroadcast.SetOutboxWriter(w)
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	grp := router.Group("/")
	notifHTTP.RegisterRoutes(grp, m.handler, m.jwtAuth, m.principalLoad)
}

func (m *Module) RegisterMessageHandlers(registry messaging.Contract) ([]messaging.Binding, error) {
	return notifMQ.RegisterHandlers(registry, m.mqDeps)
}
