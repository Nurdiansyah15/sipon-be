package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sipon-be/internal/modules/keuangan/domain/period/constant"
	"sipon-be/internal/modules/keuangan/domain/period/entity"
	"sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/shared/kernel"
)

const periodColumns = `
	id, name, start_date, end_date, status, closed_by, closed_at,
	created_by, created_at, updated_at
`

type PostgresAccountingPeriodRepository struct {
	db *sql.DB
}

func NewPostgresAccountingPeriodRepository(db *sql.DB) *PostgresAccountingPeriodRepository {
	return &PostgresAccountingPeriodRepository{db: db}
}

func (r *PostgresAccountingPeriodRepository) Save(ctx context.Context, period *entity.AccountingPeriod) error {
	execer := execerFromContext(ctx, r.db)

	query := `INSERT INTO accounting_periods (` + periodColumns + `) VALUES (
		$1,$2,$3,$4,$5,$6,$7,
		$8,$9,$10
	)`

	_, err := execer.ExecContext(ctx, query,
		period.ID, period.Name, period.StartDate, period.EndDate,
		string(period.Status), nullStr(period.ClosedBy), nullTimeVal(period.ClosedAt),
		period.CreatedBy, period.CreatedAt, period.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodePeriodOverlap, err)
		}
		return kernel.Wrap(constant.CodePeriodPersistenceFailed, fmt.Errorf("save period: %w", err))
	}
	return nil
}

func (r *PostgresAccountingPeriodRepository) Update(ctx context.Context, period *entity.AccountingPeriod) error {
	execer := execerFromContext(ctx, r.db)

	query := `UPDATE accounting_periods SET
		name=$1, start_date=$2, end_date=$3, status=$4,
		closed_by=$5, closed_at=$6, updated_at=$7
		WHERE id=$8`

	res, err := execer.ExecContext(ctx, query,
		period.Name, period.StartDate, period.EndDate, string(period.Status),
		nullStr(period.ClosedBy), nullTimeVal(period.ClosedAt),
		period.UpdatedAt, period.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodePeriodOverlap, err)
		}
		return kernel.Wrap(constant.CodePeriodPersistenceFailed, fmt.Errorf("update period: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodePeriodNotFound)
	}
	return nil
}

func (r *PostgresAccountingPeriodRepository) FindByID(ctx context.Context, id string) (*entity.AccountingPeriod, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+periodColumns+` FROM accounting_periods WHERE id=$1`, id)
	return r.scan(row)
}

func (r *PostgresAccountingPeriodRepository) FindActive(ctx context.Context) (*entity.AccountingPeriod, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+periodColumns+` FROM accounting_periods WHERE status='open' LIMIT 1`)
	return r.scan(row)
}

func (r *PostgresAccountingPeriodRepository) List(ctx context.Context, q repository.PeriodListQuery) (*repository.PeriodListResult, error) {
	execer := execerFromContext(ctx, r.db)

	where := `WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if q.Status != nil && *q.Status != "" {
		where += fmt.Sprintf(` AND status=$%d`, argIdx)
		args = append(args, *q.Status)
		argIdx++
	}

	var total int64
	countRow := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM accounting_periods `+where, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, kernel.Wrap(constant.CodePeriodQueryFailed, fmt.Errorf("count periods: %w", err))
	}

	limit := q.Limit
	offset := (q.Page - 1) * q.Limit
	listArgs := append(append([]interface{}{}, args...), limit, offset)

	query := fmt.Sprintf(`SELECT %s FROM accounting_periods %s ORDER BY start_date DESC LIMIT $%d OFFSET $%d`,
		periodColumns, where, argIdx, argIdx+1)

	rows, err := execer.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, kernel.Wrap(constant.CodePeriodQueryFailed, fmt.Errorf("list periods: %w", err))
	}
	defer rows.Close()

	items := make([]*entity.AccountingPeriod, 0)
	for rows.Next() {
		p, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(constant.CodePeriodQueryFailed, fmt.Errorf("iterate period rows: %w", err))
	}

	return &repository.PeriodListResult{Items: items, Total: total}, nil
}

func (r *PostgresAccountingPeriodRepository) FindByDate(ctx context.Context, date time.Time) (*entity.AccountingPeriod, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx,
		`SELECT `+periodColumns+` FROM accounting_periods WHERE start_date<=$1 AND end_date>=$1`,
		date,
	)
	return r.scan(row)
}

func (r *PostgresAccountingPeriodRepository) HasOverlap(ctx context.Context, startDate, endDate time.Time, excludeID string) (bool, error) {
	execer := execerFromContext(ctx, r.db)

	var exists bool
	var err error
	if excludeID == "" {
		err = execer.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM accounting_periods WHERE start_date<=$1 AND end_date>=$2)`,
			endDate, startDate,
		).Scan(&exists)
	} else {
		err = execer.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM accounting_periods WHERE id!=$1 AND start_date<=$2 AND end_date>=$3)`,
			excludeID, endDate, startDate,
		).Scan(&exists)
	}
	if err != nil {
		return false, kernel.Wrap(constant.CodePeriodQueryFailed, fmt.Errorf("has overlap: %w", err))
	}
	return exists, nil
}

func (r *PostgresAccountingPeriodRepository) scan(sc scanner) (*entity.AccountingPeriod, error) {
	var (
		id, name, periodStatus, createdBy     string
		startDate, endDate                    time.Time
		closedBy                              sql.NullString
		closedAt                              sql.NullTime
		createdAt, updatedAt                  time.Time
	)

	err := sc.Scan(
		&id, &name, &startDate, &endDate, &periodStatus, &closedBy, &closedAt,
		&createdBy, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodePeriodNotFound)
		}
		return nil, kernel.Wrap(constant.CodePeriodQueryFailed, fmt.Errorf("scan period: %w", err))
	}

	return &entity.AccountingPeriod{
		ID:        id,
		Name:      name,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    constant.PeriodStatus(periodStatus),
		ClosedBy:  strFromNull(closedBy),
		ClosedAt:  timeFromNull(closedAt),
		CreatedBy: createdBy,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
