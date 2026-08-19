package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/messaging/domain/message_job/entity"
)

const messageJobColumns = `
	id, routing_key, payload, version, correlation_id, status,
	attempt_count, max_attempts, next_attempt_at, running_at, succeeded_at,
	failed_at, locked_until, last_error, created_at, updated_at
`

type PostgresMessageJobRepository struct {
	db *sql.DB
}

func NewPostgresMessageJobRepository(db *sql.DB) *PostgresMessageJobRepository {
	return &PostgresMessageJobRepository{db: db}
}

func (r *PostgresMessageJobRepository) Save(ctx context.Context, job *entity.MessageJob) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx, `
		INSERT INTO message_jobs
			(id, routing_key, payload, version, correlation_id, status, attempt_count,
			 max_attempts, next_attempt_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO NOTHING`,
		job.ID, job.RoutingKey, jsonBytes(job.Payload), job.Version, job.CorrelationID,
		string(job.Status), job.AttemptCount, job.MaxAttempts, job.NextAttemptAt,
		job.CreatedAt, job.UpdatedAt,
	)
	return err
}

func (r *PostgresMessageJobRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.MessageJob, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+messageJobColumns+` FROM message_jobs WHERE id = $1`, id)
	return scanMessageJob(row)
}

func (r *PostgresMessageJobRepository) ClaimPending(ctx context.Context, now time.Time, limit int, leaseUntil time.Time) ([]*entity.MessageJob, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT `+messageJobColumns+`
		FROM message_jobs
		WHERE status = 'PENDING' AND next_attempt_at <= $1
		ORDER BY next_attempt_at ASC, created_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED`, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*entity.MessageJob
	var ids []uuid.UUID
	for rows.Next() {
		j, err := scanMessageJob(rows)
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
			placeholders[i] = fmt.Sprintf("$%d", i+3)
			args[i] = id
		}
		query := `UPDATE message_jobs
			SET status = 'RUNNING', attempt_count = attempt_count + 1,
			    running_at = $1, locked_until = $2, updated_at = $1
			WHERE id IN (` + strings.Join(placeholders, ",") + `)`
		if _, err := tx.ExecContext(ctx, query, append([]any{now, leaseUntil}, args...)...); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	for _, j := range jobs {
		j.StartRun(now, leaseUntil)
	}
	return jobs, nil
}

func (r *PostgresMessageJobRepository) ClaimByID(ctx context.Context, id uuid.UUID, now time.Time, leaseUntil time.Time) (*entity.MessageJob, bool, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `
		UPDATE message_jobs
		SET status = 'RUNNING', attempt_count = attempt_count + 1,
		    running_at = $2, locked_until = $3, updated_at = $2
		WHERE id = $1 AND status IN ('PENDING', 'RETRY_WAIT')
		RETURNING `+messageJobColumns,
		id, now, leaseUntil)

	job, err := scanMessageJob(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return job, true, nil
}

func (r *PostgresMessageJobRepository) Update(ctx context.Context, job *entity.MessageJob) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx, `
		UPDATE message_jobs SET
			routing_key=$2, payload=$3, version=$4, correlation_id=$5, status=$6,
			attempt_count=$7, max_attempts=$8, next_attempt_at=$9, running_at=$10,
			succeeded_at=$11, failed_at=$12, locked_until=$13, last_error=$14,
			updated_at=$15
		WHERE id=$1`,
		job.ID, job.RoutingKey, jsonBytes(job.Payload), job.Version, job.CorrelationID,
		string(job.Status), job.AttemptCount, job.MaxAttempts, job.NextAttemptAt,
		timeFromPtr(job.RunningAt), timeFromPtr(job.SucceededAt), timeFromPtr(job.FailedAt),
		timeFromPtr(job.LockedUntil), strFromPtr(job.LastError), job.UpdatedAt,
	)
	return err
}

func (r *PostgresMessageJobRepository) RecoverStuckRunning(ctx context.Context, leaseBefore time.Time) (int64, error) {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx, `
		UPDATE message_jobs
		SET status = 'PENDING', locked_until = NULL, running_at = NULL, updated_at = $2
		WHERE status = 'RUNNING' AND locked_until < $1`, leaseBefore, time.Now().UTC())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func scanMessageJob(sc scanner) (*entity.MessageJob, error) {
	var (
		j                  entity.MessageJob
		routingKey, status string
		payload            []byte
		runningAt          sql.NullTime
		succeededAt        sql.NullTime
		failedAt           sql.NullTime
		lockedUntil        sql.NullTime
		lastError          sql.NullString
	)
	err := sc.Scan(
		&j.ID, &routingKey, &payload, &j.Version, &j.CorrelationID, &status,
		&j.AttemptCount, &j.MaxAttempts, &j.NextAttemptAt, &runningAt, &succeededAt,
		&failedAt, &lockedUntil, &lastError, &j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	j.RoutingKey = routingKey
	j.Payload = json.RawMessage(payload)
	j.Status = entity.Status(status)
	j.RunningAt = timeFromNull(runningAt)
	j.SucceededAt = timeFromNull(succeededAt)
	j.FailedAt = timeFromNull(failedAt)
	j.LockedUntil = timeFromNull(lockedUntil)
	j.LastError = strFromNull(lastError)
	return &j, nil
}
