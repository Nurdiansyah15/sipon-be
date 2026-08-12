package repository

import (
	"context"
	"time"

	"sipon-be/internal/modules/fingerprint/domain/scanlog/entity"
)

// ScanPin adalah scan pertama (per pin) dalam rentang waktu tertentu.
type ScanPin struct {
	PIN         string
	SN          string
	FirstScanAt time.Time
}

// ScanLogRepository dirancang generik: selama mesin asli menulis dengan skema
// yang sama (sn, scan_date, pin, verifymode, inoutmode, deviceip), cukup ganti
// implementasi repository ini tanpa mengubah use case di atasnya.
type ScanLogRepository interface {
	Insert(ctx context.Context, log *entity.ScanLog) error
	// ListDistinctPinInRange mengembalikan scan pertama per pin dalam
	// [from, to] — dedup dilakukan di query
	// (SELECT DISTINCT ON (pin) ... ORDER BY pin, scan_date).
	ListDistinctPinInRange(ctx context.Context, from, to time.Time) ([]ScanPin, error)
	// ListInRange mengembalikan semua scan mentah dalam [from, to] — untuk
	// debug/inspeksi.
	ListInRange(ctx context.Context, from, to time.Time) ([]*entity.ScanLog, error)
}
