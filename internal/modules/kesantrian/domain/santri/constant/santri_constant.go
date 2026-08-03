package constant

import "sipon-be/internal/shared/kernel"

type SantriStatus string

const (
	SantriStatusSantri  SantriStatus = "SANTRI"
	SantriStatusAlumni  SantriStatus = "ALUMNI"
	SantriStatusDropOut SantriStatus = "DROP_OUT"
)

const (
	CodeInvalidNISFormat        kernel.Code = "INVALID_NIS_FORMAT"
	CodeSantriNotFound          kernel.Code = "SANTRI_NOT_FOUND"
	CodeSantriPersistenceFailed kernel.Code = "SANTRI_PERSISTENCE_FAILED"
	CodeSantriQueryFailed       kernel.Code = "SANTRI_QUERY_FAILED"
	CodeSantriDuplicate         kernel.Code = "SANTRI_DUPLICATE"
	CodeSantriInvalidStatus     kernel.Code = "SANTRI_INVALID_STATUS"
)
