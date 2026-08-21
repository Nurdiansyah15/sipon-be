package mq

import (
	"sipon-be/internal/modules/messaging"
	"sipon-be/internal/modules/notification/application/command"
)

const (
	RoutingLoginSucceeded = "identity.user.login_succeeded"
	QueueNotification     = "sipon.worker.notification"

	RoutingAdminBroadcast = command.RoutingAdminBroadcast

	RoutingPsbPendaftaranSubmitted         = "psb.pendaftaran.submitted"
	RoutingPsbDaftarUlangSubmitted         = "psb.daftar_ulang.submitted"
	RoutingPsbDokumenVerified              = "psb.dokumen.verified"
	RoutingPsbDokumenRejected              = "psb.dokumen.rejected"
	RoutingPsbRevisionRequested            = "psb.pendaftaran.revision_requested"
	RoutingPsbRevisionRequestedDaftarUlang = "psb.daftar_ulang.revision_requested"
	RoutingPsbPendaftaranAccepted          = "psb.pendaftaran.accepted"
	RoutingPsbPendaftaranRejected          = "psb.pendaftaran.rejected"
	RoutingPsbNISGenerated                 = "psb.pendaftaran.nis_generated"

	RoutingArticlePublished = "article.published"
	RoutingArticlesScraped  = "article.scraped"

	RoutingKeuanganInvoiceIssued    = "keuangan.invoice.issued"
	RoutingKeuanganInvoiceCancelled = "keuangan.invoice.cancelled"
	RoutingKeuanganPaymentSubmitted = "keuangan.payment.submitted"
	RoutingKeuanganPaymentVerified  = "keuangan.payment.verified"
	RoutingKeuanganPaymentRejected  = "keuangan.payment.rejected"

	RoutingAkademikSessionReminder = "akademik.session.reminder_notify"

	RoutingAkademikAttendanceRecorded = "akademik.attendance.recorded"
)

var Bindings = []messaging.Binding{
	{Queue: QueueNotification, RoutingKey: RoutingLoginSucceeded},
	{Queue: QueueNotification, RoutingKey: RoutingAdminBroadcast},
	{Queue: QueueNotification, RoutingKey: RoutingPsbPendaftaranSubmitted},
	{Queue: QueueNotification, RoutingKey: RoutingPsbDaftarUlangSubmitted},
	{Queue: QueueNotification, RoutingKey: RoutingPsbDokumenVerified},
	{Queue: QueueNotification, RoutingKey: RoutingPsbDokumenRejected},
	{Queue: QueueNotification, RoutingKey: RoutingPsbRevisionRequested},
	{Queue: QueueNotification, RoutingKey: RoutingPsbRevisionRequestedDaftarUlang},
	{Queue: QueueNotification, RoutingKey: RoutingPsbPendaftaranAccepted},
	{Queue: QueueNotification, RoutingKey: RoutingPsbPendaftaranRejected},
	{Queue: QueueNotification, RoutingKey: RoutingPsbNISGenerated},
	{Queue: QueueNotification, RoutingKey: RoutingArticlePublished},
	{Queue: QueueNotification, RoutingKey: RoutingArticlesScraped},
	{Queue: QueueNotification, RoutingKey: RoutingKeuanganInvoiceIssued},
	{Queue: QueueNotification, RoutingKey: RoutingKeuanganInvoiceCancelled},
	{Queue: QueueNotification, RoutingKey: RoutingKeuanganPaymentSubmitted},
	{Queue: QueueNotification, RoutingKey: RoutingKeuanganPaymentVerified},
	{Queue: QueueNotification, RoutingKey: RoutingKeuanganPaymentRejected},
	{Queue: QueueNotification, RoutingKey: RoutingAkademikSessionReminder},
	{Queue: QueueNotification, RoutingKey: RoutingAkademikAttendanceRecorded},
}
