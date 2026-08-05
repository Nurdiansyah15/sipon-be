package query

import (
	"context"
	"database/sql"
	"fmt"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/shared/kernel"
)

type ReportOutstandingUseCase struct {
	db *sql.DB
}

func NewReportOutstandingUseCase(db *sql.DB) *ReportOutstandingUseCase {
	return &ReportOutstandingUseCase{db: db}
}

func (uc *ReportOutstandingUseCase) Execute(ctx context.Context, query dto.OutstandingListQuery) ([]dto.OutstandingSantriResponse, *dto.Meta, error) {
	page := query.Page
	limit := query.Limit
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit

	where := "WHERE i.deleted_at IS NULL AND i.status != 'paid' AND i.status != 'cancelled'"
	args := []interface{}{}
	argIdx := 1

	if query.TahunAjaran != nil && *query.TahunAjaran != "" {
		where += fmt.Sprintf(" AND i.tahun_ajaran = $%d", argIdx)
		args = append(args, *query.TahunAjaran)
		argIdx++
	}

	countQuery := `SELECT COUNT(DISTINCT i.santri_id) FROM invoices i ` + where
	var total int64
	if err := uc.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	dataQuery := `SELECT 
		i.santri_id,
		COALESCE(SUM(i.amount - i.discount_amount - i.paid_amount), 0) as total_outstanding,
		COUNT(*) as jumlah_invoice
	FROM invoices i ` + where + `
	GROUP BY i.santri_id
	ORDER BY total_outstanding DESC
	LIMIT $` + fmt.Sprintf("%d", argIdx) + ` OFFSET $` + fmt.Sprintf("%d", argIdx+1)
	args = append(args, limit, offset)

	rows, err := uc.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	defer rows.Close()

	results := make([]dto.OutstandingSantriResponse, 0)
	for rows.Next() {
		var r dto.OutstandingSantriResponse
		if err := rows.Scan(&r.SantriID, &r.TotalOutstanding, &r.JumlahInvoice); err != nil {
			return nil, nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	meta := &dto.Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	return results, meta, nil
}
