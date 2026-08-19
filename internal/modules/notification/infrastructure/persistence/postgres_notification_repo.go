package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

notifconstant "sipon-be/internal/modules/notification/domain/notification/constant"
	notifentity "sipon-be/internal/modules/notification/domain/notification/entity"
	notifrepo "sipon-be/internal/modules/notification/domain/notification/repository"
	"sipon-be/internal/shared/kernel"
)

type PostgresNotificationRepository struct {
	db *sql.DB
}

func NewPostgresNotificationRepository(db *sql.DB) *PostgresNotificationRepository {
	return &PostgresNotificationRepository{db: db}
}

func (r *PostgresNotificationRepository) Save(ctx context.Context, n *notifentity.Notification) error {
	audienceDataJSON := toJSONB(n.AudienceData)
	payloadJSON := toJSONB(n.Payload)

	_, err := execerFromContext(ctx, r.db).ExecContext(ctx,
		`INSERT INTO notifications (id, type, title, body, payload, reference_id, reference_type, audience_type, audience_data, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		n.ID, string(n.Type), n.Title, n.Body, payloadJSON,
		nullStr(n.ReferenceID), nullStr(n.ReferenceType),
		string(n.AudienceType), audienceDataJSON, n.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(notifconstant.CodeNotificationPersistenceFailed, err)
		}
		return kernel.Wrap(notifconstant.CodeNotificationPersistenceFailed, fmt.Errorf("insert notification: %w", err))
	}
	return nil
}

func (r *PostgresNotificationRepository) FindByID(ctx context.Context, id string) (*notifentity.Notification, error) {
	var n notifentity.Notification
	var typ, audienceType string
	var payload, audienceData []byte
	var refID, refType sql.NullString

	err := r.db.QueryRowContext(ctx,
		`SELECT id, type, title, body, payload, reference_id, reference_type, audience_type, audience_data, created_at
		 FROM notifications WHERE id = $1`, id,
	).Scan(&n.ID, &typ, &n.Title, &n.Body, &payload, &refID, &refType, &audienceType, &audienceData, &n.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, kernel.New(notifconstant.CodeNotificationNotFound)
		}
		return nil, kernel.Wrap(notifconstant.CodeNotificationQueryFailed, err)
	}

	n.Type = notifconstant.NotificationType(typ)
	n.AudienceType = notifconstant.AudienceType(audienceType)
	n.ReferenceID = fromNullStr(refID)
	n.ReferenceType = fromNullStr(refType)

	return &n, nil
}

var _ notifrepo.NotificationRepository = (*PostgresNotificationRepository)(nil)

type scanner interface {
	Scan(dest ...interface{}) error
}

func nullStr(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

func fromNullStr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}

func toJSONB(v interface{}) []byte {
	b, _ := json.Marshal(v)
	if b == nil {
		return []byte("{}")
	}
	return b
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func execerFromContext(ctx context.Context, db *sql.DB) dbExecer {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok && tx != nil {
		return tx
	}
	return db
}

func scanNotification(row scanner) (notifentity.Notification, error) {
	var n notifentity.Notification
	var typ, audienceType string
	var payload, audienceData []byte
	var refID, refType sql.NullString
	err := row.Scan(&n.ID, &typ, &n.Title, &n.Body, &payload, &refID, &refType, &audienceType, &audienceData, &n.CreatedAt)
	if err != nil {
		return n, err
	}
	n.Type = notifconstant.NotificationType(typ)
	n.AudienceType = notifconstant.AudienceType(audienceType)
	n.ReferenceID = fromNullStr(refID)
	n.ReferenceType = fromNullStr(refType)
	return n, nil
}

func scanNotificationColumns() string {
	return `id, type, title, body, payload, reference_id, reference_type, audience_type, audience_data, created_at`
}

func timeFromNull(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	v := nt.Time
	return &v
}
