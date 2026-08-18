package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/shared/messaging/domain/message_job/entity"
)

type Repository interface {
	// Save menulis durable inbox dengan ON CONFLICT (id) DO NOTHING. Metode ini
	// memakai execer dari context, sehingga dapat dipanggil dalam DB transaction.
	Save(ctx context.Context, job *entity.MessageJob) error

	FindByID(ctx context.Context, id uuid.UUID) (*entity.MessageJob, error)

	// ClaimPending mengklaim batch PENDING yang jatuh tempo dengan
	// FOR UPDATE SKIP LOCKED, lalu menandainya RUNNING + lease dalam satu
	// transaksi. Entry yang dikembalikan sudah berstatus RUNNING dan
	// attempt_count sudah bertambah.
	ClaimPending(ctx context.Context, now time.Time, limit int, leaseUntil time.Time) ([]*entity.MessageJob, error)

	// ClaimByID mengklaim satu job (status PENDING/RETRY_WAIT) menjadi RUNNING
	// dengan attempt_count +1 dan lease, memakai RETURNING. Kedua nilai kembalian:
	// job yang sudah di-update, dan ok=false bila job tidak bisa diklaim (terminal
	// atau sedang diproses worker lain).
	ClaimByID(ctx context.Context, id uuid.UUID, now time.Time, leaseUntil time.Time) (*entity.MessageJob, bool, error)

	// Update menulis seluruh kolom job (dipakai untuk transisi state consumer).
	Update(ctx context.Context, job *entity.MessageJob) error

	// RecoverStuckRunning mengembalikan job RUNNING yang lease-nya sudah lewat
	// menjadi PENDING untuk di-claim ulang.
	RecoverStuckRunning(ctx context.Context, leaseBefore time.Time) (int64, error)
}
