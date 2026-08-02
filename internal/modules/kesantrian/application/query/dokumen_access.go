package query

import (
	"context"
	"time"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	dokumenconstant "sipon-be/internal/modules/kesantrian/domain/dokumen/constant"
	dokumenrepo "sipon-be/internal/modules/kesantrian/domain/dokumen/repository"
	santriconstant "sipon-be/internal/modules/kesantrian/domain/santri/constant"
	santrirepo "sipon-be/internal/modules/kesantrian/domain/santri/repository"
	"sipon-be/internal/shared/kernel"
)

const dokumenAccessTTL = 15 * time.Minute

type DokumenAccessUseCase struct {
	santriRepo   santrirepo.SantriRepository
	dokumenRepo  dokumenrepo.SantriDokumenRepository
	fileUploader ports.FileUploader
}

func NewDokumenAccessUseCase(santriRepo santrirepo.SantriRepository, dokumenRepo dokumenrepo.SantriDokumenRepository, fileUploader ports.FileUploader) *DokumenAccessUseCase {
	return &DokumenAccessUseCase{santriRepo: santriRepo, dokumenRepo: dokumenRepo, fileUploader: fileUploader}
}

func (uc *DokumenAccessUseCase) Execute(ctx context.Context, userID, dokumenID string) (*dto.DokumenAccessResponse, error) {
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

	url, err := uc.fileUploader.GeneratePresignedDownloadURL(ctx, dokumen.Key, dokumenAccessTTL, ports.PrivacyPrivate)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.DokumenAccessResponse{
		AccessURL: url,
		ExpiresIn: int(dokumenAccessTTL.Seconds()),
	}, nil
}
