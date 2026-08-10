package command

import (
	"context"

	"sipon-be/internal/modules/kesantrian/application"
	tipeconstant "sipon-be/internal/modules/kesantrian/domain/tipe_surat/constant"
	tiperepo "sipon-be/internal/modules/kesantrian/domain/tipe_surat/repository"
	"sipon-be/internal/shared/kernel"
)

type DeleteTipeSuratUseCase struct {
	tipeSuratRepo tiperepo.TipeSuratRepository
}

func NewDeleteTipeSuratUseCase(tipeSuratRepo tiperepo.TipeSuratRepository) *DeleteTipeSuratUseCase {
	return &DeleteTipeSuratUseCase{tipeSuratRepo: tipeSuratRepo}
}

func (uc *DeleteTipeSuratUseCase) Execute(ctx context.Context, id string) error {
	_, err := uc.tipeSuratRepo.FindByID(ctx, id)
	if err != nil {
		return application.WrapRepoErr(err, tipeconstant.CodeTipeSuratNotFound)
	}

	inUse, err := uc.tipeSuratRepo.IsInUse(ctx, id)
	if err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}
	if inUse {
		return kernel.New(application.ErrCodeConflict)
	}

	if err := uc.tipeSuratRepo.Delete(ctx, id); err != nil {
		return application.WrapRepoErr(err, tipeconstant.CodeTipeSuratNotFound)
	}

	return nil
}
