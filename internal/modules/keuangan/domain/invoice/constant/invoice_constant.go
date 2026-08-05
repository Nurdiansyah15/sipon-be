package constant

import "sipon-be/internal/shared/kernel"

const (
	CodeInvoiceNotFound         kernel.Code = "INVOICE_NOT_FOUND"
	CodeInvoiceDuplicate        kernel.Code = "INVOICE_DUPLICATE"
	CodeInvoiceInvalidStatus    kernel.Code = "INVOICE_INVALID_STATUS"
	CodeInvoiceAlreadyIssued    kernel.Code = "INVOICE_ALREADY_ISSUED"
	CodeInvoiceAlreadyPaid      kernel.Code = "INVOICE_ALREADY_PAID"
	CodeInvoicePersistenceFailed kernel.Code = "INVOICE_PERSISTENCE_FAILED"
	CodeInvoiceQueryFailed      kernel.Code = "INVOICE_QUERY_FAILED"
)

type InvoiceStatus string

const (
	StatusDraft     InvoiceStatus = "draft"
	StatusIssued    InvoiceStatus = "issued"
	StatusPartial   InvoiceStatus = "partial"
	StatusPaid      InvoiceStatus = "paid"
	StatusExpired   InvoiceStatus = "expired"
	StatusCancelled InvoiceStatus = "cancelled"
)
