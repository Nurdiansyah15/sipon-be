package ports

import (
	"context"
	"time"
)

// FingerprintScanPin adalah scan pertama per pin dalam rentang waktu, dari
// sudut pandang akademik.
type FingerprintScanPin struct {
	PIN         string
	FirstScanAt time.Time
}

// FingerprintReader membaca scan mesin fingerprint via modul fingerprint.
// Implementasi disediakan oleh fingerprintgateway.Gateway.
type FingerprintReader interface {
	ListDistinctPinInRange(ctx context.Context, from, to time.Time) ([]FingerprintScanPin, error)
}
