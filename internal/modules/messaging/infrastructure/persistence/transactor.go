package persistence

import (
	"context"
	"database/sql"

	"sipon-be/internal/shared/database"
)

// NewPostgresTransactor mengembalikan transactor bersama yang memakai key
// transaksi yang sama dengan repository scheduler/module lain, sehingga outbox
// dan business write dapat berada dalam satu DB transaction.
func NewPostgresTransactor(db *sql.DB) *database.Transactor {
	return database.NewTransactor(db)
}

func execerFromContext(ctx context.Context, db *sql.DB) database.Execer {
	return database.ExecerFromContext(ctx, db)
}
