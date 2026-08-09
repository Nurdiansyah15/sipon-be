package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"sipon-be/internal/modules/keuangan/application/ports"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	journalConst "sipon-be/internal/modules/keuangan/domain/journal/constant"
	"sipon-be/internal/shared/kernel"
)

type PostgresReportReader struct {
	db *sql.DB
}

func NewPostgresReportReader(db *sql.DB) *PostgresReportReader {
	return &PostgresReportReader{db: db}
}

func (r *PostgresReportReader) InvoiceSummary(ctx context.Context, billingPeriodID *string) ([]ports.InvoiceSummaryReadModel, error) {
	execer := execerFromContext(ctx, r.db)

	where := "WHERE i.deleted_at IS NULL"
	args := []interface{}{}
	argIdx := 1

	if billingPeriodID != nil && *billingPeriodID != "" {
		where += fmt.Sprintf(" AND i.billing_period_id = $%d", argIdx)
		args = append(args, *billingPeriodID)
		argIdx++
	}

	sqlQuery := `SELECT
		COALESCE(bp.id::text, '') as billing_period_id,
		COALESCE(bp.name, 'Non Periodik') as billing_period_name,
		COALESCE(SUM(i.amount), 0) as total_tagihan,
		COALESCE(SUM(i.paid_amount), 0) as total_terbayar,
		COALESCE(SUM(i.amount - i.discount_amount - i.paid_amount), 0) as total_tunggakan,
		COUNT(*) as jumlah_invoice,
		COUNT(CASE WHEN i.status = 'paid' THEN 1 END) as jumlah_lunas,
		COUNT(CASE WHEN i.status != 'paid' AND i.status != 'cancelled' THEN 1 END) as jumlah_belum
	FROM invoices i
	LEFT JOIN billing_periods bp ON bp.id = i.billing_period_id ` + where + `
	GROUP BY COALESCE(bp.id::text, ''), COALESCE(bp.name, 'Non Periodik'), COALESCE(bp.start_date, DATE '0001-01-01')
	ORDER BY COALESCE(bp.start_date, DATE '0001-01-01') DESC`

	rows, err := execer.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, kernel.WrapMsg(invConst.CodeInvoiceQueryFailed, "gagal membuat rekap tagihan", err)
	}
	defer rows.Close()

	results := make([]ports.InvoiceSummaryReadModel, 0)
	for rows.Next() {
		var rm ports.InvoiceSummaryReadModel
		if err := rows.Scan(&rm.BillingPeriodID, &rm.BillingPeriodName, &rm.TotalTagihan, &rm.TotalTerbayar, &rm.TotalTunggakan, &rm.JumlahInvoice, &rm.JumlahLunas, &rm.JumlahBelum); err != nil {
			return nil, kernel.WrapMsg(invConst.CodeInvoiceQueryFailed, "gagal membaca data rekap tagihan", err)
		}
		results = append(results, rm)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(invConst.CodeInvoiceQueryFailed, "gagal membaca data rekap tagihan", err)
	}
	return results, nil
}

func (r *PostgresReportReader) OutstandingBySantri(ctx context.Context, billingPeriodID *string, page, limit int) ([]ports.OutstandingReadModel, int64, error) {
	execer := execerFromContext(ctx, r.db)

	offset := (page - 1) * limit

	where := "WHERE i.deleted_at IS NULL AND i.status != 'paid' AND i.status != 'cancelled'"
	args := []interface{}{}
	argIdx := 1

	if billingPeriodID != nil && *billingPeriodID != "" {
		where += fmt.Sprintf(" AND i.billing_period_id = $%d", argIdx)
		args = append(args, *billingPeriodID)
		argIdx++
	}

	countQuery := `SELECT COUNT(DISTINCT i.santri_id) FROM invoices i ` + where
	var total int64
	if err := execer.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, kernel.WrapMsg(invConst.CodeInvoiceQueryFailed, "gagal menghitung jumlah santri dengan tunggakan", err)
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

	rows, err := execer.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, kernel.WrapMsg(invConst.CodeInvoiceQueryFailed, "gagal membuat laporan tunggakan", err)
	}
	defer rows.Close()

	results := make([]ports.OutstandingReadModel, 0)
	for rows.Next() {
		var rm ports.OutstandingReadModel
		if err := rows.Scan(&rm.SantriID, &rm.TotalOutstanding, &rm.JumlahInvoice); err != nil {
			return nil, 0, kernel.WrapMsg(invConst.CodeInvoiceQueryFailed, "gagal membaca data tunggakan", err)
		}
		results = append(results, rm)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, kernel.WrapMsg(invConst.CodeInvoiceQueryFailed, "gagal membaca data tunggakan", err)
	}
	return results, total, nil
}

func (r *PostgresReportReader) LedgerLines(ctx context.Context, accountID string, from, to time.Time) ([]ports.LedgerLineReadModel, error) {
	execer := execerFromContext(ctx, r.db)

	rows, err := execer.QueryContext(ctx, `
		SELECT 
			je.entry_date::date::text,
			je.journal_number,
			COALESCE(je.description, ''),
			COALESCE(jel.debit, 0),
			COALESCE(jel.credit, 0)
		FROM journal_entry_lines jel
		JOIN journal_entries je ON je.id = jel.journal_entry_id
		WHERE jel.account_id = $1 
			AND je.entry_date >= $2
			AND je.entry_date <= $3
			AND je.status = 'posted'
		ORDER BY je.entry_date ASC, je.journal_number ASC`,
		accountID, from, to,
	)
	if err != nil {
		return nil, kernel.WrapMsg(journalConst.CodeJournalQueryFailed, "gagal memuat baris buku besar", err)
	}
	defer rows.Close()

	results := make([]ports.LedgerLineReadModel, 0)
	for rows.Next() {
		var line ports.LedgerLineReadModel
		var dateStr string
		if err := rows.Scan(&dateStr, &line.JournalNumber, &line.Description, &line.Debit, &line.Credit); err != nil {
			return nil, kernel.WrapMsg(journalConst.CodeJournalQueryFailed, "gagal membaca baris buku besar", err)
		}
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, kernel.WrapMsg(journalConst.CodeJournalQueryFailed, "gagal membaca tanggal baris buku besar", err)
		}
		line.Date = date
		results = append(results, line)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(journalConst.CodeJournalQueryFailed, "gagal membaca baris buku besar", err)
	}
	return results, nil
}

func (r *PostgresReportReader) BalanceBefore(ctx context.Context, accountID string, before time.Time) (debit, credit float64, err error) {
	execer := execerFromContext(ctx, r.db)

	sqlQuery := `SELECT COALESCE(SUM(jel.debit), 0) AS total_debit, COALESCE(SUM(jel.credit), 0) AS total_credit
		FROM journal_entry_lines jel
		JOIN journal_entries je ON je.id = jel.journal_entry_id
		WHERE jel.account_id = $1
			AND je.entry_date < $2
			AND je.status = 'posted'`

	if err := execer.QueryRowContext(ctx, sqlQuery, accountID, before).Scan(&debit, &credit); err != nil {
		return 0, 0, kernel.WrapMsg(journalConst.CodeJournalQueryFailed, "gagal menghitung saldo awal akun", err)
	}
	return debit, credit, nil
}

func (r *PostgresReportReader) AccountBalancesToDate(ctx context.Context, asOfDate *string) ([]ports.AccountBalanceReadModel, error) {
	execer := execerFromContext(ctx, r.db)

	sqlQuery := `SELECT 
		jel.account_id,
		COALESCE(SUM(jel.debit), 0) as total_debit,
		COALESCE(SUM(jel.credit), 0) as total_credit
	FROM journal_entry_lines jel
	JOIN journal_entries je ON je.id = jel.journal_entry_id
	WHERE je.status = 'posted'`
	args := []interface{}{}
	if asOfDate != nil && *asOfDate != "" {
		sqlQuery += ` AND je.entry_date <= $1`
		args = append(args, *asOfDate)
	}
	sqlQuery += ` GROUP BY jel.account_id`

	rows, err := execer.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, kernel.WrapMsg(journalConst.CodeJournalQueryFailed, "gagal menghitung saldo akun", err)
	}
	defer rows.Close()

	results := make([]ports.AccountBalanceReadModel, 0)
	for rows.Next() {
		var bal ports.AccountBalanceReadModel
		if err := rows.Scan(&bal.AccountID, &bal.Debit, &bal.Credit); err != nil {
			return nil, kernel.WrapMsg(journalConst.CodeJournalQueryFailed, "gagal membaca saldo akun", err)
		}
		results = append(results, bal)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(journalConst.CodeJournalQueryFailed, "gagal membaca saldo akun", err)
	}
	return results, nil
}

func (r *PostgresReportReader) AccountBalancesByPeriod(ctx context.Context, periodID string) ([]ports.AccountBalanceReadModel, error) {
	execer := execerFromContext(ctx, r.db)

	rows, err := execer.QueryContext(ctx, `
		SELECT 
			jel.account_id,
			COALESCE(SUM(jel.debit), 0) as total_debit,
			COALESCE(SUM(jel.credit), 0) as total_credit
		FROM journal_entry_lines jel
		JOIN journal_entries je ON je.id = jel.journal_entry_id
		WHERE je.period_id = $1
			AND je.status = 'posted'
			AND je.source_type IS DISTINCT FROM 'closing'
		GROUP BY jel.account_id`,
		periodID,
	)
	if err != nil {
		return nil, kernel.WrapMsg(journalConst.CodeJournalQueryFailed, "gagal menghitung saldo akun per periode", err)
	}
	defer rows.Close()

	results := make([]ports.AccountBalanceReadModel, 0)
	for rows.Next() {
		var bal ports.AccountBalanceReadModel
		if err := rows.Scan(&bal.AccountID, &bal.Debit, &bal.Credit); err != nil {
			return nil, kernel.WrapMsg(journalConst.CodeJournalQueryFailed, "gagal membaca saldo akun per periode", err)
		}
		results = append(results, bal)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(journalConst.CodeJournalQueryFailed, "gagal membaca saldo akun per periode", err)
	}
	return results, nil
}
