package query

import (
	"context"
	"testing"
	"time"

	"sipon-be/internal/modules/fingerprint/domain/scanlog/entity"
	"sipon-be/internal/modules/fingerprint/domain/scanlog/repository"
)

// fakeScanLogRepo menyimulasikan dedup DISTINCT ON (pin) yang dilakukan query
// Postgres: per pin hanya scan pertama yang dikembalikan.
type fakeScanLogRepo struct {
	pins []repository.ScanPin
}

func (f *fakeScanLogRepo) Insert(ctx context.Context, log *entity.ScanLog) error { return nil }
func (f *fakeScanLogRepo) ListDistinctPinInRange(ctx context.Context, from, to time.Time) ([]repository.ScanPin, error) {
	return f.pins, nil
}
func (f *fakeScanLogRepo) ListInRange(ctx context.Context, from, to time.Time) ([]*entity.ScanLog, error) {
	return nil, nil
}

func TestListDistinctPinsDedupPerPin(t *testing.T) {
	from := time.Now().Add(-time.Hour)
	to := time.Now()
	uc := NewListDistinctPinsUseCase(&fakeScanLogRepo{pins: []repository.ScanPin{
		{PIN: "NIS001", SN: "DEV-1", FirstScanAt: from.Add(time.Minute)},
		{PIN: "NIS002", SN: "DEV-1", FirstScanAt: from.Add(2 * time.Minute)},
	}})

	pins, err := uc.Execute(context.Background(), from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pins) != 2 {
		t.Fatalf("expected 2 distinct pins, got %d", len(pins))
	}
	if pins[0].PIN != "NIS001" || pins[1].PIN != "NIS002" {
		t.Fatalf("unexpected pins: %+v", pins)
	}
}
