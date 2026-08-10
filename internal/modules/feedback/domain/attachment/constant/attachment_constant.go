package constant

import "sipon-be/internal/shared/kernel"

const (
	MaxAttachmentsPerFeedback = 5

	CodeAttachmentNotFound          kernel.Code = "ATTACHMENT_NOT_FOUND"
	CodeAttachmentPersistenceFailed kernel.Code = "ATTACHMENT_PERSISTENCE_FAILED"
	CodeAttachmentQueryFailed       kernel.Code = "ATTACHMENT_QUERY_FAILED"
	CodeAttachmentLimitExceeded     kernel.Code = "ATTACHMENT_LIMIT_EXCEEDED"
)

var AllowedContentTypes = map[string]bool{
	"image/jpeg":         true,
	"image/png":          true,
	"image/gif":          true,
	"image/webp":         true,
	"application/pdf":    true,
	"text/plain":         true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
}
