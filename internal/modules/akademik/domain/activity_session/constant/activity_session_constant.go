package constant

import "sipon-be/internal/shared/kernel"

type ActivitySessionStatus string

const (
	ActivitySessionStatusScheduled ActivitySessionStatus = "scheduled"
	ActivitySessionStatusOpen      ActivitySessionStatus = "open"
	ActivitySessionStatusCompleted ActivitySessionStatus = "completed"
	ActivitySessionStatusCancelled ActivitySessionStatus = "cancelled"
)

const (
	CodeActivitySessionNotFound          kernel.Code = "ACTIVITY_SESSION_NOT_FOUND"
	CodeActivitySessionInvalidStatus     kernel.Code = "ACTIVITY_SESSION_INVALID_STATUS"
	CodeActivitySessionInvalidTime       kernel.Code = "ACTIVITY_SESSION_INVALID_TIME"
	CodeActivitySessionPersistenceFailed kernel.Code = "ACTIVITY_SESSION_PERSISTENCE_FAILED"
	CodeActivitySessionQueryFailed       kernel.Code = "ACTIVITY_SESSION_QUERY_FAILED"
)
