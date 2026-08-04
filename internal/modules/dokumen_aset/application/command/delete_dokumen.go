package command

import (
	"context"

	"sipon-be/internal/modules/dokumen_aset/application"
	"sipon-be/internal/modules/dokumen_aset/application/dto"
	ports "sipon-be/internal/modules/dokumen_aset/application/ports"
	constant "sipon-be/internal/modules/dokumen_aset/domain/dokumen/constant"
	repo "sipon-be/internal/modules/dokumen_aset/domain/dokumen/repository"
	"sipon-be/internal/shared/kernel"
)

type DeleteDokumenAsetUseCase struct {
	dokumenRepo  repo.DokumenAsetRepository
	fileUploader ports.FileUploader
	transactor   ports.Transactor
}

func NewDeleteDokumenAsetUseCase(
	dokumenRepo repo.DokumenAsetRepository,
	fileUploader ports.FileUploader,
	transactor ports.Transactor,
) *DeleteDokumenAsetUseCase {
	return &DeleteDokumenAsetUseCase{
		dokumenRepo:  dokumenRepo,
		fileUploader: fileUploader,
		transactor:   transactor,
	}
}

func (uc *DeleteDokumenAsetUseCase) Execute(ctx context.Context, id string) (*dto.MessageResponse, error) {
	dokumen, err := uc.dokumenRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeDokumenNotFound)
	}

	privacy := ports.PrivacyPublic
	if !dokumen.IsPublic {
		privacy = ports.PrivacyPrivate
	}

	if err := uc.fileUploader.DeleteObject(ctx, dokumen.Key, privacy); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	dokumen.SoftDelete()
	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		return uc.dokumenRepo.Update(txCtx, dokumen)
	}); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "dokumen berhasil dihapus"}, nil
}
