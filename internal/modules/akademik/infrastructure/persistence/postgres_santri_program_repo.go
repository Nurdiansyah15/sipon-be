package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	spConst "sipon-be/internal/modules/akademik/domain/santri_program/constant"
	spEntity "sipon-be/internal/modules/akademik/domain/santri_program/entity"
	"sipon-be/internal/shared/kernel"
)

const santriProgramColumns = `
	id, santri_id, program_id, is_active, started_at, ended_at,
	created_at, updated_at, deleted_at
`

type PostgresSantriProgramRepository struct {
	db *sql.DB
}

func NewPostgresSantriProgramRepository(db *sql.DB) *PostgresSantriProgramRepository {
	return &PostgresSantriProgramRepository{db: db}
}

func (r *PostgresSantriProgramRepository) Save(ctx context.Context, sp *spEntity.SantriProgram) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx,
		`INSERT INTO santri_programs (`+santriProgramColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		sp.ID, sp.SantriID, sp.ProgramID, sp.IsActive, sp.StartedAt, nullTimeVal(sp.EndedAt),
		sp.CreatedAt, sp.UpdatedAt, nullTimeVal(sp.DeletedAt),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(spConst.CodeSantriProgramDuplicate, err)
		}
		return kernel.Wrap(spConst.CodeSantriProgramPersistenceFailed, fmt.Errorf("save santri program: %w", err))
	}
	return nil
}

func (r *PostgresSantriProgramRepository) FindActiveBySantriID(ctx context.Context, santriID string) (*spEntity.SantriProgram, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx,
		`SELECT `+santriProgramColumns+` FROM santri_programs
		 WHERE santri_id=$1 AND is_active=true AND deleted_at IS NULL`,
		santriID,
	)
	return scanSantriProgram(row)
}

func (r *PostgresSantriProgramRepository) FindBySantriID(ctx context.Context, santriID string) ([]*spEntity.SantriProgram, error) {
	execer := execerFromContext(ctx, r.db)
	rows, err := execer.QueryContext(ctx,
		`SELECT `+santriProgramColumns+` FROM santri_programs
		 WHERE santri_id=$1 AND deleted_at IS NULL ORDER BY started_at DESC`,
		santriID,
	)
	if err != nil {
		return nil, kernel.Wrap(spConst.CodeSantriProgramQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*spEntity.SantriProgram, 0)
	for rows.Next() {
		sp, err := scanSantriProgram(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, sp)
	}
	return items, rows.Err()
}

func (r *PostgresSantriProgramRepository) ListActiveSantriIDsByProgramID(ctx context.Context, programID string) ([]string, error) {
	execer := execerFromContext(ctx, r.db)
	rows, err := execer.QueryContext(ctx,
		`SELECT santri_id FROM santri_programs
		 WHERE program_id=$1 AND is_active=true AND deleted_at IS NULL`,
		programID,
	)
	if err != nil {
		return nil, kernel.Wrap(spConst.CodeSantriProgramQueryFailed, err)
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, kernel.Wrap(spConst.CodeSantriProgramQueryFailed, err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *PostgresSantriProgramRepository) ListActive(ctx context.Context) ([]*spEntity.SantriProgram, error) {
	execer := execerFromContext(ctx, r.db)
	rows, err := execer.QueryContext(ctx,
		`SELECT `+santriProgramColumns+` FROM santri_programs
		 WHERE is_active=true AND deleted_at IS NULL`,
	)
	if err != nil {
		return nil, kernel.Wrap(spConst.CodeSantriProgramQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*spEntity.SantriProgram, 0)
	for rows.Next() {
		sp, err := scanSantriProgram(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, sp)
	}
	return items, rows.Err()
}

func (r *PostgresSantriProgramRepository) DeactivateAllBySantriID(ctx context.Context, santriID string) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx,
		`UPDATE santri_programs SET is_active=false, ended_at=COALESCE(ended_at, NOW()), updated_at=NOW()
		 WHERE santri_id=$1 AND is_active=true AND deleted_at IS NULL`,
		santriID,
	)
	if err != nil {
		return kernel.Wrap(spConst.CodeSantriProgramPersistenceFailed, fmt.Errorf("deactivate santri programs: %w", err))
	}
	return nil
}

func scanSantriProgram(sc scanner) (*spEntity.SantriProgram, error) {
	var (
		id, santriID, programID string
		isActive                bool
		startedAt, createdAt    time.Time
		updatedAt               time.Time
		endedAt, deletedAt      sql.NullTime
	)
	err := sc.Scan(
		&id, &santriID, &programID, &isActive, &startedAt, &endedAt,
		&createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(spConst.CodeSantriProgramNotFound)
		}
		return nil, kernel.Wrap(spConst.CodeSantriProgramQueryFailed, fmt.Errorf("scan santri program: %w", err))
	}
	return &spEntity.SantriProgram{
		ID:        id,
		SantriID:  santriID,
		ProgramID: programID,
		IsActive:  isActive,
		StartedAt: startedAt,
		EndedAt:   timeFromNull(endedAt),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		DeletedAt: timeFromNull(deletedAt),
	}, nil
}
