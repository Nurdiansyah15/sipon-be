package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sipon-be/internal/modules/keuangan/domain/payment/constant"
	"sipon-be/internal/modules/keuangan/domain/payment/entity"
	"sipon-be/internal/modules/keuangan/domain/payment/repository"
	paymentVO "sipon-be/internal/modules/keuangan/domain/payment/valueobject"
	"sipon-be/internal/shared/kernel"
)

const paymentColumns = `
	id, payment_number, invoice_id, debit_account_id, amount, method,
	reference_number, payment_date, status, verified_by, verified_at,
	notes, proof_key, created_by, created_at, updated_at
`

type PostgresPaymentRepository struct {
	db *sql.DB
}

func NewPostgresPaymentRepository(db *sql.DB) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

func (r *PostgresPaymentRepository) NextPaymentNumber(ctx context.Context) (paymentVO.PaymentNumber, error) {
	execer := execerFromContext(ctx, r.db)
	now := time.Now()
	year := now.Year()

	seq, err := nextNumberSeq(ctx, execer, "payment", year)
	if err != nil {
		return paymentVO.PaymentNumber{}, kernel.WrapMsg(constant.CodePaymentPersistenceFailed, "gagal membuat nomor pembayaran", err)
	}

	return paymentVO.NewPaymentNumber(fmt.Sprintf("%d", year), fmt.Sprintf("%02d", int(now.Month())), seq), nil
}

func (r *PostgresPaymentRepository) Save(ctx context.Context, p *entity.Payment) error {
	execer := execerFromContext(ctx, r.db)

	query := `INSERT INTO payments (` + paymentColumns + `) VALUES (
		$1,$2,$3,$4,$5,$6,
		$7,$8,$9,$10,$11,
		$12,$13,$14,$15,$16
	)`

	_, err := execer.ExecContext(ctx, query,
		p.ID, p.PaymentNumber, p.InvoiceID, nullStr(p.DebitAccountID), p.Amount, string(p.Method),
		nullStr(p.ReferenceNumber), p.PaymentDate, string(p.Status), nullStr(p.VerifiedBy), nullTimeVal(p.VerifiedAt),
		nullStr(p.Notes), nullStr(p.ProofKey), p.CreatedBy, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return kernel.WrapMsg(constant.CodePaymentPersistenceFailed, "gagal menyimpan pembayaran", err)
	}
	return nil
}

func (r *PostgresPaymentRepository) Update(ctx context.Context, p *entity.Payment) error {
	execer := execerFromContext(ctx, r.db)

	query := `UPDATE payments SET
		debit_account_id=$1, method=$2, reference_number=$3,
		status=$4, verified_by=$5, verified_at=$6,
		notes=$7, proof_key=$8, updated_at=$9
		WHERE id=$10`

	res, err := execer.ExecContext(ctx, query,
		nullStr(p.DebitAccountID), string(p.Method), nullStr(p.ReferenceNumber),
		string(p.Status), nullStr(p.VerifiedBy), nullTimeVal(p.VerifiedAt),
		nullStr(p.Notes), nullStr(p.ProofKey), p.UpdatedAt,
		p.ID,
	)
	if err != nil {
		return kernel.WrapMsg(constant.CodePaymentPersistenceFailed, "gagal memperbarui pembayaran", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.WrapMsg(constant.CodePaymentNotFound, "Pembayaran tidak ditemukan", nil)
	}
	return nil
}

func (r *PostgresPaymentRepository) FindByID(ctx context.Context, id string) (*entity.Payment, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+paymentColumns+` FROM payments WHERE id=$1`, id)
	return r.scan(row)
}

func (r *PostgresPaymentRepository) FindByNumber(ctx context.Context, number string) (*entity.Payment, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+paymentColumns+` FROM payments WHERE payment_number=$1`, number)
	return r.scan(row)
}

func (r *PostgresPaymentRepository) List(ctx context.Context, q repository.PaymentListQuery) (*repository.PaymentListResult, error) {
	execer := execerFromContext(ctx, r.db)

	where := `WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if q.InvoiceID != nil && *q.InvoiceID != "" {
		where += fmt.Sprintf(` AND invoice_id=$%d`, argIdx)
		args = append(args, *q.InvoiceID)
		argIdx++
	}
	if q.Status != nil && *q.Status != "" {
		where += fmt.Sprintf(` AND status=$%d`, argIdx)
		args = append(args, *q.Status)
		argIdx++
	}
	if q.PeriodID != nil && *q.PeriodID != "" {
		where += fmt.Sprintf(` AND invoice_id IN (
			SELECT i.id FROM invoices i
			JOIN billing_periods bp ON bp.id = i.billing_period_id
			WHERE bp.accounting_period_id=$%d
		)`, argIdx)
		args = append(args, *q.PeriodID)
		argIdx++
	}

	var total int64
	countRow := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM payments `+where, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, kernel.WrapMsg(constant.CodePaymentQueryFailed, "gagal menghitung jumlah pembayaran", err)
	}

	limit := q.Limit
	offset := (q.Page - 1) * q.Limit
	listArgs := append(append([]interface{}{}, args...), limit, offset)

	query := fmt.Sprintf(`SELECT %s FROM payments %s ORDER BY payment_date DESC LIMIT $%d OFFSET $%d`,
		paymentColumns, where, argIdx, argIdx+1)

	rows, err := execer.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, kernel.WrapMsg(constant.CodePaymentQueryFailed, "gagal mendaftar pembayaran", err)
	}
	defer rows.Close()

	items := make([]*entity.Payment, 0)
	for rows.Next() {
		p, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(constant.CodePaymentQueryFailed, "gagal membaca data pembayaran", err)
	}

	return &repository.PaymentListResult{Items: items, Total: total}, nil
}

func (r *PostgresPaymentRepository) FindByInvoiceID(ctx context.Context, invoiceID string) ([]*entity.Payment, error) {
	execer := execerFromContext(ctx, r.db)

	rows, err := execer.QueryContext(ctx,
		`SELECT `+paymentColumns+` FROM payments WHERE invoice_id=$1 ORDER BY payment_date DESC`,
		invoiceID,
	)
	if err != nil {
		return nil, kernel.WrapMsg(constant.CodePaymentQueryFailed, "gagal mencari pembayaran berdasarkan invoice", err)
	}
	defer rows.Close()

	items := make([]*entity.Payment, 0)
	for rows.Next() {
		p, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(constant.CodePaymentQueryFailed, "gagal membaca data pembayaran invoice", err)
	}
	return items, nil
}

func (r *PostgresPaymentRepository) FindVerifiedByInvoiceID(ctx context.Context, invoiceID string) ([]*entity.Payment, error) {
	execer := execerFromContext(ctx, r.db)

	rows, err := execer.QueryContext(ctx,
		`SELECT `+paymentColumns+` FROM payments WHERE invoice_id=$1 AND status='verified' ORDER BY payment_date DESC`,
		invoiceID,
	)
	if err != nil {
		return nil, kernel.WrapMsg(constant.CodePaymentQueryFailed, "gagal mencari pembayaran terverifikasi invoice", err)
	}
	defer rows.Close()

	items := make([]*entity.Payment, 0)
	for rows.Next() {
		p, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(constant.CodePaymentQueryFailed, "gagal membaca data pembayaran terverifikasi", err)
	}
	return items, nil
}

func (r *PostgresPaymentRepository) scan(sc scanner) (*entity.Payment, error) {
	var (
		id, paymentNumber, invoiceID, method, status, createdBy                 string
		debitAccountID, referenceNumber, verifiedBy, notes, proofKey            sql.NullString
		amount                                                                  float64
		paymentDate                                                             time.Time
		verifiedAt                                                              sql.NullTime
		createdAt, updatedAt                                                    time.Time
	)

	err := sc.Scan(
		&id, &paymentNumber, &invoiceID, &debitAccountID, &amount, &method,
		&referenceNumber, &paymentDate, &status, &verifiedBy, &verifiedAt,
		&notes, &proofKey, &createdBy, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.WrapMsg(constant.CodePaymentNotFound, "Pembayaran tidak ditemukan", nil)
		}
		return nil, kernel.WrapMsg(constant.CodePaymentQueryFailed, "gagal membaca data pembayaran", err)
	}

	return &entity.Payment{
		ID:              id,
		PaymentNumber:   paymentNumber,
		InvoiceID:       invoiceID,
		DebitAccountID:  strFromNull(debitAccountID),
		Amount:          amount,
		Method:          constant.PaymentMethod(method),
		ReferenceNumber: strFromNull(referenceNumber),
		PaymentDate:     paymentDate,
		Status:          constant.PaymentStatus(status),
		VerifiedBy:      strFromNull(verifiedBy),
		VerifiedAt:      timeFromNull(verifiedAt),
		Notes:           strFromNull(notes),
		ProofKey:        strFromNull(proofKey),
		CreatedBy:       createdBy,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}, nil
}
