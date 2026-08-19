package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"sipon-be/internal/modules/notification/domain/preference/entity"
	prefRepo "sipon-be/internal/modules/notification/domain/preference/repository"
	"sipon-be/internal/shared/kernel"
)

type PostgresPreferenceRepository struct {
	db *sql.DB
}

func NewPostgresPreferenceRepository(db *sql.DB) *PostgresPreferenceRepository {
	return &PostgresPreferenceRepository{db: db}
}

func (r *PostgresPreferenceRepository) FindOrCreateByUserID(ctx context.Context, userID string) (*entity.NotificationPreference, error) {
	pref, err := r.findByUserID(ctx, userID)
	if err == nil {
		return pref, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	pref = entity.NewNotificationPreference(
		"00000000-0000-0000-0000-000000000000",
		userID,
	)
	err = r.db.QueryRowContext(ctx,
		`INSERT INTO notification_preferences (id, user_id, all_enabled, do_not_disturb, dnd_start_time, dnd_end_time, module_preferences, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, created_at, updated_at`,
		pref.UserID, pref.AllEnabled, pref.DoNotDisturbEnabled,
		pref.DNDStartTime, pref.DNDEndTime, "{}",
		pref.CreatedAt, pref.UpdatedAt,
	).Scan(&pref.ID, &pref.CreatedAt, &pref.UpdatedAt)
	if err != nil {
		return nil, kernel.Wrap(prefconstantCodePersistenceFailed, fmt.Errorf("insert preference: %w", err))
	}

	return pref, nil
}

func (r *PostgresPreferenceRepository) findByUserID(ctx context.Context, userID string) (*entity.NotificationPreference, error) {
	var p entity.NotificationPreference
	var allEnabled, dndEnabled bool
	var dndStart, dndEnd sql.NullString
	var modulePrefs []byte

	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, all_enabled, do_not_disturb, dnd_start_time, dnd_end_time, module_preferences, created_at, updated_at
		 FROM notification_preferences WHERE user_id = $1`, userID,
	).Scan(&p.ID, &p.UserID, &allEnabled, &dndEnabled, &dndStart, &dndEnd, &modulePrefs, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	p.AllEnabled = allEnabled
	p.DoNotDisturbEnabled = dndEnabled
	p.DNDStartTime = fromNullStr(dndStart)
	p.DNDEndTime = fromNullStr(dndEnd)

	return &p, nil
}

func (r *PostgresPreferenceRepository) Update(ctx context.Context, pref *entity.NotificationPreference) error {
	_, err := execerFromContext(ctx, r.db).ExecContext(ctx,
		`UPDATE notification_preferences
		 SET all_enabled = $1, do_not_disturb = $2, dnd_start_time = $3, dnd_end_time = $4,
		     module_preferences = $5, updated_at = $6
		 WHERE user_id = $7`,
		pref.AllEnabled, pref.DoNotDisturbEnabled, pref.DNDStartTime, pref.DNDEndTime,
		"{}", time.Now(), pref.UserID,
	)
	if err != nil {
		return kernel.Wrap(prefconstantCodePersistenceFailed, fmt.Errorf("update preference: %w", err))
	}
	return nil
}

func (r *PostgresPreferenceRepository) FindByUserIDs(ctx context.Context, userIDs []string) (map[string]*entity.NotificationPreference, error) {
	if len(userIDs) == 0 {
		return map[string]*entity.NotificationPreference{}, nil
	}

	query := `SELECT id, user_id, all_enabled, do_not_disturb, dnd_start_time, dnd_end_time, module_preferences, created_at, updated_at
		FROM notification_preferences WHERE user_id IN (`
	args := make([]interface{}, 0, len(userIDs))
	for i, uid := range userIDs {
		if i > 0 {
			query += ","
		}
		query += fmt.Sprintf("$%d", len(args)+1)
		args = append(args, uid)
	}
	query += ")"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, kernel.Wrap(prefconstantCodePersistenceFailed, err)
	}
	defer rows.Close()

	result := make(map[string]*entity.NotificationPreference)
	for rows.Next() {
		var p entity.NotificationPreference
		var allEnabled, dndEnabled bool
		var dndStart, dndEnd sql.NullString
		var modulePrefs []byte

		if err := rows.Scan(&p.ID, &p.UserID, &allEnabled, &dndEnabled, &dndStart, &dndEnd, &modulePrefs, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, kernel.Wrap(prefconstantCodePersistenceFailed, err)
		}
		p.AllEnabled = allEnabled
		p.DoNotDisturbEnabled = dndEnabled
		p.DNDStartTime = fromNullStr(dndStart)
		p.DNDEndTime = fromNullStr(dndEnd)
		result[p.UserID] = &p
	}
	return result, rows.Err()
}

const prefconstantCodePersistenceFailed = "PREFERENCE_PERSISTENCE_FAILED"

var _ prefRepo.NotificationPreferenceRepository = (*PostgresPreferenceRepository)(nil)
