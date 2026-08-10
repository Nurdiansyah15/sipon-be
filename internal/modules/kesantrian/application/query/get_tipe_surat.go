package query

import (
	"context"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	tipeconstant "sipon-be/internal/modules/kesantrian/domain/tipe_surat/constant"
	tiperepo "sipon-be/internal/modules/kesantrian/domain/tipe_surat/repository"
)

type GetTipeSuratUseCase struct {
	tipeSuratRepo tiperepo.TipeSuratRepository
}

func NewGetTipeSuratUseCase(tipeSuratRepo tiperepo.TipeSuratRepository) *GetTipeSuratUseCase {
	return &GetTipeSuratUseCase{tipeSuratRepo: tipeSuratRepo}
}

func (uc *GetTipeSuratUseCase) Execute(ctx context.Context, id string) (*dto.TipeSuratResponse, error) {
	ts, err := uc.tipeSuratRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, tipeconstant.CodeTipeSuratNotFound)
	}

	return &dto.TipeSuratResponse{
		ID:        ts.ID,
		Nama:      ts.Nama,
		Kode:      ts.Kode,
		CreatedBy: ts.CreatedBy,
		CreatedAt: ts.CreatedAt,
		UpdatedAt: ts.UpdatedAt,
	}, nil
}
