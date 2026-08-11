package constant

import "sipon-be/internal/shared/kernel"

type AcademicPeriodStatus string

const (
	AcademicPeriodStatusDraft    AcademicPeriodStatus = "draft"
	AcademicPeriodStatusOpen     AcademicPeriodStatus = "open"
	AcademicPeriodStatusClosed   AcademicPeriodStatus = "closed"
	AcademicPeriodStatusArchived AcademicPeriodStatus = "archived"
)

const (
	CodeAcademicPeriodNotFound          kernel.Code = "ACADEMIC_PERIOD_NOT_FOUND"
	CodeAcademicPeriodDuplicate         kernel.Code = "ACADEMIC_PERIOD_DUPLICATE"
	CodeAcademicPeriodInvalidRange      kernel.Code = "ACADEMIC_PERIOD_INVALID_RANGE"
	CodeAcademicPeriodInvalidStatus     kernel.Code = "ACADEMIC_PERIOD_INVALID_STATUS"
	CodeAcademicPeriodPersistenceFailed kernel.Code = "ACADEMIC_PERIOD_PERSISTENCE_FAILED"
	CodeAcademicPeriodQueryFailed       kernel.Code = "ACADEMIC_PERIOD_QUERY_FAILED"
)
