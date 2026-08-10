package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	setConst "sipon-be/internal/modules/keuangan/domain/setting/constant"
	setEntity "sipon-be/internal/modules/keuangan/domain/setting/entity"
	"sipon-be/internal/shared/kernel"
)

const keuanganSettingColumns = `id, settings, created_at, updated_at`

type PostgresKeuanganSettingRepository struct {
	db *sql.DB
}

func NewPostgresKeuanganSettingRepository(db *sql.DB) *PostgresKeuanganSettingRepository {
	return &PostgresKeuanganSettingRepository{db: db}
}

func (r *PostgresKeuanganSettingRepository) Find(ctx context.Context) (*setEntity.KeuanganSetting, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx,
		`SELECT `+keuanganSettingColumns+` FROM keuangan_settings WHERE id=$1`,
		setConst.SettingsRowID,
	)
	return scanKeuanganSetting(row)
}

func (r *PostgresKeuanganSettingRepository) Update(ctx context.Context, setting *setEntity.KeuanganSetting) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx,
		`UPDATE keuangan_settings SET settings=$1, updated_at=NOW() WHERE id=$2`,
		jsonBytes(setting.Settings), setting.ID,
	)
	if err != nil {
		return kernel.Wrap(setConst.CodeSettingInvalid, fmt.Errorf("update setting: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(setConst.CodeSettingNotFound)
	}
	return nil
}

func scanKeuanganSetting(sc scanner) (*setEntity.KeuanganSetting, error) {
	var (
		id                    string
		settingsRaw           []byte
		createdAt, updatedAt  time.Time
	)
	err := sc.Scan(&id, &settingsRaw, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(setConst.CodeSettingNotFound)
		}
		return nil, kernel.Wrap(setConst.CodeSettingInvalid, fmt.Errorf("scan setting: %w", err))
	}
	if len(settingsRaw) == 0 {
		settingsRaw = []byte("{}")
	}
	return &setEntity.KeuanganSetting{
		ID:        id,
		Settings:  json.RawMessage(settingsRaw),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
