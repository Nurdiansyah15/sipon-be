package command

import (
	"context"

	"sipon-be/internal/modules/psb/application"
	"sipon-be/internal/modules/psb/application/dto"
	drepo "sipon-be/internal/modules/psb/domain/dokumen/repository"
	prepo "sipon-be/internal/modules/psb/domain/pendaftar/repository"
	rrepo "sipon-be/internal/modules/psb/domain/review/repository"
	srepo "sipon-be/internal/modules/psb/domain/setting/repository"
	"sipon-be/internal/shared/kernel"
)

type PurgePeriodUseCase struct {
	settingRepo srepo.PsbSettingRepository
	pendaftarRepo prepo.PendaftarRepository
	dokumenRepo   drepo.PendaftarDokumenRepository
	reviewRepo    rrepo.PendaftarReviewRepository
}

func NewPurgePeriodUseCase(
	settingRepo srepo.PsbSettingRepository,
	pendaftarRepo prepo.PendaftarRepository,
	dokumenRepo drepo.PendaftarDokumenRepository,
	reviewRepo rrepo.PendaftarReviewRepository,
) *PurgePeriodUseCase {
	return &PurgePeriodUseCase{
		settingRepo:   settingRepo,
		pendaftarRepo: pendaftarRepo,
		dokumenRepo:   dokumenRepo,
		reviewRepo:    reviewRepo,
	}
}

func (uc *PurgePeriodUseCase) Execute(ctx context.Context, settingID string) (*dto.MessageResponse, error) {
	s, err := uc.settingRepo.FindByID(ctx, settingID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	if !s.CanPurge() {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	_, err = uc.pendaftarRepo.HardDeleteBySettingID(ctx, settingID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	s.MarkPurged()
	if err := uc.settingRepo.Update(ctx, s); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "data periode berhasil dihapus"}, nil
}
