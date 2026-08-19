package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/messaging/domain/event_outbox/entity"
)

type Repository interface {
	// Save menulis entry outbox. Metode ini memakai execer dari context, sehingga
	// dapat dipanggil di dalam DB transaction yang sama dengan perubahan bisnis.
	Save(ctx context.Context, entry *entity.OutboxEntry) error

	// ClaimDue mengklaim batch entry PENDING/FAILED yang jatuh tempo dengan
	// FOR UPDATE SKIP LOCKED, lalu menandainya PUBLISHING + lease dalam satu
	// transaksi. Entry yang dikembalikan sudah berstatus PUBLISHING.
	ClaimDue(ctx context.Context, now time.Time, limit int) ([]*entity.OutboxEntry, error)

	MarkPublished(ctx context.Context, id uuid.UUID, publishedAt time.Time) error
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string, nextAttemptAt time.Time) error

	// RecoverStuckPublishing mengembalikan entry PUBLISHING yang lease-nya sudah
	// lewat menjadi PENDING untuk di-claim ulang.
	RecoverStuckPublishing(ctx context.Context, leaseBefore time.Time) (int64, error)

	FindByID(ctx context.Context, id uuid.UUID) (*entity.OutboxEntry, error)
}
