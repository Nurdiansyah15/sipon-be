package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"sipon-be/internal/modules/notification/domain/device/constant"
	"sipon-be/internal/modules/notification/domain/device/entity"
	"sipon-be/internal/shared/kernel"
)

type PostgresDeviceRegistrationRepository struct {
	db *sql.DB
}

func NewPostgresDeviceRegistrationRepository(db *sql.DB) *PostgresDeviceRegistrationRepository {
	return &PostgresDeviceRegistrationRepository{db: db}
}

func (r *PostgresDeviceRegistrationRepository) Save(ctx context.Context, dr *entity.DeviceRegistration) error {
	_, err := execerFromContext(ctx, r.db).ExecContext(ctx,
		`INSERT INTO device_registrations (id, user_id, provider_token, push_provider, platform, device_id, device_name, device_model, os_version, app_version, timezone, active, last_seen_at, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		dr.ID, dr.UserID, dr.ProviderToken, string(dr.PushProvider), string(dr.Platform),
		dr.DeviceID, dr.DeviceName, dr.DeviceModel, dr.OSVersion, dr.AppVersion, dr.Timezone,
		dr.Active, dr.LastSeenAt, dr.CreatedAt, dr.UpdatedAt,
	)
	if err != nil {
		return kernel.Wrap(constant.CodeDevicePersistenceFailed, fmt.Errorf("insert device: %w", err))
	}
	return nil
}

func (r *PostgresDeviceRegistrationRepository) Update(ctx context.Context, dr *entity.DeviceRegistration) error {
	_, err := execerFromContext(ctx, r.db).ExecContext(ctx,
		`UPDATE device_registrations
		 SET user_id=$1, provider_token=$2, active=$3, last_seen_at=$4, updated_at=$5, timezone=$6
		 WHERE id=$7`,
		dr.UserID, dr.ProviderToken, dr.Active, dr.LastSeenAt, dr.UpdatedAt, dr.Timezone, dr.ID,
	)
	if err != nil {
		return kernel.Wrap(constant.CodeDevicePersistenceFailed, fmt.Errorf("update device: %w", err))
	}
	return nil
}

func (r *PostgresDeviceRegistrationRepository) FindByID(ctx context.Context, id string) (*entity.DeviceRegistration, error) {
	return r.scanOne(ctx,
		`SELECT id, user_id, provider_token, push_provider, platform, device_id, device_name, device_model, os_version, app_version, timezone, active, last_seen_at, created_at, updated_at
		 FROM device_registrations WHERE id=$1`, id)
}

func (r *PostgresDeviceRegistrationRepository) FindByToken(ctx context.Context, token string) (*entity.DeviceRegistration, error) {
	return r.scanOne(ctx,
		`SELECT id, user_id, provider_token, push_provider, platform, device_id, device_name, device_model, os_version, app_version, timezone, active, last_seen_at, created_at, updated_at
		 FROM device_registrations WHERE provider_token=$1`, token)
}

func (r *PostgresDeviceRegistrationRepository) FindByUserIDAndToken(ctx context.Context, userID, token string) (*entity.DeviceRegistration, error) {
	return r.scanOne(ctx,
		`SELECT id, user_id, provider_token, push_provider, platform, device_id, device_name, device_model, os_version, app_version, timezone, active, last_seen_at, created_at, updated_at
		 FROM device_registrations WHERE user_id=$1 AND provider_token=$2`, userID, token)
}

func (r *PostgresDeviceRegistrationRepository) FindByUserID(ctx context.Context, userID string, includeInactive bool) ([]*entity.DeviceRegistration, error) {
	query := `SELECT id, user_id, provider_token, push_provider, platform, device_id, device_name, device_model, os_version, app_version, timezone, active, last_seen_at, created_at, updated_at
		FROM device_registrations WHERE user_id=$1`
	if !includeInactive {
		query += ` AND active=TRUE`
	}
	query += ` ORDER BY last_seen_at DESC`
	return r.scanList(ctx, query, userID)
}

func (r *PostgresDeviceRegistrationRepository) FindActiveByUserID(ctx context.Context, userID string) ([]*entity.DeviceRegistration, error) {
	return r.scanList(ctx,
		`SELECT id, user_id, provider_token, push_provider, platform, device_id, device_name, device_model, os_version, app_version, timezone, active, last_seen_at, created_at, updated_at
		 FROM device_registrations WHERE user_id=$1 AND active=TRUE ORDER BY last_seen_at DESC`, userID)
}

func (r *PostgresDeviceRegistrationRepository) FindActiveByUserIDs(ctx context.Context, userIDs []string) (map[string][]*entity.DeviceRegistration, error) {
	if len(userIDs) == 0 {
		return map[string][]*entity.DeviceRegistration{}, nil
	}
	query := `SELECT id, user_id, provider_token, push_provider, platform, device_id, device_name, device_model, os_version, app_version, timezone, active, last_seen_at, created_at, updated_at
		FROM device_registrations WHERE user_id IN (`
	args := make([]interface{}, 0, len(userIDs))
	for i, uid := range userIDs {
		if i > 0 {
			query += ","
		}
		query += fmt.Sprintf("$%d", len(args)+1)
		args = append(args, uid)
	}
	query += `) AND active=TRUE ORDER BY last_seen_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeDevicePersistenceFailed, err)
	}
	defer rows.Close()

	result := make(map[string][]*entity.DeviceRegistration)
	for rows.Next() {
		dr, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		result[dr.UserID] = append(result[dr.UserID], dr)
	}
	return result, rows.Err()
}

func (r *PostgresDeviceRegistrationRepository) DeactivateByToken(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE device_registrations SET active=FALSE, updated_at=NOW() WHERE provider_token=$1`, token)
	return err
}

func (r *PostgresDeviceRegistrationRepository) DeactivateAllByUserID(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE device_registrations SET active=FALSE, updated_at=NOW() WHERE user_id=$1`, userID)
	return err
}

func (r *PostgresDeviceRegistrationRepository) scanOne(ctx context.Context, query string, args ...interface{}) (*entity.DeviceRegistration, error) {
	row := r.db.QueryRowContext(ctx, query, args...)
	return r.scanRow(row)
}

func (r *PostgresDeviceRegistrationRepository) scanList(ctx context.Context, query string, args ...interface{}) ([]*entity.DeviceRegistration, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeDevicePersistenceFailed, err)
	}
	defer rows.Close()

	var result []*entity.DeviceRegistration
	for rows.Next() {
		dr, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, dr)
	}
	return result, rows.Err()
}

func (r *PostgresDeviceRegistrationRepository) scanRow(row scanner) (*entity.DeviceRegistration, error) {
	var dr entity.DeviceRegistration
	var platform, pushProvider string
	var deviceID, deviceName, deviceModel, osVersion, appVersion, timezone sql.NullString
	var lastSeenAt time.Time

	if err := row.Scan(
		&dr.ID, &dr.UserID, &dr.ProviderToken, &pushProvider, &platform,
		&deviceID, &deviceName, &deviceModel, &osVersion, &appVersion, &timezone,
		&dr.Active, &lastSeenAt, &dr.CreatedAt, &dr.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, kernel.New(constant.CodeDeviceNotFound)
		}
		return nil, kernel.Wrap(constant.CodeDevicePersistenceFailed, err)
	}

	dr.Platform = constant.Platform(platform)
	dr.PushProvider = constant.PushProvider(pushProvider)
	dr.DeviceID = fromNullStr(deviceID)
	dr.DeviceName = fromNullStr(deviceName)
	dr.DeviceModel = fromNullStr(deviceModel)
	dr.OSVersion = fromNullStr(osVersion)
	dr.AppVersion = fromNullStr(appVersion)
	dr.Timezone = fromNullStr(timezone)
	dr.LastSeenAt = lastSeenAt

	return &dr, nil
}
