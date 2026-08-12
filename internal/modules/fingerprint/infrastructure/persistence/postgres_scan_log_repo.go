package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"sipon-be/internal/modules/fingerprint/domain/scanlog/constant"
	"sipon-be/internal/modules/fingerprint/domain/scanlog/entity"
	"sipon-be/internal/modules/fingerprint/domain/scanlog/repository"
	"sipon-be/internal/shared/kernel"
)

type PostgresScanLogRepository struct {
	db *sql.DB
}

func NewPostgresScanLogRepository(db *sql.DB) *PostgresScanLogRepository {
	return &PostgresScanLogRepository{db: db}
}

func (r *PostgresScanLogRepository) Insert(ctx context.Context, log *entity.ScanLog) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO fingerprint_scan_logs (id, sn, scan_date, pin, verifymode, inoutmode, deviceip, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		log.ID, log.SN, log.ScanDate, log.PIN, nullableInt(log.VerifyMode), nullableInt(log.InOutMode), nullableStr(log.DeviceIP), log.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeScanLogNotFound, err)
		}
		return kernel.Wrap(constant.CodeScanLogPersistenceFailed, fmt.Errorf("insert scan log: %w", err))
	}
	return nil
}

func (r *PostgresScanLogRepository) ListDistinctPinInRange(ctx context.Context, from, to time.Time) ([]repository.ScanPin, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT DISTINCT ON (pin) pin, sn, scan_date
		 FROM fingerprint_scan_logs
		 WHERE scan_date >= $1 AND scan_date <= $2
		 ORDER BY pin, scan_date ASC`,
		from, to)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeScanLogQueryFailed, err)
	}
	defer rows.Close()

	pins := make([]repository.ScanPin, 0)
	for rows.Next() {
		var p repository.ScanPin
		if err := rows.Scan(&p.PIN, &p.SN, &p.FirstScanAt); err != nil {
			return nil, kernel.Wrap(constant.CodeScanLogQueryFailed, fmt.Errorf("scan distinct pin: %w", err))
		}
		pins = append(pins, p)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(constant.CodeScanLogQueryFailed, err)
	}
	return pins, nil
}

func (r *PostgresScanLogRepository) ListInRange(ctx context.Context, from, to time.Time) ([]*entity.ScanLog, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, sn, scan_date, pin, verifymode, inoutmode, deviceip, created_at
		 FROM fingerprint_scan_logs
		 WHERE scan_date >= $1 AND scan_date <= $2
		 ORDER BY scan_date ASC`,
		from, to)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeScanLogQueryFailed, err)
	}
	defer rows.Close()

	logs := make([]*entity.ScanLog, 0)
	for rows.Next() {
		var (
			l        entity.ScanLog
			verify   sql.NullInt64
			inout    sql.NullInt64
			deviceIP sql.NullString
		)
		if err := rows.Scan(&l.ID, &l.SN, &l.ScanDate, &l.PIN, &verify, &inout, &deviceIP, &l.CreatedAt); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, kernel.New(constant.CodeScanLogNotFound)
			}
			return nil, kernel.Wrap(constant.CodeScanLogQueryFailed, fmt.Errorf("scan scan log: %w", err))
		}
		l.VerifyMode = int(verify.Int64)
		l.InOutMode = int(inout.Int64)
		l.DeviceIP = deviceIP.String
		logs = append(logs, &l)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(constant.CodeScanLogQueryFailed, err)
	}
	return logs, nil
}

func nullableStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullableInt(v int) interface{} {
	if v == 0 {
		return nil
	}
	return v
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
