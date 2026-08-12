package query

import (
	"context"

	"sipon-be/internal/modules/kesantrian/domain/santri/repository"
)

type GetSantriByNISUseCase struct {
	santriRepo repository.SantriRepository
}

func NewGetSantriByNISUseCase(santriRepo repository.SantriRepository) *GetSantriByNISUseCase {
	return &GetSantriByNISUseCase{santriRepo: santriRepo}
}

func (uc *GetSantriByNISUseCase) Execute(ctx context.Context, nis string) (*repository.SantriBasicInfo, error) {
	return uc.santriRepo.FindBasicByNIS(ctx, nis)
}
