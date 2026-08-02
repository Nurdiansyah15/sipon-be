package constant

import "sipon-be/internal/shared/kernel"

type SantriRequestStatus string

const (
	SantriRequestStatusPending  SantriRequestStatus = "pending"
	SantriRequestStatusApproved SantriRequestStatus = "approved"
	SantriRequestStatusRejected SantriRequestStatus = "rejected"
)

const (
	CodeSantriRequestNotFound          kernel.Code = "SANTRI_REQUEST_NOT_FOUND"
	CodeSantriRequestPersistenceFailed kernel.Code = "SANTRI_REQUEST_PERSISTENCE_FAILED"
	CodeSantriRequestQueryFailed       kernel.Code = "SANTRI_REQUEST_QUERY_FAILED"
	CodeSantriRequestInvalidStatus     kernel.Code = "SANTRI_REQUEST_INVALID_STATUS"
	CodeSantriRequestAlreadyExists     kernel.Code = "SANTRI_REQUEST_ALREADY_EXISTS"
)
