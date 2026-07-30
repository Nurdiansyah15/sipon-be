package external

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"sipon-be/internal/modules/identity/application"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioFileUploader struct {
	client        *minio.Client
	bucket        string
	privateBucket string
}

func NewMinioFileUploader(endpoint, accessKey, secretKey, bucket, privateBucket string, useSSL bool) (*MinioFileUploader, error) {
	if endpoint == "" || accessKey == "" || secretKey == "" {
		return nil, nil
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	return &MinioFileUploader{
		client:        client,
		bucket:        bucket,
		privateBucket: privateBucket,
	}, nil
}

func (u *MinioFileUploader) RequestUpload(ctx context.Context, objectName, contentType string, expiry time.Duration, privacy application.PrivacyRule) (presignURL, key, publicURL string, err error) {
	if u == nil {
		return "", "", "", nil
	}

	bucket := u.bucket
	if privacy == application.PrivacyPrivate {
		bucket = u.privateBucket
	}

	presignResp, err := u.client.PresignedPutObject(ctx, bucket, objectName, expiry)
	if err != nil {
		return "", "", "", fmt.Errorf("presign upload: %w", err)
	}

	publicURL = u.PublicURL(objectName)

	return presignResp.String(), objectName, publicURL, nil
}

func (u *MinioFileUploader) ConfirmUpload(ctx context.Context, key string) error {
	if u == nil {
		return nil
	}

	_, err := u.client.StatObject(ctx, u.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		_, err = u.client.StatObject(ctx, u.privateBucket, key, minio.StatObjectOptions{})
		if err != nil {
			return fmt.Errorf("confirm upload: %w", err)
		}
	}
	return nil
}

func (u *MinioFileUploader) MarkDeleted(ctx context.Context, key string) error {
	if u == nil {
		return nil
	}

	err1 := u.client.RemoveObject(ctx, u.bucket, key, minio.RemoveObjectOptions{})
	err2 := u.client.RemoveObject(ctx, u.privateBucket, key, minio.RemoveObjectOptions{})
	if err1 != nil && err2 != nil {
		return fmt.Errorf("mark deleted: %w", err1)
	}
	return nil
}

func (u *MinioFileUploader) DeleteObject(ctx context.Context, key string, privacy application.PrivacyRule) error {
	if u == nil {
		return nil
	}

	bucket := u.bucket
	if privacy == application.PrivacyPrivate {
		bucket = u.privateBucket
	}

	return u.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
}

func (u *MinioFileUploader) PublicURL(key string) string {
	if u == nil || key == "" {
		return ""
	}

	scheme := "https"
	if !strings.Contains(u.client.EndpointURL().String(), "https") {
		scheme = "http"
	}

	return fmt.Sprintf("%s://%s/%s/%s", scheme, u.client.EndpointURL().Host, u.bucket, key)
}

func (u *MinioFileUploader) KeyFromURL(raw string) string {
	if raw == "" {
		return ""
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	parts := strings.SplitN(strings.TrimPrefix(parsed.Path, "/"), "/", 2)
	if len(parts) == 2 {
		return parts[1]
	}

	return ""
}

type NoopFileUploader struct{}

func NewNoopFileUploader() *NoopFileUploader {
	return &NoopFileUploader{}
}

func (n *NoopFileUploader) RequestUpload(ctx context.Context, objectName, contentType string, expiry time.Duration, privacy application.PrivacyRule) (string, string, string, error) {
	return "", "", "", nil
}

func (n *NoopFileUploader) ConfirmUpload(ctx context.Context, key string) error {
	return nil
}

func (n *NoopFileUploader) MarkDeleted(ctx context.Context, key string) error {
	return nil
}

func (n *NoopFileUploader) DeleteObject(ctx context.Context, key string, privacy application.PrivacyRule) error {
	return nil
}

func (n *NoopFileUploader) PublicURL(key string) string {
	return ""
}

func (n *NoopFileUploader) KeyFromURL(raw string) string {
	return ""
}
