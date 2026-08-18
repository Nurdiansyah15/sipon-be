package database

import (
	"context"
	"database/sql"
	"fmt"
)

type txKey struct{}

// Execer adalah subset *sql.DB / *sql.Tx yang cukup untuk query repository.
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// ExecerFromContext mengembalikan *sql.Tx bila context membawa transaksi yang
// dibuka oleh Transactor.WithTx, selain itu mengembalikan *sql.DB. Semua
// repository yang ingin ikut dalam transaksi bisnis memakai helper ini.
func ExecerFromContext(ctx context.Context, db *sql.DB) Execer {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok && tx != nil {
		return tx
	}
	return db
}

// Transactor membuka satu DB transaction dan membaginya ke fn melalui context.
type Transactor struct {
	db *sql.DB
}

func NewTransactor(db *sql.DB) *Transactor {
	return &Transactor{db: db}
}

func (t *Transactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback tx: %w (original: %v)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
