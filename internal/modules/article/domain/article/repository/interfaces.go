package repository

import (
	"context"
	"time"

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

type SaveScrapedInput struct {
	SourceID    string
	CategoryKey string
	Title       string
	Content     string
	Summary     string
	Author      string
	Thumbnail   *string
	Tags        []string
	PubDate     *time.Time
	OriginalURL string
	CategoryID  *string
	AutoPublish bool
}

type ArticleRepository interface {
	Save(ctx context.Context, article *entity.Article) error
	Update(ctx context.Context, article *entity.Article) error
	FindByID(ctx context.Context, id string) (*entity.Article, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, query ArticleListQuery) (*ArticleListResult, error)
	SaveScraped(ctx context.Context, input SaveScrapedInput) (articleID string, err error)
	ExistingURLs(ctx context.Context, urls []string) (map[string]bool, error)
}

type CategoryRepository interface {
	Save(ctx context.Context, cat *entity.Category) error
	Update(ctx context.Context, cat *entity.Category) error
	FindByID(ctx context.Context, id string) (*entity.Category, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Category, error)
	FindAll(ctx context.Context, activeOnly bool) ([]*entity.Category, error)
	Delete(ctx context.Context, id string) error
}

type SourceRepository interface {
	Save(ctx context.Context, s *entity.Source) error
	Update(ctx context.Context, s *entity.Source) error
	FindByID(ctx context.Context, id string) (*entity.Source, error)
	FindAll(ctx context.Context) ([]*entity.Source, error)
	Delete(ctx context.Context, id string) error
}

type SourceSelectorRepository interface {
	Save(ctx context.Context, sel *entity.SourceSelector) error
	SaveOrUpdate(ctx context.Context, sel *entity.SourceSelector) error
	FindBySourceID(ctx context.Context, sourceID string) (*entity.SourceSelector, error)
}

type SourceCategoryRepository interface {
	Save(ctx context.Context, cat *entity.SourceCategory) error
	Update(ctx context.Context, cat *entity.SourceCategory) error
	FindByID(ctx context.Context, id string) (*entity.SourceCategory, error)
	FindAllBySource(ctx context.Context, sourceID string) ([]*entity.SourceCategory, error)
	Delete(ctx context.Context, id string) error
	UpdateLastScraped(ctx context.Context, id string, t time.Time) error
}

type ScraperSourceRepo interface {
	FindActiveCategoryByID(ctx context.Context, categoryID string) (*ScraperSource, error)
	UpdateLastScrapedCategory(ctx context.Context, categoryID string, t time.Time) error
}

type ScraperSelectors struct {
	ContentSelector *string
	AuthorSelector  *string
	TagsSelector    *string
}

type ScraperCategory struct {
	ID            string
	CategoryKey   string
	URLSuffix     *string
	URLOverride   *string
	ArticleLimit  int
	Keywords      []string
	LastScrapedAt *time.Time
}

type ScraperSource struct {
	ID        string
	Key       string
	Name      string
	BaseURL   string
	AutoPublish bool
	Selectors   ScraperSelectors
	Category    ScraperCategory
}
