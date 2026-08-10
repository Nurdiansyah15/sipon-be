package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/kesantrian/application"
	tipeconstant "sipon-be/internal/modules/kesantrian/domain/tipe_surat/constant"
	tiperepo "sipon-be/internal/modules/kesantrian/domain/tipe_surat/repository"
	"sipon-be/internal/shared/kernel"
)

type UpdateTipeSuratUseCase struct {
	tipeSuratRepo tiperepo.TipeSuratRepository
}

func NewUpdateTipeSuratUseCase(tipeSuratRepo tiperepo.TipeSuratRepository) *UpdateTipeSuratUseCase {
	return &UpdateTipeSuratUseCase{tipeSuratRepo: tipeSuratRepo}
}

func (uc *UpdateTipeSuratUseCase) Execute(ctx context.Context, id, nama, kode string) error {
	ts, err := uc.tipeSuratRepo.FindByID(ctx, id)
	if err != nil {
		return application.WrapRepoErr(err, tipeconstant.CodeTipeSuratNotFound)
	}

	inUse, err := uc.tipeSuratRepo.IsInUse(ctx, id)
	if err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}

	if inUse {
		if ts.Kode != kode {
			return kernel.New(application.ErrCodeConflict)
		}
		ts.UpdateNama(nama)
	} else {
		ts.Update(nama, kode)
	}

	if err := uc.tipeSuratRepo.Update(ctx, ts); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == tipeconstant.CodeTipeSuratKodeDuplicate {
			return kernel.Wrap(application.ErrCodeConflict, err)
		}
		return kernel.Wrap(application.ErrCodeInternal, err)
	}

	return nil
}
