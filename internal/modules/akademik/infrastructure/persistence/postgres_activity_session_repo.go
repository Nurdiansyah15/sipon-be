package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"sipon-be/internal/modules/akademik/domain/activity_session/constant"
	"sipon-be/internal/modules/akademik/domain/activity_session/entity"
	"sipon-be/internal/modules/akademik/domain/activity_session/repository"
	"sipon-be/internal/shared/kernel"
)

const activitySessionColumns = `
	id, activity_schedule_id, starts_at, ends_at, status, created_at, updated_at, deleted_at
`

type PostgresActivitySessionRepository struct {
	db *sql.DB
}

func NewPostgresActivitySessionRepository(db *sql.DB) *PostgresActivitySessionRepository {
	return &PostgresActivitySessionRepository{db: db}
}

func (r *PostgresActivitySessionRepository) Save(ctx context.Context, s *entity.ActivitySession) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx,
		`INSERT INTO activity_sessions (`+activitySessionColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		s.ID, s.ActivityScheduleID, s.StartsAt, s.EndsAt, string(s.Status), s.CreatedAt, s.UpdatedAt, nullTimeVal(s.DeletedAt),
	)
	if err != nil {
		return kernel.Wrap(constant.CodeActivitySessionPersistenceFailed, fmt.Errorf("save activity session: %w", err))
	}
	return nil
}

func (r *PostgresActivitySessionRepository) Update(ctx context.Context, s *entity.ActivitySession) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx,
		`UPDATE activity_sessions SET status=$1, updated_at=$2, deleted_at=$3 WHERE id=$4 AND deleted_at IS NULL`,
		string(s.Status), s.UpdatedAt, nullTimeVal(s.DeletedAt), s.ID,
	)
	if err != nil {
		return kernel.Wrap(constant.CodeActivitySessionPersistenceFailed, fmt.Errorf("update activity session: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeActivitySessionNotFound)
	}
	return nil
}

func (r *PostgresActivitySessionRepository) FindByID(ctx context.Context, id string) (*entity.ActivitySession, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+activitySessionColumns+` FROM activity_sessions WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanActivitySession(row)
}

func (r *PostgresActivitySessionRepository) ListByScheduleIDs(ctx context.Context, scheduleIDs []string) ([]*entity.ActivitySession, error) {
	if len(scheduleIDs) == 0 {
		return nil, nil
	}
	execer := execerFromContext(ctx, r.db)
	placeholders := make([]string, len(scheduleIDs))
	args := make([]interface{}, len(scheduleIDs))
	for i, id := range scheduleIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	rows, err := execer.QueryContext(ctx,
		`SELECT `+activitySessionColumns+` FROM activity_sessions WHERE deleted_at IS NULL AND activity_schedule_id IN (`+strings.Join(placeholders, ",")+`)`,
		args...)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeActivitySessionQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*entity.ActivitySession, 0)
	for rows.Next() {
		s, err := scanActivitySession(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

func (r *PostgresActivitySessionRepository) List(ctx context.Context, q repository.ActivitySessionListQuery) (*repository.ActivitySessionListResult, error) {
	execer := execerFromContext(ctx, r.db)

	where := `s.deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 1
	if q.ActivityScheduleID != nil && *q.ActivityScheduleID != "" {
		where += fmt.Sprintf(` AND s.activity_schedule_id=$%d`, argIdx)
		args = append(args, *q.ActivityScheduleID)
		argIdx++
	}
	if q.AcademicPeriodID != nil && *q.AcademicPeriodID != "" {
		where += fmt.Sprintf(` AND ap.academic_period_id=$%d`, argIdx)
		args = append(args, *q.AcademicPeriodID)
		argIdx++
	}
	if q.Status != nil && *q.Status != "" {
		where += fmt.Sprintf(` AND s.status=$%d`, argIdx)
		args = append(args, *q.Status)
		argIdx++
	}
	if q.StartDate != nil && *q.StartDate != "" {
		where += fmt.Sprintf(` AND s.starts_at >= $%d::timestamptz`, argIdx)
		args = append(args, *q.StartDate)
		argIdx++
	}
	if q.EndDate != nil && *q.EndDate != "" {
		where += fmt.Sprintf(` AND s.ends_at <= $%d::timestamptz`, argIdx)
		args = append(args, *q.EndDate)
		argIdx++
	}

	from := `activity_sessions s
		JOIN activity_schedules sc ON sc.id = s.activity_schedule_id
		JOIN activity_periods ap ON ap.id = sc.activity_period_id`

	var total int64
	if err := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+from+` WHERE `+where, args...).Scan(&total); err != nil {
		return nil, kernel.Wrap(constant.CodeActivitySessionQueryFailed, err)
	}

	limit := q.Limit
	offset := (q.Page - 1) * q.Limit
	listArgs := append(append([]interface{}{}, args...), limit, offset)
	rows, err := execer.QueryContext(ctx,
		fmt.Sprintf(`SELECT s.id, s.activity_schedule_id, s.starts_at, s.ends_at, s.status, s.created_at, s.updated_at, s.deleted_at
			FROM %s WHERE %s ORDER BY s.starts_at DESC LIMIT $%d OFFSET $%d`,
			from, where, argIdx, argIdx+1), listArgs...)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeActivitySessionQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*entity.ActivitySession, 0)
	for rows.Next() {
		s, err := scanActivitySession(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return &repository.ActivitySessionListResult{Items: items, Total: total}, rows.Err()
}

func scanActivitySession(sc scanner) (*entity.ActivitySession, error) {
	var (
		id, scheduleID, status string
		startsAt, endsAt       time.Time
		createdAt, updatedAt   time.Time
		deletedAt              sql.NullTime
	)
	err := sc.Scan(&id, &scheduleID, &startsAt, &endsAt, &status, &createdAt, &updatedAt, &deletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeActivitySessionNotFound)
		}
		return nil, kernel.Wrap(constant.CodeActivitySessionQueryFailed, fmt.Errorf("scan activity session: %w", err))
	}
	return &entity.ActivitySession{
		ID:                 id,
		ActivityScheduleID: scheduleID,
		StartsAt:           startsAt,
		EndsAt:             endsAt,
		Status:             constant.ActivitySessionStatus(status),
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
		DeletedAt:          timeFromNull(deletedAt),
	}, nil
}
