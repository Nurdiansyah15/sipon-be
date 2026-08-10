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

type FileUploader interface {
	RequestUpload(ctx context.Context, objectName, contentType string, expiry time.Duration, privacy PrivacyRule) (presignURL, key, publicURL string, err error)
	ConfirmUpload(ctx context.Context, key string) error
	DeleteObject(ctx context.Context, key string, privacy PrivacyRule) error
	PromoteUpload(ctx context.Context, stagingKey, finalKey string, privacy PrivacyRule) error
	EnsurePendingUploadLifecycle(ctx context.Context, expireDays int) error
	GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration, privacy PrivacyRule) (string, error)
	PublicURL(key string) string
}
