package persistence

import (
	"context"
	"database/sql"
	"fmt"
)

type PostgresTransactor struct {
	db *sql.DB
}

func NewPostgresTransactor(db *sql.DB) *PostgresTransactor {
	return &PostgresTransactor{db: db}
}

type txKey struct{}

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
