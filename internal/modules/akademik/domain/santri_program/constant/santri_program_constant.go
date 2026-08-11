package constant

import "sipon-be/internal/shared/kernel"

const (
	CodeSantriProgramNotFound          kernel.Code = "SANTRI_PROGRAM_NOT_FOUND"
	CodeSantriProgramDuplicate         kernel.Code = "SANTRI_PROGRAM_DUPLICATE"
	CodeSantriProgramAlreadyActive     kernel.Code = "SANTRI_PROGRAM_ALREADY_ACTIVE"
	CodeSantriProgramInvalidProgram    kernel.Code = "SANTRI_PROGRAM_INVALID_PROGRAM"
	CodeSantriProgramPersistenceFailed kernel.Code = "SANTRI_PROGRAM_PERSISTENCE_FAILED"
	CodeSantriProgramQueryFailed       kernel.Code = "SANTRI_PROGRAM_QUERY_FAILED"
)
