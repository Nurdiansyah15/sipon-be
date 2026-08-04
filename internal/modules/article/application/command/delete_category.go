package command

import (
	"context"

	"sipon-be/internal/modules/article/application"
	articleconst "sipon-be/internal/modules/article/domain/article/constant"
	articlerepo "sipon-be/internal/modules/article/domain/article/repository"
)

type DeleteCategoryUseCase struct {
	categoryRepo articlerepo.CategoryRepository
}

func NewDeleteCategoryUseCase(categoryRepo articlerepo.CategoryRepository) *DeleteCategoryUseCase {
	return &DeleteCategoryUseCase{categoryRepo: categoryRepo}
}

func (uc *DeleteCategoryUseCase) Execute(ctx context.Context, categoryID string) error {
	if _, err := uc.categoryRepo.FindByID(ctx, categoryID); err != nil {
		return application.WrapRepoErr(err, articleconst.CodeCategoryNotFound)
	}
	if err := uc.categoryRepo.Delete(ctx, categoryID); err != nil {
		return application.WrapRepoErr(err, articleconst.CodeCategoryPersistenceFailed)
	}
	return nil
}
