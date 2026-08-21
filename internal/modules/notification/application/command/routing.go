package command

// RoutingAdminBroadcast adalah routing key broadcast yang dipublish endpoint
// admin dan dikonsumsi worker notification untuk fanout ke semua user aktif.
const RoutingAdminBroadcast = "notification.broadcast"
