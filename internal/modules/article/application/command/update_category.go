package command

import (
	"context"

	"sipon-be/internal/modules/article/application"
	"sipon-be/internal/modules/article/application/dto"
	articleconst "sipon-be/internal/modules/article/domain/article/constant"
	articlerepo "sipon-be/internal/modules/article/domain/article/repository"
)

type UpdateCategoryUseCase struct {
	categoryRepo articlerepo.CategoryRepository
}

func NewUpdateCategoryUseCase(categoryRepo articlerepo.CategoryRepository) *UpdateCategoryUseCase {
	return &UpdateCategoryUseCase{categoryRepo: categoryRepo}
}

func (uc *UpdateCategoryUseCase) Execute(ctx context.Context, categoryID, userID string, req dto.UpdateCategoryRequest) (*dto.CategoryMutationResponse, error) {
	cat, err := uc.categoryRepo.FindByID(ctx, categoryID)
	if err != nil {
		return nil, application.WrapRepoErr(err, articleconst.CodeCategoryNotFound)
	}

	if req.Name != nil {
		cat.Name = *req.Name
	}
	if req.Slug != nil {
		cat.Slug = *req.Slug
	}
	if req.IsActive != nil {
		cat.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		cat.SortOrder = *req.SortOrder
	}
	cat.UpdatedBy = &userID

	if err := uc.categoryRepo.Update(ctx, cat); err != nil {
		return nil, application.WrapConflictErr(err, articleconst.CodeCategoryDuplicateSlug)
	}
	return &dto.CategoryMutationResponse{ID: categoryID}, nil
}
