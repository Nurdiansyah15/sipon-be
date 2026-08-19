package scheduler

import (
	"context"
	"encoding/json"
	"time"
)

// Contract adalah permukaan outward-facing modul scheduler yang dikonsumsi oleh
// modul bisnis (mis. akademik) dan proses worker. Modul bisnis memakai operasi
// penjadwalan; proses worker memakai Run untuk menjalankan dispatcher. Dibuat
// konsisten dengan pola Contract pada modul bisnis (internal/modules/*).
type Contract interface {
	// Run menjalankan dispatcher sampai context dibatalkan (dipakai worker).
	Run(ctx context.Context)

	// ScheduleRecurring mendaftarkan job recurring (cron). referenceID dipakai
	// untuk menemukan/memperbarui job milik sebuah resource (mis. sesi).
	ScheduleRecurring(ctx context.Context, in ScheduleRecurringInput) error

	// ScheduleOneOff mendaftarkan job sekali jalan pada waktu runAt.
	ScheduleOneOff(ctx context.Context, in ScheduleOneOffInput) error

	// PauseByTypeAndReferenceID menghentikan job recurring milik referenceID bila
	// masih aktif (dipakai untuk menghentikan job setelah resource selesai).
	PauseByTypeAndReferenceID(ctx context.Context, jobType, referenceID string) error
}

// ScheduleRecurringInput parameter job recurring.
type ScheduleRecurringInput struct {
	JobType     string
	Payload     json.RawMessage
	CronExpr    string
	ReferenceID string
}

// ScheduleOneOffInput parameter job sekali jalan.
type ScheduleOneOffInput struct {
	JobType     string
	Payload     json.RawMessage
	RunAt       time.Time
	ReferenceID string
}

var _ Contract = (*Module)(nil)
