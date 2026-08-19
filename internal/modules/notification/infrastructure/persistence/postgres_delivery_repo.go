package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

notifconstant "sipon-be/internal/modules/notification/domain/notification/constant"
	deliveryEntity "sipon-be/internal/modules/notification/domain/delivery/entity"
	deliveryRepo "sipon-be/internal/modules/notification/domain/delivery/repository"
	"sipon-be/internal/shared/kernel"
)

type PostgresDeliveryAttemptRepository struct {
	db *sql.DB
}

func NewPostgresDeliveryAttemptRepository(db *sql.DB) *PostgresDeliveryAttemptRepository {
	return &PostgresDeliveryAttemptRepository{db: db}
}

func (r *PostgresDeliveryAttemptRepository) Save(ctx context.Context, da *deliveryEntity.DeliveryAttempt) error {
	_, err := execerFromContext(ctx, r.db).ExecContext(ctx,
		`INSERT INTO delivery_attempts (id, notification_id, user_id, channel, status, provider_code, retry_count, next_retry_at, read_at, attempted_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		da.ID, da.NotificationID, da.UserID, string(da.Channel), string(da.Status),
		nullStr(da.ProviderCode), da.RetryCount, nullTime(da.NextRetryAt), nullTime(da.ReadAt), da.AttemptedAt,
	)
	if err != nil {
		return kernel.Wrap(notifconstant.CodeDeliveryAttemptPersistenceFailed, fmt.Errorf("insert delivery attempt: %w", err))
	}
	return nil
}

func (r *PostgresDeliveryAttemptRepository) FindByID(ctx context.Context, id string) (*deliveryEntity.DeliveryAttempt, error) {
	var da deliveryEntity.DeliveryAttempt
	var channel, status string
	var providerCode sql.NullString
	var nextRetryAt, readAt sql.NullTime

	err := r.db.QueryRowContext(ctx,
		`SELECT id, notification_id, user_id, channel, status, provider_code, retry_count, next_retry_at, read_at, attempted_at
		 FROM delivery_attempts WHERE id = $1`, id,
	).Scan(&da.ID, &da.NotificationID, &da.UserID, &channel, &status, &providerCode,
		&da.RetryCount, &nextRetryAt, &readAt, &da.AttemptedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, kernel.New(notifconstant.CodeDeliveryAttemptNotFound)
		}
		return nil, kernel.Wrap(notifconstant.CodeDeliveryAttemptPersistenceFailed, err)
	}

	da.Channel = notifconstant.NotificationChannel(channel)
	da.Status = notifconstant.DeliveryStatus(status)
	da.ProviderCode = fromNullStr(providerCode)
	da.NextRetryAt = timeFromNull(nextRetryAt)
	da.ReadAt = timeFromNull(readAt)

	return &da, nil
}

func (r *PostgresDeliveryAttemptRepository) ListInApp(ctx context.Context, q deliveryRepo.ListInAppQuery) ([]deliveryRepo.InboxReadItem, deliveryRepo.Meta, error) {
	offset := (q.Page - 1) * q.Limit

	var total int
	countQuery := `SELECT COUNT(*) FROM delivery_attempts da
		JOIN notifications n ON n.id = da.notification_id
		WHERE da.user_id = $1 AND da.channel = 'in_app'`
	countArgs := []interface{}{q.UserID}
	if q.UnreadOnly {
		countQuery += ` AND da.read_at IS NULL`
	}
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, deliveryRepo.Meta{}, kernel.Wrap(notifconstant.CodeDeliveryAttemptPersistenceFailed, err)
	}

	query := `SELECT da.id, n.type, n.title, n.body,
		COALESCE((n.payload->>'image_url'), '') as image_url,
		COALESCE((n.payload->>'module'), '') as module,
		COALESCE((n.payload->>'event_type'), '') as event_type,
		COALESCE((n.payload->>'entity_id'), '') as entity_id,
		COALESCE((n.payload->>'click_action'), '') as click_action,
		COALESCE((n.payload->>'bypass')::boolean, false) as bypass,
		n.payload->'extra' as extra,
		n.reference_id, n.reference_type,
		da.attempted_at, da.read_at
		FROM delivery_attempts da
		JOIN notifications n ON n.id = da.notification_id
		WHERE da.user_id = $1 AND da.channel = 'in_app'`
	args := []interface{}{q.UserID}
	if q.UnreadOnly {
		query += ` AND da.read_at IS NULL`
	}
	query += ` ORDER BY da.attempted_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, q.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, deliveryRepo.Meta{}, kernel.Wrap(notifconstant.CodeDeliveryAttemptPersistenceFailed, err)
	}
	defer rows.Close()

	var items []deliveryRepo.InboxReadItem
	for rows.Next() {
		var item deliveryRepo.InboxReadItem
		var extraJSON sql.NullString
		var imageURL, module, eventType, entityID, clickAction string
		var bypass bool
		var refID, refType sql.NullString
		var readAt sql.NullTime

		if err := rows.Scan(
			&item.DeliveryAttemptID, &item.Type, &item.Title, &item.Body,
			&imageURL, &module, &eventType, &entityID, &clickAction, &bypass,
			&extraJSON, &refID, &refType,
			&item.AttemptedAt, &readAt,
		); err != nil {
			return nil, deliveryRepo.Meta{}, kernel.Wrap(notifconstant.CodeDeliveryAttemptPersistenceFailed, err)
		}

		item.ImageURL = imageURL
		item.Module = module
		item.EventType = eventType
		item.EntityID = entityID
		item.ClickAction = clickAction
		item.Bypass = bypass
		item.ReferenceID = fromNullStr(refID)
		item.ReferenceType = fromNullStr(refType)
		item.ReadAt = timeFromNull(readAt)

		if extraJSON.Valid && extraJSON.String != "null" {
			var extra map[string]string
			if err := json.Unmarshal([]byte(extraJSON.String), &extra); err == nil {
				item.Extra = extra
			}
		}

		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, deliveryRepo.Meta{}, kernel.Wrap(notifconstant.CodeDeliveryAttemptPersistenceFailed, err)
	}

	totalPages := total / q.Limit
	if total%q.Limit > 0 {
		totalPages++
	}

	return items, deliveryRepo.Meta{
		CurrentPage: q.Page,
		PerPage:     q.Limit,
		Total:       total,
		TotalPages:  totalPages,
	}, nil
}

func (r *PostgresDeliveryAttemptRepository) CountUnreadInApp(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM delivery_attempts
		 WHERE user_id = $1 AND channel = 'in_app' AND read_at IS NULL`, userID,
	).Scan(&count)
	if err != nil {
		return 0, kernel.Wrap(notifconstant.CodeDeliveryAttemptPersistenceFailed, err)
	}
	return count, nil
}

func (r *PostgresDeliveryAttemptRepository) MarkRead(ctx context.Context, id, userID string) error {
	result, err := execerFromContext(ctx, r.db).ExecContext(ctx,
		`UPDATE delivery_attempts SET read_at = NOW()
		 WHERE id = $1 AND user_id = $2 AND read_at IS NULL`, id, userID,
	)
	if err != nil {
		return kernel.Wrap(notifconstant.CodeDeliveryAttemptPersistenceFailed, err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return kernel.New(notifconstant.CodeDeliveryAttemptNotFound)
	}
	return nil
}

func (r *PostgresDeliveryAttemptRepository) MarkAllRead(ctx context.Context, userID string) (int, error) {
	result, err := execerFromContext(ctx, r.db).ExecContext(ctx,
		`UPDATE delivery_attempts SET read_at = NOW()
		 WHERE user_id = $1 AND channel = 'in_app' AND read_at IS NULL`, userID,
	)
	if err != nil {
		return 0, kernel.Wrap(notifconstant.CodeDeliveryAttemptPersistenceFailed, err)
	}
	rows, _ := result.RowsAffected()
	return int(rows), nil
}

func (r *PostgresDeliveryAttemptRepository) FindPendingByChannel(ctx context.Context, channel notifconstant.NotificationChannel) ([]*deliveryEntity.DeliveryAttempt, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, notification_id, user_id, channel, status, provider_code, retry_count, next_retry_at, read_at, attempted_at
		 FROM delivery_attempts WHERE channel = $1 AND status IN ('pending', 'retrying')
		 ORDER BY next_retry_at ASC LIMIT 50`, string(channel),
	)
	if err != nil {
		return nil, kernel.Wrap(notifconstant.CodeDeliveryAttemptPersistenceFailed, err)
	}
	defer rows.Close()

	var result []*deliveryEntity.DeliveryAttempt
	for rows.Next() {
		da, err := scanDeliveryAttempt(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, da)
	}
	return result, rows.Err()
}

func scanDeliveryAttempt(row scanner) (*deliveryEntity.DeliveryAttempt, error) {
	var da deliveryEntity.DeliveryAttempt
	var channel, status string
	var providerCode sql.NullString
	var nextRetryAt, readAt sql.NullTime

	if err := row.Scan(
		&da.ID, &da.NotificationID, &da.UserID, &channel, &status,
		&providerCode, &da.RetryCount, &nextRetryAt, &readAt, &da.AttemptedAt,
	); err != nil {
		return nil, err
	}

	da.Channel = notifconstant.NotificationChannel(channel)
	da.Status = notifconstant.DeliveryStatus(status)
	da.ProviderCode = fromNullStr(providerCode)
	da.NextRetryAt = timeFromNull(nextRetryAt)
	da.ReadAt = timeFromNull(readAt)

	return &da, nil
}

var _ deliveryRepo.DeliveryAttemptRepository = (*PostgresDeliveryAttemptRepository)(nil)

func nullTime(t interface{}) interface{} {
	switch v := t.(type) {
	case *interface{}:
		return v
	default:
		_ = v
		return nil
	}
}
