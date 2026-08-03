package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sconstant "sipon-be/internal/modules/psb/domain/setting/constant"
	sentity "sipon-be/internal/modules/psb/domain/setting/entity"
	"sipon-be/internal/shared/kernel"
)

const settingColumns = `
	id, name, start_period, end_period, status, quota, reg_fee, bank_accounts,
	data_purged_at, created_at, updated_at, deleted_at
`

type PostgresSettingRepository struct {
	db *sql.DB
}

func NewPostgresSettingRepository(db *sql.DB) *PostgresSettingRepository {
	return &PostgresSettingRepository{db: db}
}

func (r *PostgresSettingRepository) Save(ctx context.Context, s *sentity.PsbSetting) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx,
		`INSERT INTO psb_settings (`+settingColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		s.ID, s.Name, s.StartPeriod, s.EndPeriod, string(s.Status),
		jsonBytes(s.Quota), s.RegFee, jsonBytes(s.BankAccounts),
		nullTimeVal(s.DataPurgedAt), s.CreatedAt, s.UpdatedAt, nullTimeVal(s.DeletedAt),
	)
	if err != nil {
		return kernel.Wrap(sconstant.ErrCodeInvalidSetting, fmt.Errorf("save setting: %w", err))
	}
	return nil
}

func (r *PostgresSettingRepository) Update(ctx context.Context, s *sentity.PsbSetting) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx,
		`UPDATE psb_settings SET name=$1, start_period=$2, end_period=$3, status=$4, quota=$5, reg_fee=$6, bank_accounts=$7, data_purged_at=$8, updated_at=$9, deleted_at=$10 WHERE id=$11 AND deleted_at IS NULL`,
		s.Name, s.StartPeriod, s.EndPeriod, string(s.Status),
		jsonBytes(s.Quota), s.RegFee, jsonBytes(s.BankAccounts),
		nullTimeVal(s.DataPurgedAt), s.UpdatedAt, nullTimeVal(s.DeletedAt), s.ID,
	)
	if err != nil {
		return kernel.Wrap(sconstant.ErrCodeInvalidSetting, fmt.Errorf("update setting: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(sconstant.ErrCodeInvalidSetting)
	}
	return nil
}

func (r *PostgresSettingRepository) FindByID(ctx context.Context, id string) (*sentity.PsbSetting, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+settingColumns+` FROM psb_settings WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanSetting(row)
}

func (r *PostgresSettingRepository) FindActive(ctx context.Context) (*sentity.PsbSetting, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+settingColumns+` FROM psb_settings WHERE status='active' AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 1`)
	return scanSetting(row)
}

func (r *PostgresSettingRepository) List(ctx context.Context) ([]*sentity.PsbSetting, error) {
	execer := execerFromContext(ctx, r.db)
	rows, err := execer.QueryContext(ctx, `SELECT `+settingColumns+` FROM psb_settings WHERE deleted_at IS NULL ORDER BY created_at DESC`)
	if err != nil {
		return nil, kernel.Wrap(sconstant.ErrCodeInvalidSetting, fmt.Errorf("list settings: %w", err))
	}
	defer rows.Close()

	var items []*sentity.PsbSetting
	for rows.Next() {
		s, err := scanSetting(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

func scanSetting(sc scanner) (*sentity.PsbSetting, error) {
	var (
		id                                          string
		name                                        string
		startPeriod, endPeriod                      time.Time
		status                                      string
		quota, bankAccounts                         sql.NullString
		regFee                                      sql.NullFloat64
		dataPurgedAt                                sql.NullTime
		createdAt, updatedAt                        time.Time
		deletedAt                                   sql.NullTime
	)
	err := sc.Scan(&id, &name, &startPeriod, &endPeriod, &status,
		&quota, &regFee, &bankAccounts,
		&dataPurgedAt, &createdAt, &updatedAt, &deletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(sconstant.ErrCodeInvalidSetting)
		}
		return nil, kernel.Wrap(sconstant.ErrCodeInvalidSetting, fmt.Errorf("scan setting: %w", err))
	}

	var fee float64
	if regFee.Valid {
		fee = regFee.Float64
	}

	return &sentity.PsbSetting{
		ID:           id,
		Name:         name,
		StartPeriod:  startPeriod,
		EndPeriod:    endPeriod,
		Status:       sconstant.SettingStatus(status),
		Quota:        jsonBytesPtr(quota),
		RegFee:       fee,
		BankAccounts: jsonBytesPtr(bankAccounts),
		DataPurgedAt: timeFromNull(dataPurgedAt),
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		DeletedAt:    timeFromNull(deletedAt),
	}, nil
}

func jsonBytes(v json.RawMessage) []byte {
	if len(v) == 0 {
		return []byte("{}")
	}
	return []byte(v)
}

func jsonBytesPtr(ns sql.NullString) json.RawMessage {
	if !ns.Valid || ns.String == "" {
		return json.RawMessage("{}")
	}
	return json.RawMessage(ns.String)
}
