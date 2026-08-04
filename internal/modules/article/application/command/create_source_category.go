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

type CreateSourceCategoryUseCase struct {
	categoryRepo articlerepo.SourceCategoryRepository
}

func NewCreateSourceCategoryUseCase(categoryRepo articlerepo.SourceCategoryRepository) *CreateSourceCategoryUseCase {
	return &CreateSourceCategoryUseCase{categoryRepo: categoryRepo}
}

func (uc *CreateSourceCategoryUseCase) Execute(ctx context.Context, sourceID string, req dto.CreateSourceCategoryRequest) (*dto.SourceCategoryMutationResponse, error) {
	cat, err := articleentity.NewSourceCategory(articleentity.SourceCategoryParams{
		ID:                uuid.NewString(),
		SourceID:          sourceID,
		CategoryKey:       req.CategoryKey,
		URLSuffix:         req.URLSuffix,
		URLOverride:       req.URLOverride,
		ArticleLimit:      req.ArticleLimit,
		IsActive:          req.IsActive,
		ArticleCategoryID: req.ArticleCategoryID,
		Keywords:          req.Keywords,
	})
	if err != nil {
		return nil, application.WrapRepoErr(err, articleconst.CodeSourceCategoryKeyRequired)
	}

	if err := uc.categoryRepo.Save(ctx, cat); err != nil {
		return nil, application.WrapConflictErr(err, articleconst.CodeSourceCategoryDuplicate)
	}
	return &dto.SourceCategoryMutationResponse{ID: cat.ID}, nil
}
