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
	"sipon-be/internal/shared/kernel"
)

const invoiceColumns = `
	id, invoice_number, santri_id, user_id, billing_scheme_id, fee_component_id,
	periode, tahun_ajaran, amount, discount_amount, paid_amount, status,
	due_date, issued_at, notes, created_by, created_at, updated_at, deleted_at
`

type PostgresInvoiceRepository struct {
	db *sql.DB
}

func NewPostgresInvoiceRepository(db *sql.DB) *PostgresInvoiceRepository {
	return &PostgresInvoiceRepository{db: db}
}

func (r *PostgresInvoiceRepository) Save(ctx context.Context, inv *entity.Invoice) error {
	execer := execerFromContext(ctx, r.db)

	query := `INSERT INTO invoices (` + invoiceColumns + `) VALUES (
		$1,$2,$3,$4,$5,$6,
		$7,$8,$9,$10,$11,$12,
		$13,$14,$15,$16,$17,$18,$19
	)`

	_, err := execer.ExecContext(ctx, query,
		inv.ID, inv.InvoiceNumber, inv.SantriID, inv.UserID,
		nullStr(inv.BillingSchemeID), inv.FeeComponentID,
		inv.Periode, inv.TahunAjaran, inv.Amount, inv.DiscountAmount, inv.PaidAmount,
		string(inv.Status), inv.DueDate, nullTimeVal(inv.IssuedAt), nullStr(inv.Notes),
		inv.CreatedBy, inv.CreatedAt, inv.UpdatedAt, nullTimeVal(inv.DeletedAt),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeInvoiceDuplicate, err)
		}
		return kernel.Wrap(constant.CodeInvoicePersistenceFailed, fmt.Errorf("save invoice: %w", err))
	}
	return nil
}

func (r *PostgresInvoiceRepository) Update(ctx context.Context, inv *entity.Invoice) error {
	execer := execerFromContext(ctx, r.db)

	query := `UPDATE invoices SET
		santri_id=$1, user_id=$2, billing_scheme_id=$3, fee_component_id=$4,
		periode=$5, tahun_ajaran=$6, amount=$7, discount_amount=$8, paid_amount=$9,
		status=$10, due_date=$11, issued_at=$12, notes=$13,
		updated_at=$14, deleted_at=$15
		WHERE id=$16 AND deleted_at IS NULL`

	res, err := execer.ExecContext(ctx, query,
		inv.SantriID, inv.UserID, nullStr(inv.BillingSchemeID), inv.FeeComponentID,
		inv.Periode, inv.TahunAjaran, inv.Amount, inv.DiscountAmount, inv.PaidAmount,
		string(inv.Status), inv.DueDate, nullTimeVal(inv.IssuedAt), nullStr(inv.Notes),
		inv.UpdatedAt, nullTimeVal(inv.DeletedAt),
		inv.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeInvoiceDuplicate, err)
		}
		return kernel.Wrap(constant.CodeInvoicePersistenceFailed, fmt.Errorf("update invoice: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeInvoiceNotFound)
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
	if q.Periode != nil && *q.Periode != "" {
		where += fmt.Sprintf(` AND periode=$%d`, argIdx)
		args = append(args, *q.Periode)
		argIdx++
	}
	if q.TahunAjaran != nil && *q.TahunAjaran != "" {
		where += fmt.Sprintf(` AND tahun_ajaran=$%d`, argIdx)
		args = append(args, *q.TahunAjaran)
		argIdx++
	}

	var total int64
	countRow := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM invoices `+where, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, kernel.Wrap(constant.CodeInvoiceQueryFailed, fmt.Errorf("count invoices: %w", err))
	}

	limit := q.Limit
	offset := (q.Page - 1) * q.Limit
	listArgs := append(append([]interface{}{}, args...), limit, offset)

	query := fmt.Sprintf(`SELECT %s FROM invoices %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		invoiceColumns, where, argIdx, argIdx+1)

	rows, err := execer.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeInvoiceQueryFailed, fmt.Errorf("list invoices: %w", err))
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
		return nil, kernel.Wrap(constant.CodeInvoiceQueryFailed, fmt.Errorf("iterate invoice rows: %w", err))
	}

	return &repository.InvoiceListResult{Items: items, Total: total}, nil
}

func (r *PostgresInvoiceRepository) FindBySantriComponentPeriode(ctx context.Context, santriID, feeComponentID, periode string) (*entity.Invoice, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx,
		`SELECT `+invoiceColumns+` FROM invoices WHERE santri_id=$1 AND fee_component_id=$2 AND periode=$3 AND deleted_at IS NULL`,
		santriID, feeComponentID, periode,
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
		return nil, kernel.Wrap(constant.CodeInvoiceQueryFailed, fmt.Errorf("find outstanding by santri id: %w", err))
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
		return nil, kernel.Wrap(constant.CodeInvoiceQueryFailed, fmt.Errorf("iterate outstanding invoices: %w", err))
	}
	return items, nil
}

func (r *PostgresInvoiceRepository) HasPaidComponent(ctx context.Context, santriID, componentCode, periode string) (bool, error) {
	execer := execerFromContext(ctx, r.db)

	var exists bool
	err := execer.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM invoices i JOIN fee_components fc ON fc.id=i.fee_component_id WHERE i.santri_id=$1 AND fc.code=$2 AND i.periode=$3 AND i.status IN ('paid','partial') AND i.deleted_at IS NULL)`,
		santriID, componentCode, periode,
	).Scan(&exists)
	if err != nil {
		return false, kernel.Wrap(constant.CodeInvoiceQueryFailed, fmt.Errorf("has paid component: %w", err))
	}
	return exists, nil
}

func (r *PostgresInvoiceRepository) scan(sc scanner) (*entity.Invoice, error) {
	var (
		id, invoiceNumber, santriID, userID, feeComponentID, periode, tahunAjaran string
		billingSchemeID, notes                                                     sql.NullString
		amount, discountAmount, paidAmount                                         float64
		status                                                                     string
		dueDate                                                                    time.Time
		issuedAt                                                                   sql.NullTime
		createdBy                                                                  string
		createdAt, updatedAt                                                       time.Time
		deletedAt                                                                  sql.NullTime
	)

	err := sc.Scan(
		&id, &invoiceNumber, &santriID, &userID, &billingSchemeID, &feeComponentID,
		&periode, &tahunAjaran, &amount, &discountAmount, &paidAmount, &status,
		&dueDate, &issuedAt, &notes, &createdBy, &createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeInvoiceNotFound)
		}
		return nil, kernel.Wrap(constant.CodeInvoiceQueryFailed, fmt.Errorf("scan invoice: %w", err))
	}

	return &entity.Invoice{
		ID:              id,
		InvoiceNumber:   invoiceNumber,
		SantriID:        santriID,
		UserID:          userID,
		BillingSchemeID: strFromNull(billingSchemeID),
		FeeComponentID:  feeComponentID,
		Periode:         periode,
		TahunAjaran:     tahunAjaran,
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
