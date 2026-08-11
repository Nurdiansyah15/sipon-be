package constant

import "sipon-be/internal/shared/kernel"

type ActivityStatus string

const (
	ActivityStatusActive   ActivityStatus = "active"
	ActivityStatusInactive ActivityStatus = "inactive"
)

const (
	CodeActivityNotFound          kernel.Code = "ACTIVITY_NOT_FOUND"
	CodeActivityDuplicate         kernel.Code = "ACTIVITY_DUPLICATE"
	CodeActivityPersistenceFailed kernel.Code = "ACTIVITY_PERSISTENCE_FAILED"
	CodeActivityQueryFailed       kernel.Code = "ACTIVITY_QUERY_FAILED"
	CodeActivityInvalidStatus     kernel.Code = "ACTIVITY_INVALID_STATUS"
)
