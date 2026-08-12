package constant

import "sipon-be/internal/shared/kernel"

type HerregistrasiDocumentStatus string

const (
	HerregistrasiDocumentStatusPending  HerregistrasiDocumentStatus = "pending"
	HerregistrasiDocumentStatusVerified HerregistrasiDocumentStatus = "verified"
	HerregistrasiDocumentStatusRejected HerregistrasiDocumentStatus = "rejected"
)

// AllowedContentTypes adalah whitelist MIME yang boleh di-upload.
var AllowedContentTypes = map[string]bool{
	"image/jpeg":        true,
	"image/png":         true,
	"application/pdf":   true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
}

const (
	CodeHerregistrasiDocumentNotFound          kernel.Code = "HERREG_DOCUMENT_NOT_FOUND"
	CodeHerregistrasiDocumentDuplicate         kernel.Code = "HERREG_DOCUMENT_DUPLICATE"
	CodeHerregistrasiDocumentInvalidStatus     kernel.Code = "HERREG_DOCUMENT_INVALID_STATUS"
	CodeHerregistrasiDocumentInvalidKind       kernel.Code = "HERREG_DOCUMENT_INVALID_KIND"
	CodeHerregistrasiDocumentInvalidRegistration kernel.Code = "HERREG_DOCUMENT_INVALID_REGISTRATION"
	CodeHerregistrasiDocumentPersistenceFailed kernel.Code = "HERREG_DOCUMENT_PERSISTENCE_FAILED"
	CodeHerregistrasiDocumentQueryFailed       kernel.Code = "HERREG_DOCUMENT_QUERY_FAILED"
)
