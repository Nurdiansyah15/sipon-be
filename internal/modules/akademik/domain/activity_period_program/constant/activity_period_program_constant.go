package constant

import "sipon-be/internal/shared/kernel"

const (
	CodeActivityPeriodProgramNotFound          kernel.Code = "ACTIVITY_PERIOD_PROGRAM_NOT_FOUND"
	CodeActivityPeriodProgramDuplicate         kernel.Code = "ACTIVITY_PERIOD_PROGRAM_DUPLICATE"
	CodeActivityPeriodProgramInvalid           kernel.Code = "ACTIVITY_PERIOD_PROGRAM_INVALID"
	CodeActivityPeriodProgramPersistenceFailed kernel.Code = "ACTIVITY_PERIOD_PROGRAM_PERSISTENCE_FAILED"
	CodeActivityPeriodProgramQueryFailed       kernel.Code = "ACTIVITY_PERIOD_PROGRAM_QUERY_FAILED"
)
