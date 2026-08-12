package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sipon-be/internal/modules/akademik/domain/herregistrasi_document_requirement/constant"
	"sipon-be/internal/modules/akademik/domain/herregistrasi_document_requirement/entity"
	"sipon-be/internal/shared/kernel"
)

const herregDocRequirementColumns = `
	id, academic_period_id, kind, label, is_required, description, created_at, updated_at, deleted_at
`

type PostgresHerregistrasiDocumentRequirementRepository struct {
	db *sql.DB
}

func NewPostgresHerregistrasiDocumentRequirementRepository(db *sql.DB) *PostgresHerregistrasiDocumentRequirementRepository {
	return &PostgresHerregistrasiDocumentRequirementRepository{db: db}
}

func (r *PostgresHerregistrasiDocumentRequirementRepository) Save(ctx context.Context, req *entity.HerregistrasiDocumentRequirement) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx,
		`INSERT INTO herregistrasi_document_requirements (`+herregDocRequirementColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		req.ID, req.AcademicPeriodID, req.Kind, req.Label, req.IsRequired, nullStr(req.Description),
		req.CreatedAt, req.UpdatedAt, nullTimeVal(req.DeletedAt),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeHerregistrasiDocumentRequirementDuplicate, err)
		}
		return kernel.Wrap(constant.CodeHerregistrasiDocumentRequirementPersistenceFailed, fmt.Errorf("save herreg doc requirement: %w", err))
	}
	return nil
}

func (r *PostgresHerregistrasiDocumentRequirementRepository) Update(ctx context.Context, req *entity.HerregistrasiDocumentRequirement) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx,
		`UPDATE herregistrasi_document_requirements SET label=$1, is_required=$2, description=$3, updated_at=$4, deleted_at=$5 WHERE id=$6 AND deleted_at IS NULL`,
		req.Label, req.IsRequired, nullStr(req.Description), req.UpdatedAt, nullTimeVal(req.DeletedAt), req.ID,
	)
	if err != nil {
		return kernel.Wrap(constant.CodeHerregistrasiDocumentRequirementPersistenceFailed, fmt.Errorf("update herreg doc requirement: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeHerregistrasiDocumentRequirementNotFound)
	}
	return nil
}

func (r *PostgresHerregistrasiDocumentRequirementRepository) FindByID(ctx context.Context, id string) (*entity.HerregistrasiDocumentRequirement, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx,
		`SELECT `+herregDocRequirementColumns+` FROM herregistrasi_document_requirements WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanHerregDocRequirement(row)
}

func (r *PostgresHerregistrasiDocumentRequirementRepository) FindByAcademicPeriod(ctx context.Context, academicPeriodID string) ([]*entity.HerregistrasiDocumentRequirement, error) {
	execer := execerFromContext(ctx, r.db)
	rows, err := execer.QueryContext(ctx,
		`SELECT `+herregDocRequirementColumns+` FROM herregistrasi_document_requirements
		 WHERE academic_period_id=$1 AND deleted_at IS NULL ORDER BY created_at ASC`, academicPeriodID)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeHerregistrasiDocumentRequirementQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*entity.HerregistrasiDocumentRequirement, 0)
	for rows.Next() {
		req, err := scanHerregDocRequirement(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, req)
	}
	return items, rows.Err()
}

func (r *PostgresHerregistrasiDocumentRequirementRepository) Delete(ctx context.Context, id string) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx,
		`UPDATE herregistrasi_document_requirements SET deleted_at=$1, updated_at=$2 WHERE id=$3 AND deleted_at IS NULL`,
		time.Now(), time.Now(), id)
	if err != nil {
		return kernel.Wrap(constant.CodeHerregistrasiDocumentRequirementPersistenceFailed, fmt.Errorf("delete herreg doc requirement: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeHerregistrasiDocumentRequirementNotFound)
	}
	return nil
}

func scanHerregDocRequirement(sc scanner) (*entity.HerregistrasiDocumentRequirement, error) {
	var (
		id, academicPeriodID, kind, label string
		isRequired                        bool
		description                       sql.NullString
		createdAt, updatedAt              time.Time
		deletedAt                         sql.NullTime
	)
	err := sc.Scan(&id, &academicPeriodID, &kind, &label, &isRequired, &description, &createdAt, &updatedAt, &deletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeHerregistrasiDocumentRequirementNotFound)
		}
		return nil, kernel.Wrap(constant.CodeHerregistrasiDocumentRequirementQueryFailed, fmt.Errorf("scan herreg doc requirement: %w", err))
	}
	return &entity.HerregistrasiDocumentRequirement{
		ID:               id,
		AcademicPeriodID: academicPeriodID,
		Kind:             kind,
		Label:            label,
		IsRequired:       isRequired,
		Description:      strFromNull(description),
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		DeletedAt:        timeFromNull(deletedAt),
	}, nil
}
