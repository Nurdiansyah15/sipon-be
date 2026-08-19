package notification

import (
	"database/sql"
	"log/slog"

	"github.com/gin-gonic/gin"

notifApp "sipon-be/internal/modules/notification/application"
	"sipon-be/internal/modules/notification/application/command"
	"sipon-be/internal/modules/notification/application/query"
notifMQ "sipon-be/internal/modules/notification/interfaces/mq"

	"sipon-be/internal/modules/notification/infrastructure/persistence"
	notifHTTP "sipon-be/internal/modules/notification/interfaces/http"
	"sipon-be/internal/shared/config"
	messaging "sipon-be/internal/modules/messaging"
)

type Module struct {
	handler *notifHTTP.NotificationHandler
	mqDeps  notifMQ.Dependencies
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

	dispatcher := notifApp.NewDispatcher(notifRepo, deliveryRepo, prefRepo, slog.Default())

	listInbox := query.NewListInboxUseCase(deliveryRepo)
	unreadCount := query.NewUnreadCountUseCase(deliveryRepo)
	getPreference := query.NewGetPreferenceUseCase(prefRepo)
	updatePref := command.NewUpdatePreferenceUseCase(prefRepo)
	markRead := command.NewMarkNotificationReadUseCase(deliveryRepo)
	markAllRead := command.NewMarkAllNotificationsReadUseCase(deliveryRepo)
	sendBroadcast := command.NewSendNotificationUseCase(dispatcher)

	handler := notifHTTP.NewNotificationHandler(
		listInbox, unreadCount, getPreference, updatePref,
		markRead, markAllRead, sendBroadcast,
	)

	mqDeps := notifMQ.Dependencies{Dispatcher: dispatcher}

	return &Module{
		handler:       handler,
		mqDeps:        mqDeps,
		jwtAuth:       jwtAuth,
		principalLoad: principalLoad,
	}
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	grp := router.Group("/")
	notifHTTP.RegisterRoutes(grp, m.handler, m.jwtAuth, m.principalLoad)
}

func (m *Module) RegisterMessageHandlers(registry messaging.Contract) ([]messaging.Binding, error) {
	return notifMQ.RegisterHandlers(registry, m.mqDeps)
}
