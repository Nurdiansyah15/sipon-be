package ports

import (
	"context"
	"time"
)

type FileUploader interface {
	RequestUpload(ctx context.Context, objectName, contentType string, expiry time.Duration) (presignURL, key, publicURL string, err error)
	ConfirmUpload(ctx context.Context, key string) error
	PublicURL(key string) string
}
