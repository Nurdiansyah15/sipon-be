package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	tipeconstant "sipon-be/internal/modules/kesantrian/domain/tipe_surat/constant"
	"sipon-be/internal/modules/kesantrian/domain/tipe_surat/entity"
	"sipon-be/internal/modules/kesantrian/domain/tipe_surat/repository"
	"sipon-be/internal/shared/kernel"
)

const tipeSuratColumns = `id, nama, kode, created_by, created_at, updated_at`

var tipeSuratSortColumns = map[string]string{
	"nama":       "nama",
	"kode":       "kode",
	"created_at": "created_at",
}

type PostgresTipeSuratRepository struct {
	db *sql.DB
}

func NewPostgresTipeSuratRepository(db *sql.DB) *PostgresTipeSuratRepository {
	return &PostgresTipeSuratRepository{db: db}
}

func (r *PostgresTipeSuratRepository) Save(ctx context.Context, ts *entity.TipeSurat) error {
	execer := execerFromContext(ctx, r.db)
	query := `INSERT INTO master_tipe_surat (` + tipeSuratColumns + `) VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := execer.ExecContext(ctx, query,
		ts.ID, ts.Nama, ts.Kode, nullStr(ts.CreatedBy), ts.CreatedAt, ts.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(tipeconstant.CodeTipeSuratKodeDuplicate, err)
		}
		return kernel.Wrap(tipeconstant.CodeTipeSuratPersistenceFailed, fmt.Errorf("save tipe_surat: %w", err))
	}
	return nil
}

func (r *PostgresTipeSuratRepository) Update(ctx context.Context, ts *entity.TipeSurat) error {
	execer := execerFromContext(ctx, r.db)
	query := `UPDATE master_tipe_surat SET nama=$1, kode=$2, updated_at=$3 WHERE id=$4`
	res, err := execer.ExecContext(ctx, query, ts.Nama, ts.Kode, ts.UpdatedAt, ts.ID)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(tipeconstant.CodeTipeSuratKodeDuplicate, err)
		}
		return kernel.Wrap(tipeconstant.CodeTipeSuratPersistenceFailed, fmt.Errorf("update tipe_surat: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(tipeconstant.CodeTipeSuratNotFound)
	}
	return nil
}

func (r *PostgresTipeSuratRepository) Delete(ctx context.Context, id string) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx, `DELETE FROM master_tipe_surat WHERE id=$1`, id)
	if err != nil {
		return kernel.Wrap(tipeconstant.CodeTipeSuratPersistenceFailed, fmt.Errorf("delete tipe_surat: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(tipeconstant.CodeTipeSuratNotFound)
	}
	return nil
}

func (r *PostgresTipeSuratRepository) FindByID(ctx context.Context, id string) (*entity.TipeSurat, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+tipeSuratColumns+` FROM master_tipe_surat WHERE id=$1`, id)
	return scanTipeSurat(row)
}

func (r *PostgresTipeSuratRepository) List(ctx context.Context, q repository.TipeSuratListQuery) (*repository.TipeSuratListResult, error) {
	execer := execerFromContext(ctx, r.db)
	where := `WHERE 1=1`
	args := []interface{}{}

	var total int64
	if err := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM master_tipe_surat `+where, args...).Scan(&total); err != nil {
		return nil, kernel.Wrap(tipeconstant.CodeTipeSuratQueryFailed, fmt.Errorf("count tipe_surat: %w", err))
	}

	sortCol, ok := tipeSuratSortColumns[q.SortBy]
	if !ok {
		sortCol = "created_at"
	}
	sortDir := "DESC"
	if q.SortType == "asc" {
		sortDir = "ASC"
	}

	listArgs := append(args, q.Limit, (q.Page-1)*q.Limit)
	query := fmt.Sprintf(`SELECT %s FROM master_tipe_surat %s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		tipeSuratColumns, where, sortCol, sortDir, len(args)+1, len(args)+2)

	rows, err := execer.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, kernel.Wrap(tipeconstant.CodeTipeSuratQueryFailed, fmt.Errorf("list tipe_surat: %w", err))
	}
	defer rows.Close()

	items := make([]*entity.TipeSurat, 0)
	for rows.Next() {
		item, err := scanTipeSurat(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(tipeconstant.CodeTipeSuratQueryFailed, fmt.Errorf("iterate tipe_surat rows: %w", err))
	}

	return &repository.TipeSuratListResult{Items: items, Total: total}, nil
}

func (r *PostgresTipeSuratRepository) IsInUse(ctx context.Context, tipeSuratID string) (bool, error) {
	execer := execerFromContext(ctx, r.db)
	var count int
	err := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM surat WHERE tipe_surat_id=$1`, tipeSuratID).Scan(&count)
	if err != nil {
		return false, kernel.Wrap(tipeconstant.CodeTipeSuratQueryFailed, fmt.Errorf("count surat for tipe: %w", err))
	}
	return count > 0, nil
}

func scanTipeSurat(sc scanner) (*entity.TipeSurat, error) {
	var (
		id, nama, kode       string
		createdBy            sql.NullString
		createdAt, updatedAt time.Time
	)
	err := sc.Scan(&id, &nama, &kode, &createdBy, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(tipeconstant.CodeTipeSuratNotFound)
		}
		return nil, kernel.Wrap(tipeconstant.CodeTipeSuratQueryFailed, fmt.Errorf("scan tipe_surat: %w", err))
	}
	return &entity.TipeSurat{
		ID:        id,
		Nama:      nama,
		Kode:      kode,
		CreatedBy: strFromNull(createdBy),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
