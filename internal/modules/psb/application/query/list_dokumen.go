package query

import (
	"context"

	"sipon-be/internal/modules/psb/application"
	"sipon-be/internal/modules/psb/application/dto"
	dentity "sipon-be/internal/modules/psb/domain/dokumen/entity"
	drepo "sipon-be/internal/modules/psb/domain/dokumen/repository"
	pconstant "sipon-be/internal/modules/psb/domain/pendaftar/constant"
	prepo "sipon-be/internal/modules/psb/domain/pendaftar/repository"
	srepo "sipon-be/internal/modules/psb/domain/setting/repository"
	"sipon-be/internal/shared/kernel"
)

type DokumenListUseCase struct {
	pendaftarRepo prepo.PendaftarRepository
	settingRepo   srepo.PsbSettingRepository
	dokumenRepo   drepo.PendaftarDokumenRepository
}

func NewDokumenListUseCase(
	pendaftarRepo prepo.PendaftarRepository,
	settingRepo srepo.PsbSettingRepository,
	dokumenRepo drepo.PendaftarDokumenRepository,
) *DokumenListUseCase {
	return &DokumenListUseCase{
		pendaftarRepo: pendaftarRepo,
		settingRepo:   settingRepo,
		dokumenRepo:   dokumenRepo,
	}
}

func (uc *DokumenListUseCase) Execute(ctx context.Context, userID string) ([]dto.DokumenItemResponse, error) {
	setting, err := uc.settingRepo.FindActive(ctx)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	p, err := uc.pendaftarRepo.FindByUserIDAndSetting(ctx, userID, setting.ID)
	if err != nil {
		return nil, application.WrapRepoErr(err, pconstant.CodePendaftarNotFound)
	}

	docs, err := uc.dokumenRepo.FindByPendaftarID(ctx, p.ID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := make([]dto.DokumenItemResponse, 0, len(docs))
	for _, d := range docs {
		if d.DeletedAt != nil {
			continue
		}
		items = append(items, mapDokumenToResponse(d))
	}

	return items, nil
}

func (uc *DokumenListUseCase) ExecuteByPendaftarID(ctx context.Context, pendaftarID string) ([]dto.DokumenItemResponse, error) {
	docs, err := uc.dokumenRepo.FindByPendaftarID(ctx, pendaftarID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := make([]dto.DokumenItemResponse, 0, len(docs))
	for _, d := range docs {
		if d.DeletedAt != nil {
			continue
		}
		items = append(items, mapDokumenToResponse(d))
	}

	return items, nil
}

func mapDokumenToResponse(d *dentity.PendaftarDokumen) dto.DokumenItemResponse {
	return dto.DokumenItemResponse{
		ID:               d.ID,
		Stage:            string(d.Stage),
		Kind:             string(d.Kind),
		Status:           string(d.Status),
		OriginalFilename: d.OriginalFilename,
		MimeType:         d.MimeType,
		Size:             d.Size,
		Notes:            d.Notes,
		VerifiedBy:       d.VerifiedBy,
		VerifiedAt:       d.VerifiedAt,
		CreatedAt:        d.CreatedAt,
	}
}
