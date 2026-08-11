package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"sipon-be/internal/modules/akademik/domain/program/constant"
	"sipon-be/internal/modules/akademik/domain/program/entity"
	"sipon-be/internal/modules/akademik/domain/program/repository"
	"sipon-be/internal/shared/kernel"
)

const programColumns = `
	id, code, name, status, created_at, updated_at, deleted_at
`

type PostgresProgramRepository struct {
	db *sql.DB
}

func NewPostgresProgramRepository(db *sql.DB) *PostgresProgramRepository {
	return &PostgresProgramRepository{db: db}
}

func (r *PostgresProgramRepository) Save(ctx context.Context, p *entity.Program) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx,
		`INSERT INTO programs (`+programColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		p.ID, p.Code, p.Name, string(p.Status), p.CreatedAt, p.UpdatedAt, nullTimeVal(p.DeletedAt),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeProgramDuplicate, err)
		}
		return kernel.Wrap(constant.CodeProgramPersistenceFailed, fmt.Errorf("save program: %w", err))
	}
	return nil
}

func (r *PostgresProgramRepository) Update(ctx context.Context, p *entity.Program) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx,
		`UPDATE programs SET code=$1, name=$2, status=$3, updated_at=$4, deleted_at=$5 WHERE id=$6 AND deleted_at IS NULL`,
		p.Code, p.Name, string(p.Status), p.UpdatedAt, nullTimeVal(p.DeletedAt), p.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeProgramDuplicate, err)
		}
		return kernel.Wrap(constant.CodeProgramPersistenceFailed, fmt.Errorf("update program: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeProgramNotFound)
	}
	return nil
}

func (r *PostgresProgramRepository) FindByID(ctx context.Context, id string) (*entity.Program, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+programColumns+` FROM programs WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanProgram(row)
}

func (r *PostgresProgramRepository) FindByCode(ctx context.Context, code string) (*entity.Program, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+programColumns+` FROM programs WHERE code=$1 AND deleted_at IS NULL`, code)
	return scanProgram(row)
}

func (r *PostgresProgramRepository) FindByIDs(ctx context.Context, ids []string) ([]*entity.Program, error) {
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
		`SELECT `+programColumns+` FROM programs WHERE deleted_at IS NULL AND id IN (`+strings.Join(placeholders, ",")+`)`, args...)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeProgramQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*entity.Program, 0)
	for rows.Next() {
		p, err := scanProgram(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func (r *PostgresProgramRepository) List(ctx context.Context, q repository.ProgramListQuery) (*repository.ProgramListResult, error) {
	execer := execerFromContext(ctx, r.db)

	where := `deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 1
	if q.Status != nil && *q.Status != "" {
		where += fmt.Sprintf(` AND status=$%d`, argIdx)
		args = append(args, *q.Status)
		argIdx++
	}
	if q.Search != nil && *q.Search != "" {
		where += fmt.Sprintf(` AND (code ILIKE $%d OR name ILIKE $%d)`, argIdx, argIdx+1)
		args = append(args, "%"+*q.Search+"%", "%"+*q.Search+"%")
		argIdx += 2
	}

	var total int64
	if err := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM programs WHERE `+where, args...).Scan(&total); err != nil {
		return nil, kernel.Wrap(constant.CodeProgramQueryFailed, err)
	}

	limit := q.Limit
	offset := (q.Page - 1) * q.Limit
	listArgs := append(append([]interface{}{}, args...), limit, offset)
	rows, err := execer.QueryContext(ctx,
		fmt.Sprintf(`SELECT %s FROM programs WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
			programColumns, where, argIdx, argIdx+1), listArgs...)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeProgramQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*entity.Program, 0)
	for rows.Next() {
		p, err := scanProgram(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return &repository.ProgramListResult{Items: items, Total: total}, rows.Err()
}

func scanProgram(sc scanner) (*entity.Program, error) {
	var (
		id, code, name       string
		status               string
		createdAt, updatedAt time.Time
		deletedAt            sql.NullTime
	)
	err := sc.Scan(&id, &code, &name, &status, &createdAt, &updatedAt, &deletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeProgramNotFound)
		}
		return nil, kernel.Wrap(constant.CodeProgramQueryFailed, fmt.Errorf("scan program: %w", err))
	}
	return &entity.Program{
		ID:        id,
		Code:      code,
		Name:      name,
		Status:    constant.ProgramStatus(status),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		DeletedAt: timeFromNull(deletedAt),
	}, nil
}
