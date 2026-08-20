package command

// Routing key untuk event keuangan yang dipublish ke outbox, dikonsumsi oleh
// modul notification untuk mengirim notifikasi in-app/push.
const (
	RoutingInvoiceIssued       = "keuangan.invoice.issued"
	RoutingInvoiceBatchCreated = "keuangan.invoice.batch_created"
	RoutingInvoiceCancelled    = "keuangan.invoice.cancelled"
	RoutingPaymentSubmitted    = "keuangan.payment.submitted"
	RoutingPaymentVerified     = "keuangan.payment.verified"
	RoutingPaymentRejected     = "keuangan.payment.rejected"
)
