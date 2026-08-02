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

// FileUploader mirrors identity's ports.FileUploader shape, plus
// GeneratePresignedDownloadURL — needed here because santri documents live
// in the private bucket and must be served through short-lived presigned
// GET URLs rather than a public URL.
type FileUploader interface {
	RequestUpload(ctx context.Context, objectName, contentType string, expiry time.Duration, privacy PrivacyRule) (presignURL, key, publicURL string, err error)
	ConfirmUpload(ctx context.Context, key string) error
	DeleteObject(ctx context.Context, key string, privacy PrivacyRule) error
	GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration, privacy PrivacyRule) (string, error)
	// PublicURL builds a public URL for a key in the shared public bucket —
	// used only to render a santri's avatar (owned/uploaded via identity's
	// own avatar flow, but the public bucket itself is shared infra, not
	// identity-exclusive).
	PublicURL(key string) string
}
