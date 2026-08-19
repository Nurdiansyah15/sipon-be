package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"sipon-be/internal/modules/scheduler/application/ports"
	"sipon-be/internal/modules/scheduler/domain/scheduled_job/constant"
	"sipon-be/internal/modules/scheduler/domain/scheduled_job/entity"
	"sipon-be/internal/modules/scheduler/domain/scheduled_job/repository"
	"sipon-be/internal/shared/database"
)

// SchedulerDispatcher mengambil scheduled job yang jatuh tempo dan menuliskan
// event ke event_outbox dalam satu DB transaction dengan state scheduled job
// (pola outbox). Eksekusi handler diserahkan ke Message Consumer via RabbitMQ.
// Tidak ada ticker/polling lain di proses API — hanya proses worker yang
// menjalankan dispatcher ini.
type SchedulerDispatcher struct {
	repo       repository.Repository
	outboxRepo ports.OutboxWriter
	transactor *database.Transactor
	parser     cron.Parser
	tick       time.Duration
	lease      time.Duration
	batchSize  int
	logger     *slog.Logger
}

func NewDispatcher(
	repo repository.Repository,
	tick time.Duration,
	lease time.Duration,
	logger *slog.Logger,
) *SchedulerDispatcher {
	if lease <= 0 {
		lease = 5 * time.Minute
	}
	return &SchedulerDispatcher{
		repo:      repo,
		parser:    cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
		tick:      tick,
		lease:     lease,
		batchSize: 50,
		logger:    logger,
	}
}

// WithOutbox memasang outbox writer dan transactor. Dispatcher selalu memakai
// pola outbox — tanpa pasangan ini, dispatchJob akan error.
func (d *SchedulerDispatcher) WithOutbox(outboxRepo ports.OutboxWriter, transactor *database.Transactor) *SchedulerDispatcher {
	d.outboxRepo = outboxRepo
	d.transactor = transactor
	return d
}

func (d *SchedulerDispatcher) Run(ctx context.Context) {
	d.logger.Info("scheduler dispatcher started (outbox mode)",
		slog.Duration("tick", d.tick),
		slog.Duration("lease", d.lease),
	)
	ticker := time.NewTicker(d.tick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			d.logger.Info("scheduler dispatcher stopped")
			return
		case <-ticker.C:
			d.processBatch(ctx)
		}
	}
}

func (d *SchedulerDispatcher) processBatch(ctx context.Context) {
	now := time.Now()
	jobs, err := d.repo.FindDueAndClaim(ctx, now, d.batchSize, now.Add(d.lease))
	if err != nil {
		d.logger.Error("scheduler: gagal klaim due jobs", slog.Any("error", err))
		return
	}

	for _, j := range jobs {
		if err := d.dispatchJob(ctx, j, now); err != nil {
			d.logger.Error("scheduler: gagal dispatch job",
				slog.String("job_id", j.ID.String()),
				slog.String("type", j.Type),
				slog.Any("error", err),
			)
		}
	}
}

// dispatchJob menulis event ke event_outbox dan memajukan state scheduled job
// dalam satu DB transaction. Handler TIDAK dipanggil di sini.
func (d *SchedulerDispatcher) dispatchJob(ctx context.Context, j *entity.ScheduledJob, now time.Time) error {
	if d.transactor == nil || d.outboxRepo == nil {
		return fmt.Errorf("scheduler: outbox mode butuh transactor dan outbox repository")
	}
	return d.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := d.outboxRepo.Save(txCtx, j.Type, j.Payload); err != nil {
			return err
		}
		if err := d.advanceJob(j, now); err != nil {
			return err
		}
		j.LeaseUntil = nil
		return d.repo.Update(txCtx, j)
	})
}

// advanceJob memajukan state job setelah eksekusi berhasil.
func (d *SchedulerDispatcher) advanceJob(j *entity.ScheduledJob, now time.Time) error {
	if j.ScheduleType == constant.ScheduleTypeRecurring {
		if j.CronExpr == nil {
			return fmt.Errorf("recurring job tanpa cron expression")
		}
		sched, err := d.parser.Parse(*j.CronExpr)
		if err != nil {
			return err
		}
		j.MarkFired(sched.Next(now))
		return nil
	}
	j.MarkCompleted()
	return nil
}
