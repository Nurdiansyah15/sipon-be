package command

import (
	"context"

	"sipon-be/internal/modules/psb/application"
	"sipon-be/internal/modules/psb/application/dto"
	dconstant "sipon-be/internal/modules/psb/domain/dokumen/constant"
	drepo "sipon-be/internal/modules/psb/domain/dokumen/repository"
	"sipon-be/internal/shared/kernel"
)

type DokumenVerifyUseCase struct {
	dokumenRepo drepo.PendaftarDokumenRepository
}

func NewDokumenVerifyUseCase(dokumenRepo drepo.PendaftarDokumenRepository) *DokumenVerifyUseCase {
	return &DokumenVerifyUseCase{dokumenRepo: dokumenRepo}
}

func (uc *DokumenVerifyUseCase) Execute(ctx context.Context, verifierID, dokumenID string) (*dto.MessageResponse, error) {
	doc, err := uc.dokumenRepo.FindByID(ctx, dokumenID)
	if err != nil {
		return nil, application.WrapRepoErr(err, dconstant.CodeDokumenNotFound)
	}

	if err := doc.Verify(verifierID); err != nil {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	if err := uc.dokumenRepo.Update(ctx, doc); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "dokumen diverifikasi"}, nil
}

type DokumenRejectUseCase struct {
	dokumenRepo drepo.PendaftarDokumenRepository
}

func NewDokumenRejectUseCase(dokumenRepo drepo.PendaftarDokumenRepository) *DokumenRejectUseCase {
	return &DokumenRejectUseCase{dokumenRepo: dokumenRepo}
}

func (uc *DokumenRejectUseCase) Execute(ctx context.Context, verifierID, dokumenID string, notes *string) (*dto.MessageResponse, error) {
	doc, err := uc.dokumenRepo.FindByID(ctx, dokumenID)
	if err != nil {
		return nil, application.WrapRepoErr(err, dconstant.CodeDokumenNotFound)
	}

	doc.Reject(verifierID, notes)

	if err := uc.dokumenRepo.Update(ctx, doc); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "dokumen ditolak"}, nil
}
