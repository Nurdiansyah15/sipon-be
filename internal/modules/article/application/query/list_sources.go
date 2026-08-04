package query

import (
	"context"

	"sipon-be/internal/modules/article/application"
	"sipon-be/internal/modules/article/application/dto"
	articleconst "sipon-be/internal/modules/article/domain/article/constant"
	articlerepo "sipon-be/internal/modules/article/domain/article/repository"
)

type ListSourcesUseCase struct {
	sourceRepo    articlerepo.SourceRepository
	selectorRepo  articlerepo.SourceSelectorRepository
	categoryRepo  articlerepo.SourceCategoryRepository
}

func NewListSourcesUseCase(
	sourceRepo articlerepo.SourceRepository,
	selectorRepo articlerepo.SourceSelectorRepository,
	categoryRepo articlerepo.SourceCategoryRepository,
) *ListSourcesUseCase {
	return &ListSourcesUseCase{
		sourceRepo:   sourceRepo,
		selectorRepo: selectorRepo,
		categoryRepo: categoryRepo,
	}
}

func (uc *ListSourcesUseCase) Execute(ctx context.Context) ([]dto.SourceListItem, error) {
	sources, err := uc.sourceRepo.FindAll(ctx)
	if err != nil {
		return nil, application.WrapRepoErr(err, articleconst.CodeSourceQueryFailed)
	}

	items := make([]dto.SourceListItem, 0, len(sources))
	for _, s := range sources {
		item := dto.SourceListItem{
			ID:            s.ID,
			Key:           s.Key,
			Name:          s.Name,
			BaseURL:       s.BaseURL,
			AutoPublish:   s.AutoPublish,
			IsActive:      s.IsActive,
			LastScrapedAt: s.LastScrapedAt,
			CreatedAt:     s.CreatedAt,
			Categories:    []dto.SourceCategoryItem{},
		}

		sel, _ := uc.selectorRepo.FindBySourceID(ctx, s.ID)
		if sel != nil {
			item.Selectors = &dto.SourceSelectorItem{
				ContentSelector: sel.ContentSelector,
				AuthorSelector:  sel.AuthorSelector,
				TagsSelector:    sel.TagsSelector,
			}
		}

		cats, _ := uc.categoryRepo.FindAllBySource(ctx, s.ID)
		for _, c := range cats {
			item.Categories = append(item.Categories, dto.SourceCategoryItem{
				ID:                c.ID,
				CategoryKey:       c.CategoryKey,
				URLSuffix:         c.URLSuffix,
				URLOverride:       c.URLOverride,
				ArticleLimit:      c.ArticleLimit,
				IsActive:          c.IsActive,
				ArticleCategoryID: c.ArticleCategoryID,
				Keywords:          c.Keywords,
				LastScrapedAt:     c.LastScrapedAt,
			})
		}

		items = append(items, item)
	}
	return items, nil
}
