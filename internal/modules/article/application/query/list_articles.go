package query

import (
	"context"
	"math"

	"sipon-be/internal/modules/article/application"
	"sipon-be/internal/modules/article/application/dto"
	ports "sipon-be/internal/modules/article/application/ports"
	articleconst "sipon-be/internal/modules/article/domain/article/constant"
	articlerepo "sipon-be/internal/modules/article/domain/article/repository"
)

type ListArticlesUseCase struct {
	articleRepo  articlerepo.ArticleRepository
	fileUploader ports.FileUploader
}

func NewListArticlesUseCase(articleRepo articlerepo.ArticleRepository, fileUploader ports.FileUploader) *ListArticlesUseCase {
	return &ListArticlesUseCase{articleRepo: articleRepo, fileUploader: fileUploader}
}

func (uc *ListArticlesUseCase) Execute(ctx context.Context, req dto.ListArticlesQuery) ([]dto.ArticleListItem, dto.Meta, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 20
	}

	result, err := uc.articleRepo.List(ctx, articlerepo.ArticleListQuery{
		Status:     req.Status,
		CategoryID: req.CategoryID,
		Search:     req.Search,
		Page:       page,
		Limit:      limit,
		SortBy:     req.SortBy,
		SortType:   req.SortType,
	})
	if err != nil {
		return nil, dto.Meta{}, application.WrapRepoErr(err, articleconst.CodeArticleQueryFailed)
	}

	items := make([]dto.ArticleListItem, 0, len(result.Items))
	for _, a := range result.Items {
		items = append(items, dto.ArticleListItem{
			ID:           a.ID,
			Title:        a.Title,
			Status:       string(a.Status),
			CategoryID:   a.CategoryID,
			CategoryName: a.CategoryName,
			Author:       a.Author,
			ThumbnailURL: resolveThumbnailURL(uc.fileUploader, a.ThumbnailURL),
			IsFeatured:   a.IsFeatured,
			ViewCount:    a.ViewCount,
			PublishedAt:  a.PublishedAt,
			CreatedAt:    a.CreatedAt,
		})
	}

	totalPages := int(math.Ceil(float64(result.Total) / float64(limit)))
	meta := dto.Meta{
		CurrentPage: page,
		PerPage:     limit,
		Total:       int(result.Total),
		TotalPages:  totalPages,
	}

	return items, meta, nil
}
