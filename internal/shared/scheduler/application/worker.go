package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"sipon-be/internal/shared/scheduler/domain/scheduled_job/constant"
	"sipon-be/internal/shared/scheduler/domain/scheduled_job/entity"
	"sipon-be/internal/shared/scheduler/domain/scheduled_job/repository"
)

type Worker struct {
	repo      repository.Repository
	registry  *Registry
	parser    cron.Parser
	tick      time.Duration
	batchSize int
	logger    *slog.Logger
}

func NewWorker(repo repository.Repository, registry *Registry, tick time.Duration, logger *slog.Logger) *Worker {
	return &Worker{
		repo:      repo,
		registry:  registry,
		parser:    cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
		tick:      tick,
		batchSize: 50,
		logger:    logger,
	}
}

func (w *Worker) Run(ctx context.Context) {
	w.logger.Info("scheduler worker started", slog.Duration("tick", w.tick))
	ticker := time.NewTicker(w.tick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("scheduler worker stopped")
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *Worker) processBatch(ctx context.Context) {
	now := time.Now()
	jobs, err := w.repo.FindDueAndClaim(ctx, now, w.batchSize)
	if err != nil {
		w.logger.Error("scheduler: gagal klaim due jobs", slog.Any("error", err))
		return
	}

	for _, j := range jobs {
		w.executeJob(ctx, j, now)
	}
}

func (w *Worker) executeJob(ctx context.Context, j *entity.ScheduledJob, now time.Time) {
	start := time.Now()

	if err := w.registry.Dispatch(ctx, j.Type, j.Payload); err != nil {
		w.handleJobError(ctx, j, err, start)
		return
	}

	w.completeJob(ctx, j, now)

	w.logger.Info("scheduler: job selesai",
		slog.String("job_id", j.ID.String()),
		slog.String("type", j.Type),
		slog.Duration("durasi", time.Since(start)),
	)
}

func (w *Worker) completeJob(ctx context.Context, j *entity.ScheduledJob, now time.Time) {
	if j.ScheduleType == constant.ScheduleTypeRecurring {
		sched, err := w.parser.Parse(*j.CronExpr)
		if err != nil {
			w.logger.Error("scheduler: cron parse error pada recurring job",
				slog.String("job_id", j.ID.String()),
				slog.Any("error", err),
			)
			return
		}
		j.MarkFired(sched.Next(now))
	} else {
		j.MarkCompleted()
	}

	if err := w.repo.Update(ctx, j); err != nil {
		w.logger.Error("scheduler: gagal update state setelah eksekusi",
			slog.String("job_id", j.ID.String()),
			slog.Any("error", err),
		)
	}
}

func (w *Worker) handleJobError(ctx context.Context, j *entity.ScheduledJob, err error, start time.Time) {
	w.logger.Error("scheduler: job gagal",
		slog.String("job_id", j.ID.String()),
		slog.String("type", j.Type),
		slog.Duration("durasi", time.Since(start)),
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

	if updateErr := w.repo.Update(ctx, j); updateErr != nil {
		w.logger.Error("scheduler: gagal update error state",
			slog.String("job_id", j.ID.String()),
			slog.Any("error", updateErr),
		)
	}
}
