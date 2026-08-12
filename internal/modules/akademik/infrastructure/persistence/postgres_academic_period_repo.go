package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"sipon-be/internal/modules/akademik/domain/academic_period/constant"
	"sipon-be/internal/modules/akademik/domain/academic_period/entity"
	"sipon-be/internal/modules/akademik/domain/academic_period/repository"
	"sipon-be/internal/shared/kernel"
)

const academicPeriodColumns = `
	id, code, name, start_date, end_date, status, created_at, updated_at, deleted_at
`

type PostgresAcademicPeriodRepository struct {
	db *sql.DB
}

func NewPostgresAcademicPeriodRepository(db *sql.DB) *PostgresAcademicPeriodRepository {
	return &PostgresAcademicPeriodRepository{db: db}
}

func (r *PostgresAcademicPeriodRepository) Save(ctx context.Context, p *entity.AcademicPeriod) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx,
		`INSERT INTO academic_periods (`+academicPeriodColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		p.ID, p.Code, p.Name, p.StartDate, p.EndDate, string(p.Status), p.CreatedAt, p.UpdatedAt, nullTimeVal(p.DeletedAt),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeAcademicPeriodDuplicate, err)
		}
		return kernel.Wrap(constant.CodeAcademicPeriodPersistenceFailed, fmt.Errorf("save academic period: %w", err))
	}
	return nil
}

func (r *PostgresAcademicPeriodRepository) Update(ctx context.Context, p *entity.AcademicPeriod) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx,
		`UPDATE academic_periods SET code=$1, name=$2, start_date=$3, end_date=$4, status=$5, updated_at=$6, deleted_at=$7 WHERE id=$8 AND deleted_at IS NULL`,
		p.Code, p.Name, p.StartDate, p.EndDate, string(p.Status), p.UpdatedAt, nullTimeVal(p.DeletedAt), p.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeAcademicPeriodDuplicate, err)
		}
		return kernel.Wrap(constant.CodeAcademicPeriodPersistenceFailed, fmt.Errorf("update academic period: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeAcademicPeriodNotFound)
	}
	return nil
}

func (r *PostgresAcademicPeriodRepository) FindByID(ctx context.Context, id string) (*entity.AcademicPeriod, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+academicPeriodColumns+` FROM academic_periods WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanAcademicPeriod(row)
}

func (r *PostgresAcademicPeriodRepository) FindByCode(ctx context.Context, code string) (*entity.AcademicPeriod, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+academicPeriodColumns+` FROM academic_periods WHERE code=$1 AND deleted_at IS NULL`, code)
	return scanAcademicPeriod(row)
}

func (r *PostgresAcademicPeriodRepository) FindByIDs(ctx context.Context, ids []string) ([]*entity.AcademicPeriod, error) {
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
		`SELECT `+academicPeriodColumns+` FROM academic_periods WHERE deleted_at IS NULL AND id IN (`+strings.Join(placeholders, ",")+`)`, args...)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeAcademicPeriodQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*entity.AcademicPeriod, 0)
	for rows.Next() {
		p, err := scanAcademicPeriod(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func (r *PostgresAcademicPeriodRepository) FindOpen(ctx context.Context) (*entity.AcademicPeriod, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx,
		`SELECT `+academicPeriodColumns+` FROM academic_periods WHERE status='open' AND deleted_at IS NULL ORDER BY start_date DESC LIMIT 1`)
	return scanAcademicPeriod(row)
}

func (r *PostgresAcademicPeriodRepository) List(ctx context.Context, q repository.AcademicPeriodListQuery) (*repository.AcademicPeriodListResult, error) {
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
	if err := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM academic_periods WHERE `+where, args...).Scan(&total); err != nil {
		return nil, kernel.Wrap(constant.CodeAcademicPeriodQueryFailed, err)
	}

	limit := q.Limit
	offset := (q.Page - 1) * q.Limit
	listArgs := append(append([]interface{}{}, args...), limit, offset)
	rows, err := execer.QueryContext(ctx,
		fmt.Sprintf(`SELECT %s FROM academic_periods WHERE %s ORDER BY start_date DESC LIMIT $%d OFFSET $%d`,
			academicPeriodColumns, where, argIdx, argIdx+1), listArgs...)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeAcademicPeriodQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*entity.AcademicPeriod, 0)
	for rows.Next() {
		p, err := scanAcademicPeriod(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return &repository.AcademicPeriodListResult{Items: items, Total: total}, rows.Err()
}

func (r *PostgresAcademicPeriodRepository) HasData(ctx context.Context, id string) (bool, error) {
	execer := execerFromContext(ctx, r.db)
	var count int
	if err := execer.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM (
			SELECT 1 FROM santri_registrations WHERE academic_period_id=$1
			UNION ALL
			SELECT 1 FROM activity_periods WHERE academic_period_id=$1
		) t`, id).Scan(&count); err != nil {
		return false, kernel.Wrap(constant.CodeAcademicPeriodQueryFailed, err)
	}
	return count > 0, nil
}

func scanAcademicPeriod(sc scanner) (*entity.AcademicPeriod, error) {
	var (
		id, code, name       string
		startDate, endDate   time.Time
		status               string
		createdAt, updatedAt time.Time
		deletedAt            sql.NullTime
	)
	err := sc.Scan(&id, &code, &name, &startDate, &endDate, &status, &createdAt, &updatedAt, &deletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeAcademicPeriodNotFound)
		}
		return nil, kernel.Wrap(constant.CodeAcademicPeriodQueryFailed, fmt.Errorf("scan academic period: %w", err))
	}
	return &entity.AcademicPeriod{
		ID:        id,
		Code:      code,
		Name:      name,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    constant.AcademicPeriodStatus(status),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		DeletedAt: timeFromNull(deletedAt),
	}, nil
}
