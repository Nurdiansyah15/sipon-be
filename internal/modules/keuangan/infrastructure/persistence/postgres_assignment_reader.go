package persistence

import (
	"context"
	"database/sql"
	"time"

	"sipon-be/internal/modules/keuangan/application/ports"
	"sipon-be/internal/modules/keuangan/domain/billingscheme/constant"
	"sipon-be/internal/shared/kernel"
)

type PostgresAssignmentReader struct {
	db *sql.DB
}

func NewPostgresAssignmentReader(db *sql.DB) *PostgresAssignmentReader {
	return &PostgresAssignmentReader{db: db}
}

func (r *PostgresAssignmentReader) ListActive(ctx context.Context) ([]ports.AssignmentReadModel, error) {
	execer := execerFromContext(ctx, r.db)

	rows, err := execer.QueryContext(ctx, `
		SELECT id, santri_id, billing_scheme_id, effective_from, effective_until, assigned_by, created_at
		FROM santri_billing_assignments
		WHERE effective_until IS NULL OR effective_until >= CURRENT_DATE
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, kernel.WrapMsg(constant.CodeBillingSchemeQueryFailed, "gagal mendaftar penugasan skema tagihan", err)
	}
	defer rows.Close()

	results := make([]ports.AssignmentReadModel, 0)
	for rows.Next() {
		var a ports.AssignmentReadModel
		var effectiveFrom sql.NullTime
		var effectiveUntil sql.NullTime

		if err := rows.Scan(&a.ID, &a.SantriID, &a.BillingSchemeID, &effectiveFrom, &effectiveUntil, &a.AssignedBy, &a.CreatedAt); err != nil {
			return nil, kernel.WrapMsg(constant.CodeBillingSchemeQueryFailed, "gagal membaca data penugasan skema tagihan", err)
		}

		if effectiveFrom.Valid {
			a.EffectiveFrom = effectiveFrom.Time
		} else {
			a.EffectiveFrom = time.Time{}
		}
		if effectiveUntil.Valid {
			until := effectiveUntil.Time
			a.EffectiveUntil = &until
		}

		results = append(results, a)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(constant.CodeBillingSchemeQueryFailed, "gagal membaca data penugasan skema tagihan", err)
	}
	return results, nil
}
