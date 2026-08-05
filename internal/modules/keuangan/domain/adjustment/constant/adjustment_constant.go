package constant

import "sipon-be/internal/shared/kernel"

const (
	CodeAdjustmentNotFound         kernel.Code = "ADJUSTMENT_NOT_FOUND"
	CodeAdjustmentPersistenceFailed kernel.Code = "ADJUSTMENT_PERSISTENCE_FAILED"
	CodeAdjustmentQueryFailed      kernel.Code = "ADJUSTMENT_QUERY_FAILED"
)

type AdjustmentType string

const (
	AdjBeasiswa    AdjustmentType = "beasiswa"
	AdjDiskon      AdjustmentType = "diskon"
	AdjPenyesuaian AdjustmentType = "penyesuaian"
)
