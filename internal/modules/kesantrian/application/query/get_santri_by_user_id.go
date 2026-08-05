package query

import (
	"context"

	"sipon-be/internal/modules/kesantrian/domain/santri/repository"
)

type GetSantriByUserIDUseCase struct {
	santriRepo repository.SantriRepository
}

func NewGetSantriByUserIDUseCase(santriRepo repository.SantriRepository) *GetSantriByUserIDUseCase {
	return &GetSantriByUserIDUseCase{santriRepo: santriRepo}
}

func (uc *GetSantriByUserIDUseCase) Execute(ctx context.Context, userID string) (*repository.SantriBasicInfo, error) {
	return uc.santriRepo.FindBasicByUserID(ctx, userID)
}
