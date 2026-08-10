package ports

import (
	"context"
	"time"
)

type PrivacyRule string

const (
	PrivacyPublic  PrivacyRule = "PUBLIC"
	PrivacyPrivate PrivacyRule = "PRIVATE"
)

// FileUploader mirrors kesantrian's ports.FileUploader — needed here because
// payment proofs live in the private bucket and must be served through
// short-lived presigned GET URLs.
type FileUploader interface {
	RequestUpload(ctx context.Context, objectName, contentType string, expiry time.Duration, privacy PrivacyRule) (presignURL, key, publicURL string, err error)
	ConfirmUpload(ctx context.Context, key string) error
	DeleteObject(ctx context.Context, key string, privacy PrivacyRule) error
	GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration, privacy PrivacyRule) (string, error)
}
