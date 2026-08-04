package external

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const defaultSigningRegion = "us-east-1"

type MinioFileUploader struct {
	client        *minio.Client
	presignClient *minio.Client
	bucket        string
}

func NewMinioFileUploader(endpoint, publicEndpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinioFileUploader, error) {
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
	}, nil
}

func (u *MinioFileUploader) RequestUpload(ctx context.Context, objectName, contentType string, expiry time.Duration) (presignURL, key, publicURL string, err error) {
	if u == nil {
		return "", "", "", nil
	}

	presignResp, err := u.presignClient.PresignedPutObject(ctx, u.bucket, objectName, expiry)
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
		return fmt.Errorf("confirm upload: %w", err)
	}
	return nil
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

func (n *NoopFileUploader) RequestUpload(ctx context.Context, objectName, contentType string, expiry time.Duration) (string, string, string, error) {
	return "", "", "", nil
}

func (n *NoopFileUploader) ConfirmUpload(ctx context.Context, key string) error {
	return nil
}

func (n *NoopFileUploader) PublicURL(key string) string {
	return ""
}
