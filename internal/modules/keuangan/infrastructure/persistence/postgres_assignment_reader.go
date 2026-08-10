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

func (r *PostgresAssignmentReader) ListAll(ctx context.Context, santriID *string) ([]ports.AssignmentReadModel, error) {
	execer := execerFromContext(ctx, r.db)

	where := ""
	args := []interface{}{}
	if santriID != nil {
		where = ` WHERE santri_id=$1`
		args = append(args, *santriID)
	}

	rows, err := execer.QueryContext(ctx, `
		SELECT id, santri_id, billing_scheme_id, effective_from, effective_until, assigned_by, created_at
		FROM santri_billing_assignments`+where+`
		ORDER BY created_at DESC
	`, args...)
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
