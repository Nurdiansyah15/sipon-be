package constant

import "sipon-be/internal/shared/kernel"

type SantriRegistrationStatus string

const (
	SantriRegistrationStatusDraft     SantriRegistrationStatus = "draft"
	SantriRegistrationStatusPending   SantriRegistrationStatus = "pending"
	SantriRegistrationStatusRevision  SantriRegistrationStatus = "revision"
	SantriRegistrationStatusCompleted SantriRegistrationStatus = "completed"
	SantriRegistrationStatusCancelled SantriRegistrationStatus = "cancelled"
)

const (
	CodeSantriRegistrationNotFound          kernel.Code = "SANTRI_REGISTRATION_NOT_FOUND"
	CodeSantriRegistrationDuplicate         kernel.Code = "SANTRI_REGISTRATION_DUPLICATE"
	CodeSantriRegistrationInvalidStatus     kernel.Code = "SANTRI_REGISTRATION_INVALID_STATUS"
	CodeSantriRegistrationInvalidSantri     kernel.Code = "SANTRI_REGISTRATION_INVALID_SANTRI"
	CodeSantriRegistrationInvalidPeriod     kernel.Code = "SANTRI_REGISTRATION_INVALID_PERIOD"
	CodeSantriRegistrationMissingDocuments  kernel.Code = "SANTRI_REGISTRATION_MISSING_DOCUMENTS"
	CodeSantriRegistrationPersistenceFailed kernel.Code = "SANTRI_REGISTRATION_PERSISTENCE_FAILED"
	CodeSantriRegistrationQueryFailed       kernel.Code = "SANTRI_REGISTRATION_QUERY_FAILED"
)
