package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"sipon-be/internal/modules/article/domain/article/constant"
	"sipon-be/internal/modules/article/domain/article/entity"
	"sipon-be/internal/shared/kernel"
)

type PostgresSourceRepository struct {
	db *sql.DB
}

func NewPostgresSourceRepository(db *sql.DB) *PostgresSourceRepository {
	return &PostgresSourceRepository{db: db}
}

func (r *PostgresSourceRepository) Save(ctx context.Context, s *entity.Source) error {
	execer := execerFromContext(ctx, r.db)
	query := `INSERT INTO article_sources (id, key, name, base_url, auto_publish, is_active, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`

	_, err := execer.ExecContext(ctx, query,
		s.ID, s.Key, s.Name, s.BaseURL, s.AutoPublish, s.IsActive,
		nullStr(s.CreatedBy), s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeSourceDuplicateKey, err)
		}
		return kernel.Wrap(constant.CodeSourcePersistenceFailed, fmt.Errorf("save source: %w", err))
	}
	return nil
}

func (r *PostgresSourceRepository) Update(ctx context.Context, s *entity.Source) error {
	execer := execerFromContext(ctx, r.db)
	query := `UPDATE article_sources SET
		key=$1, name=$2, base_url=$3, auto_publish=$4, is_active=$5,
		updated_by=$6, updated_at=$7
		WHERE id=$8 AND deleted_at IS NULL`

	res, err := execer.ExecContext(ctx, query,
		s.Key, s.Name, s.BaseURL, s.AutoPublish, s.IsActive,
		nullStr(s.UpdatedBy), s.UpdatedAt, s.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeSourceDuplicateKey, err)
		}
		return kernel.Wrap(constant.CodeSourcePersistenceFailed, fmt.Errorf("update source: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeSourceNotFound)
	}
	return nil
}

func (r *PostgresSourceRepository) FindByID(ctx context.Context, id string) (*entity.Source, error) {
	execer := execerFromContext(ctx, r.db)
	query := `SELECT id, key, name, base_url, auto_publish, is_active, last_scraped_at,
		created_by, updated_by, created_at, updated_at, deleted_at
		FROM article_sources WHERE id=$1 AND deleted_at IS NULL`
	return r.scanSource(execer.QueryRowContext(ctx, query, id))
}

func (r *PostgresSourceRepository) FindAll(ctx context.Context) ([]*entity.Source, error) {
	execer := execerFromContext(ctx, r.db)
	query := `SELECT id, key, name, base_url, auto_publish, is_active, last_scraped_at,
		created_by, updated_by, created_at, updated_at, deleted_at
		FROM article_sources WHERE deleted_at IS NULL ORDER BY created_at DESC`

	rows, err := execer.QueryContext(ctx, query)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeSourceQueryFailed, fmt.Errorf("find all sources: %w", err))
	}
	defer rows.Close()

	sources := make([]*entity.Source, 0)
	for rows.Next() {
		s, err := r.scanSource(rows)
		if err != nil {
			return nil, err
		}
		sources = append(sources, s)
	}
	return sources, rows.Err()
}

func (r *PostgresSourceRepository) Delete(ctx context.Context, id string) error {
	execer := execerFromContext(ctx, r.db)
	query := `UPDATE article_sources SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`
	res, err := execer.ExecContext(ctx, query, id)
	if err != nil {
		return kernel.Wrap(constant.CodeSourcePersistenceFailed, fmt.Errorf("delete source: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeSourceNotFound)
	}
	return nil
}

func (r *PostgresSourceRepository) scanSource(sc scanner) (*entity.Source, error) {
	var (
		id, key, name, baseURL                string
		autoPublish, isActive                 bool
		lastScrapedAt                         sql.NullTime
		createdBy, updatedBy                  sql.NullString
		createdAt, updatedAt                  time.Time
		deletedAt                             sql.NullTime
	)

	err := sc.Scan(
		&id, &key, &name, &baseURL, &autoPublish, &isActive, &lastScrapedAt,
		&createdBy, &updatedBy, &createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeSourceNotFound)
		}
		return nil, kernel.Wrap(constant.CodeSourceQueryFailed, fmt.Errorf("scan source: %w", err))
	}

	return &entity.Source{
		ID:            id,
		Key:           key,
		Name:          name,
		BaseURL:       baseURL,
		AutoPublish:   autoPublish,
		IsActive:      isActive,
		LastScrapedAt: timeFromNull(lastScrapedAt),
		CreatedBy:     strFromNull(createdBy),
		UpdatedBy:     strFromNull(updatedBy),
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		DeletedAt:     timeFromNull(deletedAt),
	}, nil
}

type PostgresSourceSelectorRepository struct {
	db *sql.DB
}

func NewPostgresSourceSelectorRepository(db *sql.DB) *PostgresSourceSelectorRepository {
	return &PostgresSourceSelectorRepository{db: db}
}

func (r *PostgresSourceSelectorRepository) Save(ctx context.Context, sel *entity.SourceSelector) error {
	execer := execerFromContext(ctx, r.db)
	query := `INSERT INTO article_source_selectors (id, source_id, content_selector, author_selector, tags_selector, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := execer.ExecContext(ctx, query,
		sel.ID, sel.SourceID, nullStr(sel.ContentSelector), nullStr(sel.AuthorSelector),
		nullStr(sel.TagsSelector), sel.CreatedAt, sel.UpdatedAt,
	)
	if err != nil {
		return kernel.Wrap(constant.CodeSelectorPersistenceFailed, fmt.Errorf("save selector: %w", err))
	}
	return nil
}

func (r *PostgresSourceSelectorRepository) SaveOrUpdate(ctx context.Context, sel *entity.SourceSelector) error {
	execer := execerFromContext(ctx, r.db)
	query := `INSERT INTO article_source_selectors (id, source_id, content_selector, author_selector, tags_selector, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (source_id) DO UPDATE SET
		content_selector=EXCLUDED.content_selector,
		author_selector=EXCLUDED.author_selector,
		tags_selector=EXCLUDED.tags_selector,
		updated_at=EXCLUDED.updated_at`
	_, err := execer.ExecContext(ctx, query,
		sel.ID, sel.SourceID, nullStr(sel.ContentSelector), nullStr(sel.AuthorSelector),
		nullStr(sel.TagsSelector), sel.CreatedAt, sel.UpdatedAt,
	)
	if err != nil {
		return kernel.Wrap(constant.CodeSelectorPersistenceFailed, fmt.Errorf("save or update selector: %w", err))
	}
	return nil
}

func (r *PostgresSourceSelectorRepository) FindBySourceID(ctx context.Context, sourceID string) (*entity.SourceSelector, error) {
	execer := execerFromContext(ctx, r.db)
	query := `SELECT id, source_id, content_selector, author_selector, tags_selector, created_at, updated_at
		FROM article_source_selectors WHERE source_id=$1`
	row := execer.QueryRowContext(ctx, query, sourceID)

	var (
		id, sid                               string
		contentSelector, authorSelector, tagsSelector sql.NullString
		createdAt, updatedAt                   time.Time
	)
	err := row.Scan(&id, &sid, &contentSelector, &authorSelector, &tagsSelector, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeSelectorNotFound)
		}
		return nil, kernel.Wrap(constant.CodeSelectorPersistenceFailed, fmt.Errorf("find selector: %w", err))
	}
	return &entity.SourceSelector{
		ID:              id,
		SourceID:        sid,
		ContentSelector: strFromNull(contentSelector),
		AuthorSelector:  strFromNull(authorSelector),
		TagsSelector:    strFromNull(tagsSelector),
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}, nil
}

type PostgresSourceCategoryRepository struct {
	db *sql.DB
}

func NewPostgresSourceCategoryRepository(db *sql.DB) *PostgresSourceCategoryRepository {
	return &PostgresSourceCategoryRepository{db: db}
}

func (r *PostgresSourceCategoryRepository) Save(ctx context.Context, cat *entity.SourceCategory) error {
	execer := execerFromContext(ctx, r.db)
	query := `INSERT INTO article_source_categories (
		id, source_id, category_key, url_suffix, url_override, article_limit, is_active,
		article_category_id, keywords, updated_by, created_at, updated_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`

	_, err := execer.ExecContext(ctx, query,
		cat.ID, cat.SourceID, cat.CategoryKey, nullStr(cat.URLSuffix), nullStr(cat.URLOverride),
		cat.ArticleLimit, cat.IsActive, nullStr(cat.ArticleCategoryID),
		marshalKeywords(cat.Keywords), nullStr(cat.UpdatedBy), cat.CreatedAt, cat.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeSourceCategoryDuplicate, err)
		}
		return kernel.Wrap(constant.CodeSourceCategoryPersistenceFailed, fmt.Errorf("save source category: %w", err))
	}
	return nil
}

func (r *PostgresSourceCategoryRepository) Update(ctx context.Context, cat *entity.SourceCategory) error {
	execer := execerFromContext(ctx, r.db)
	query := `UPDATE article_source_categories SET
		category_key=$1, url_suffix=$2, url_override=$3, article_limit=$4, is_active=$5,
		article_category_id=$6, keywords=$7, updated_by=$8, updated_at=$9
		WHERE id=$10`

	res, err := execer.ExecContext(ctx, query,
		cat.CategoryKey, nullStr(cat.URLSuffix), nullStr(cat.URLOverride),
		cat.ArticleLimit, cat.IsActive, nullStr(cat.ArticleCategoryID),
		marshalKeywords(cat.Keywords), nullStr(cat.UpdatedBy), cat.UpdatedAt, cat.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeSourceCategoryDuplicate, err)
		}
		return kernel.Wrap(constant.CodeSourceCategoryPersistenceFailed, fmt.Errorf("update source category: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeSourceCategoryNotFound)
	}
	return nil
}

func (r *PostgresSourceCategoryRepository) FindByID(ctx context.Context, id string) (*entity.SourceCategory, error) {
	execer := execerFromContext(ctx, r.db)
	query := `SELECT id, source_id, category_key, url_suffix, url_override, article_limit, is_active,
		article_category_id, keywords, last_scraped_at, updated_by, created_at, updated_at
		FROM article_source_categories WHERE id=$1`
	return r.scanSourceCategory(execer.QueryRowContext(ctx, query, id))
}

func (r *PostgresSourceCategoryRepository) FindAllBySource(ctx context.Context, sourceID string) ([]*entity.SourceCategory, error) {
	execer := execerFromContext(ctx, r.db)
	query := `SELECT id, source_id, category_key, url_suffix, url_override, article_limit, is_active,
		article_category_id, keywords, last_scraped_at, updated_by, created_at, updated_at
		FROM article_source_categories WHERE source_id=$1 ORDER BY created_at ASC`

	rows, err := execer.QueryContext(ctx, query, sourceID)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeSourceCategoryQueryFailed, fmt.Errorf("find source categories: %w", err))
	}
	defer rows.Close()

	cats := make([]*entity.SourceCategory, 0)
	for rows.Next() {
		c, err := r.scanSourceCategory(rows)
		if err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *PostgresSourceCategoryRepository) Delete(ctx context.Context, id string) error {
	execer := execerFromContext(ctx, r.db)
	query := `DELETE FROM article_source_categories WHERE id=$1`
	res, err := execer.ExecContext(ctx, query, id)
	if err != nil {
		return kernel.Wrap(constant.CodeSourceCategoryPersistenceFailed, fmt.Errorf("delete source category: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeSourceCategoryNotFound)
	}
	return nil
}

func (r *PostgresSourceCategoryRepository) UpdateLastScraped(ctx context.Context, id string, t time.Time) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx, `UPDATE article_source_categories SET last_scraped_at=$1 WHERE id=$2`, t, id)
	if err != nil {
		return kernel.Wrap(constant.CodeSourceCategoryPersistenceFailed, fmt.Errorf("update last scraped: %w", err))
	}
	return nil
}

func (r *PostgresSourceCategoryRepository) scanSourceCategory(sc scanner) (*entity.SourceCategory, error) {
	var (
		id, sourceID, categoryKey       string
		urlSuffix, urlOverride         sql.NullString
		articleLimit                   int
		isActive                      bool
		articleCategoryID              sql.NullString
		keywords                      sql.NullString
		lastScrapedAt                 sql.NullTime
		updatedBy                     sql.NullString
		createdAt, updatedAt          time.Time
	)
	err := sc.Scan(
		&id, &sourceID, &categoryKey, &urlSuffix, &urlOverride, &articleLimit, &isActive,
		&articleCategoryID, &keywords, &lastScrapedAt, &updatedBy, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeSourceCategoryNotFound)
		}
		return nil, kernel.Wrap(constant.CodeSourceCategoryQueryFailed, fmt.Errorf("scan source category: %w", err))
	}

	var kw []string
	if keywords.Valid && keywords.String != "" {
		_ = parseKeywords(keywords.String, &kw)
	}

	return &entity.SourceCategory{
		ID:                id,
		SourceID:          sourceID,
		CategoryKey:       categoryKey,
		URLSuffix:         strFromNull(urlSuffix),
		URLOverride:       strFromNull(urlOverride),
		ArticleLimit:      articleLimit,
		IsActive:          isActive,
		ArticleCategoryID: strFromNull(articleCategoryID),
		Keywords:          kw,
		LastScrapedAt:     timeFromNull(lastScrapedAt),
		UpdatedBy:         strFromNull(updatedBy),
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}, nil
}

func marshalKeywords(kw []string) interface{} {
	if len(kw) == 0 {
		return nil
	}
	b, err := json.Marshal(kw)
	if err != nil {
		return nil
	}
	return string(b)
}

func parseKeywords(raw string, out *[]string) error {
	return json.Unmarshal([]byte(raw), out)
}
