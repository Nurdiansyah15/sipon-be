package query

import (
	"context"
	"time"

	"sipon-be/internal/modules/fingerprint/application"
	"sipon-be/internal/modules/fingerprint/application/dto"
	"sipon-be/internal/modules/fingerprint/domain/scanlog/repository"
	"sipon-be/internal/shared/kernel"
)

// ListDistinctPinsUseCase (get scan info) mengembalikan scan pertama per pin
// dalam rentang waktu. Dipakai oleh HTTP handler (debug) dan diekspos lewat
// Contract untuk dikonsumsi module lain (akademik).
type ListDistinctPinsUseCase struct {
	repo repository.ScanLogRepository
}

func NewListDistinctPinsUseCase(repo repository.ScanLogRepository) *ListDistinctPinsUseCase {
	return &ListDistinctPinsUseCase{repo: repo}
}

func (uc *ListDistinctPinsUseCase) Execute(ctx context.Context, from, to time.Time) ([]dto.ScanPin, error) {
	pins, err := uc.repo.ListDistinctPinInRange(ctx, from, to)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	result := make([]dto.ScanPin, 0, len(pins))
	for _, p := range pins {
		result = append(result, dto.ScanPin{PIN: p.PIN, SN: p.SN, FirstScanAt: p.FirstScanAt})
	}
	return result, nil
}
