package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/kesantrian/application"
	tipeconstant "sipon-be/internal/modules/kesantrian/domain/tipe_surat/constant"
	"sipon-be/internal/modules/kesantrian/domain/tipe_surat/entity"
	tiperepo "sipon-be/internal/modules/kesantrian/domain/tipe_surat/repository"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type CreateTipeSuratUseCase struct {
	tipeSuratRepo tiperepo.TipeSuratRepository
}

func NewCreateTipeSuratUseCase(tipeSuratRepo tiperepo.TipeSuratRepository) *CreateTipeSuratUseCase {
	return &CreateTipeSuratUseCase{tipeSuratRepo: tipeSuratRepo}
}

func (uc *CreateTipeSuratUseCase) Execute(ctx context.Context, createdBy *string, nama, kode string) (*entity.TipeSurat, error) {
	ts, err := entity.NewTipeSurat(uuid.NewString(), nama, kode, createdBy)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeBadRequest, err)
	}

	if err := uc.tipeSuratRepo.Save(ctx, ts); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == tipeconstant.CodeTipeSuratKodeDuplicate {
			return nil, kernel.Wrap(application.ErrCodeConflict, err)
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return ts, nil
}
