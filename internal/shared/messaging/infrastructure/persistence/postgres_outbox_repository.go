package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/shared/messaging/domain/event_outbox/entity"
)

const outboxColumns = `
	id, routing_key, payload, version, correlation_id, causation_id, status,
	attempt_count, next_attempt_at, locked_at, published_at, last_error,
	created_at, updated_at
`

const outboxInsertColumns = `
	id, routing_key, payload, version, correlation_id, causation_id, status,
	attempt_count, next_attempt_at, created_at, updated_at
`

type PostgresOutboxRepository struct {
	db *sql.DB
}

func NewPostgresOutboxRepository(db *sql.DB) *PostgresOutboxRepository {
	return &PostgresOutboxRepository{db: db}
}

func (r *PostgresOutboxRepository) Save(ctx context.Context, entry *entity.OutboxEntry) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx, `
		INSERT INTO event_outbox (`+outboxInsertColumns+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		entry.ID, entry.RoutingKey, jsonBytes(entry.Payload), entry.Version,
		entry.CorrelationID, nullUUID(entry.CausationID), string(entry.Status),
		entry.AttemptCount, entry.NextAttemptAt, entry.CreatedAt, entry.UpdatedAt,
	)
	return err
}

func (r *PostgresOutboxRepository) ClaimDue(ctx context.Context, now time.Time, limit int) ([]*entity.OutboxEntry, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT `+outboxColumns+`
		FROM event_outbox
		WHERE status IN ('PENDING', 'FAILED') AND next_attempt_at <= $1
		ORDER BY next_attempt_at ASC, created_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED`, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*entity.OutboxEntry
	var ids []uuid.UUID
	for rows.Next() {
		e, err := scanOutboxEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
		ids = append(ids, e.ID)
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
		query := `UPDATE event_outbox
			SET status = 'PUBLISHING', locked_at = $1, updated_at = $1
			WHERE id IN (` + strings.Join(placeholders, ",") + `)`
		if _, err := tx.ExecContext(ctx, query, append([]any{now}, args...)...); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	for _, e := range entries {
		e.MarkPublishing(now)
	}
	return entries, nil
}

func (r *PostgresOutboxRepository) MarkPublished(ctx context.Context, id uuid.UUID, publishedAt time.Time) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx, `
		UPDATE event_outbox
		SET status = 'PUBLISHED', published_at = $2, locked_at = NULL, updated_at = $2
		WHERE id = $1 AND status = 'PUBLISHING'`, id, publishedAt)
	return err
}

func (r *PostgresOutboxRepository) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string, nextAttemptAt time.Time) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx, `
		UPDATE event_outbox
		SET status = 'FAILED', attempt_count = attempt_count + 1, last_error = $2,
		    next_attempt_at = $3, locked_at = NULL, updated_at = $4
		WHERE id = $1`, id, errMsg, nextAttemptAt, time.Now().UTC())
	return err
}

func (r *PostgresOutboxRepository) RecoverStuckPublishing(ctx context.Context, leaseBefore time.Time) (int64, error) {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx, `
		UPDATE event_outbox
		SET status = 'PENDING', locked_at = NULL, updated_at = $2
		WHERE status = 'PUBLISHING' AND locked_at < $1`, leaseBefore, time.Now().UTC())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *PostgresOutboxRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.OutboxEntry, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+outboxColumns+` FROM event_outbox WHERE id = $1`, id)
	return scanOutboxEntry(row)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanOutboxEntry(sc scanner) (*entity.OutboxEntry, error) {
	var (
		e                  entity.OutboxEntry
		routingKey, status string
		payload            []byte
		causationID        sql.NullString
		lockedAt           sql.NullTime
		publishedAt        sql.NullTime
		lastError          sql.NullString
	)
	err := sc.Scan(
		&e.ID, &routingKey, &payload, &e.Version, &e.CorrelationID, &causationID, &status,
		&e.AttemptCount, &e.NextAttemptAt, &lockedAt, &publishedAt, &lastError,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	e.RoutingKey = routingKey
	e.Payload = json.RawMessage(payload)
	e.Status = entity.Status(status)
	e.CausationID, err = uuidFromNull(causationID)
	if err != nil {
		return nil, err
	}
	e.LockedAt = timeFromNull(lockedAt)
	e.PublishedAt = timeFromNull(publishedAt)
	e.LastError = strFromNull(lastError)
	return &e, nil
}
