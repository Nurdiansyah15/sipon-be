package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sipon-be/internal/modules/akademik/domain/attendance/constant"
	"sipon-be/internal/modules/akademik/domain/attendance/entity"
	"sipon-be/internal/shared/kernel"
)

const attendanceColumns = `
	id, activity_session_id, santri_id, status, recorded_at, created_at, updated_at, deleted_at
`

type PostgresAttendanceRepository struct {
	db *sql.DB
}

func NewPostgresAttendanceRepository(db *sql.DB) *PostgresAttendanceRepository {
	return &PostgresAttendanceRepository{db: db}
}

func (r *PostgresAttendanceRepository) Save(ctx context.Context, a *entity.Attendance) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx,
		`INSERT INTO attendances (`+attendanceColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		a.ID, a.ActivitySessionID, a.SantriID, string(a.Status), a.RecordedAt, a.CreatedAt, a.UpdatedAt, nullTimeVal(a.DeletedAt),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeAttendanceDuplicate, err)
		}
		return kernel.Wrap(constant.CodeAttendancePersistenceFailed, fmt.Errorf("save attendance: %w", err))
	}
	return nil
}

func (r *PostgresAttendanceRepository) Update(ctx context.Context, a *entity.Attendance) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx,
		`UPDATE attendances SET status=$1, updated_at=$2, deleted_at=$3 WHERE id=$4 AND deleted_at IS NULL`,
		string(a.Status), a.UpdatedAt, nullTimeVal(a.DeletedAt), a.ID,
	)
	if err != nil {
		return kernel.Wrap(constant.CodeAttendancePersistenceFailed, fmt.Errorf("update attendance: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeAttendanceNotFound)
	}
	return nil
}

func (r *PostgresAttendanceRepository) FindByID(ctx context.Context, id string) (*entity.Attendance, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+attendanceColumns+` FROM attendances WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanAttendance(row)
}

func (r *PostgresAttendanceRepository) FindBySessionAndSantri(ctx context.Context, sessionID, santriID string) (*entity.Attendance, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx,
		`SELECT `+attendanceColumns+` FROM attendances WHERE activity_session_id=$1 AND santri_id=$2 AND deleted_at IS NULL`,
		sessionID, santriID)
	return scanAttendance(row)
}

func (r *PostgresAttendanceRepository) ListBySession(ctx context.Context, sessionID string) ([]*entity.Attendance, error) {
	execer := execerFromContext(ctx, r.db)
	rows, err := execer.QueryContext(ctx,
		`SELECT `+attendanceColumns+` FROM attendances WHERE activity_session_id=$1 AND deleted_at IS NULL ORDER BY created_at DESC`,
		sessionID)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeAttendanceQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*entity.Attendance, 0)
	for rows.Next() {
		a, err := scanAttendance(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	return items, rows.Err()
}

func scanAttendance(sc scanner) (*entity.Attendance, error) {
	var (
		id, sessionID, santriID, status  string
		recordedAt, createdAt, updatedAt time.Time
		deletedAt                        sql.NullTime
	)
	err := sc.Scan(&id, &sessionID, &santriID, &status, &recordedAt, &createdAt, &updatedAt, &deletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeAttendanceNotFound)
		}
		return nil, kernel.Wrap(constant.CodeAttendanceQueryFailed, fmt.Errorf("scan attendance: %w", err))
	}
	return &entity.Attendance{
		ID:                id,
		ActivitySessionID: sessionID,
		SantriID:          santriID,
		Status:            constant.AttendanceStatus(status),
		RecordedAt:        recordedAt,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		DeletedAt:         timeFromNull(deletedAt),
	}, nil
}
