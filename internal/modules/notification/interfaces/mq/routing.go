package mq

import "sipon-be/internal/modules/messaging"

const (
	RoutingLoginSucceeded = "identity.user.login_succeeded"
	QueueNotification     = "sipon.worker.notification"
)

var Bindings = []messaging.Binding{
	{Queue: QueueNotification, RoutingKey: RoutingLoginSucceeded},
}
