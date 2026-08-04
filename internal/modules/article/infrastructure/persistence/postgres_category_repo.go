package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"sipon-be/internal/modules/article/domain/article/constant"
	"sipon-be/internal/modules/article/domain/article/entity"
	"sipon-be/internal/shared/kernel"
)

type PostgresCategoryRepository struct {
	db *sql.DB
}

func NewPostgresCategoryRepository(db *sql.DB) *PostgresCategoryRepository {
	return &PostgresCategoryRepository{db: db}
}

func (r *PostgresCategoryRepository) Save(ctx context.Context, c *entity.Category) error {
	execer := execerFromContext(ctx, r.db)

	query := `INSERT INTO article_categories (
		id, name, slug, is_active, sort_order, created_by, created_at, updated_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`

	_, err := execer.ExecContext(ctx, query,
		c.ID, c.Name, c.Slug, c.IsActive, c.SortOrder,
		nullStr(c.CreatedBy), c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeCategoryDuplicateSlug, err)
		}
		return kernel.Wrap(constant.CodeCategoryPersistenceFailed, fmt.Errorf("save category: %w", err))
	}
	return nil
}

func (r *PostgresCategoryRepository) Update(ctx context.Context, c *entity.Category) error {
	execer := execerFromContext(ctx, r.db)

	query := `UPDATE article_categories SET
		name=$1, slug=$2, is_active=$3, sort_order=$4, updated_by=$5, updated_at=$6
		WHERE id=$7 AND deleted_at IS NULL`

	res, err := execer.ExecContext(ctx, query,
		c.Name, c.Slug, c.IsActive, c.SortOrder,
		nullStr(c.UpdatedBy), c.UpdatedAt, c.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeCategoryDuplicateSlug, err)
		}
		return kernel.Wrap(constant.CodeCategoryPersistenceFailed, fmt.Errorf("update category: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeCategoryNotFound)
	}
	return nil
}

func (r *PostgresCategoryRepository) FindByID(ctx context.Context, id string) (*entity.Category, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx,
		`SELECT id, name, slug, is_active, sort_order, created_by, updated_by, created_at, updated_at, deleted_at
		 FROM article_categories WHERE id=$1 AND deleted_at IS NULL`, id)
	return r.scan(row)
}

func (r *PostgresCategoryRepository) FindBySlug(ctx context.Context, slug string) (*entity.Category, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx,
		`SELECT id, name, slug, is_active, sort_order, created_by, updated_by, created_at, updated_at, deleted_at
		 FROM article_categories WHERE slug=$1 AND deleted_at IS NULL`, slug)
	return r.scan(row)
}

func (r *PostgresCategoryRepository) FindAll(ctx context.Context, activeOnly bool) ([]*entity.Category, error) {
	execer := execerFromContext(ctx, r.db)

	query := `SELECT id, name, slug, is_active, sort_order, created_by, updated_by, created_at, updated_at, deleted_at
		FROM article_categories WHERE deleted_at IS NULL`
	if activeOnly {
		query += ` AND is_active = TRUE`
	}
	query += ` ORDER BY sort_order ASC, created_at ASC`

	rows, err := execer.QueryContext(ctx, query)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeCategoryQueryFailed, fmt.Errorf("find all categories: %w", err))
	}
	defer rows.Close()

	cats := make([]*entity.Category, 0)
	for rows.Next() {
		c, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(constant.CodeCategoryQueryFailed, fmt.Errorf("iterate categories: %w", err))
	}

	return cats, nil
}

func (r *PostgresCategoryRepository) Delete(ctx context.Context, id string) error {
	execer := execerFromContext(ctx, r.db)
	query := `UPDATE article_categories SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`
	res, err := execer.ExecContext(ctx, query, id)
	if err != nil {
		return kernel.Wrap(constant.CodeCategoryPersistenceFailed, fmt.Errorf("delete category: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeCategoryNotFound)
	}
	return nil
}

func (r *PostgresCategoryRepository) scan(sc scanner) (*entity.Category, error) {
	var (
		id                                        string
		name, slug                                string
		isActive                                  bool
		sortOrder                                 int
		createdBy, updatedBy                      sql.NullString
		createdAt, updatedAt                      sql.NullTime
		deletedAt                                 sql.NullTime
	)

	err := sc.Scan(
		&id, &name, &slug, &isActive, &sortOrder,
		&createdBy, &updatedBy, &createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeCategoryNotFound)
		}
		return nil, kernel.Wrap(constant.CodeCategoryQueryFailed, fmt.Errorf("scan category: %w", err))
	}

	cat := &entity.Category{
		ID:        id,
		Name:      name,
		Slug:      slug,
		IsActive:  isActive,
		SortOrder: sortOrder,
		CreatedBy: strFromNull(createdBy),
		UpdatedBy: strFromNull(updatedBy),
		CreatedAt: createdAt.Time,
		UpdatedAt: updatedAt.Time,
		DeletedAt: timeFromNull(deletedAt),
	}
	return cat, nil
}
