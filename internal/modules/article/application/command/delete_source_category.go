package command

import (
	"context"

	"sipon-be/internal/modules/article/application"
	articleconst "sipon-be/internal/modules/article/domain/article/constant"
	articlerepo "sipon-be/internal/modules/article/domain/article/repository"
)

type DeleteSourceCategoryUseCase struct {
	categoryRepo articlerepo.SourceCategoryRepository
}

func NewDeleteSourceCategoryUseCase(categoryRepo articlerepo.SourceCategoryRepository) *DeleteSourceCategoryUseCase {
	return &DeleteSourceCategoryUseCase{categoryRepo: categoryRepo}
}

func (uc *DeleteSourceCategoryUseCase) Execute(ctx context.Context, categoryID string) error {
	if _, err := uc.categoryRepo.FindByID(ctx, categoryID); err != nil {
		return application.WrapRepoErr(err, articleconst.CodeSourceCategoryNotFound)
	}
	if err := uc.categoryRepo.Delete(ctx, categoryID); err != nil {
		return application.WrapRepoErr(err, articleconst.CodeSourceCategoryPersistenceFailed)
	}
	return nil
}
