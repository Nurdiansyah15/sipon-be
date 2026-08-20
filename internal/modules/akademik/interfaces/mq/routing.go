package mq

import "sipon-be/internal/modules/messaging"

// Routing key canonical dengan format <module>.<resource>.<action>.
const (
	RoutingFingerprintSync  = "akademik.fingerprint.sync"
	RoutingSessionAutoClose = "akademik.session.auto_close"
	RoutingSessionAutoOpen  = "akademik.session.auto_open"

	// Queue untuk consumer role scheduler.
	QueueScheduler = "sipon.worker.scheduler"
)

// Legacy routing key untuk row scheduled_jobs yang masih menyimpan format lama.
// Hanya dipakai selama migration/backfill data; dihapus setelah seluruh row lama
// ter-backfill (lihat plan Phase 2 dan Phase 10).
const (
	LegacyRoutingFingerprintSync  = "akademik.fingerprint_sync"
	LegacyRoutingSessionAutoClose = "akademik.session_auto_close"
)

// Bindings memetakan queue consumer role ke routing key yang dilayani module ini.
// Legacy key tetap di-binding agar job lama tetap sampai ke handler yang sama.
var Bindings = []messaging.Binding{
	{Queue: QueueScheduler, RoutingKey: RoutingFingerprintSync},
	{Queue: QueueScheduler, RoutingKey: RoutingSessionAutoClose},
	{Queue: QueueScheduler, RoutingKey: RoutingSessionAutoOpen},
	{Queue: QueueScheduler, RoutingKey: LegacyRoutingFingerprintSync},
	{Queue: QueueScheduler, RoutingKey: LegacyRoutingSessionAutoClose},
}
