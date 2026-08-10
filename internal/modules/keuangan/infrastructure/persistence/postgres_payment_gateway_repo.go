package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"sipon-be/internal/modules/keuangan/domain/paymentgateway/constant"
	"sipon-be/internal/modules/keuangan/domain/paymentgateway/entity"
	"sipon-be/internal/shared/kernel"
)

const paymentGatewayColumns = `
	id, transaction_id, invoice_id, payment_id, amount, status,
	payment_method, snap_token, redirect_url, raw_notification, metadata,
	expired_at, created_at, updated_at
`

type PostgresPaymentGatewayRepository struct {
	db *sql.DB
}

func NewPostgresPaymentGatewayRepository(db *sql.DB) *PostgresPaymentGatewayRepository {
	return &PostgresPaymentGatewayRepository{db: db}
}

func (r *PostgresPaymentGatewayRepository) Save(ctx context.Context, tx *entity.PaymentGatewayTransaction) error {
	execer := execerFromContext(ctx, r.db)

	query := `INSERT INTO payment_gateway_transactions (
		id, transaction_id, invoice_id, payment_id, amount, status,
		payment_method, snap_token, redirect_url, raw_notification, metadata,
		expired_at, created_at, updated_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb,$11::jsonb,$12,$13,$14)`

	_, err := execer.ExecContext(ctx, query,
		tx.ID, tx.TransactionID, tx.InvoiceID, nullStr(tx.PaymentID), tx.Amount, string(tx.Status),
		nullStr(tx.PaymentMethod), tx.SnapToken, tx.RedirectURL,
		stringOrNil(tx.RawNotification), stringOrNil(tx.Metadata),
		tx.ExpiredAt, tx.CreatedAt, tx.UpdatedAt,
	)
	if err != nil {
		return kernel.WrapMsg(constant.CodePaymentGatewayPersistenceFailed, "gagal menyimpan transaksi payment gateway", err)
	}
	return nil
}

func (r *PostgresPaymentGatewayRepository) Update(ctx context.Context, tx *entity.PaymentGatewayTransaction) error {
	execer := execerFromContext(ctx, r.db)

	query := `UPDATE payment_gateway_transactions SET
		payment_id=$1, status=$2, payment_method=$3,
		raw_notification=$4::jsonb, updated_at=$5
		WHERE id=$6`

	res, err := execer.ExecContext(ctx, query,
		nullStr(tx.PaymentID), string(tx.Status), nullStr(tx.PaymentMethod),
		stringOrNil(tx.RawNotification), tx.UpdatedAt, tx.ID,
	)
	if err != nil {
		return kernel.WrapMsg(constant.CodePaymentGatewayPersistenceFailed, "gagal memperbarui transaksi payment gateway", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.WrapMsg(constant.CodePaymentGatewayNotFound, "Transaksi payment gateway tidak ditemukan", nil)
	}
	return nil
}

func (r *PostgresPaymentGatewayRepository) FindByTransactionID(ctx context.Context, transactionID string) (*entity.PaymentGatewayTransaction, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx,
		`SELECT `+paymentGatewayColumns+` FROM payment_gateway_transactions WHERE transaction_id=$1`,
		transactionID,
	)
	return r.scan(row)
}

func (r *PostgresPaymentGatewayRepository) FindByInvoiceID(ctx context.Context, invoiceID string) ([]*entity.PaymentGatewayTransaction, error) {
	execer := execerFromContext(ctx, r.db)
	rows, err := execer.QueryContext(ctx,
		`SELECT `+paymentGatewayColumns+` FROM payment_gateway_transactions WHERE invoice_id=$1 ORDER BY created_at DESC`,
		invoiceID,
	)
	if err != nil {
		return nil, kernel.WrapMsg(constant.CodePaymentGatewayQueryFailed, "gagal mencari transaksi payment gateway", err)
	}
	defer rows.Close()

	items := make([]*entity.PaymentGatewayTransaction, 0)
	for rows.Next() {
		tx, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(constant.CodePaymentGatewayQueryFailed, "gagal membaca transaksi payment gateway", err)
	}
	return items, nil
}

func (r *PostgresPaymentGatewayRepository) FindActiveByInvoiceID(ctx context.Context, invoiceID string) (*entity.PaymentGatewayTransaction, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx,
		`SELECT `+paymentGatewayColumns+` FROM payment_gateway_transactions
		 WHERE invoice_id=$1 AND status IN ('pending', 'pending_challenge', 'capture', 'settlement')
		 ORDER BY created_at DESC LIMIT 1`,
		invoiceID,
	)
	return r.scan(row)
}

func (r *PostgresPaymentGatewayRepository) scan(sc scanner) (*entity.PaymentGatewayTransaction, error) {
	var (
		id, transactionID, invoiceID, status, snapToken, redirectURL string
		paymentID, paymentMethod                                     sql.NullString
		rawNotification, metadata                                    sql.NullString
		amount                                                       float64
		expiredAt, createdAt, updatedAt                              time.Time
	)

	err := sc.Scan(
		&id, &transactionID, &invoiceID, &paymentID, &amount, &status,
		&paymentMethod, &snapToken, &redirectURL, &rawNotification, &metadata,
		&expiredAt, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.WrapMsg(constant.CodePaymentGatewayNotFound, "Transaksi payment gateway tidak ditemukan", nil)
		}
		return nil, kernel.WrapMsg(constant.CodePaymentGatewayQueryFailed, "gagal membaca transaksi payment gateway", err)
	}

	return &entity.PaymentGatewayTransaction{
		ID:              id,
		TransactionID:   transactionID,
		InvoiceID:       invoiceID,
		PaymentID:       strFromNull(paymentID),
		Amount:          amount,
		Status:          constant.PaymentGatewayStatus(status),
		PaymentMethod:   strFromNull(paymentMethod),
		SnapToken:       snapToken,
		RedirectURL:     redirectURL,
		RawNotification: bytesFromNull(rawNotification),
		Metadata:        bytesFromNull(metadata),
		ExpiredAt:       expiredAt,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}, nil
}

func stringOrNil(b []byte) interface{} {
	if len(b) == 0 {
		return nil
	}
	return string(b)
}

func bytesFromNull(ns sql.NullString) []byte {
	if !ns.Valid {
		return nil
	}
	return []byte(ns.String)
}
