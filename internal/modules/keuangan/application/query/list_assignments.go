package query

import (
	"context"
	"database/sql"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/shared/kernel"
)

type ListAssignmentsUseCase struct {
	db *sql.DB
}

func NewListAssignmentsUseCase(db *sql.DB) *ListAssignmentsUseCase {
	return &ListAssignmentsUseCase{db: db}
}

func (uc *ListAssignmentsUseCase) Execute(ctx context.Context) ([]dto.AssignmentResponse, error) {
	query := `
		SELECT id, santri_id, billing_scheme_id, effective_from, effective_until, assigned_by, created_at
		FROM santri_billing_assignments
		WHERE effective_until IS NULL OR effective_until >= CURRENT_DATE
		ORDER BY created_at DESC
	`

	rows, err := uc.db.QueryContext(ctx, query)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	defer rows.Close()

	var results []dto.AssignmentResponse
	for rows.Next() {
		var a dto.AssignmentResponse
		var effectiveFrom sql.NullTime
		var effectiveUntil sql.NullTime

		if err := rows.Scan(&a.ID, &a.SantriID, &a.BillingSchemeID, &effectiveFrom, &effectiveUntil, &a.AssignedBy, &a.CreatedAt); err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}

		if effectiveFrom.Valid {
			a.EffectiveFrom = effectiveFrom.Time.Format("2006-01-02")
		}
		if effectiveUntil.Valid {
			s := effectiveUntil.Time.Format("2006-01-02")
			a.EffectiveUntil = &s
		}

		results = append(results, a)
	}

	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if results == nil {
		results = []dto.AssignmentResponse{}
	}

	return results, nil
}
