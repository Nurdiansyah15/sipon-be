package constant

import "sipon-be/internal/shared/kernel"

type ProgramStatus string

const (
	ProgramStatusActive   ProgramStatus = "active"
	ProgramStatusInactive ProgramStatus = "inactive"
)

// ProgramCode are predefined/default codes for the master program table.
// They are NOT a DB enum — new programs can be added later via the table.
const (
	ProgramCodeTahfidz string = "TAHFIDZ"
	ProgramCodeKitab   string = "KITAB"
)

const (
	CodeProgramNotFound          kernel.Code = "PROGRAM_NOT_FOUND"
	CodeProgramDuplicate         kernel.Code = "PROGRAM_DUPLICATE"
	CodeProgramPersistenceFailed kernel.Code = "PROGRAM_PERSISTENCE_FAILED"
	CodeProgramQueryFailed       kernel.Code = "PROGRAM_QUERY_FAILED"
	CodeProgramInvalidStatus     kernel.Code = "PROGRAM_INVALID_STATUS"
)
