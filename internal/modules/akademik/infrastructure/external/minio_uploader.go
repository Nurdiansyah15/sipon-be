package external

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	ports "sipon-be/internal/modules/akademik/application/ports"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
)

const defaultSigningRegion = "us-east-1"

type MinioFileUploader struct {
	client        *minio.Client
	presignClient *minio.Client
	bucket        string
	privateBucket string
}

func NewMinioFileUploader(endpoint, publicEndpoint, accessKey, secretKey, bucket, privateBucket string, useSSL, internalUseSSL bool) (*MinioFileUploader, error) {
	if endpoint == "" || accessKey == "" || secretKey == "" {
		return nil, nil
	}

	if publicEndpoint == "" {
		publicEndpoint = endpoint
	}

	creds := credentials.NewStaticV4(accessKey, secretKey, "")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  creds,
		Secure: internalUseSSL,
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

func (u *MinioFileUploader) PromoteUpload(ctx context.Context, stagingKey, finalKey string, privacy ports.PrivacyRule) error {
	if u == nil {
		return nil
	}
	bucket := u.bucket
	if privacy == ports.PrivacyPrivate {
		bucket = u.privateBucket
	}
	dst := minio.CopyDestOptions{Bucket: bucket, Object: finalKey}
	src := minio.CopySrcOptions{Bucket: bucket, Object: stagingKey}
	if _, err := u.client.CopyObject(ctx, dst, src); err != nil {
		return fmt.Errorf("promote upload copy: %w", err)
	}
	if err := u.client.RemoveObject(ctx, bucket, stagingKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("promote upload cleanup staging: %w", err)
	}
	return nil
}

func (u *MinioFileUploader) EnsurePendingUploadLifecycle(ctx context.Context, expireDays int) error {
	if u == nil {
		return nil
	}
	cfg := lifecycle.NewConfiguration()
	cfg.Rules = []lifecycle.Rule{{
		ID:         "expire-pending-uploads",
		Status:     "Enabled",
		RuleFilter: lifecycle.Filter{Prefix: "pending/"},
		Expiration: lifecycle.Expiration{Days: lifecycle.ExpirationDays(expireDays)},
	}}
	return u.client.SetBucketLifecycle(ctx, u.privateBucket, cfg)
}

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
