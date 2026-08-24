package query

import (
	"context"

	"sipon-be/internal/modules/psb/application"
	prepo "sipon-be/internal/modules/psb/domain/pendaftar/repository"
	"sipon-be/internal/modules/psb/domain/quota"
	srepo "sipon-be/internal/modules/psb/domain/setting/repository"
	"sipon-be/internal/shared/kernel"
)

// QuotaUsageItem adalah pemakaian & sisa kuota untuk satu kombinasi program ×
// gender. Limit/Remaining nil berarti kombinasi tidak dibatasi (unlimited).
type QuotaUsageItem struct {
	Key       string `json:"key"`
	Program   string `json:"program"`
	Gender    string `json:"gender"`
	Limit     *int   `json:"limit"`
	Used      int64  `json:"used"`
	Remaining *int   `json:"remaining"`
}

type QuotaUsageUseCase struct {
	settingRepo   srepo.PsbSettingRepository
	pendaftarRepo prepo.PendaftarRepository
}

func NewQuotaUsageUseCase(settingRepo srepo.PsbSettingRepository, pendaftarRepo prepo.PendaftarRepository) *QuotaUsageUseCase {
	return &QuotaUsageUseCase{settingRepo: settingRepo, pendaftarRepo: pendaftarRepo}
}

// Execute menghitung pemakaian & sisa kuota untuk satu setting (biasanya yang
// berjalan) dengan merekalkulasi langsung dari data pendaftar.
func (uc *QuotaUsageUseCase) Execute(ctx context.Context, settingID string) ([]QuotaUsageItem, error) {
	setting, err := uc.settingRepo.FindByID(ctx, settingID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	q, err := quota.Parse(setting.Quota)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	usage, err := uc.pendaftarRepo.QuotaUsage(ctx, settingID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := make([]QuotaUsageItem, 0, len(quota.Combinations))
	for _, combo := range quota.Combinations {
		used := usage[combo.Key]
		item := QuotaUsageItem{
			Key:     combo.Key,
			Program: combo.ProgramCode,
			Gender:  combo.Gender,
			Used:    used,
		}
		if remaining, ok := q.Remaining(combo.Key, used); ok {
			item.Remaining = remaining
		}
		if limit, ok := q.Limit(combo.Key); ok {
			item.Limit = &limit
		}
		items = append(items, item)
	}
	return items, nil
}

// ExecuteActive menghitung pemakaian & sisa kuota untuk setting yang sedang
// berjalan (aktif) — dipakai endpoint publik agar pendaftar tahu sisa kuota.
func (uc *QuotaUsageUseCase) ExecuteActive(ctx context.Context) ([]QuotaUsageItem, error) {
	setting, err := uc.settingRepo.FindActive(ctx)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}
	return uc.Execute(ctx, setting.ID)
}