package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"sipon-be/internal/modules/akademik/domain/activity_period/constant"
	"sipon-be/internal/modules/akademik/domain/activity_period/entity"
	"sipon-be/internal/modules/akademik/domain/activity_period/repository"
	"sipon-be/internal/shared/kernel"
)

const activityPeriodColumns = `
	id, activity_id, academic_period_id, status, created_at, updated_at, deleted_at
`

type PostgresActivityPeriodRepository struct {
	db *sql.DB
}

func NewPostgresActivityPeriodRepository(db *sql.DB) *PostgresActivityPeriodRepository {
	return &PostgresActivityPeriodRepository{db: db}
}

func (r *PostgresActivityPeriodRepository) Save(ctx context.Context, p *entity.ActivityPeriod) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx,
		`INSERT INTO activity_periods (`+activityPeriodColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		p.ID, p.ActivityID, p.AcademicPeriodID, string(p.Status), p.CreatedAt, p.UpdatedAt, nullTimeVal(p.DeletedAt),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeActivityPeriodDuplicate, err)
		}
		return kernel.Wrap(constant.CodeActivityPeriodPersistenceFailed, fmt.Errorf("save activity period: %w", err))
	}
	return nil
}

func (r *PostgresActivityPeriodRepository) Update(ctx context.Context, p *entity.ActivityPeriod) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx,
		`UPDATE activity_periods SET status=$1, updated_at=$2, deleted_at=$3 WHERE id=$4 AND deleted_at IS NULL`,
		string(p.Status), p.UpdatedAt, nullTimeVal(p.DeletedAt), p.ID,
	)
	if err != nil {
		return kernel.Wrap(constant.CodeActivityPeriodPersistenceFailed, fmt.Errorf("update activity period: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeActivityPeriodNotFound)
	}
	return nil
}

func (r *PostgresActivityPeriodRepository) FindByID(ctx context.Context, id string) (*entity.ActivityPeriod, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+activityPeriodColumns+` FROM activity_periods WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanActivityPeriod(row)
}

func (r *PostgresActivityPeriodRepository) FindByActivityAndPeriod(ctx context.Context, activityID, academicPeriodID string) (*entity.ActivityPeriod, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx,
		`SELECT `+activityPeriodColumns+` FROM activity_periods WHERE activity_id=$1 AND academic_period_id=$2 AND deleted_at IS NULL`,
		activityID, academicPeriodID)
	return scanActivityPeriod(row)
}

func (r *PostgresActivityPeriodRepository) FindByIDs(ctx context.Context, ids []string) ([]*entity.ActivityPeriod, error) {
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
		`SELECT `+activityPeriodColumns+` FROM activity_periods WHERE deleted_at IS NULL AND id IN (`+strings.Join(placeholders, ",")+`)`, args...)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeActivityPeriodQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*entity.ActivityPeriod, 0)
	for rows.Next() {
		p, err := scanActivityPeriod(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func (r *PostgresActivityPeriodRepository) List(ctx context.Context, q repository.ActivityPeriodListQuery) (*repository.ActivityPeriodListResult, error) {
	execer := execerFromContext(ctx, r.db)

	where := `deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 1
	if q.ActivityID != nil && *q.ActivityID != "" {
		where += fmt.Sprintf(` AND activity_id=$%d`, argIdx)
		args = append(args, *q.ActivityID)
		argIdx++
	}
	if q.AcademicPeriodID != nil && *q.AcademicPeriodID != "" {
		where += fmt.Sprintf(` AND academic_period_id=$%d`, argIdx)
		args = append(args, *q.AcademicPeriodID)
		argIdx++
	}
	if q.Status != nil && *q.Status != "" {
		where += fmt.Sprintf(` AND status=$%d`, argIdx)
		args = append(args, *q.Status)
		argIdx++
	}

	var total int64
	if err := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM activity_periods WHERE `+where, args...).Scan(&total); err != nil {
		return nil, kernel.Wrap(constant.CodeActivityPeriodQueryFailed, err)
	}

	limit := q.Limit
	offset := (q.Page - 1) * q.Limit
	listArgs := append(append([]interface{}{}, args...), limit, offset)
	rows, err := execer.QueryContext(ctx,
		fmt.Sprintf(`SELECT %s FROM activity_periods WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
			activityPeriodColumns, where, argIdx, argIdx+1), listArgs...)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeActivityPeriodQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*entity.ActivityPeriod, 0)
	for rows.Next() {
		p, err := scanActivityPeriod(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return &repository.ActivityPeriodListResult{Items: items, Total: total}, rows.Err()
}

func scanActivityPeriod(sc scanner) (*entity.ActivityPeriod, error) {
	var (
		id, activityID, academicPeriodID, status string
		createdAt, updatedAt                     time.Time
		deletedAt                                sql.NullTime
	)
	err := sc.Scan(&id, &activityID, &academicPeriodID, &status, &createdAt, &updatedAt, &deletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeActivityPeriodNotFound)
		}
		return nil, kernel.Wrap(constant.CodeActivityPeriodQueryFailed, fmt.Errorf("scan activity period: %w", err))
	}
	return &entity.ActivityPeriod{
		ID:               id,
		ActivityID:       activityID,
		AcademicPeriodID: academicPeriodID,
		Status:           constant.ActivityPeriodStatus(status),
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		DeletedAt:        timeFromNull(deletedAt),
	}, nil
}
