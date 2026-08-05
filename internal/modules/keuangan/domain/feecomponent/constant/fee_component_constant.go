package constant

import "sipon-be/internal/shared/kernel"

const (
	CodeFeeComponentNotFound         kernel.Code = "FEE_COMPONENT_NOT_FOUND"
	CodeFeeComponentDuplicate        kernel.Code = "FEE_COMPONENT_DUPLICATE"
	CodeFeeComponentInvalidType      kernel.Code = "FEE_COMPONENT_INVALID_TYPE"
	CodeFeeComponentPersistenceFailed kernel.Code = "FEE_COMPONENT_PERSISTENCE_FAILED"
	CodeFeeComponentQueryFailed      kernel.Code = "FEE_COMPONENT_QUERY_FAILED"
)

type FeeComponentType string

const (
	FeeTypeUKT         FeeComponentType = "ukt"
	FeeTypeSPP         FeeComponentType = "spp"
	FeeTypeDaftarUlang FeeComponentType = "daftar_ulang"
	FeeTypeInsidental  FeeComponentType = "insidental"
)

type PeriodType string

const (
	PeriodMonthly    PeriodType = "monthly"
	PeriodSemesterly PeriodType = "semesterly"
	PeriodYearly     PeriodType = "yearly"
	PeriodOnce       PeriodType = "once"
)

var ValidFeeTypes = map[FeeComponentType]bool{
	FeeTypeUKT:         true,
	FeeTypeSPP:         true,
	FeeTypeDaftarUlang: true,
	FeeTypeInsidental:  true,
}

func IsValidFeeType(t FeeComponentType) bool {
	return ValidFeeTypes[t]
}
