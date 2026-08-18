package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sipon-be/internal/modules/keuangan/domain/billingperiod/constant"
	"sipon-be/internal/modules/keuangan/domain/billingperiod/entity"
	"sipon-be/internal/modules/keuangan/domain/billingperiod/repository"
	feeConst "sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	"sipon-be/internal/shared/kernel"
)

const billingPeriodColumns = `
	id, name, period_type, accounting_period_id, start_date, end_date, status, created_by, created_at, updated_at
`

type PostgresBillingPeriodRepository struct {
	db *sql.DB
}

func NewPostgresBillingPeriodRepository(db *sql.DB) *PostgresBillingPeriodRepository {
	return &PostgresBillingPeriodRepository{db: db}
}

func (r *PostgresBillingPeriodRepository) Save(ctx context.Context, period *entity.BillingPeriod) error {
	execer := execerFromContext(ctx, r.db)

	query := `INSERT INTO billing_periods (` + billingPeriodColumns + `) VALUES (
		$1,$2,$3,$4,$5,$6,$7,$8,$9,$10
	)`

	_, err := execer.ExecContext(ctx, query,
		period.ID, period.Name, string(period.PeriodType), period.AccountingPeriodID,
		period.StartDate, period.EndDate,
		string(period.Status), period.CreatedBy, period.CreatedAt, period.UpdatedAt,
	)
	if err != nil {
		return kernel.WrapMsg(constant.CodeBillingPeriodPersistenceFailed, "gagal menyimpan periode tagihan", err)
	}
	return nil
}

func (r *PostgresBillingPeriodRepository) Update(ctx context.Context, period *entity.BillingPeriod) error {
	execer := execerFromContext(ctx, r.db)

	query := `UPDATE billing_periods SET
		name=$1, period_type=$2, start_date=$3, end_date=$4, status=$5, updated_at=$6
		WHERE id=$7`

	res, err := execer.ExecContext(ctx, query,
		period.Name, string(period.PeriodType), period.StartDate, period.EndDate,
		string(period.Status), period.UpdatedAt, period.ID,
	)
	if err != nil {
		return kernel.WrapMsg(constant.CodeBillingPeriodPersistenceFailed, "gagal memperbarui periode tagihan", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.WrapMsg(constant.CodeBillingPeriodNotFound, "Periode tagihan tidak ditemukan", nil)
	}
	return nil
}

func (r *PostgresBillingPeriodRepository) FindByID(ctx context.Context, id string) (*entity.BillingPeriod, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+billingPeriodColumns+` FROM billing_periods WHERE id=$1`, id)
	return r.scan(row)
}

func (r *PostgresBillingPeriodRepository) List(ctx context.Context, q repository.BillingPeriodListQuery) (*repository.BillingPeriodListResult, error) {
	execer := execerFromContext(ctx, r.db)

	where := `WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if q.Status != nil && *q.Status != "" {
		where += fmt.Sprintf(` AND status=$%d`, argIdx)
		args = append(args, *q.Status)
		argIdx++
	}
	if q.AccountingPeriodID != nil && *q.AccountingPeriodID != "" {
		where += fmt.Sprintf(` AND accounting_period_id=$%d`, argIdx)
		args = append(args, *q.AccountingPeriodID)
		argIdx++
	}

	var total int64
	countRow := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM billing_periods `+where, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, kernel.WrapMsg(constant.CodeBillingPeriodQueryFailed, "gagal menghitung jumlah periode tagihan", err)
	}

	limit := q.Limit
	offset := (q.Page - 1) * q.Limit
	listArgs := append(append([]interface{}{}, args...), limit, offset)

	query := fmt.Sprintf(`SELECT %s FROM billing_periods %s ORDER BY start_date DESC LIMIT $%d OFFSET $%d`,
		billingPeriodColumns, where, argIdx, argIdx+1)

	rows, err := execer.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, kernel.WrapMsg(constant.CodeBillingPeriodQueryFailed, "gagal mendaftar periode tagihan", err)
	}
	defer rows.Close()

	items := make([]*entity.BillingPeriod, 0)
	for rows.Next() {
		period, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, period)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(constant.CodeBillingPeriodQueryFailed, "gagal membaca data periode tagihan", err)
	}

	return &repository.BillingPeriodListResult{Items: items, Total: total}, nil
}

func (r *PostgresBillingPeriodRepository) scan(sc scanner) (*entity.BillingPeriod, error) {
	var (
		id, name, periodType, accountingPeriodID, status, createdBy string
		startDate, endDate                                           time.Time
		createdAt, updatedAt                                         time.Time
	)

	err := sc.Scan(&id, &name, &periodType, &accountingPeriodID, &startDate, &endDate, &status, &createdBy, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.WrapMsg(constant.CodeBillingPeriodNotFound, "Periode tagihan tidak ditemukan", nil)
		}
		return nil, kernel.WrapMsg(constant.CodeBillingPeriodQueryFailed, "gagal membaca data periode tagihan", err)
	}

	return &entity.BillingPeriod{
		ID:                 id,
		Name:               name,
		PeriodType:         feeConst.PeriodType(periodType),
		AccountingPeriodID: accountingPeriodID,
		StartDate:          startDate,
		EndDate:            endDate,
		Status:             constant.BillingPeriodStatus(status),
		CreatedBy:          createdBy,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}, nil
}
