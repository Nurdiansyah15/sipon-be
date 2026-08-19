package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"sipon-be/internal/shared/database"
	"sipon-be/internal/modules/scheduler/application/ports"
	"sipon-be/internal/modules/scheduler/domain/scheduled_job/constant"
	"sipon-be/internal/modules/scheduler/domain/scheduled_job/entity"
	"sipon-be/internal/modules/scheduler/domain/scheduled_job/repository"
)

// Mode menentukan jalur eksekusi scheduler.
type Mode string

const (
	// ModeDirect mengeksekusi handler langsung lewat DispatchFunc. Dipakai selama
	// transisi (compatibility bridge) sebelum outbox/RabbitMQ aktif.
	ModeDirect Mode = "direct"
	// ModeOutbox hanya menulis event ke event_outbox dalam transaksi yang sama
	// dengan state scheduled job; eksekusi handler diserahkan ke consumer MQ.
	ModeOutbox Mode = "outbox"
)

// DispatchFunc dipakai pada ModeDirect. Error yang dikembalikan diklasifikasi:
// FatalError = gagal permanen (FAILED), selain itu dianggap retryable.
type DispatchFunc func(ctx context.Context, jobType string, payload json.RawMessage) error

// SchedulerDispatcher mengambil scheduled job yang jatuh tempo dan men-dispatchnya.
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
	mode       Mode
	dispatch   DispatchFunc
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
		mode:      ModeDirect,
		logger:    logger,
	}
}

// WithDirectMode memasang handler eksekusi langsung (mode default).
func (d *SchedulerDispatcher) WithDirectMode(dispatch DispatchFunc) *SchedulerDispatcher {
	d.mode = ModeDirect
	d.dispatch = dispatch
	return d
}

// WithOutboxMode mengalihkan dispatcher untuk hanya menulis event_outbox dalam
// satu DB transaction dengan state scheduled job.
func (d *SchedulerDispatcher) WithOutboxMode(repo ports.OutboxWriter, transactor *database.Transactor) *SchedulerDispatcher {
	d.mode = ModeOutbox
	d.outboxRepo = repo
	d.transactor = transactor
	return d
}

func (d *SchedulerDispatcher) Run(ctx context.Context) {
	d.logger.Info("scheduler dispatcher started",
		slog.String("mode", string(d.mode)),
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

func (d *SchedulerDispatcher) dispatchJob(ctx context.Context, j *entity.ScheduledJob, now time.Time) error {
	if d.mode == ModeOutbox {
		return d.dispatchOutbox(ctx, j, now)
	}
	return d.dispatchDirect(ctx, j, now)
}

// dispatchOutbox menulis event ke event_outbox dan memajukan state scheduled job
// dalam satu DB transaction. Handler TIDAK dipanggil di sini.
func (d *SchedulerDispatcher) dispatchOutbox(ctx context.Context, j *entity.ScheduledJob, now time.Time) error {
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

// dispatchDirect mengeksekusi handler langsung lewat DispatchFunc (compatibility
// bridge). Dipakai sampai outbox/RabbitMQ consumer aktif.
func (d *SchedulerDispatcher) dispatchDirect(ctx context.Context, j *entity.ScheduledJob, now time.Time) error {
	if d.dispatch == nil {
		return fmt.Errorf("scheduler: handler belum dipasang pada mode direct")
	}

	start := time.Now()

	if err := d.dispatch(ctx, j.Type, j.Payload); err != nil {
		d.handleJobError(j, err)
		if uerr := d.repo.Update(ctx, j); uerr != nil {
			return uerr
		}
		return nil
	}

	if err := d.advanceJob(j, now); err != nil {
		d.logger.Error("scheduler: cron parse error pada recurring job",
			slog.String("job_id", j.ID.String()),
			slog.Any("error", err),
		)
		return err
	}
	j.LeaseUntil = nil
	if err := d.repo.Update(ctx, j); err != nil {
		return err
	}

	d.logger.Info("scheduler: job selesai",
		slog.String("job_id", j.ID.String()),
		slog.String("type", j.Type),
		slog.Duration("durasi", time.Since(start)),
	)
	return nil
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

func (d *SchedulerDispatcher) handleJobError(j *entity.ScheduledJob, err error) {
	d.logger.Error("scheduler: job gagal",
		slog.String("job_id", j.ID.String()),
		slog.String("type", j.Type),
		slog.Any("error", err),
	)

	if IsFatal(err) {
		j.Status = constant.StatusFailed
		errMsg := err.Error()
		j.LastError = &errMsg
		j.UpdatedAt = time.Now()
	} else {
		j.MarkFailed(err.Error())
	}
	j.LeaseUntil = nil
}
