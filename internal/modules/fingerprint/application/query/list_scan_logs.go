package query

import (
	"context"
	"time"

	"sipon-be/internal/modules/fingerprint/application"
	"sipon-be/internal/modules/fingerprint/application/dto"
	"sipon-be/internal/modules/fingerprint/domain/scanlog/repository"
	"sipon-be/internal/shared/kernel"
)

// ListScanLogsUseCase mengembalikan seluruh scan mentah dalam rentang waktu —
// untuk debug/inspeksi (endpoint GET /scans).
type ListScanLogsUseCase struct {
	repo repository.ScanLogRepository
}

func NewListScanLogsUseCase(repo repository.ScanLogRepository) *ListScanLogsUseCase {
	return &ListScanLogsUseCase{repo: repo}
}

func (uc *ListScanLogsUseCase) Execute(ctx context.Context, from, to time.Time) ([]dto.ScanLogResponse, error) {
	logs, err := uc.repo.ListInRange(ctx, from, to)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	result := make([]dto.ScanLogResponse, 0, len(logs))
	for _, l := range logs {
		result = append(result, dto.ScanLogResponse{
			ID:         l.ID,
			SN:         l.SN,
			ScanDate:   l.ScanDate,
			PIN:        l.PIN,
			VerifyMode: l.VerifyMode,
			InOutMode:  l.InOutMode,
			DeviceIP:   l.DeviceIP,
			CreatedAt:  l.CreatedAt,
		})
	}
	return result, nil
}
