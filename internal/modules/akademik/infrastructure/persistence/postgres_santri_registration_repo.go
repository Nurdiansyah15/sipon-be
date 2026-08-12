package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sipon-be/internal/modules/akademik/domain/santri_registration/constant"
	"sipon-be/internal/modules/akademik/domain/santri_registration/entity"
	"sipon-be/internal/modules/akademik/domain/santri_registration/repository"
	"sipon-be/internal/shared/kernel"
)

const santriRegistrationColumns = `
	id, santri_id, academic_period_id, status, registered_at, revision_notes, created_at, updated_at, deleted_at
`

type PostgresSantriRegistrationRepository struct {
	db *sql.DB
}

func NewPostgresSantriRegistrationRepository(db *sql.DB) *PostgresSantriRegistrationRepository {
	return &PostgresSantriRegistrationRepository{db: db}
}

func (r *PostgresSantriRegistrationRepository) Save(ctx context.Context, reg *entity.SantriRegistration) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx,
		`INSERT INTO santri_registrations (`+santriRegistrationColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		reg.ID, reg.SantriID, reg.AcademicPeriodID, string(reg.Status), nullTimeVal(reg.RegisteredAt),
		nullStr(reg.RevisionNotes), reg.CreatedAt, reg.UpdatedAt, nullTimeVal(reg.DeletedAt),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeSantriRegistrationDuplicate, err)
		}
		return kernel.Wrap(constant.CodeSantriRegistrationPersistenceFailed, fmt.Errorf("save santri registration: %w", err))
	}
	return nil
}

func (r *PostgresSantriRegistrationRepository) Update(ctx context.Context, reg *entity.SantriRegistration) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx,
		`UPDATE santri_registrations SET status=$1, registered_at=$2, revision_notes=$3, updated_at=$4, deleted_at=$5 WHERE id=$6 AND deleted_at IS NULL`,
		string(reg.Status), nullTimeVal(reg.RegisteredAt), nullStr(reg.RevisionNotes), reg.UpdatedAt, nullTimeVal(reg.DeletedAt), reg.ID,
	)
	if err != nil {
		return kernel.Wrap(constant.CodeSantriRegistrationPersistenceFailed, fmt.Errorf("update santri registration: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeSantriRegistrationNotFound)
	}
	return nil
}

func (r *PostgresSantriRegistrationRepository) FindByID(ctx context.Context, id string) (*entity.SantriRegistration, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+santriRegistrationColumns+` FROM santri_registrations WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanSantriRegistration(row)
}

func (r *PostgresSantriRegistrationRepository) FindBySantriAndPeriod(ctx context.Context, santriID, academicPeriodID string) (*entity.SantriRegistration, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx,
		`SELECT `+santriRegistrationColumns+` FROM santri_registrations WHERE santri_id=$1 AND academic_period_id=$2 AND deleted_at IS NULL`,
		santriID, academicPeriodID)
	return scanSantriRegistration(row)
}

func (r *PostgresSantriRegistrationRepository) List(ctx context.Context, q repository.SantriRegistrationListQuery) (*repository.SantriRegistrationListResult, error) {
	execer := execerFromContext(ctx, r.db)

	where := `deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 1
	if q.AcademicPeriodID != nil && *q.AcademicPeriodID != "" {
		where += fmt.Sprintf(` AND academic_period_id=$%d`, argIdx)
		args = append(args, *q.AcademicPeriodID)
		argIdx++
	}
	if q.SantriID != nil && *q.SantriID != "" {
		where += fmt.Sprintf(` AND santri_id=$%d`, argIdx)
		args = append(args, *q.SantriID)
		argIdx++
	}
	if q.Status != nil && *q.Status != "" {
		where += fmt.Sprintf(` AND status=$%d`, argIdx)
		args = append(args, *q.Status)
		argIdx++
	}

	var total int64
	if err := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM santri_registrations WHERE `+where, args...).Scan(&total); err != nil {
		return nil, kernel.Wrap(constant.CodeSantriRegistrationQueryFailed, err)
	}

	limit := q.Limit
	offset := (q.Page - 1) * q.Limit
	listArgs := append(append([]interface{}{}, args...), limit, offset)
	rows, err := execer.QueryContext(ctx,
		fmt.Sprintf(`SELECT %s FROM santri_registrations WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
			santriRegistrationColumns, where, argIdx, argIdx+1), listArgs...)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeSantriRegistrationQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*entity.SantriRegistration, 0)
	for rows.Next() {
		reg, err := scanSantriRegistration(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, reg)
	}
	return &repository.SantriRegistrationListResult{Items: items, Total: total}, rows.Err()
}

func (r *PostgresSantriRegistrationRepository) ListCompletedByAcademicPeriod(ctx context.Context, academicPeriodID string) ([]*entity.SantriRegistration, error) {
	execer := execerFromContext(ctx, r.db)
	rows, err := execer.QueryContext(ctx,
		`SELECT `+santriRegistrationColumns+` FROM santri_registrations
		 WHERE academic_period_id=$1 AND status=$2 AND deleted_at IS NULL
		 ORDER BY created_at ASC`,
		academicPeriodID, string(constant.SantriRegistrationStatusCompleted),
	)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeSantriRegistrationQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*entity.SantriRegistration, 0)
	for rows.Next() {
		reg, err := scanSantriRegistration(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, reg)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(constant.CodeSantriRegistrationQueryFailed, err)
	}
	return items, nil
}

func scanSantriRegistration(sc scanner) (*entity.SantriRegistration, error) {
	var (
		id, santriID, academicPeriodID, status string
		registeredAt                           sql.NullTime
		revisionNotes                          sql.NullString
		createdAt, updatedAt                   time.Time
		deletedAt                              sql.NullTime
	)
	err := sc.Scan(&id, &santriID, &academicPeriodID, &status, &registeredAt, &revisionNotes, &createdAt, &updatedAt, &deletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeSantriRegistrationNotFound)
		}
		return nil, kernel.Wrap(constant.CodeSantriRegistrationQueryFailed, fmt.Errorf("scan santri registration: %w", err))
	}
	return &entity.SantriRegistration{
		ID:               id,
		SantriID:         santriID,
		AcademicPeriodID: academicPeriodID,
		Status:           constant.SantriRegistrationStatus(status),
		RegisteredAt:     timeFromNull(registeredAt),
		RevisionNotes:    strFromNull(revisionNotes),
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		DeletedAt:        timeFromNull(deletedAt),
	}, nil
}
