package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"sipon-be/internal/modules/akademik/domain/activity_schedule/constant"
	"sipon-be/internal/modules/akademik/domain/activity_schedule/entity"
	"sipon-be/internal/shared/kernel"
)

const activityScheduleColumns = `
	id, activity_period_id, type, start_date, end_date, start_time::text, end_time::text,
	created_at, updated_at, deleted_at
`

const activityScheduleInsertColumns = `
	id, activity_period_id, type, start_date, end_date, start_time, end_time,
	created_at, updated_at, deleted_at
`

type PostgresActivityScheduleRepository struct {
	db *sql.DB
}

func NewPostgresActivityScheduleRepository(db *sql.DB) *PostgresActivityScheduleRepository {
	return &PostgresActivityScheduleRepository{db: db}
}

func (r *PostgresActivityScheduleRepository) Save(ctx context.Context, s *entity.ActivitySchedule) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx,
		`INSERT INTO activity_schedules (`+activityScheduleInsertColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		s.ID, s.ActivityPeriodID, string(s.Type), nullTimeVal(s.StartDate), nullTimeVal(s.EndDate),
		s.StartTime, s.EndTime, s.CreatedAt, s.UpdatedAt, nullTimeVal(s.DeletedAt),
	)
	if err != nil {
		return kernel.Wrap(constant.CodeActivitySchedulePersistenceFailed, fmt.Errorf("save activity schedule: %w", err))
	}
	return nil
}

func (r *PostgresActivityScheduleRepository) Update(ctx context.Context, s *entity.ActivitySchedule) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx,
		`UPDATE activity_schedules SET start_date=$1, end_date=$2, start_time=$3, end_time=$4, updated_at=$5, deleted_at=$6 WHERE id=$7 AND deleted_at IS NULL`,
		nullTimeVal(s.StartDate), nullTimeVal(s.EndDate), s.StartTime, s.EndTime, s.UpdatedAt, nullTimeVal(s.DeletedAt), s.ID,
	)
	if err != nil {
		return kernel.Wrap(constant.CodeActivitySchedulePersistenceFailed, fmt.Errorf("update activity schedule: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeActivityScheduleNotFound)
	}
	return nil
}

func (r *PostgresActivityScheduleRepository) FindByID(ctx context.Context, id string) (*entity.ActivitySchedule, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+activityScheduleColumns+` FROM activity_schedules WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanActivitySchedule(row)
}

func (r *PostgresActivityScheduleRepository) FindByIDs(ctx context.Context, ids []string) ([]*entity.ActivitySchedule, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	execer := execerFromContext(ctx, r.db)
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	rows, err := execer.QueryContext(ctx,
		`SELECT `+activityScheduleColumns+` FROM activity_schedules WHERE deleted_at IS NULL AND id IN (`+strings.Join(placeholders, ",")+`)`, args...)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeActivityScheduleQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*entity.ActivitySchedule, 0)
	for rows.Next() {
		s, err := scanActivitySchedule(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

func (r *PostgresActivityScheduleRepository) ListByActivityPeriod(ctx context.Context, activityPeriodID string) ([]*entity.ActivitySchedule, error) {
	execer := execerFromContext(ctx, r.db)
	rows, err := execer.QueryContext(ctx,
		`SELECT `+activityScheduleColumns+` FROM activity_schedules WHERE activity_period_id=$1 AND deleted_at IS NULL ORDER BY created_at DESC`,
		activityPeriodID)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeActivityScheduleQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*entity.ActivitySchedule, 0)
	for rows.Next() {
		s, err := scanActivitySchedule(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

func (r *PostgresActivityScheduleRepository) ReplaceWeeklies(ctx context.Context, scheduleID string, days []constant.DayOfWeek) error {
	execer := execerFromContext(ctx, r.db)
	if _, err := execer.ExecContext(ctx, `DELETE FROM activity_schedule_weeklies WHERE schedule_id=$1`, scheduleID); err != nil {
		return kernel.Wrap(constant.CodeActivitySchedulePersistenceFailed, err)
	}
	for _, day := range days {
		if _, err := execer.ExecContext(ctx,
			`INSERT INTO activity_schedule_weeklies (id, schedule_id, day_of_week) VALUES ($1,$2,$3)`,
			newUUID(), scheduleID, string(day)); err != nil {
			return kernel.Wrap(constant.CodeActivitySchedulePersistenceFailed, err)
		}
	}
	return nil
}

func (r *PostgresActivityScheduleRepository) ReplaceMonthlies(ctx context.Context, scheduleID string, days []int) error {
	execer := execerFromContext(ctx, r.db)
	if _, err := execer.ExecContext(ctx, `DELETE FROM activity_schedule_monthlies WHERE schedule_id=$1`, scheduleID); err != nil {
		return kernel.Wrap(constant.CodeActivitySchedulePersistenceFailed, err)
	}
	for _, day := range days {
		if _, err := execer.ExecContext(ctx,
			`INSERT INTO activity_schedule_monthlies (id, schedule_id, day_of_month) VALUES ($1,$2,$3)`,
			newUUID(), scheduleID, day); err != nil {
			return kernel.Wrap(constant.CodeActivitySchedulePersistenceFailed, err)
		}
	}
	return nil
}

func (r *PostgresActivityScheduleRepository) ReplaceYearlies(ctx context.Context, scheduleID string, dates []entity.YearlyDate) error {
	execer := execerFromContext(ctx, r.db)
	if _, err := execer.ExecContext(ctx, `DELETE FROM activity_schedule_yearlies WHERE schedule_id=$1`, scheduleID); err != nil {
		return kernel.Wrap(constant.CodeActivitySchedulePersistenceFailed, err)
	}
	for _, d := range dates {
		if _, err := execer.ExecContext(ctx,
			`INSERT INTO activity_schedule_yearlies (id, schedule_id, month, day) VALUES ($1,$2,$3,$4)`,
			newUUID(), scheduleID, d.Month, d.Day); err != nil {
			return kernel.Wrap(constant.CodeActivitySchedulePersistenceFailed, err)
		}
	}
	return nil
}

func (r *PostgresActivityScheduleRepository) ListWeeklies(ctx context.Context, scheduleID string) ([]entity.ActivityScheduleWeekly, error) {
	execer := execerFromContext(ctx, r.db)
	rows, err := execer.QueryContext(ctx,
		`SELECT id, schedule_id, day_of_week FROM activity_schedule_weeklies WHERE schedule_id=$1 ORDER BY day_of_week`, scheduleID)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeActivityScheduleQueryFailed, err)
	}
	defer rows.Close()

	items := make([]entity.ActivityScheduleWeekly, 0)
	for rows.Next() {
		var w entity.ActivityScheduleWeekly
		var day string
		if err := rows.Scan(&w.ID, &w.ScheduleID, &day); err != nil {
			return nil, err
		}
		w.DayOfWeek = constant.DayOfWeek(day)
		items = append(items, w)
	}
	return items, rows.Err()
}

func (r *PostgresActivityScheduleRepository) ListMonthlies(ctx context.Context, scheduleID string) ([]entity.ActivityScheduleMonthly, error) {
	execer := execerFromContext(ctx, r.db)
	rows, err := execer.QueryContext(ctx,
		`SELECT id, schedule_id, day_of_month FROM activity_schedule_monthlies WHERE schedule_id=$1 ORDER BY day_of_month`, scheduleID)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeActivityScheduleQueryFailed, err)
	}
	defer rows.Close()

	items := make([]entity.ActivityScheduleMonthly, 0)
	for rows.Next() {
		var m entity.ActivityScheduleMonthly
		if err := rows.Scan(&m.ID, &m.ScheduleID, &m.DayOfMonth); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

func (r *PostgresActivityScheduleRepository) ListYearlies(ctx context.Context, scheduleID string) ([]entity.ActivityScheduleYearly, error) {
	execer := execerFromContext(ctx, r.db)
	rows, err := execer.QueryContext(ctx,
		`SELECT id, schedule_id, month, day FROM activity_schedule_yearlies WHERE schedule_id=$1 ORDER BY month, day`, scheduleID)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeActivityScheduleQueryFailed, err)
	}
	defer rows.Close()

	items := make([]entity.ActivityScheduleYearly, 0)
	for rows.Next() {
		var y entity.ActivityScheduleYearly
		if err := rows.Scan(&y.ID, &y.ScheduleID, &y.Month, &y.Day); err != nil {
			return nil, err
		}
		items = append(items, y)
	}
	return items, rows.Err()
}

func scanActivitySchedule(sc scanner) (*entity.ActivitySchedule, error) {
	var (
		id, activityPeriodID, typ, startTime, endTime string
		startDate, endDate                            sql.NullTime
		createdAt, updatedAt                          time.Time
		deletedAt                                     sql.NullTime
	)
	err := sc.Scan(&id, &activityPeriodID, &typ, &startDate, &endDate, &startTime, &endTime,
		&createdAt, &updatedAt, &deletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeActivityScheduleNotFound)
		}
		return nil, kernel.Wrap(constant.CodeActivityScheduleQueryFailed, fmt.Errorf("scan activity schedule: %w", err))
	}
	return &entity.ActivitySchedule{
		ID:               id,
		ActivityPeriodID: activityPeriodID,
		Type:             constant.ActivityScheduleType(typ),
		StartDate:        timeFromNull(startDate),
		EndDate:          timeFromNull(endDate),
		StartTime:        startTime,
		EndTime:          endTime,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		DeletedAt:        timeFromNull(deletedAt),
	}, nil
}
