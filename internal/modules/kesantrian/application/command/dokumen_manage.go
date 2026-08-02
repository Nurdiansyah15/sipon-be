package command

import (
	"context"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	dokumenconstant "sipon-be/internal/modules/kesantrian/domain/dokumen/constant"
	dokumenrepo "sipon-be/internal/modules/kesantrian/domain/dokumen/repository"
	santriconstant "sipon-be/internal/modules/kesantrian/domain/santri/constant"
	santrirepo "sipon-be/internal/modules/kesantrian/domain/santri/repository"
	"sipon-be/internal/shared/kernel"
)

type DokumenDeleteUseCase struct {
	santriRepo   santrirepo.SantriRepository
	dokumenRepo  dokumenrepo.SantriDokumenRepository
	fileUploader ports.FileUploader
	transactor   ports.Transactor
}

func NewDokumenDeleteUseCase(
	santriRepo santrirepo.SantriRepository,
	dokumenRepo dokumenrepo.SantriDokumenRepository,
	fileUploader ports.FileUploader,
	transactor ports.Transactor,
) *DokumenDeleteUseCase {
	return &DokumenDeleteUseCase{
		santriRepo:   santriRepo,
		dokumenRepo:  dokumenRepo,
		fileUploader: fileUploader,
		transactor:   transactor,
	}
}

func (uc *DokumenDeleteUseCase) Execute(ctx context.Context, userID, dokumenID string) (*dto.MessageResponse, error) {
	santri, err := uc.santriRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, application.WrapRepoErr(err, santriconstant.CodeSantriNotFound)
	}

	dokumen, err := uc.dokumenRepo.FindByID(ctx, dokumenID)
	if err != nil {
		return nil, application.WrapRepoErr(err, dokumenconstant.CodeDokumenNotFound)
	}

	if dokumen.SantriID != santri.ID {
		return nil, kernel.New(application.ErrCodeForbidden)
	}

	dokumen.SoftDelete()

	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		return uc.dokumenRepo.Update(txCtx, dokumen)
	}); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	_ = uc.fileUploader.DeleteObject(ctx, dokumen.Key, ports.PrivacyPrivate)

	return &dto.MessageResponse{Message: "dokumen berhasil dihapus"}, nil
}
