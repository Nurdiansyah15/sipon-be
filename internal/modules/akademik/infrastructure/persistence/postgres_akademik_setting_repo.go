package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	setConst "sipon-be/internal/modules/akademik/domain/setting/constant"
	setEntity "sipon-be/internal/modules/akademik/domain/setting/entity"
	"sipon-be/internal/shared/kernel"
)

const akademikSettingColumns = `id, settings, created_at, updated_at`

type PostgresAkademikSettingRepository struct {
	db *sql.DB
}

func NewPostgresAkademikSettingRepository(db *sql.DB) *PostgresAkademikSettingRepository {
	return &PostgresAkademikSettingRepository{db: db}
}

func (r *PostgresAkademikSettingRepository) Find(ctx context.Context) (*setEntity.AkademikSetting, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx,
		`SELECT `+akademikSettingColumns+` FROM akademik_settings WHERE id=$1`,
		setConst.SettingsRowID,
	)
	return scanAkademikSetting(row)
}

func (r *PostgresAkademikSettingRepository) Update(ctx context.Context, setting *setEntity.AkademikSetting) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx,
		`UPDATE akademik_settings SET settings=$1, updated_at=NOW() WHERE id=$2`,
		jsonBytes(setting.Settings), setting.ID,
	)
	if err != nil {
		return kernel.Wrap(setConst.CodeSettingInvalid, fmt.Errorf("update akademik setting: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(setConst.CodeSettingNotFound)
	}
	return nil
}

func scanAkademikSetting(sc scanner) (*setEntity.AkademikSetting, error) {
	var (
		id                   string
		settingsRaw          []byte
		createdAt, updatedAt time.Time
	)
	err := sc.Scan(&id, &settingsRaw, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(setConst.CodeSettingNotFound)
		}
		return nil, kernel.Wrap(setConst.CodeSettingInvalid, fmt.Errorf("scan akademik setting: %w", err))
	}
	if len(settingsRaw) == 0 {
		settingsRaw = []byte("{}")
	}
	return &setEntity.AkademikSetting{
		ID:        id,
		Settings:  json.RawMessage(settingsRaw),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
