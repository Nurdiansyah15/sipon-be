package query

import (
	"context"

	"sipon-be/internal/modules/dokumen_aset/application"
	"sipon-be/internal/modules/dokumen_aset/application/dto"
	ports "sipon-be/internal/modules/dokumen_aset/application/ports"
	constant "sipon-be/internal/modules/dokumen_aset/domain/dokumen/constant"
	repo "sipon-be/internal/modules/dokumen_aset/domain/dokumen/repository"
	"sipon-be/internal/shared/kernel"
)

type GetDokumenAsetUseCase struct {
	dokumenRepo repo.DokumenAsetRepository
}

func NewGetDokumenAsetUseCase(dokumenRepo repo.DokumenAsetRepository) *GetDokumenAsetUseCase {
	return &GetDokumenAsetUseCase{dokumenRepo: dokumenRepo}
}

func (uc *GetDokumenAsetUseCase) Execute(ctx context.Context, id string, fileUploader ports.FileUploader, isAuthenticated bool) (*dto.DokumenAsetDetail, error) {
	dokumen, err := uc.dokumenRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeDokumenNotFound)
	}

	if !dokumen.IsPublic && !isAuthenticated {
		return nil, kernel.New(application.ErrCodeNotFound)
	}

	detail := &dto.DokumenAsetDetail{
		DokumenAsetItem: dto.DokumenAsetItem{
			ID:        dokumen.ID,
			Judul:     dokumen.Judul,
			Deskripsi: dokumen.Deskripsi,
			Kategori:  string(dokumen.Kategori),
			Filename:  dokumen.Filename,
			MimeType:  dokumen.MimeType,
			Size:      dokumen.Size,
			IsPublic:  dokumen.IsPublic,
			CreatedBy: dokumen.CreatedBy,
			CreatedAt: dokumen.CreatedAt,
			UpdatedAt: dokumen.UpdatedAt,
		},
	}

	return detail, nil
}
