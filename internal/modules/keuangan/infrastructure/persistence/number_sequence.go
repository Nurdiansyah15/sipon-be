package persistence

import (
	"context"
	"fmt"
)

// nextNumberSeq atomically increments the counter for (docType, year) in the
// shared finance_number_sequences table and returns the new value. The
// INSERT ... ON CONFLICT DO UPDATE is a single statement, so Postgres
// serializes concurrent callers on the same row without an explicit
// SELECT ... FOR UPDATE — safe to call from any document type that needs a
// yearly-reset running number (invoice, payment, ...).
func nextNumberSeq(ctx context.Context, execer dbExecer, docType string, year int) (int, error) {
	var seq int
	err := execer.QueryRowContext(ctx,
		`INSERT INTO finance_number_sequences (doc_type, year, seq) VALUES ($1, $2, 1)
		 ON CONFLICT (doc_type, year) DO UPDATE SET seq = finance_number_sequences.seq + 1
		 RETURNING seq`,
		docType, year,
	).Scan(&seq)
	if err != nil {
		return 0, fmt.Errorf("next %s number: %w", docType, err)
	}
	return seq, nil
}
