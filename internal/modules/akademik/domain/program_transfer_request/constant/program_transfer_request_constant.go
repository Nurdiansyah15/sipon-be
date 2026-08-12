package constant

import "sipon-be/internal/shared/kernel"

type ProgramTransferRequestStatus string

const (
	ProgramTransferRequestStatusPending  ProgramTransferRequestStatus = "pending"
	ProgramTransferRequestStatusApproved ProgramTransferRequestStatus = "approved"
	ProgramTransferRequestStatusRejected ProgramTransferRequestStatus = "rejected"
)

const (
	CodeProgramTransferRequestNotFound          kernel.Code = "PROGRAM_TRANSFER_REQUEST_NOT_FOUND"
	CodeProgramTransferRequestInvalidStatus     kernel.Code = "PROGRAM_TRANSFER_REQUEST_INVALID_STATUS"
	CodeProgramTransferRequestDuplicate         kernel.Code = "PROGRAM_TRANSFER_REQUEST_DUPLICATE"
	CodeProgramTransferRequestSameProgram       kernel.Code = "PROGRAM_TRANSFER_REQUEST_SAME_PROGRAM"
	CodeProgramTransferRequestPersistenceFailed kernel.Code = "PROGRAM_TRANSFER_REQUEST_PERSISTENCE_FAILED"
	CodeProgramTransferRequestQueryFailed       kernel.Code = "PROGRAM_TRANSFER_REQUEST_QUERY_FAILED"
)
