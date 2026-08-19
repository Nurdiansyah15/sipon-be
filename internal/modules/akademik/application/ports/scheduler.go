package ports

import (
	"context"
	"time"
)

// ScheduleRecurringInput parameter job recurring yang didaftarkan ke scheduler.
type ScheduleRecurringInput struct {
	JobType     string
	Payload     []byte
	CronExpr    string
	ReferenceID string
}

// ScheduleOneOffInput parameter job sekali jalan yang didaftarkan ke scheduler.
type ScheduleOneOffInput struct {
	JobType     string
	Payload     []byte
	RunAt       time.Time
	ReferenceID string
}

// Scheduler adalah boundary untuk penjadwalan job background (recurring/one-off)
// dan penghentian job milik sebuah resource. Dipenuhi oleh adapter di lapisan
// infrastructure (schedulergateway) yang membungkus modul scheduler.
type Scheduler interface {
	// ScheduleRecurring mendaftarkan job recurring (cron) milik referenceID.
	ScheduleRecurring(ctx context.Context, in ScheduleRecurringInput) error
	// ScheduleOneOff mendaftarkan job sekali jalan pada waktu RunAt.
	ScheduleOneOff(ctx context.Context, in ScheduleOneOffInput) error
	// PauseByTypeAndReferenceID menghentikan job recurring milik referenceID
	// bila masih aktif.
	PauseByTypeAndReferenceID(ctx context.Context, jobType, referenceID string) error
}
