package constant

import "sipon-be/internal/shared/kernel"

const (
	CodePeriodNotFound        kernel.Code = "PERIOD_NOT_FOUND"
	CodePeriodOverlap         kernel.Code = "PERIOD_OVERLAP"
	CodePeriodInvalidStatus   kernel.Code = "PERIOD_INVALID_STATUS"
	CodePeriodPersistenceFailed kernel.Code = "PERIOD_PERSISTENCE_FAILED"
	CodePeriodQueryFailed     kernel.Code = "PERIOD_QUERY_FAILED"
	CodePeriodClosingAccountMissing kernel.Code = "PERIOD_CLOSING_ACCOUNT_MISSING"
)

type PeriodStatus string

const (
	PeriodOpen   PeriodStatus = "open"
	PeriodClosed PeriodStatus = "closed"
	PeriodLocked PeriodStatus = "locked"
)
