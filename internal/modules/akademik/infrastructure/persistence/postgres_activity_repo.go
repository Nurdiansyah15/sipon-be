package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"sipon-be/internal/modules/akademik/domain/activity/constant"
	"sipon-be/internal/modules/akademik/domain/activity/entity"
	"sipon-be/internal/modules/akademik/domain/activity/repository"
	"sipon-be/internal/shared/kernel"
)

const activityColumns = `
	id, code, name, status, created_at, updated_at, deleted_at
`

type PostgresActivityRepository struct {
	db *sql.DB
}

func NewPostgresActivityRepository(db *sql.DB) *PostgresActivityRepository {
	return &PostgresActivityRepository{db: db}
}

func (r *PostgresActivityRepository) Save(ctx context.Context, a *entity.Activity) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx,
		`INSERT INTO activities (`+activityColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		a.ID, a.Code, a.Name, string(a.Status), a.CreatedAt, a.UpdatedAt, nullTimeVal(a.DeletedAt),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeActivityDuplicate, err)
		}
		return kernel.Wrap(constant.CodeActivityPersistenceFailed, fmt.Errorf("save activity: %w", err))
	}
	return nil
}

func (r *PostgresActivityRepository) Update(ctx context.Context, a *entity.Activity) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx,
		`UPDATE activities SET code=$1, name=$2, status=$3, updated_at=$4, deleted_at=$5 WHERE id=$6 AND deleted_at IS NULL`,
		a.Code, a.Name, string(a.Status), a.UpdatedAt, nullTimeVal(a.DeletedAt), a.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeActivityDuplicate, err)
		}
		return kernel.Wrap(constant.CodeActivityPersistenceFailed, fmt.Errorf("update activity: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeActivityNotFound)
	}
	return nil
}

func (r *PostgresActivityRepository) FindByID(ctx context.Context, id string) (*entity.Activity, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+activityColumns+` FROM activities WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanActivity(row)
}

func (r *PostgresActivityRepository) FindByCode(ctx context.Context, code string) (*entity.Activity, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+activityColumns+` FROM activities WHERE code=$1 AND deleted_at IS NULL`, code)
	return scanActivity(row)
}

func (r *PostgresActivityRepository) FindByIDs(ctx context.Context, ids []string) ([]*entity.Activity, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	execer := execerFromContext(ctx, r.db)
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	rows, err := execer.QueryContext(ctx,
		`SELECT `+activityColumns+` FROM activities WHERE deleted_at IS NULL AND id IN (`+strings.Join(placeholders, ",")+`)`, args...)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeActivityQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*entity.Activity, 0)
	for rows.Next() {
		a, err := scanActivity(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	return items, rows.Err()
}

func (r *PostgresActivityRepository) List(ctx context.Context, q repository.ActivityListQuery) (*repository.ActivityListResult, error) {
	execer := execerFromContext(ctx, r.db)

	where := `deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 1
	if q.Status != nil && *q.Status != "" {
		where += fmt.Sprintf(` AND status=$%d`, argIdx)
		args = append(args, *q.Status)
		argIdx++
	}
	if q.Search != nil && *q.Search != "" {
		where += fmt.Sprintf(` AND (code ILIKE $%d OR name ILIKE $%d)`, argIdx, argIdx+1)
		args = append(args, "%"+*q.Search+"%", "%"+*q.Search+"%")
		argIdx += 2
	}

	var total int64
	if err := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM activities WHERE `+where, args...).Scan(&total); err != nil {
		return nil, kernel.Wrap(constant.CodeActivityQueryFailed, err)
	}

	limit := q.Limit
	offset := (q.Page - 1) * q.Limit
	listArgs := append(append([]interface{}{}, args...), limit, offset)
	rows, err := execer.QueryContext(ctx,
		fmt.Sprintf(`SELECT %s FROM activities WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
			activityColumns, where, argIdx, argIdx+1), listArgs...)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeActivityQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*entity.Activity, 0)
	for rows.Next() {
		a, err := scanActivity(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	return &repository.ActivityListResult{Items: items, Total: total}, rows.Err()
}

func scanActivity(sc scanner) (*entity.Activity, error) {
	var (
		id, code, name, status string
		createdAt, updatedAt   time.Time
		deletedAt              sql.NullTime
	)
	err := sc.Scan(&id, &code, &name, &status, &createdAt, &updatedAt, &deletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeActivityNotFound)
		}
		return nil, kernel.Wrap(constant.CodeActivityQueryFailed, fmt.Errorf("scan activity: %w", err))
	}
	return &entity.Activity{
		ID:        id,
		Code:      code,
		Name:      name,
		Status:    constant.ActivityStatus(status),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		DeletedAt: timeFromNull(deletedAt),
	}, nil
}
