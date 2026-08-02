package persistence

import (
	"context"
	"database/sql"
	"fmt"
)

type txKey struct{}

type PostgresTransactor struct {
	db *sql.DB
}

func NewPostgresTransactor(db *sql.DB) *PostgresTransactor {
	return &PostgresTransactor{db: db}
}

func (t *PostgresTransactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
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

// dbExecer is satisfied by both *sql.DB and *sql.Tx — every repo in this
// package goes through execerFromContext instead of holding r.db directly,
// so that a Transactor.WithTx wrapping multiple repo calls is genuinely
// atomic. This is the deliberate fix for the bug where identity's own
// PostgresTransactor stuffs a *sql.Tx into ctx but no identity repo ever
// reads it back out — see docs/architecture notes / the kesantrian port plan.
type dbExecer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

func execerFromContext(ctx context.Context, db *sql.DB) dbExecer {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok && tx != nil {
		return tx
	}
	return db
}
