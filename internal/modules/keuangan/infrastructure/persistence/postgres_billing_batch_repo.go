package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sipon-be/internal/modules/keuangan/domain/billingbatch/constant"
	"sipon-be/internal/modules/keuangan/domain/billingbatch/entity"
	"sipon-be/internal/modules/keuangan/domain/billingbatch/repository"
	"sipon-be/internal/shared/kernel"
)

const billingBatchColumns = `
	id, name, billing_scheme_id, billing_period_id, status, created_by, created_at,
	completed_at, total_created, total_skipped, total_error
`

type PostgresBillingBatchRepository struct {
	db *sql.DB
}

func NewPostgresBillingBatchRepository(db *sql.DB) *PostgresBillingBatchRepository {
	return &PostgresBillingBatchRepository{db: db}
}

func (r *PostgresBillingBatchRepository) Save(ctx context.Context, batch *entity.BillingBatch) error {
	execer := execerFromContext(ctx, r.db)

	query := `INSERT INTO billing_batches (` + billingBatchColumns + `) VALUES (
		$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11
	)`

	_, err := execer.ExecContext(ctx, query,
		batch.ID, batch.Name, batch.BillingSchemeID, batch.BillingPeriodID,
		string(batch.Status), batch.CreatedBy, batch.CreatedAt, nullTimeVal(batch.CompletedAt),
		batch.TotalCreated, batch.TotalSkipped, batch.TotalError,
	)
	if err != nil {
		return kernel.WrapMsg(constant.CodeBillingBatchPersistenceFailed, "gagal menyimpan batch tagihan", err)
	}
	return nil
}

func (r *PostgresBillingBatchRepository) Update(ctx context.Context, batch *entity.BillingBatch) error {
	execer := execerFromContext(ctx, r.db)

	query := `UPDATE billing_batches SET
		status=$1, completed_at=$2, total_created=$3, total_skipped=$4, total_error=$5
		WHERE id=$6`

	res, err := execer.ExecContext(ctx, query,
		string(batch.Status), nullTimeVal(batch.CompletedAt),
		batch.TotalCreated, batch.TotalSkipped, batch.TotalError, batch.ID,
	)
	if err != nil {
		return kernel.WrapMsg(constant.CodeBillingBatchPersistenceFailed, "gagal memperbarui batch tagihan", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.WrapMsg(constant.CodeBillingBatchNotFound, "Batch tagihan tidak ditemukan", nil)
	}
	return nil
}

func (r *PostgresBillingBatchRepository) FindByID(ctx context.Context, id string) (*entity.BillingBatch, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+billingBatchColumns+` FROM billing_batches WHERE id=$1`, id)
	return r.scan(row)
}

func (r *PostgresBillingBatchRepository) List(ctx context.Context, q repository.BillingBatchListQuery) (*repository.BillingBatchListResult, error) {
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
	countRow := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM billing_batches `+where, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, kernel.WrapMsg(constant.CodeBillingBatchQueryFailed, "gagal menghitung jumlah batch tagihan", err)
	}

	limit := q.Limit
	offset := (q.Page - 1) * q.Limit
	listArgs := append(append([]interface{}{}, args...), limit, offset)

	query := fmt.Sprintf(`SELECT %s FROM billing_batches %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		billingBatchColumns, where, argIdx, argIdx+1)

	rows, err := execer.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, kernel.WrapMsg(constant.CodeBillingBatchQueryFailed, "gagal mendaftar batch tagihan", err)
	}
	defer rows.Close()

	items := make([]*entity.BillingBatch, 0)
	for rows.Next() {
		batch, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, batch)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(constant.CodeBillingBatchQueryFailed, "gagal membaca data batch tagihan", err)
	}

	return &repository.BillingBatchListResult{Items: items, Total: total}, nil
}

func (r *PostgresBillingBatchRepository) scan(sc scanner) (*entity.BillingBatch, error) {
	var (
		id, name, billingSchemeID, billingPeriodID, status, createdBy string
		createdAt                                                    time.Time
		completedAt                                                  sql.NullTime
		totalCreated, totalSkipped, totalError                       int
	)

	err := sc.Scan(&id, &name, &billingSchemeID, &billingPeriodID, &status, &createdBy, &createdAt,
		&completedAt, &totalCreated, &totalSkipped, &totalError)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.WrapMsg(constant.CodeBillingBatchNotFound, "Batch tagihan tidak ditemukan", nil)
		}
		return nil, kernel.WrapMsg(constant.CodeBillingBatchQueryFailed, "gagal membaca data batch tagihan", err)
	}

	return &entity.BillingBatch{
		ID:              id,
		Name:            name,
		BillingSchemeID: billingSchemeID,
		BillingPeriodID: billingPeriodID,
		Status:          constant.BillingBatchStatus(status),
		CreatedBy:       createdBy,
		CreatedAt:       createdAt,
		CompletedAt:     timeFromNull(completedAt),
		TotalCreated:    totalCreated,
		TotalSkipped:    totalSkipped,
		TotalError:      totalError,
	}, nil
}

// ——— BillingBatchTargetRepository ———

const billingBatchTargetColumns = `
	id, batch_id, santri_id, status, invoice_id, reason, processed_at
`

type PostgresBillingBatchTargetRepository struct {
	db *sql.DB
}

func NewPostgresBillingBatchTargetRepository(db *sql.DB) *PostgresBillingBatchTargetRepository {
	return &PostgresBillingBatchTargetRepository{db: db}
}

func (r *PostgresBillingBatchTargetRepository) SaveMany(ctx context.Context, targets []*entity.BillingBatchTarget) error {
	execer := execerFromContext(ctx, r.db)

	query := `INSERT INTO billing_batch_targets (` + billingBatchTargetColumns + `) VALUES ($1,$2,$3,$4,$5,$6,$7)`

	for _, t := range targets {
		_, err := execer.ExecContext(ctx, query,
			t.ID, t.BatchID, t.SantriID, string(t.Status), nullStr(t.InvoiceID), nullStr(t.Reason), nullTimeVal(t.ProcessedAt),
		)
		if err != nil {
			return kernel.WrapMsg(constant.CodeBillingBatchPersistenceFailed, "gagal menyimpan target batch tagihan", err)
		}
	}
	return nil
}

func (r *PostgresBillingBatchTargetRepository) UpdateTarget(ctx context.Context, target *entity.BillingBatchTarget) error {
	execer := execerFromContext(ctx, r.db)

	query := `UPDATE billing_batch_targets SET status=$1, invoice_id=$2, reason=$3, processed_at=$4 WHERE id=$5`

	res, err := execer.ExecContext(ctx, query,
		string(target.Status), nullStr(target.InvoiceID), nullStr(target.Reason), nullTimeVal(target.ProcessedAt), target.ID,
	)
	if err != nil {
		return kernel.WrapMsg(constant.CodeBillingBatchPersistenceFailed, "gagal memperbarui target batch tagihan", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.WrapMsg(constant.CodeBillingBatchNotFound, "Batch tagihan tidak ditemukan", nil)
	}
	return nil
}

func (r *PostgresBillingBatchTargetRepository) FindByBatchID(ctx context.Context, batchID string) ([]*entity.BillingBatchTarget, error) {
	execer := execerFromContext(ctx, r.db)

	rows, err := execer.QueryContext(ctx,
		`SELECT `+billingBatchTargetColumns+` FROM billing_batch_targets WHERE batch_id=$1 ORDER BY id ASC`,
		batchID,
	)
	if err != nil {
		return nil, kernel.WrapMsg(constant.CodeBillingBatchQueryFailed, "gagal mencari target batch tagihan", err)
	}
	defer rows.Close()

	items := make([]*entity.BillingBatchTarget, 0)
	for rows.Next() {
		var (
			id, batchID, santriID, status string
			invoiceID, reason             sql.NullString
			processedAt                   sql.NullTime
		)
		if err := rows.Scan(&id, &batchID, &santriID, &status, &invoiceID, &reason, &processedAt); err != nil {
			return nil, kernel.WrapMsg(constant.CodeBillingBatchQueryFailed, "gagal membaca data target batch tagihan", err)
		}
		items = append(items, &entity.BillingBatchTarget{
			ID:          id,
			BatchID:     batchID,
			SantriID:    santriID,
			Status:      constant.BillingBatchTargetStatus(status),
			InvoiceID:   strFromNull(invoiceID),
			Reason:      strFromNull(reason),
			ProcessedAt: timeFromNull(processedAt),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(constant.CodeBillingBatchQueryFailed, "gagal membaca data target batch tagihan", err)
	}
	return items, nil
}
