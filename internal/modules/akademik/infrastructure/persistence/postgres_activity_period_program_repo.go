package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"sipon-be/internal/modules/akademik/domain/activity_period_program/constant"
	"sipon-be/internal/modules/akademik/domain/activity_period_program/entity"
	"sipon-be/internal/shared/kernel"
)

type PostgresActivityPeriodProgramRepository struct {
	db *sql.DB
}

func NewPostgresActivityPeriodProgramRepository(db *sql.DB) *PostgresActivityPeriodProgramRepository {
	return &PostgresActivityPeriodProgramRepository{db: db}
}

func (r *PostgresActivityPeriodProgramRepository) Save(ctx context.Context, p *entity.ActivityPeriodProgram) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx,
		`INSERT INTO activity_period_programs (id, activity_period_id, program_id) VALUES ($1,$2,$3)`,
		p.ID, p.ActivityPeriodID, p.ProgramID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeActivityPeriodProgramDuplicate, err)
		}
		return kernel.Wrap(constant.CodeActivityPeriodProgramPersistenceFailed, fmt.Errorf("save activity period program: %w", err))
	}
	return nil
}

func (r *PostgresActivityPeriodProgramRepository) Delete(ctx context.Context, id string) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx, `DELETE FROM activity_period_programs WHERE id=$1`, id)
	if err != nil {
		return kernel.Wrap(constant.CodeActivityPeriodProgramPersistenceFailed, fmt.Errorf("delete activity period program: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeActivityPeriodProgramNotFound)
	}
	return nil
}

func (r *PostgresActivityPeriodProgramRepository) FindByActivityPeriodAndProgram(ctx context.Context, activityPeriodID, programID string) (*entity.ActivityPeriodProgram, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx,
		`SELECT id, activity_period_id, program_id FROM activity_period_programs WHERE activity_period_id=$1 AND program_id=$2`,
		activityPeriodID, programID)
	return scanActivityPeriodProgram(row)
}

func (r *PostgresActivityPeriodProgramRepository) ListByActivityPeriod(ctx context.Context, activityPeriodID string) ([]*entity.ActivityPeriodProgram, error) {
	execer := execerFromContext(ctx, r.db)
	rows, err := execer.QueryContext(ctx,
		`SELECT id, activity_period_id, program_id FROM activity_period_programs WHERE activity_period_id=$1`,
		activityPeriodID)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeActivityPeriodProgramQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*entity.ActivityPeriodProgram, 0)
	for rows.Next() {
		p, err := scanActivityPeriodProgram(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func scanActivityPeriodProgram(sc scanner) (*entity.ActivityPeriodProgram, error) {
	var id, activityPeriodID, programID string
	err := sc.Scan(&id, &activityPeriodID, &programID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeActivityPeriodProgramNotFound)
		}
		return nil, kernel.Wrap(constant.CodeActivityPeriodProgramQueryFailed, fmt.Errorf("scan activity period program: %w", err))
	}
	return &entity.ActivityPeriodProgram{
		ID:               id,
		ActivityPeriodID: activityPeriodID,
		ProgramID:        programID,
	}, nil
}
