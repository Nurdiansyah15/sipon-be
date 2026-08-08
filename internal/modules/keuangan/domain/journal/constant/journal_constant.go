package constant

import "sipon-be/internal/shared/kernel"

const (
	CodeJournalNotFound        kernel.Code = "JOURNAL_NOT_FOUND"
	CodeJournalNotBalanced     kernel.Code = "JOURNAL_NOT_BALANCED"
	CodeJournalMinLines        kernel.Code = "JOURNAL_MIN_LINES"
	CodeJournalPeriodClosed    kernel.Code = "JOURNAL_PERIOD_CLOSED"
	CodeJournalAutoCannotCancel kernel.Code = "JOURNAL_AUTO_CANNOT_CANCEL"
	CodeJournalInvalidStatus   kernel.Code = "JOURNAL_INVALID_STATUS"
	CodeJournalPersistenceFailed kernel.Code = "JOURNAL_PERSISTENCE_FAILED"
	CodeJournalQueryFailed     kernel.Code = "JOURNAL_QUERY_FAILED"
	CodeJournalAccountMappingNotFound kernel.Code = "JOURNAL_ACCOUNT_MAPPING_NOT_FOUND"
)

type JournalStatus string

const (
	JournalDraft     JournalStatus = "draft"
	JournalPosted    JournalStatus = "posted"
	JournalCancelled JournalStatus = "cancelled"
)

type SourceType string

const (
	SourceInvoiceIssued    SourceType = "invoice_issued"
	SourcePaymentVerified  SourceType = "payment_verified"
	SourceInvoiceCancelled SourceType = "invoice_cancelled"
	SourceAdjustment       SourceType = "adjustment"
	SourceClosing          SourceType = "closing"
	SourceManual           SourceType = "manual"
)
