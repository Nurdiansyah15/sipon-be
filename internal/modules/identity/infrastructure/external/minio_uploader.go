package external

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	ports "sipon-be/internal/modules/identity/application/ports"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// defaultSigningRegion dipasang eksplisit di kedua client supaya minio-go
// tidak melakukan GetBucketLocation (round-trip network) sebelum presign.
// Tanpa ini, presignClient yang endpoint-nya publik (kerap tak terjangkau
// dari dalam container) akan hang menunggu response yang tak pernah datang.
const defaultSigningRegion = "us-east-1"

type MinioFileUploader struct {
	// client memakai endpoint internal (docker network) — dipakai untuk
	// operasi yang butuh koneksi nyata ke MinIO: StatObject, RemoveObject.
	client *minio.Client
	// presignClient memakai endpoint publik (dijangkau browser/FE) — dipakai
	// murni untuk menandatangani presigned URL & membangun PublicURL. Berkat
	// Region eksplisit, client ini tidak pernah benar-benar melakukan koneksi
	// jaringan, jadi endpoint-nya boleh tidak terjangkau dari container.
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

// KeyFromURL menormalisasi input avatar confirm ke object key kanonis.
// Endpoint confirm menerima balik nilai `key` yang sama persis dengan yang
// dikembalikan presign (object key polos, mis. "avatars/uuid.jpg") — BUKAN
// URL. Jika raw sudah berupa key polos (tidak ada skema), kembalikan apa
// adanya; jangan asumsikan "segmen pertama = nama bucket" seperti sebelumnya,
// karena itu salah memotong prefix "avatars/" dan hanya menyisakan UUID-nya.
// Hanya kalau raw benar-benar URL (presigned/public URL), potong bucket-nya.
func (u *MinioFileUploader) KeyFromURL(raw string) string {
	if u == nil || raw == "" {
		return ""
	}

	if !strings.Contains(raw, "://") {
		return strings.TrimPrefix(raw, "/")
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	trimmed := strings.TrimPrefix(parsed.Path, "/")
	for _, bucket := range []string{u.bucket, u.privateBucket} {
		if rest, ok := strings.CutPrefix(trimmed, bucket+"/"); ok {
			return rest
		}
	}

	return trimmed
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

func (n *NoopFileUploader) MarkDeleted(ctx context.Context, key string) error {
	return nil
}

func (n *NoopFileUploader) DeleteObject(ctx context.Context, key string, privacy ports.PrivacyRule) error {
	return nil
}

func (n *NoopFileUploader) PublicURL(key string) string {
	return ""
}

func (n *NoopFileUploader) KeyFromURL(raw string) string {
	return ""
}
