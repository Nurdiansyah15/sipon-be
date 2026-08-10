package constant

import "sipon-be/internal/shared/kernel"

const (
	CodeFeeComponentNotFound          kernel.Code = "FEE_COMPONENT_NOT_FOUND"
	CodeFeeComponentDuplicate         kernel.Code = "FEE_COMPONENT_DUPLICATE"
	CodeFeeComponentPersistenceFailed kernel.Code = "FEE_COMPONENT_PERSISTENCE_FAILED"
	CodeFeeComponentQueryFailed       kernel.Code = "FEE_COMPONENT_QUERY_FAILED"
)

type PeriodType string

const (
	PeriodMonthly    PeriodType = "monthly"
	PeriodSemesterly PeriodType = "semesterly"
	PeriodYearly     PeriodType = "yearly"
	PeriodOnce       PeriodType = "once"
)
