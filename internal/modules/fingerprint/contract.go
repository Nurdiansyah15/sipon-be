package fingerprint

import (
	"context"
	"time"
)

// ScanPin adalah scan pertama per pin dalam rentang waktu, versi contract
// (tanpa SN — SN hanya relevan internal/debug).
type ScanPin struct {
	PIN         string
	FirstScanAt time.Time
}

// Contract diekspos untuk dikonsumsi module lain (akademik). Dirancang
// generik: jika nanti mesin asli menulis ke DB eksternal, implementasi
// internal repository bisa diganti tanpa mengubah kontrak ini.
type Contract interface {
	ListDistinctPinInRange(ctx context.Context, from, to time.Time) ([]ScanPin, error)
}

var _ Contract = (*Module)(nil)
