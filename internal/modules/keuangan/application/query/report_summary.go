package query

import (
	"context"
	"database/sql"
	"fmt"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/shared/kernel"
)

type ReportSummaryUseCase struct {
	db *sql.DB
}

func NewReportSummaryUseCase(db *sql.DB) *ReportSummaryUseCase {
	return &ReportSummaryUseCase{db: db}
}

func (uc *ReportSummaryUseCase) Execute(ctx context.Context, query dto.InvoiceSummaryQuery) ([]dto.InvoiceSummaryResponse, error) {
	where := "WHERE i.deleted_at IS NULL"
	args := []interface{}{}
	argIdx := 1

	if query.TahunAjaran != nil && *query.TahunAjaran != "" {
		where += fmt.Sprintf(" AND i.tahun_ajaran = $%d", argIdx)
		args = append(args, *query.TahunAjaran)
		argIdx++
	}
	if query.Periode != nil && *query.Periode != "" {
		where += fmt.Sprintf(" AND i.periode = $%d", argIdx)
		args = append(args, *query.Periode)
		argIdx++
	}

	sqlQuery := `SELECT 
		i.tahun_ajaran, i.periode,
		COALESCE(SUM(i.amount), 0) as total_tagihan,
		COALESCE(SUM(i.paid_amount), 0) as total_terbayar,
		COALESCE(SUM(i.amount - i.discount_amount - i.paid_amount), 0) as total_tunggakan,
		COUNT(*) as jumlah_invoice,
		COUNT(CASE WHEN i.status = 'paid' THEN 1 END) as jumlah_lunas,
		COUNT(CASE WHEN i.status != 'paid' AND i.status != 'cancelled' THEN 1 END) as jumlah_belum
	FROM invoices i ` + where + `
	GROUP BY i.tahun_ajaran, i.periode
	ORDER BY i.tahun_ajaran DESC, i.periode DESC`

	rows, err := uc.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	defer rows.Close()

	results := make([]dto.InvoiceSummaryResponse, 0)
	for rows.Next() {
		var r dto.InvoiceSummaryResponse
		if err := rows.Scan(&r.TahunAjaran, &r.Periode, &r.TotalTagihan, &r.TotalTerbayar, &r.TotalTunggakan, &r.JumlahInvoice, &r.JumlahLunas, &r.JumlahBelum); err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	return results, nil
}
