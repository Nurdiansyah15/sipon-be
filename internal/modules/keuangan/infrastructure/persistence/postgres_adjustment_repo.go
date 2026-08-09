package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"sipon-be/internal/modules/keuangan/domain/adjustment/constant"
	"sipon-be/internal/modules/keuangan/domain/adjustment/entity"
	"sipon-be/internal/shared/kernel"
)

const adjustmentColumns = `
	id, invoice_id, type, amount, percentage, description, applied_by, applied_at
`

type PostgresAdjustmentRepository struct {
	db *sql.DB
}

func NewPostgresAdjustmentRepository(db *sql.DB) *PostgresAdjustmentRepository {
	return &PostgresAdjustmentRepository{db: db}
}

func (r *PostgresAdjustmentRepository) Save(ctx context.Context, adj *entity.InvoiceAdjustment) error {
	execer := execerFromContext(ctx, r.db)

	query := `INSERT INTO invoice_adjustments (` + adjustmentColumns + `) VALUES (
		$1,$2,$3,$4,$5,$6,$7,$8
	)`

	_, err := execer.ExecContext(ctx, query,
		adj.ID, adj.InvoiceID, string(adj.Type), adj.Amount,
		nullFloat64(adj.Percentage), nullStr(adj.Description), adj.AppliedBy, adj.AppliedAt,
	)
	if err != nil {
		return kernel.WrapMsg(constant.CodeAdjustmentPersistenceFailed, "gagal menyimpan penyesuaian", err)
	}
	return nil
}

func (r *PostgresAdjustmentRepository) FindByInvoiceID(ctx context.Context, invoiceID string) ([]*entity.InvoiceAdjustment, error) {
	execer := execerFromContext(ctx, r.db)

	rows, err := execer.QueryContext(ctx,
		`SELECT `+adjustmentColumns+` FROM invoice_adjustments WHERE invoice_id=$1 ORDER BY applied_at DESC`,
		invoiceID,
	)
	if err != nil {
		return nil, kernel.WrapMsg(constant.CodeAdjustmentQueryFailed, "gagal mencari penyesuaian invoice", err)
	}
	defer rows.Close()

	items := make([]*entity.InvoiceAdjustment, 0)
	for rows.Next() {
		adj, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, adj)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(constant.CodeAdjustmentQueryFailed, "gagal membaca data penyesuaian", err)
	}
	return items, nil
}

func (r *PostgresAdjustmentRepository) scan(sc scanner) (*entity.InvoiceAdjustment, error) {
	var (
		id, invoiceID, adjType, appliedBy    string
		amount                                float64
		percentage                            sql.NullFloat64
		description                           sql.NullString
		appliedAt                             time.Time
	)

	err := sc.Scan(&id, &invoiceID, &adjType, &amount, &percentage, &description, &appliedBy, &appliedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.WrapMsg(constant.CodeAdjustmentNotFound, "Penyesuaian tidak ditemukan", nil)
		}
		return nil, kernel.WrapMsg(constant.CodeAdjustmentQueryFailed, "gagal membaca data penyesuaian", err)
	}

	return &entity.InvoiceAdjustment{
		ID:          id,
		InvoiceID:   invoiceID,
		Type:        constant.AdjustmentType(adjType),
		Amount:      amount,
		Percentage:  float64FromNull(percentage),
		Description: strFromNull(description),
		AppliedBy:   appliedBy,
		AppliedAt:   appliedAt,
	}, nil
}
