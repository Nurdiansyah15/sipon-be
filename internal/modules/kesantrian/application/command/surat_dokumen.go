package command

import (
	"context"
	"errors"
	"time"

	"sipon-be/internal/modules/kesantrian/application"
	suratconstant "sipon-be/internal/modules/kesantrian/domain/surat/constant"
	"sipon-be/internal/modules/kesantrian/domain/surat/entity"
	suratrepo "sipon-be/internal/modules/kesantrian/domain/surat/repository"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type AddSuratDokumenUseCase struct {
	suratRepo suratrepo.SuratRepository
}

func NewAddSuratDokumenUseCase(suratRepo suratrepo.SuratRepository) *AddSuratDokumenUseCase {
	return &AddSuratDokumenUseCase{suratRepo: suratRepo}
}

func (uc *AddSuratDokumenUseCase) Execute(ctx context.Context, suratID, dokumenAsetID string) error {
	_, err := uc.suratRepo.FindByID(ctx, suratID)
	if err != nil {
		return application.WrapRepoErr(err, suratconstant.CodeSuratNotFound)
	}

	link := &entity.SuratDokumenAset{
		ID:            uuid.NewString(),
		SuratID:       suratID,
		DokumenAsetID: dokumenAsetID,
		CreatedAt:     time.Now(),
	}

	if err := uc.suratRepo.SaveDokumenLink(ctx, link); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == suratconstant.CodeSuratDokumenExists {
			return kernel.Wrap(application.ErrCodeConflict, err)
		}
		return kernel.Wrap(application.ErrCodeInternal, err)
	}

	return nil
}

type RemoveSuratDokumenUseCase struct {
	suratRepo suratrepo.SuratRepository
}

func NewRemoveSuratDokumenUseCase(suratRepo suratrepo.SuratRepository) *RemoveSuratDokumenUseCase {
	return &RemoveSuratDokumenUseCase{suratRepo: suratRepo}
}

func (uc *RemoveSuratDokumenUseCase) Execute(ctx context.Context, suratID, dokumenAsetID string) error {
	if err := uc.suratRepo.DeleteDokumenLink(ctx, suratID, dokumenAsetID); err != nil {
		return application.WrapRepoErr(err, suratconstant.CodeSuratDokumenNotFound)
	}
	return nil
}
