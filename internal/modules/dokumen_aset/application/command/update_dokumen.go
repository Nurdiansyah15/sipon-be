package command

import (
	"context"

	"sipon-be/internal/modules/dokumen_aset/application"
	"sipon-be/internal/modules/dokumen_aset/application/dto"
	constant "sipon-be/internal/modules/dokumen_aset/domain/dokumen/constant"
	repo "sipon-be/internal/modules/dokumen_aset/domain/dokumen/repository"
	"sipon-be/internal/shared/kernel"
)

type UpdateDokumenAsetUseCase struct {
	dokumenRepo repo.DokumenAsetRepository
}

func NewUpdateDokumenAsetUseCase(dokumenRepo repo.DokumenAsetRepository) *UpdateDokumenAsetUseCase {
	return &UpdateDokumenAsetUseCase{dokumenRepo: dokumenRepo}
}

func (uc *UpdateDokumenAsetUseCase) Execute(ctx context.Context, id string, req dto.DokumenAsetUpdateRequest) (*dto.MessageResponse, error) {
	dokumen, err := uc.dokumenRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeDokumenNotFound)
	}

	var kategori *constant.Kategori
	if req.Kategori != nil {
		k := constant.Kategori(*req.Kategori)
		kategori = &k
	}

	if err := dokumen.UpdateMetadata(req.Judul, req.Deskripsi, kategori, req.IsPublic); err != nil {
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	if err := uc.dokumenRepo.Update(ctx, dokumen); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "dokumen berhasil diperbarui"}, nil
}
