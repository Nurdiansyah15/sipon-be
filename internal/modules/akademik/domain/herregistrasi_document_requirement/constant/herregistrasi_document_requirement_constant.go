package constant

import "sipon-be/internal/shared/kernel"

const (
	CodeHerregistrasiDocumentRequirementNotFound          kernel.Code = "HERREG_DOC_REQUIREMENT_NOT_FOUND"
	CodeHerregistrasiDocumentRequirementDuplicate         kernel.Code = "HERREG_DOC_REQUIREMENT_DUPLICATE"
	CodeHerregistrasiDocumentRequirementInvalidPeriod     kernel.Code = "HERREG_DOC_REQUIREMENT_INVALID_PERIOD"
	CodeHerregistrasiDocumentRequirementInvalidKind       kernel.Code = "HERREG_DOC_REQUIREMENT_INVALID_KIND"
	CodeHerregistrasiDocumentRequirementPersistenceFailed kernel.Code = "HERREG_DOC_REQUIREMENT_PERSISTENCE_FAILED"
	CodeHerregistrasiDocumentRequirementQueryFailed       kernel.Code = "HERREG_DOC_REQUIREMENT_QUERY_FAILED"
)
