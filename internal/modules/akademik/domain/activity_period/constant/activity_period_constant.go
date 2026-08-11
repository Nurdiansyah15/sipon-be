package constant

import "sipon-be/internal/shared/kernel"

type ActivityPeriodStatus string

const (
	ActivityPeriodStatusActive   ActivityPeriodStatus = "active"
	ActivityPeriodStatusInactive ActivityPeriodStatus = "inactive"
)

const (
	CodeActivityPeriodNotFound          kernel.Code = "ACTIVITY_PERIOD_NOT_FOUND"
	CodeActivityPeriodDuplicate         kernel.Code = "ACTIVITY_PERIOD_DUPLICATE"
	CodeActivityPeriodInvalidStatus     kernel.Code = "ACTIVITY_PERIOD_INVALID_STATUS"
	CodeActivityPeriodPersistenceFailed kernel.Code = "ACTIVITY_PERIOD_PERSISTENCE_FAILED"
	CodeActivityPeriodQueryFailed       kernel.Code = "ACTIVITY_PERIOD_QUERY_FAILED"
)
