package query

import (
	"context"

	"sipon-be/internal/modules/kesantrian/domain/santri/repository"
)

type ListActiveSantriWithUserIDUseCase struct {
	santriRepo repository.SantriRepository
}

func NewListActiveSantriWithUserIDUseCase(santriRepo repository.SantriRepository) *ListActiveSantriWithUserIDUseCase {
	return &ListActiveSantriWithUserIDUseCase{santriRepo: santriRepo}
}

func (uc *ListActiveSantriWithUserIDUseCase) Execute(ctx context.Context) ([]repository.SantriBasicInfo, error) {
	return uc.santriRepo.ListActiveWithUserID(ctx)
}
