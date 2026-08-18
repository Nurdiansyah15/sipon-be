package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sipon-be/internal/modules/keuangan/domain/invoice/constant"
	"sipon-be/internal/modules/keuangan/domain/invoice/entity"
	"sipon-be/internal/modules/keuangan/domain/invoice/repository"
	invVO "sipon-be/internal/modules/keuangan/domain/invoice/valueobject"
	"sipon-be/internal/shared/kernel"
)

const invoiceColumns = `
	id, invoice_number, santri_id, user_id, billing_scheme_id, fee_component_id,
	billing_period_id, amount, discount_amount, paid_amount, status,
	due_date, issued_at, notes, created_by, created_at, updated_at, deleted_at
`

type PostgresInvoiceRepository struct {
	db *sql.DB
}

func NewPostgresInvoiceRepository(db *sql.DB) *PostgresInvoiceRepository {
	return &PostgresInvoiceRepository{db: db}
}

func (r *PostgresInvoiceRepository) NextInvoiceNumber(ctx context.Context) (invVO.InvoiceNumber, error) {
	execer := execerFromContext(ctx, r.db)
	now := time.Now()
	year := now.Year()

	seq, err := nextNumberSeq(ctx, execer, "invoice", year)
	if err != nil {
		return invVO.InvoiceNumber{}, kernel.WrapMsg(constant.CodeInvoicePersistenceFailed, "gagal membuat nomor invoice", err)
	}

	return invVO.NewInvoiceNumber(fmt.Sprintf("%d", year), fmt.Sprintf("%02d", int(now.Month())), seq), nil
}

func (r *PostgresInvoiceRepository) Save(ctx context.Context, inv *entity.Invoice) error {
	execer := execerFromContext(ctx, r.db)

	query := `INSERT INTO invoices (` + invoiceColumns + `) VALUES (
		$1,$2,$3,$4,$5,$6,
		$7,$8,$9,$10,$11,
		$12,$13,$14,$15,$16,$17,$18
	)`

	_, err := execer.ExecContext(ctx, query,
		inv.ID, inv.InvoiceNumber, inv.SantriID, inv.UserID,
		nullStr(inv.BillingSchemeID), inv.FeeComponentID,
		nullStr(inv.BillingPeriodID), inv.Amount, inv.DiscountAmount, inv.PaidAmount,
		string(inv.Status), inv.DueDate, nullTimeVal(inv.IssuedAt), nullStr(inv.Notes),
		inv.CreatedBy, inv.CreatedAt, inv.UpdatedAt, nullTimeVal(inv.DeletedAt),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.WrapMsg(constant.CodeInvoiceDuplicate, "Invoice duplikat", err)
		}
		return kernel.WrapMsg(constant.CodeInvoicePersistenceFailed, "gagal menyimpan invoice", err)
	}
	return nil
}

func (r *PostgresInvoiceRepository) Update(ctx context.Context, inv *entity.Invoice) error {
	execer := execerFromContext(ctx, r.db)

	query := `UPDATE invoices SET
		santri_id=$1, user_id=$2, billing_scheme_id=$3, fee_component_id=$4,
		billing_period_id=$5, amount=$6, discount_amount=$7, paid_amount=$8,
		status=$9, due_date=$10, issued_at=$11, notes=$12,
		updated_at=$13, deleted_at=$14
		WHERE id=$15 AND deleted_at IS NULL`

	res, err := execer.ExecContext(ctx, query,
		inv.SantriID, inv.UserID, nullStr(inv.BillingSchemeID), inv.FeeComponentID,
		nullStr(inv.BillingPeriodID), inv.Amount, inv.DiscountAmount, inv.PaidAmount,
		string(inv.Status), inv.DueDate, nullTimeVal(inv.IssuedAt), nullStr(inv.Notes),
		inv.UpdatedAt, nullTimeVal(inv.DeletedAt),
		inv.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.WrapMsg(constant.CodeInvoiceDuplicate, "Invoice duplikat", err)
		}
		return kernel.WrapMsg(constant.CodeInvoicePersistenceFailed, "gagal memperbarui invoice", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.WrapMsg(constant.CodeInvoiceNotFound, "Invoice tidak ditemukan", nil)
	}
	return nil
}

func (r *PostgresInvoiceRepository) FindByID(ctx context.Context, id string) (*entity.Invoice, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+invoiceColumns+` FROM invoices WHERE id=$1 AND deleted_at IS NULL`, id)
	return r.scan(row)
}

func (r *PostgresInvoiceRepository) FindByNumber(ctx context.Context, number string) (*entity.Invoice, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+invoiceColumns+` FROM invoices WHERE invoice_number=$1 AND deleted_at IS NULL`, number)
	return r.scan(row)
}

func (r *PostgresInvoiceRepository) List(ctx context.Context, q repository.InvoiceListQuery) (*repository.InvoiceListResult, error) {
	execer := execerFromContext(ctx, r.db)

	where := `WHERE deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 1
	if q.SantriID != nil && *q.SantriID != "" {
		where += fmt.Sprintf(` AND santri_id=$%d`, argIdx)
		args = append(args, *q.SantriID)
		argIdx++
	}
	if q.UserID != nil && *q.UserID != "" {
		where += fmt.Sprintf(` AND user_id=$%d`, argIdx)
		args = append(args, *q.UserID)
		argIdx++
	}
	if q.Status != nil && *q.Status != "" {
		where += fmt.Sprintf(` AND status=$%d`, argIdx)
		args = append(args, *q.Status)
		argIdx++
	}
	if q.BillingPeriodID != nil && *q.BillingPeriodID != "" {
		where += fmt.Sprintf(` AND billing_period_id=$%d`, argIdx)
		args = append(args, *q.BillingPeriodID)
		argIdx++
	}
	if q.PeriodID != nil && *q.PeriodID != "" {
		where += fmt.Sprintf(` AND billing_period_id IN (
			SELECT id FROM billing_periods WHERE accounting_period_id=$%d
		)`, argIdx)
		args = append(args, *q.PeriodID)
		argIdx++
	}

	var total int64
	countRow := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM invoices `+where, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, kernel.WrapMsg(constant.CodeInvoiceQueryFailed, "gagal menghitung jumlah invoice", err)
	}

	limit := q.Limit
	offset := (q.Page - 1) * q.Limit
	listArgs := append(append([]interface{}{}, args...), limit, offset)

	query := fmt.Sprintf(`SELECT %s FROM invoices %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		invoiceColumns, where, argIdx, argIdx+1)

	rows, err := execer.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, kernel.WrapMsg(constant.CodeInvoiceQueryFailed, "gagal mendaftar invoice", err)
	}
	defer rows.Close()

	items := make([]*entity.Invoice, 0)
	for rows.Next() {
		inv, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, inv)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(constant.CodeInvoiceQueryFailed, "gagal membaca data invoice", err)
	}

	return &repository.InvoiceListResult{Items: items, Total: total}, nil
}

func (r *PostgresInvoiceRepository) FindBySantriComponentPeriod(ctx context.Context, santriID, feeComponentID, billingPeriodID string) (*entity.Invoice, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx,
		`SELECT `+invoiceColumns+` FROM invoices WHERE santri_id=$1 AND fee_component_id=$2 AND billing_period_id=$3 AND deleted_at IS NULL`,
		santriID, feeComponentID, billingPeriodID,
	)
	return r.scan(row)
}

func (r *PostgresInvoiceRepository) FindOutstandingBySantriID(ctx context.Context, santriID string) ([]*entity.Invoice, error) {
	execer := execerFromContext(ctx, r.db)

	rows, err := execer.QueryContext(ctx,
		`SELECT `+invoiceColumns+` FROM invoices WHERE santri_id=$1 AND status IN ('issued','partial','expired') AND deleted_at IS NULL ORDER BY due_date ASC`,
		santriID,
	)
	if err != nil {
		return nil, kernel.WrapMsg(constant.CodeInvoiceQueryFailed, "gagal mencari invoice outstanding santri", err)
	}
	defer rows.Close()

	items := make([]*entity.Invoice, 0)
	for rows.Next() {
		inv, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, inv)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(constant.CodeInvoiceQueryFailed, "gagal membaca data invoice outstanding", err)
	}
	return items, nil
}

func (r *PostgresInvoiceRepository) HasPaidComponent(ctx context.Context, santriID, componentCode, billingPeriodID string) (bool, error) {
	execer := execerFromContext(ctx, r.db)

	var exists bool
	err := execer.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM invoices i JOIN fee_components fc ON fc.id=i.fee_component_id WHERE i.santri_id=$1 AND fc.code=$2 AND i.billing_period_id=$3 AND i.status IN ('paid','partial') AND i.deleted_at IS NULL)`,
		santriID, componentCode, billingPeriodID,
	).Scan(&exists)
	if err != nil {
		return false, kernel.WrapMsg(constant.CodeInvoiceQueryFailed, "gagal memeriksa komponen yang telah dibayar", err)
	}
	return exists, nil
}

func (r *PostgresInvoiceRepository) FindSummaryByUserID(ctx context.Context, userID string) (*repository.InvoiceSummary, error) {
	execer := execerFromContext(ctx, r.db)

	row := execer.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN status != 'cancelled' THEN amount ELSE 0 END), 0) AS total_tagihan,
			COALESCE(SUM(CASE WHEN status != 'cancelled' THEN paid_amount ELSE 0 END), 0) AS total_terbayar,
			COALESCE(SUM(CASE WHEN status IN ('issued','partial','expired') THEN GREATEST(amount - discount_amount - paid_amount, 0) ELSE 0 END), 0) AS total_tunggakan,
			COUNT(CASE WHEN status != 'cancelled' THEN 1 END) AS jumlah_invoice,
			COUNT(CASE WHEN status = 'paid' THEN 1 END) AS jumlah_lunas,
			COUNT(CASE WHEN status NOT IN ('paid','cancelled') THEN 1 END) AS jumlah_belum
		FROM invoices
		WHERE user_id=$1 AND deleted_at IS NULL
	`, userID)

	var s repository.InvoiceSummary
	if err := row.Scan(&s.TotalTagihan, &s.TotalTerbayar, &s.TotalTunggakan, &s.JumlahInvoice, &s.JumlahLunas, &s.JumlahBelum); err != nil {
		return nil, kernel.WrapMsg(constant.CodeInvoiceQueryFailed, "gagal menghitung ringkasan invoice", err)
	}
	return &s, nil
}

func (r *PostgresInvoiceRepository) scan(sc scanner) (*entity.Invoice, error) {
	var (
		id, invoiceNumber, santriID, userID, feeComponentID string
		billingSchemeID, billingPeriodID, notes             sql.NullString
		amount, discountAmount, paidAmount                  float64
		status                                              string
		dueDate                                             time.Time
		issuedAt                                            sql.NullTime
		createdBy                                           string
		createdAt, updatedAt                                time.Time
		deletedAt                                           sql.NullTime
	)

	err := sc.Scan(
		&id, &invoiceNumber, &santriID, &userID, &billingSchemeID, &feeComponentID,
		&billingPeriodID, &amount, &discountAmount, &paidAmount, &status,
		&dueDate, &issuedAt, &notes, &createdBy, &createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.WrapMsg(constant.CodeInvoiceNotFound, "Invoice tidak ditemukan", nil)
		}
		return nil, kernel.WrapMsg(constant.CodeInvoiceQueryFailed, "gagal membaca data invoice", err)
	}

	return &entity.Invoice{
		ID:              id,
		InvoiceNumber:   invoiceNumber,
		SantriID:        santriID,
		UserID:          userID,
		BillingSchemeID: strFromNull(billingSchemeID),
		FeeComponentID:  feeComponentID,
		BillingPeriodID: strFromNull(billingPeriodID),
		Amount:          amount,
		DiscountAmount:  discountAmount,
		PaidAmount:      paidAmount,
		Status:          constant.InvoiceStatus(status),
		DueDate:         dueDate,
		IssuedAt:        timeFromNull(issuedAt),
		Notes:           strFromNull(notes),
		CreatedBy:       createdBy,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
		DeletedAt:       timeFromNull(deletedAt),
	}, nil
}
