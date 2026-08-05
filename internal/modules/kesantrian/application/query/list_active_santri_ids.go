package query

import (
	"context"

	"sipon-be/internal/modules/kesantrian/domain/santri/repository"
)

type ListActiveSantriIDsUseCase struct {
	santriRepo repository.SantriRepository
}

func NewListActiveSantriIDsUseCase(santriRepo repository.SantriRepository) *ListActiveSantriIDsUseCase {
	return &ListActiveSantriIDsUseCase{santriRepo: santriRepo}
}

func (uc *ListActiveSantriIDsUseCase) Execute(ctx context.Context) ([]string, error) {
	return uc.santriRepo.ListActiveIDs(ctx)
}
