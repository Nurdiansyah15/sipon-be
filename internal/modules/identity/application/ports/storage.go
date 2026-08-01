package ports

import (
	"context"
	"time"
)

type FileUploader interface {
	RequestUpload(ctx context.Context, objectName, contentType string, expiry time.Duration, privacy PrivacyRule) (presignURL, key, publicURL string, err error)
	ConfirmUpload(ctx context.Context, key string) error
	MarkDeleted(ctx context.Context, key string) error
	DeleteObject(ctx context.Context, key string, privacy PrivacyRule) error
	PublicURL(key string) string
	KeyFromURL(url string) string
}

type PrivacyRule string

const (
	PrivacyPublic  PrivacyRule = "PUBLIC"
	PrivacyPrivate PrivacyRule = "PRIVATE"
)
