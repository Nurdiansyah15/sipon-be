package mq

import "sipon-be/internal/modules/messaging"

const (
	RoutingLoginSucceeded = "identity.user.login_succeeded"
	QueueNotification     = "sipon.worker.notification"

	RoutingPsbPendaftaranSubmitted         = "psb.pendaftaran.submitted"
	RoutingPsbDaftarUlangSubmitted         = "psb.daftar_ulang.submitted"
	RoutingPsbDokumenVerified              = "psb.dokumen.verified"
	RoutingPsbDokumenRejected              = "psb.dokumen.rejected"
	RoutingPsbRevisionRequested            = "psb.pendaftaran.revision_requested"
	RoutingPsbRevisionRequestedDaftarUlang = "psb.daftar_ulang.revision_requested"
	RoutingPsbPendaftaranAccepted          = "psb.pendaftaran.accepted"
	RoutingPsbPendaftaranRejected          = "psb.pendaftaran.rejected"
	RoutingPsbNISGenerated                 = "psb.pendaftaran.nis_generated"
)

var Bindings = []messaging.Binding{
	{Queue: QueueNotification, RoutingKey: RoutingLoginSucceeded},
	{Queue: QueueNotification, RoutingKey: RoutingPsbPendaftaranSubmitted},
	{Queue: QueueNotification, RoutingKey: RoutingPsbDaftarUlangSubmitted},
	{Queue: QueueNotification, RoutingKey: RoutingPsbDokumenVerified},
	{Queue: QueueNotification, RoutingKey: RoutingPsbDokumenRejected},
	{Queue: QueueNotification, RoutingKey: RoutingPsbRevisionRequested},
	{Queue: QueueNotification, RoutingKey: RoutingPsbRevisionRequestedDaftarUlang},
	{Queue: QueueNotification, RoutingKey: RoutingPsbPendaftaranAccepted},
	{Queue: QueueNotification, RoutingKey: RoutingPsbPendaftaranRejected},
	{Queue: QueueNotification, RoutingKey: RoutingPsbNISGenerated},
}
