package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/shared/database"
	"sipon-be/internal/shared/scheduler/domain/scheduled_job/constant"
	"sipon-be/internal/shared/scheduler/domain/scheduled_job/entity"
)

type PostgresScheduledJobRepository struct {
	db *sql.DB
}

func NewPostgresScheduledJobRepository(db *sql.DB) *PostgresScheduledJobRepository {
	return &PostgresScheduledJobRepository{db: db}
}

func (r *PostgresScheduledJobRepository) Save(ctx context.Context, job *entity.ScheduledJob) error {
	execer := database.ExecerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx, `
		INSERT INTO scheduled_jobs
			(id, type, payload, schedule_type, cron_expr, run_at, next_run_at, status, retry_count, max_retry, reference_id, lease_until, created_at, updated_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		job.ID, job.Type, job.Payload, job.ScheduleType, job.CronExpr, job.RunAt,
		job.NextRunAt, job.Status, job.RetryCount, job.MaxRetry, job.ReferenceID,
		timeFromPtr(job.LeaseUntil), job.CreatedAt, job.UpdatedAt,
	)
	return err
}

// FindDueAndClaim mengklaim job yang jatuh tempo: ACTIVE dengan next_run_at <= now,
// ATAU job PROCESSING yang lease-nya sudah expired (recovery crash). Klaim
// dilakukan dengan FOR UPDATE SKIP LOCKED, lalu status/lease di-update dalam
// transaksi yang sama.
func (r *PostgresScheduledJobRepository) FindDueAndClaim(ctx context.Context, now time.Time, limit int, leaseUntil time.Time) ([]*entity.ScheduledJob, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT id, type, payload, schedule_type, cron_expr, run_at, next_run_at,
		       last_run_at, status, retry_count, max_retry, last_error, reference_id,
		       lease_until, created_at, updated_at
		FROM scheduled_jobs
		WHERE (status = 'ACTIVE' AND next_run_at <= $1)
		   OR (status = 'PROCESSING' AND lease_until <= $1)
		ORDER BY next_run_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED`, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*entity.ScheduledJob
	var ids []uuid.UUID
	for rows.Next() {
		j, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
		ids = append(ids, j.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(ids) > 0 {
		placeholders := make([]string, len(ids))
		args := make([]any, len(ids))
		for i, id := range ids {
			placeholders[i] = fmt.Sprintf("$%d", i+2)
			args[i] = id
		}
		query := `UPDATE scheduled_jobs
			SET status = 'PROCESSING', lease_until = $2, updated_at = $2
			WHERE id IN (` + strings.Join(placeholders, ",") + `)`
		if _, err := tx.ExecContext(ctx, query, append([]any{now, leaseUntil}, args...)...); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	for _, j := range jobs {
		j.Status = constant.StatusProcessing
		lease := leaseUntil
		j.LeaseUntil = &lease
		j.UpdatedAt = now
	}
	return jobs, nil
}

func (r *PostgresScheduledJobRepository) Update(ctx context.Context, job *entity.ScheduledJob) error {
	execer := database.ExecerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx, `
		UPDATE scheduled_jobs SET
			payload = $2, schedule_type = $3, cron_expr = $4, run_at = $5,
			next_run_at = $6, last_run_at = $7, status = $8,
			retry_count = $9, max_retry = $10, last_error = $11,
			reference_id = $12, lease_until = $13, updated_at = $14
		WHERE id = $1`,
		job.ID, job.Payload, job.ScheduleType, job.CronExpr, job.RunAt,
		job.NextRunAt, job.LastRunAt, job.Status,
		job.RetryCount, job.MaxRetry, job.LastError,
		job.ReferenceID, timeFromPtr(job.LeaseUntil), job.UpdatedAt,
	)
	return err
}

func (r *PostgresScheduledJobRepository) FindByTypeAndReferenceID(ctx context.Context, jobType string, referenceID string) (*entity.ScheduledJob, error) {
	execer := database.ExecerFromContext(ctx, r.db)
	rows, err := execer.QueryContext(ctx, `
		SELECT id, type, payload, schedule_type, cron_expr, run_at, next_run_at,
		       last_run_at, status, retry_count, max_retry, last_error, reference_id,
		       lease_until, created_at, updated_at
		FROM scheduled_jobs
		WHERE type = $1 AND reference_id = $2
		LIMIT 1`, jobType, referenceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}
	return scanJob(rows)
}

type scanner interface {
	Scan(dest ...any) error
}

func timeFromPtr(t *time.Time) any {
	if t == nil {
		return nil
	}
	return *t
}

func scanJob(s scanner) (*entity.ScheduledJob, error) {
	var j entity.ScheduledJob
	var cronExpr, lastError, referenceID sql.NullString
	var runAt, lastRunAt, leaseUntil sql.NullTime
	var payload []byte

	err := s.Scan(
		&j.ID, &j.Type, &payload, &j.ScheduleType, &cronExpr, &runAt,
		&j.NextRunAt, &lastRunAt, &j.Status, &j.RetryCount, &j.MaxRetry,
		&lastError, &referenceID, &leaseUntil, &j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	j.Payload = json.RawMessage(payload)
	if cronExpr.Valid {
		j.CronExpr = &cronExpr.String
	}
	if runAt.Valid {
		j.RunAt = &runAt.Time
	}
	if lastRunAt.Valid {
		j.LastRunAt = &lastRunAt.Time
	}
	if lastError.Valid {
		j.LastError = &lastError.String
	}
	if referenceID.Valid {
		j.ReferenceID = &referenceID.String
	}
	if leaseUntil.Valid {
		j.LeaseUntil = &leaseUntil.Time
	}
	return &j, nil
}
