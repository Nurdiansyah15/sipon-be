package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	"sipon-be/internal/modules/keuangan/domain/feecomponent/entity"
	"sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	"sipon-be/internal/shared/kernel"
)

const feeComponentColumns = `
	id, code, name, type, amount, is_periodic, period_type, description, is_active,
	created_by, created_at, updated_at, deleted_at
`

type PostgresFeeComponentRepository struct {
	db *sql.DB
}

func NewPostgresFeeComponentRepository(db *sql.DB) *PostgresFeeComponentRepository {
	return &PostgresFeeComponentRepository{db: db}
}

func (r *PostgresFeeComponentRepository) Save(ctx context.Context, fc *entity.FeeComponent) error {
	execer := execerFromContext(ctx, r.db)

	query := `INSERT INTO fee_components (` + feeComponentColumns + `) VALUES (
		$1,$2,$3,$4,$5,$6,$7,$8,$9,
		$10,$11,$12,$13
	)`

	_, err := execer.ExecContext(ctx, query,
		fc.ID, fc.Code, fc.Name, string(fc.Type), fc.Amount, fc.IsPeriodic,
		nullStr((*string)(fc.PeriodType)), nullStr(fc.Description), fc.IsActive,
		fc.CreatedBy, fc.CreatedAt, fc.UpdatedAt, nullTimeVal(fc.DeletedAt),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.WrapMsg(constant.CodeFeeComponentDuplicate, "Komponen biaya dengan kode yang sama sudah ada", err)
		}
		return kernel.WrapMsg(constant.CodeFeeComponentPersistenceFailed, "gagal menyimpan komponen biaya", err)
	}
	return nil
}

func (r *PostgresFeeComponentRepository) Update(ctx context.Context, fc *entity.FeeComponent) error {
	execer := execerFromContext(ctx, r.db)

	query := `UPDATE fee_components SET
		code=$1, name=$2, type=$3, amount=$4, is_periodic=$5, period_type=$6,
		description=$7, is_active=$8, updated_at=$9, deleted_at=$10
		WHERE id=$11 AND deleted_at IS NULL`

	res, err := execer.ExecContext(ctx, query,
		fc.Code, fc.Name, string(fc.Type), fc.Amount, fc.IsPeriodic,
		nullStr((*string)(fc.PeriodType)), nullStr(fc.Description), fc.IsActive,
		fc.UpdatedAt, nullTimeVal(fc.DeletedAt),
		fc.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.WrapMsg(constant.CodeFeeComponentDuplicate, "Komponen biaya dengan kode yang sama sudah ada", err)
		}
		return kernel.WrapMsg(constant.CodeFeeComponentPersistenceFailed, "gagal memperbarui komponen biaya", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.WrapMsg(constant.CodeFeeComponentNotFound, "Komponen biaya tidak ditemukan", nil)
	}
	return nil
}

func (r *PostgresFeeComponentRepository) FindByID(ctx context.Context, id string) (*entity.FeeComponent, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+feeComponentColumns+` FROM fee_components WHERE id=$1 AND deleted_at IS NULL`, id)
	return r.scan(row)
}

func (r *PostgresFeeComponentRepository) FindByCode(ctx context.Context, code string) (*entity.FeeComponent, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+feeComponentColumns+` FROM fee_components WHERE code=$1 AND deleted_at IS NULL`, code)
	return r.scan(row)
}

func (r *PostgresFeeComponentRepository) List(ctx context.Context, q repository.FeeComponentListQuery) (*repository.FeeComponentListResult, error) {
	execer := execerFromContext(ctx, r.db)

	where := `WHERE deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 1
	if q.Type != nil && *q.Type != "" {
		where += fmt.Sprintf(` AND type=$%d`, argIdx)
		args = append(args, *q.Type)
		argIdx++
	}
	if q.Active != nil {
		where += fmt.Sprintf(` AND is_active=$%d`, argIdx)
		args = append(args, *q.Active)
		argIdx++
	}

	var total int64
	countRow := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM fee_components `+where, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, kernel.WrapMsg(constant.CodeFeeComponentQueryFailed, "gagal menghitung jumlah komponen biaya", err)
	}

	limit := q.Limit
	offset := (q.Page - 1) * q.Limit
	listArgs := append(append([]interface{}{}, args...), limit, offset)

	query := fmt.Sprintf(`SELECT %s FROM fee_components %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		feeComponentColumns, where, argIdx, argIdx+1)

	rows, err := execer.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, kernel.WrapMsg(constant.CodeFeeComponentQueryFailed, "gagal mendaftar komponen biaya", err)
	}
	defer rows.Close()

	items := make([]*entity.FeeComponent, 0)
	for rows.Next() {
		fc, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, fc)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(constant.CodeFeeComponentQueryFailed, "gagal membaca data komponen biaya", err)
	}

	return &repository.FeeComponentListResult{Items: items, Total: total}, nil
}

func (r *PostgresFeeComponentRepository) ExistsByCode(ctx context.Context, code string, excludeID string) (bool, error) {
	execer := execerFromContext(ctx, r.db)

	var exists bool
	var err error
	if excludeID == "" {
		query := `SELECT EXISTS(SELECT 1 FROM fee_components WHERE code=$1 AND deleted_at IS NULL)`
		err = execer.QueryRowContext(ctx, query, code).Scan(&exists)
	} else {
		query := `SELECT EXISTS(SELECT 1 FROM fee_components WHERE code=$1 AND deleted_at IS NULL AND id!=$2)`
		err = execer.QueryRowContext(ctx, query, code, excludeID).Scan(&exists)
	}
	if err != nil {
		return false, kernel.WrapMsg(constant.CodeFeeComponentQueryFailed, "gagal memeriksa ketersediaan kode", err)
	}
	return exists, nil
}

func (r *PostgresFeeComponentRepository) scan(sc scanner) (*entity.FeeComponent, error) {
	var (
		id, code, name, feeType, createdBy                               string
		amount                                                           float64
		isPeriodic                                                       bool
		periodType, description                                          sql.NullString
		isActive                                                         bool
		createdAt, updatedAt                                             time.Time
		deletedAt                                                        sql.NullTime
	)

	err := sc.Scan(
		&id, &code, &name, &feeType, &amount, &isPeriodic, &periodType, &description, &isActive,
		&createdBy, &createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.WrapMsg(constant.CodeFeeComponentNotFound, "Komponen biaya tidak ditemukan", nil)
		}
		return nil, kernel.WrapMsg(constant.CodeFeeComponentQueryFailed, "gagal membaca data komponen biaya", err)
	}

	var pt *constant.PeriodType
	if periodType.Valid {
		v := constant.PeriodType(periodType.String)
		pt = &v
	}

	return &entity.FeeComponent{
		ID:          id,
		Code:        code,
		Name:        name,
		Type:        constant.FeeComponentType(feeType),
		Amount:      amount,
		IsPeriodic:  isPeriodic,
		PeriodType:  pt,
		Description: strFromNull(description),
		IsActive:    isActive,
		CreatedBy:   createdBy,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		DeletedAt:   timeFromNull(deletedAt),
	}, nil
}
