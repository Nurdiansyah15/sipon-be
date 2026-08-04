package command

import (
	"context"

	"sipon-be/internal/modules/article/application"
	articleconst "sipon-be/internal/modules/article/domain/article/constant"
	articlerepo "sipon-be/internal/modules/article/domain/article/repository"
)

type DeleteSourceUseCase struct {
	sourceRepo   articlerepo.SourceRepository
}

func NewDeleteSourceUseCase(sourceRepo articlerepo.SourceRepository) *DeleteSourceUseCase {
	return &DeleteSourceUseCase{sourceRepo: sourceRepo}
}

func (uc *DeleteSourceUseCase) Execute(ctx context.Context, sourceID string) error {
	if _, err := uc.sourceRepo.FindByID(ctx, sourceID); err != nil {
		return application.WrapRepoErr(err, articleconst.CodeSourceNotFound)
	}
	if err := uc.sourceRepo.Delete(ctx, sourceID); err != nil {
		return application.WrapRepoErr(err, articleconst.CodeSourcePersistenceFailed)
	}
	return nil
}
