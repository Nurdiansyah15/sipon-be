package ports

import (
	"context"
	"encoding/json"
)

// OutboxWriter menulis event ke event_outbox dalam transaksi yang sama dengan
// perubahan bisnis (memakai execer dari context). Diimplementasikan oleh modul
// messaging (infrastructure/persistence). Port ini membalik arah dependensi agar
// scheduler tidak perlu tahu implementasi outbox messaging secara langsung.
type OutboxWriter interface {
	// Save menulis satu entry outbox. Dapat dipanggil di dalam DB transaction.
	Save(ctx context.Context, routingKey string, payload json.RawMessage) error
}
