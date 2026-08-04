package query

import (
	"context"
	"time"

	"sipon-be/internal/modules/dokumen_aset/application"
	"sipon-be/internal/modules/dokumen_aset/application/dto"
	ports "sipon-be/internal/modules/dokumen_aset/application/ports"
	constant "sipon-be/internal/modules/dokumen_aset/domain/dokumen/constant"
	repo "sipon-be/internal/modules/dokumen_aset/domain/dokumen/repository"
	"sipon-be/internal/shared/kernel"
)

const dokumenAsetDownloadTTL = 5 * time.Minute

type DownloadDokumenAsetUseCase struct {
	dokumenRepo  repo.DokumenAsetRepository
	fileUploader ports.FileUploader
}

func NewDownloadDokumenAsetUseCase(dokumenRepo repo.DokumenAsetRepository, fileUploader ports.FileUploader) *DownloadDokumenAsetUseCase {
	return &DownloadDokumenAsetUseCase{dokumenRepo: dokumenRepo, fileUploader: fileUploader}
}

func (uc *DownloadDokumenAsetUseCase) Execute(ctx context.Context, id string, isAuthenticated bool) (*dto.DokumenAsetDownloadResponse, error) {
	dokumen, err := uc.dokumenRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeDokumenNotFound)
	}

	if !dokumen.IsPublic && !isAuthenticated {
		return nil, kernel.New(application.ErrCodeForbidden)
	}

	privacy := ports.PrivacyPublic
	if !dokumen.IsPublic {
		privacy = ports.PrivacyPrivate
	}

	url, err := uc.fileUploader.GeneratePresignedDownloadURL(ctx, dokumen.Key, dokumenAsetDownloadTTL, privacy)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.DokumenAsetDownloadResponse{
		AccessURL: url,
		ExpiresIn: int(dokumenAsetDownloadTTL.Seconds()),
	}, nil
}
