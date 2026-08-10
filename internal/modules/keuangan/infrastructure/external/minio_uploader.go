package external

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	ports "sipon-be/internal/modules/keuangan/application/ports"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// defaultSigningRegion — see identity's own minio_uploader.go for why this
// must be set explicitly (avoids a GetBucketLocation round-trip that would
// hang against an unreachable public endpoint).
const defaultSigningRegion = "us-east-1"

// MinioFileUploader is keuangan's own MinIO adapter instance — a separate Go
// object from other modules', even though all point at the same shared MinIO
// config (each module builds its own infrastructure adapters; only the
// underlying MinIO server/bucket config is shared infra).
type MinioFileUploader struct {
	client        *minio.Client
	presignClient *minio.Client
	bucket        string
	privateBucket string
}

func NewMinioFileUploader(endpoint, publicEndpoint, accessKey, secretKey, bucket, privateBucket string, useSSL bool) (*MinioFileUploader, error) {
	if endpoint == "" || accessKey == "" || secretKey == "" {
		return nil, nil
	}

	if publicEndpoint == "" {
		publicEndpoint = endpoint
	}

	creds := credentials.NewStaticV4(accessKey, secretKey, "")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  creds,
		Secure: useSSL,
		Region: defaultSigningRegion,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	presignClient, err := minio.New(publicEndpoint, &minio.Options{
		Creds:  creds,
		Secure: useSSL,
		Region: defaultSigningRegion,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio presign client: %w", err)
	}

	return &MinioFileUploader{
		client:        client,
		presignClient: presignClient,
		bucket:        bucket,
		privateBucket: privateBucket,
	}, nil
}

func (u *MinioFileUploader) RequestUpload(ctx context.Context, objectName, contentType string, expiry time.Duration, privacy ports.PrivacyRule) (presignURL, key, publicURL string, err error) {
	if u == nil {
		return "", "", "", nil
	}

	bucket := u.bucket
	if privacy == ports.PrivacyPrivate {
		bucket = u.privateBucket
	}

	presignResp, err := u.presignClient.PresignedPutObject(ctx, bucket, objectName, expiry)
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

func (u *MinioFileUploader) DeleteObject(ctx context.Context, key string, privacy ports.PrivacyRule) error {
	if u == nil {
		return nil
	}

	bucket := u.bucket
	if privacy == ports.PrivacyPrivate {
		bucket = u.privateBucket
	}

	return u.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
}

// GeneratePresignedDownloadURL — payment proofs live in the private bucket and
// must be served via short-lived presigned GET URLs, never a bare public URL.
func (u *MinioFileUploader) GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration, privacy ports.PrivacyRule) (string, error) {
	if u == nil {
		return "", nil
	}

	bucket := u.bucket
	if privacy == ports.PrivacyPrivate {
		bucket = u.privateBucket
	}

	presignResp, err := u.presignClient.PresignedGetObject(ctx, bucket, key, expiry, url.Values{})
	if err != nil {
		return "", fmt.Errorf("presign download: %w", err)
	}
	return presignResp.String(), nil
}

func (u *MinioFileUploader) PublicURL(key string) string {
	if u == nil || key == "" {
		return ""
	}

	scheme := "https"
	if !strings.Contains(u.presignClient.EndpointURL().String(), "https") {
		scheme = "http"
	}

	return fmt.Sprintf("%s://%s/%s/%s", scheme, u.presignClient.EndpointURL().Host, u.bucket, key)
}

type NoopFileUploader struct{}

func NewNoopFileUploader() *NoopFileUploader {
	return &NoopFileUploader{}
}

func (n *NoopFileUploader) RequestUpload(ctx context.Context, objectName, contentType string, expiry time.Duration, privacy ports.PrivacyRule) (string, string, string, error) {
	return "", "", "", nil
}

func (n *NoopFileUploader) ConfirmUpload(ctx context.Context, key string) error {
	return nil
}

func (n *NoopFileUploader) DeleteObject(ctx context.Context, key string, privacy ports.PrivacyRule) error {
	return nil
}

func (n *NoopFileUploader) GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration, privacy ports.PrivacyRule) (string, error) {
	return "", nil
}
