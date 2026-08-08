package constant

import "sipon-be/internal/shared/kernel"

const (
	CodeBillingPeriodNotFound         kernel.Code = "BILLING_PERIOD_NOT_FOUND"
	CodeBillingPeriodInvalidStatus    kernel.Code = "BILLING_PERIOD_INVALID_STATUS"
	CodeBillingPeriodPersistenceFailed kernel.Code = "BILLING_PERIOD_PERSISTENCE_FAILED"
	CodeBillingPeriodQueryFailed      kernel.Code = "BILLING_PERIOD_QUERY_FAILED"
)

type BillingPeriodStatus string

const (
	BillingPeriodDraft  BillingPeriodStatus = "draft"
	BillingPeriodOpen   BillingPeriodStatus = "open"
	BillingPeriodClosed BillingPeriodStatus = "closed"
)
