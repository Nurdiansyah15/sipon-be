package command

import (
	"context"

	"sipon-be/internal/modules/article/application"
	"sipon-be/internal/modules/article/application/dto"
	articleconst "sipon-be/internal/modules/article/domain/article/constant"
	articleentity "sipon-be/internal/modules/article/domain/article/entity"
	articlerepo "sipon-be/internal/modules/article/domain/article/repository"

	"github.com/google/uuid"
)

type CreateCategoryUseCase struct {
	categoryRepo articlerepo.CategoryRepository
}

func NewCreateCategoryUseCase(categoryRepo articlerepo.CategoryRepository) *CreateCategoryUseCase {
	return &CreateCategoryUseCase{categoryRepo: categoryRepo}
}

func (uc *CreateCategoryUseCase) Execute(ctx context.Context, req dto.CreateCategoryRequest, userID string) (*dto.CategoryMutationResponse, error) {
	cat, err := articleentity.NewCategory(uuid.NewString(), req.Name, req.Slug, &userID)
	if err != nil {
		return nil, application.WrapRepoErr(err, articleconst.CodeCategoryNameRequired)
	}
	if req.SortOrder != nil {
		cat.SortOrder = *req.SortOrder
	}

	if err := uc.categoryRepo.Save(ctx, cat); err != nil {
		return nil, application.WrapConflictErr(err, articleconst.CodeCategoryDuplicateSlug)
	}
	return &dto.CategoryMutationResponse{ID: cat.ID}, nil
}
