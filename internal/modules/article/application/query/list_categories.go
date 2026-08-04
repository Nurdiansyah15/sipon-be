package query

import (
	"context"

	"sipon-be/internal/modules/article/application"
	"sipon-be/internal/modules/article/application/dto"
	articleconst "sipon-be/internal/modules/article/domain/article/constant"
	articlerepo "sipon-be/internal/modules/article/domain/article/repository"
)

type ListCategoriesUseCase struct {
	categoryRepo articlerepo.CategoryRepository
}

func NewListCategoriesUseCase(categoryRepo articlerepo.CategoryRepository) *ListCategoriesUseCase {
	return &ListCategoriesUseCase{categoryRepo: categoryRepo}
}

func (uc *ListCategoriesUseCase) Execute(ctx context.Context, activeOnly bool) ([]dto.CategoryItem, error) {
	cats, err := uc.categoryRepo.FindAll(ctx, activeOnly)
	if err != nil {
		return nil, application.WrapRepoErr(err, articleconst.CodeCategoryQueryFailed)
	}

	items := make([]dto.CategoryItem, 0, len(cats))
	for _, c := range cats {
		items = append(items, dto.CategoryItem{
			ID:        c.ID,
			Name:      c.Name,
			Slug:      c.Slug,
			IsActive:  c.IsActive,
			SortOrder: c.SortOrder,
		})
	}
	return items, nil
}
