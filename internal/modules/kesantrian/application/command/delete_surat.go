package command

import (
	"context"

	"sipon-be/internal/modules/kesantrian/application"
	suratrepo "sipon-be/internal/modules/kesantrian/domain/surat/repository"
)

type DeleteSuratUseCase struct {
	suratRepo suratrepo.SuratRepository
}

func NewDeleteSuratUseCase(suratRepo suratrepo.SuratRepository) *DeleteSuratUseCase {
	return &DeleteSuratUseCase{suratRepo: suratRepo}
}

func (uc *DeleteSuratUseCase) Execute(ctx context.Context, id string) error {
	if err := uc.suratRepo.Delete(ctx, id); err != nil {
		return application.WrapRepoErr(err, "SURAT_NOT_FOUND")
	}
	return nil
}
