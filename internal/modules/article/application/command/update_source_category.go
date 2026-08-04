package command

import (
	"context"
	"time"

	"sipon-be/internal/modules/article/application"
	"sipon-be/internal/modules/article/application/dto"
	articleconst "sipon-be/internal/modules/article/domain/article/constant"
	articlerepo "sipon-be/internal/modules/article/domain/article/repository"
)

type UpdateSourceCategoryUseCase struct {
	categoryRepo articlerepo.SourceCategoryRepository
}

func NewUpdateSourceCategoryUseCase(categoryRepo articlerepo.SourceCategoryRepository) *UpdateSourceCategoryUseCase {
	return &UpdateSourceCategoryUseCase{categoryRepo: categoryRepo}
}

func (uc *UpdateSourceCategoryUseCase) Execute(ctx context.Context, categoryID, userID string, req dto.UpdateSourceCategoryRequest) (*dto.SourceCategoryMutationResponse, error) {
	cat, err := uc.categoryRepo.FindByID(ctx, categoryID)
	if err != nil {
		return nil, application.WrapRepoErr(err, articleconst.CodeSourceCategoryNotFound)
	}

	if req.CategoryKey != nil {
		cat.CategoryKey = *req.CategoryKey
	}
	if req.URLSuffix != nil {
		cat.URLSuffix = req.URLSuffix
	}
	if req.URLOverride != nil {
		cat.URLOverride = req.URLOverride
	}
	if req.ArticleLimit != nil {
		cat.ArticleLimit = *req.ArticleLimit
	}
	if req.IsActive != nil {
		cat.IsActive = *req.IsActive
	}
	if req.ArticleCategoryID != nil {
		cat.ArticleCategoryID = req.ArticleCategoryID
	}
	if req.Keywords != nil {
		cat.Keywords = *req.Keywords
	}
	cat.UpdatedBy = &userID
	cat.UpdatedAt = time.Now()

	if err := uc.categoryRepo.Update(ctx, cat); err != nil {
		return nil, application.WrapConflictErr(err, articleconst.CodeSourceCategoryDuplicate)
	}
	return &dto.SourceCategoryMutationResponse{ID: categoryID}, nil
}
