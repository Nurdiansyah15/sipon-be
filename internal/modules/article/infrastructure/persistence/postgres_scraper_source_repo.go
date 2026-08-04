package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"sipon-be/internal/modules/article/domain/article/constant"
	"sipon-be/internal/modules/article/domain/article/repository"
	"sipon-be/internal/shared/kernel"
)

type PostgresScraperSourceRepo struct {
	db *sql.DB
}

func NewPostgresScraperSourceRepo(db *sql.DB) *PostgresScraperSourceRepo {
	return &PostgresScraperSourceRepo{db: db}
}

func (r *PostgresScraperSourceRepo) FindActiveCategoryByID(ctx context.Context, categoryID string) (*repository.ScraperSource, error) {
	query := `SELECT
		sc.id, sc.category_key, sc.url_suffix, sc.url_override, sc.article_limit, sc.article_category_id, sc.keywords, sc.last_scraped_at,
		ns.id, ns.key, ns.name, ns.base_url, ns.auto_publish,
		sel.content_selector, sel.author_selector, sel.tags_selector
	FROM article_source_categories sc
	JOIN article_sources ns ON ns.id = sc.source_id AND ns.deleted_at IS NULL
	LEFT JOIN article_source_selectors sel ON sel.source_id = ns.id
	WHERE sc.id = $1 AND sc.is_active = TRUE AND ns.is_active = TRUE`

	row := r.db.QueryRowContext(ctx, query, categoryID)

	var (
		catID, catKey, nsID, nsKey, nsName, nsBaseURL                              string
		urlSuffix, urlOverride, articleCategoryID                                    sql.NullString
		articleLimit                                                                 int
		keywords                                                                     sql.NullString
		lastScrapedAt                                                                sql.NullTime
		autoPublish                                                                  bool
		contentSelector, authorSelector, tagsSelector                                 sql.NullString
	)

	err := row.Scan(
		&catID, &catKey, &urlSuffix, &urlOverride, &articleLimit, &articleCategoryID, &keywords, &lastScrapedAt,
		&nsID, &nsKey, &nsName, &nsBaseURL, &autoPublish,
		&contentSelector, &authorSelector, &tagsSelector,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, kernel.Wrap(constant.CodeSourceQueryFailed, fmt.Errorf("find active category: %w", err))
	}

	var kw []string
	if keywords.Valid && keywords.String != "" {
		_ = json.Unmarshal([]byte(keywords.String), &kw)
	}

	return &repository.ScraperSource{
		ID:          nsID,
		Key:         nsKey,
		Name:        nsName,
		BaseURL:     nsBaseURL,
		AutoPublish: autoPublish,
		Selectors: repository.ScraperSelectors{
			ContentSelector: strFromNull(contentSelector),
			AuthorSelector:  strFromNull(authorSelector),
			TagsSelector:    strFromNull(tagsSelector),
		},
		Category: repository.ScraperCategory{
			ID:            catID,
			CategoryKey:   catKey,
			URLSuffix:     strFromNull(urlSuffix),
			URLOverride:   strFromNull(urlOverride),
			ArticleLimit:  articleLimit,
			Keywords:      kw,
			LastScrapedAt: timeFromNull(lastScrapedAt),
		},
	}, nil
}

func (r *PostgresScraperSourceRepo) UpdateLastScrapedCategory(ctx context.Context, categoryID string, t time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE article_source_categories SET last_scraped_at = $1 WHERE id = $2`, t, categoryID)
	if err != nil {
		return kernel.Wrap(constant.CodeSourceCategoryPersistenceFailed, fmt.Errorf("update last scraped: %w", err))
	}
	return nil
}
