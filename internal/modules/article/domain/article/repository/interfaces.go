package repository

import (
	"context"

	"sipon-be/internal/modules/article/domain/article/entity"
)

type ArticleListQuery struct {
	Status     *string
	CategoryID *string
	Search     *string
	Page       int
	Limit      int
	SortBy     string
	SortType   string
}

type ArticleListResult struct {
	Items []*entity.Article
	Total int64
}

type ArticleRepository interface {
	Save(ctx context.Context, article *entity.Article) error
	Update(ctx context.Context, article *entity.Article) error
	FindByID(ctx context.Context, id string) (*entity.Article, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, query ArticleListQuery) (*ArticleListResult, error)
}

type CategoryRepository interface {
	Save(ctx context.Context, cat *entity.Category) error
	Update(ctx context.Context, cat *entity.Category) error
	FindByID(ctx context.Context, id string) (*entity.Category, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Category, error)
	FindAll(ctx context.Context, activeOnly bool) ([]*entity.Category, error)
	Delete(ctx context.Context, id string) error
}
