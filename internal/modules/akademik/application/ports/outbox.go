package ports

import (
	"context"
	"encoding/json"
)

// OutboxWriter menulis event ke event_outbox dalam transaksi yang sama dengan
// perubahan bisnis. Diimplementasikan oleh modul messaging (infrastructure/persistence).
type OutboxWriter interface {
	Save(ctx context.Context, routingKey string, payload json.RawMessage) error
}
