package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"sipon-be/internal/modules/article/domain/article/constant"
	"sipon-be/internal/modules/article/domain/article/entity"
	"sipon-be/internal/modules/article/domain/article/repository"
	"sipon-be/internal/shared/kernel"
)

const articleColumns = `
	a.id, a.title, a.content, a.summary, a.category_id,
	ac.name AS category_name,
	a.status, a.author, a.thumbnail_url, a.view_count, a.is_featured,
	a.source_id, a.original_url, a.is_manual,
	a.created_by, a.updated_by, a.published_at, a.archived_at,
	a.created_at, a.updated_at, a.deleted_at
`

const articleFrom = `FROM articles a
LEFT JOIN article_categories ac ON a.category_id = ac.id AND ac.deleted_at IS NULL
WHERE a.deleted_at IS NULL`

var articleSortColumns = map[string]string{
	"created_at":   "a.created_at",
	"published_at": "a.published_at",
	"title":        "a.title",
	"view_count":   "a.view_count",
}

type PostgresArticleRepository struct {
	db *sql.DB
}

func NewPostgresArticleRepository(db *sql.DB) *PostgresArticleRepository {
	return &PostgresArticleRepository{db: db}
}

func (r *PostgresArticleRepository) Save(ctx context.Context, a *entity.Article) error {
	execer := execerFromContext(ctx, r.db)

	query := `INSERT INTO articles (
		id, title, content, summary, category_id, status, author,
		thumbnail_url, view_count, is_featured, source_id, original_url, is_manual,
		created_by, published_at, created_at, updated_at
	) VALUES (
		$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17
	)`

	_, err := execer.ExecContext(ctx, query,
		a.ID, a.Title, a.Content, nullStr(a.Summary), nullStr(a.CategoryID),
		string(a.Status), a.Author, nullStr(a.ThumbnailURL),
		a.ViewCount, a.IsFeatured, nullStr(a.SourceID), nullStr(a.OriginalURL), a.IsManual,
		nullStr(a.CreatedBy), nullTimeVal(a.PublishedAt), a.CreatedAt, a.UpdatedAt,
	)
	if err != nil {
		return kernel.Wrap(constant.CodeArticlePersistenceFailed, fmt.Errorf("save article: %w", err))
	}
	return nil
}

func (r *PostgresArticleRepository) Update(ctx context.Context, a *entity.Article) error {
	execer := execerFromContext(ctx, r.db)

	query := `UPDATE articles SET
		title=$1, content=$2, summary=$3, category_id=$4, status=$5, author=$6,
		thumbnail_url=$7, is_featured=$8, updated_by=$9,
		published_at=$10, archived_at=$11, updated_at=$12
		WHERE id=$13 AND deleted_at IS NULL`

	res, err := execer.ExecContext(ctx, query,
		a.Title, a.Content, nullStr(a.Summary), nullStr(a.CategoryID),
		string(a.Status), a.Author, nullStr(a.ThumbnailURL),
		a.IsFeatured, nullStr(a.UpdatedBy),
		nullTimeVal(a.PublishedAt), nullTimeVal(a.ArchivedAt),
		a.UpdatedAt, a.ID,
	)
	if err != nil {
		return kernel.Wrap(constant.CodeArticlePersistenceFailed, fmt.Errorf("update article: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeArticleNotFound)
	}
	return nil
}

func (r *PostgresArticleRepository) FindByID(ctx context.Context, id string) (*entity.Article, error) {
	execer := execerFromContext(ctx, r.db)
	query := `SELECT ` + articleColumns + ` ` + articleFrom + ` AND a.id=$1`
	row := execer.QueryRowContext(ctx, query, id)
	return r.scan(row)
}

func (r *PostgresArticleRepository) Delete(ctx context.Context, id string) error {
	execer := execerFromContext(ctx, r.db)
	query := `UPDATE articles SET deleted_at=$1, updated_at=$1 WHERE id=$2 AND deleted_at IS NULL`
	now := time.Now()
	res, err := execer.ExecContext(ctx, query, now, id)
	if err != nil {
		return kernel.Wrap(constant.CodeArticlePersistenceFailed, fmt.Errorf("delete article: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeArticleNotFound)
	}
	return nil
}

func (r *PostgresArticleRepository) List(ctx context.Context, q repository.ArticleListQuery) (*repository.ArticleListResult, error) {
	execer := execerFromContext(ctx, r.db)

	where := ``
	args := []interface{}{}
	argIdx := 1

	if q.Status != nil && *q.Status != "" {
		where += fmt.Sprintf(` AND a.status=$%d`, argIdx)
		args = append(args, *q.Status)
		argIdx++
	}
	if q.CategoryID != nil && *q.CategoryID != "" {
		where += fmt.Sprintf(` AND a.category_id=$%d`, argIdx)
		args = append(args, *q.CategoryID)
		argIdx++
	}
	if q.Search != nil && *q.Search != "" {
		where += fmt.Sprintf(` AND a.title ILIKE $%d`, argIdx)
		args = append(args, "%"+*q.Search+"%")
		argIdx++
	}

	sortCol, ok := articleSortColumns[q.SortBy]
	if !ok {
		sortCol = "a.created_at"
	}
	sortDir := "DESC"
	if q.SortType == "asc" {
		sortDir = "ASC"
	}

	var total int64
	countQuery := `SELECT COUNT(*) ` + articleFrom + where
	countRow := execer.QueryRowContext(ctx, countQuery, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, kernel.Wrap(constant.CodeArticleQueryFailed, fmt.Errorf("count articles: %w", err))
	}

	limit := q.Limit
	if limit < 1 {
		limit = 20
	}
	offset := (q.Page - 1) * limit
	listArgs := append(append([]interface{}{}, args...), limit, offset)

	query := fmt.Sprintf(`SELECT %s %s%s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		articleColumns, articleFrom, where, sortCol, sortDir, argIdx, argIdx+1)

	rows, err := execer.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeArticleQueryFailed, fmt.Errorf("list articles: %w", err))
	}
	defer rows.Close()

	items := make([]*entity.Article, 0)
	for rows.Next() {
		a, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(constant.CodeArticleQueryFailed, fmt.Errorf("iterate article rows: %w", err))
	}

	return &repository.ArticleListResult{Items: items, Total: total}, nil
}

func (r *PostgresArticleRepository) scan(sc scanner) (*entity.Article, error) {
	var (
		id                                     string
		title, content, author                  string
		summary                                sql.NullString
		categoryID, categoryName                sql.NullString
		status                                 string
		thumbnailURL                           sql.NullString
		viewCount                              int
		isFeatured                             bool
		sourceID, originalURL                  sql.NullString
		isManual                               bool
		createdBy, updatedBy                   sql.NullString
		publishedAt, archivedAt                sql.NullTime
		createdAt, updatedAt                   time.Time
		deletedAt                              sql.NullTime
	)

	err := sc.Scan(
		&id, &title, &content, &summary, &categoryID, &categoryName,
		&status, &author, &thumbnailURL, &viewCount, &isFeatured,
		&sourceID, &originalURL, &isManual,
		&createdBy, &updatedBy, &publishedAt, &archivedAt,
		&createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeArticleNotFound)
		}
		return nil, kernel.Wrap(constant.CodeArticleQueryFailed, fmt.Errorf("scan article: %w", err))
	}

	return &entity.Article{
		ID:           id,
		Title:        title,
		Content:      content,
		Summary:      strFromNull(summary),
		CategoryID:   strFromNull(categoryID),
		CategoryName: strFromNull(categoryName),
		Status:       constant.ArticleStatus(status),
		Author:       author,
		ThumbnailURL: strFromNull(thumbnailURL),
		ViewCount:    viewCount,
		IsFeatured:   isFeatured,
		SourceID:     strFromNull(sourceID),
		OriginalURL:  strFromNull(originalURL),
		IsManual:     isManual,
		CreatedBy:    strFromNull(createdBy),
		UpdatedBy:    strFromNull(updatedBy),
		PublishedAt:  timeFromNull(publishedAt),
		ArchivedAt:   timeFromNull(archivedAt),
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		DeletedAt:    timeFromNull(deletedAt),
	}, nil
}

func (r *PostgresArticleRepository) SaveScraped(ctx context.Context, input repository.SaveScrapedInput) (string, error) {
	execer := execerFromContext(ctx, r.db)

	articleID := input.Title + "-" + input.OriginalURL // placeholder
	_ = articleID

	status := "draft"
	pubDate := time.Now()
	if input.PubDate != nil {
		pubDate = *input.PubDate
	}

	var publishedAt *time.Time
	if input.AutoPublish {
		status = "published"
		publishedAt = &pubDate
	}

	query := `INSERT INTO articles (
		id, title, content, summary, category_id, status, author,
		thumbnail_url, view_count, is_featured, source_id, original_url, is_manual,
		created_by, published_at, created_at, updated_at
	) VALUES (
		gen_random_uuid(),$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16
	) ON CONFLICT (original_url) WHERE original_url IS NOT NULL AND deleted_at IS NULL DO NOTHING RETURNING id`

	row := execer.QueryRowContext(ctx, query,
		input.Title, input.Content, summaryOrNil(input.Summary), nullStr(input.CategoryID),
		status, input.Author, nullStr(input.Thumbnail), 0, false,
		nullStr(&input.SourceID), nullStr(&input.OriginalURL), false,
		nil, nullTimeVal(publishedAt), time.Now(), time.Now(),
	)

	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", kernel.Wrap(constant.CodeArticlePersistenceFailed, fmt.Errorf("save scraped article: %w", err))
	}
	return id, nil
}

func (r *PostgresArticleRepository) ExistingURLs(ctx context.Context, urls []string) (map[string]bool, error) {
	result := make(map[string]bool)
	if len(urls) == 0 {
		return result, nil
	}

	execer := execerFromContext(ctx, r.db)
	placeholders := make([]string, len(urls))
	args := make([]interface{}, len(urls))
	for i, u := range urls {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = u
	}

	query := fmt.Sprintf(`SELECT original_url FROM articles WHERE original_url IN (%s) AND deleted_at IS NULL`,
		strings.Join(placeholders, ","))

	rows, err := execer.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, err
		}
		result[url] = true
	}
	return result, rows.Err()
}

func summaryOrNil(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
