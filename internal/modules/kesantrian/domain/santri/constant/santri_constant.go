package constant

import "sipon-be/internal/shared/kernel"

const (
	CodeInvalidNISFormat        kernel.Code = "INVALID_NIS_FORMAT"
	CodeSantriNotFound          kernel.Code = "SANTRI_NOT_FOUND"
	CodeSantriPersistenceFailed kernel.Code = "SANTRI_PERSISTENCE_FAILED"
	CodeSantriQueryFailed       kernel.Code = "SANTRI_QUERY_FAILED"
	CodeSantriDuplicate         kernel.Code = "SANTRI_DUPLICATE"
)
