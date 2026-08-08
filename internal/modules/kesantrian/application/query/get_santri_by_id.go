package query

import (
	"context"

	"sipon-be/internal/modules/kesantrian/domain/santri/repository"
)

type GetSantriByIDUseCase struct {
	santriRepo repository.SantriRepository
}

func NewGetSantriByIDUseCase(santriRepo repository.SantriRepository) *GetSantriByIDUseCase {
	return &GetSantriByIDUseCase{santriRepo: santriRepo}
}

func (uc *GetSantriByIDUseCase) Execute(ctx context.Context, santriID string) (*repository.SantriBasicInfo, error) {
	return uc.santriRepo.FindBasicByID(ctx, santriID)
}
