package scheduler

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"sipon-be/internal/modules/scheduler/application"
	"sipon-be/internal/modules/scheduler/domain/scheduled_job/constant"
	"sipon-be/internal/modules/scheduler/domain/scheduled_job/entity"
	"sipon-be/internal/modules/scheduler/domain/scheduled_job/repository"
	"sipon-be/internal/modules/scheduler/infrastructure/persistence"
	"sipon-be/internal/shared/timeutil"
)

// Module adalah facade komposisi scheduler: repository + dispatcher + operasi
// penjadwalan yang diekspos lewat Contract. Dibuat konsisten dengan pola modul
// (module.go + contract.go) pada internal/modules/*.
type Module struct {
	repo       repository.Repository
	dispatcher *application.SchedulerDispatcher
	parser     cron.Parser
	loc        *time.Location
}

// NewModule membuat facade scheduler dengan repository PostgreSQL dan dispatcher
// yang berjalan pada interval tick dan lease yang diberikan.
func NewModule(db *sql.DB, tick, lease time.Duration, logger *slog.Logger) *Module {
	repo := persistence.NewPostgresScheduledJobRepository(db)
	loc := timeutil.Loc()
	return &Module{
		repo:       repo,
		dispatcher: application.NewDispatcher(repo, tick, lease, logger),
		parser:     cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
		loc:        loc,
	}
}

// Dispatcher mengembalikan dispatcher internal untuk konfigurasi lanjutan
// (WithDirectMode / WithOutboxMode).
func (m *Module) Dispatcher() *application.SchedulerDispatcher {
	return m.dispatcher
}

// Run menjalankan dispatcher sampai context dibatalkan.
func (m *Module) Run(ctx context.Context) {
	m.dispatcher.Run(ctx)
}

// ScheduleRecurring mendaftarkan job recurring (cron) milik referenceID.
func (m *Module) ScheduleRecurring(ctx context.Context, in ScheduleRecurringInput) error {
	job, err := entity.NewRecurringJob(in.JobType, in.Payload, in.CronExpr, m.parser, m.loc)
	if err != nil {
		return err
	}
	job.ReferenceID = &in.ReferenceID
	return m.repo.Save(ctx, job)
}

// ScheduleOneOff mendaftarkan job sekali jalan pada waktu runAt.
func (m *Module) ScheduleOneOff(ctx context.Context, in ScheduleOneOffInput) error {
	job := entity.NewOneOffJob(in.JobType, in.Payload, in.RunAt)
	job.ReferenceID = &in.ReferenceID
	return m.repo.Save(ctx, job)
}

// PauseByTypeAndReferenceID menghentikan job recurring milik referenceID bila
// masih aktif.
func (m *Module) PauseByTypeAndReferenceID(ctx context.Context, jobType, referenceID string) error {
	job, err := m.repo.FindByTypeAndReferenceID(ctx, jobType, referenceID)
	if err != nil || job == nil {
		return err
	}
	if job.Status == constant.StatusActive {
		job.Pause()
		return m.repo.Update(ctx, job)
	}
	return nil
}
