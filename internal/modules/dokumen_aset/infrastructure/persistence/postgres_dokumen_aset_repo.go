package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	constant "sipon-be/internal/modules/dokumen_aset/domain/dokumen/constant"
	entity "sipon-be/internal/modules/dokumen_aset/domain/dokumen/entity"
	repo "sipon-be/internal/modules/dokumen_aset/domain/dokumen/repository"
	"sipon-be/internal/shared/kernel"
)

const dokumenAsetColumns = `
	id, judul, deskripsi, kategori, key, filename, mime_type, size,
	is_public, created_by, created_at, updated_at, deleted_at
`

type PostgresDokumenAsetRepository struct {
	db *sql.DB
}

func NewPostgresDokumenAsetRepository(db *sql.DB) *PostgresDokumenAsetRepository {
	return &PostgresDokumenAsetRepository{db: db}
}

func (r *PostgresDokumenAsetRepository) Save(ctx context.Context, d *entity.DokumenAset) error {
	execer := execerFromContext(ctx, r.db)
	query := `INSERT INTO dokumen_aset (` + dokumenAsetColumns + `) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`
	_, err := execer.ExecContext(ctx, query,
		d.ID, d.Judul, nullStr(d.Deskripsi), string(d.Kategori), d.Key, d.Filename, d.MimeType, d.Size,
		d.IsPublic, d.CreatedBy, d.CreatedAt, d.UpdatedAt, nullTimeVal(d.DeletedAt),
	)
	if err != nil {
		return kernel.Wrap(constant.CodeDokumenPersistenceFailed, fmt.Errorf("save dokumen aset: %w", err))
	}
	return nil
}

func (r *PostgresDokumenAsetRepository) Update(ctx context.Context, d *entity.DokumenAset) error {
	execer := execerFromContext(ctx, r.db)
	query := `UPDATE dokumen_aset SET
		judul=$1, deskripsi=$2, kategori=$3, is_public=$4,
		updated_at=$5, deleted_at=$6
		WHERE id=$7 AND deleted_at IS NULL`
	res, err := execer.ExecContext(ctx, query,
		d.Judul, nullStr(d.Deskripsi), string(d.Kategori), d.IsPublic,
		d.UpdatedAt, nullTimeVal(d.DeletedAt), d.ID,
	)
	if err != nil {
		return kernel.Wrap(constant.CodeDokumenPersistenceFailed, fmt.Errorf("update dokumen aset: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeDokumenNotFound)
	}
	return nil
}

func (r *PostgresDokumenAsetRepository) FindByID(ctx context.Context, id string) (*entity.DokumenAset, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+dokumenAsetColumns+` FROM dokumen_aset WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanDokumenAset(row)
}

func (r *PostgresDokumenAsetRepository) List(ctx context.Context, filter repo.DokumenAsetFilter) ([]*entity.DokumenAset, int, error) {
	execer := execerFromContext(ctx, r.db)

	where := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	argIdx := 1

	if filter.PublicOnly {
		where = append(where, fmt.Sprintf("is_public = true"))
	}

	if filter.Kategori != nil {
		where = append(where, fmt.Sprintf("kategori = $%d", argIdx))
		args = append(args, string(*filter.Kategori))
		argIdx++
	}

	if filter.Search != "" {
		where = append(where, fmt.Sprintf("(judul ILIKE $%d OR deskripsi ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := `SELECT COUNT(*) FROM dokumen_aset WHERE ` + whereClause
	if err := execer.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, kernel.Wrap(constant.CodeDokumenQueryFailed, fmt.Errorf("count dokumen aset: %w", err))
	}

	offset := (filter.Page - 1) * filter.Limit
	dataQuery := fmt.Sprintf(`SELECT %s FROM dokumen_aset WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		dokumenAsetColumns, whereClause, argIdx, argIdx+1)
	args = append(args, filter.Limit, offset)

	rows, err := execer.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, kernel.Wrap(constant.CodeDokumenQueryFailed, fmt.Errorf("list dokumen aset: %w", err))
	}
	defer rows.Close()

	items := make([]*entity.DokumenAset, 0)
	for rows.Next() {
		d, err := scanDokumenAset(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, kernel.Wrap(constant.CodeDokumenQueryFailed, fmt.Errorf("iterate dokumen aset rows: %w", err))
	}

	return items, total, nil
}

func scanDokumenAset(sc scanner) (*entity.DokumenAset, error) {
	var (
		id, judul, kategori, key, filename, mimeType, createdBy string
		deskripsi                                                sql.NullString
		size                                                     int64
		isPublic                                                 bool
		deletedAt                                                sql.NullTime
		createdAt, updatedAt                                     time.Time
	)

	err := sc.Scan(
		&id, &judul, &deskripsi, &kategori, &key, &filename, &mimeType, &size,
		&isPublic, &createdBy, &createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeDokumenNotFound)
		}
		return nil, kernel.Wrap(constant.CodeDokumenQueryFailed, fmt.Errorf("scan dokumen aset: %w", err))
	}

	return &entity.DokumenAset{
		ID:        id,
		Judul:     judul,
		Deskripsi: strFromNull(deskripsi),
		Kategori:  constant.Kategori(kategori),
		Key:       key,
		Filename:  filename,
		MimeType:  mimeType,
		Size:      size,
		IsPublic:  isPublic,
		CreatedBy: createdBy,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		DeletedAt: timeFromNull(deletedAt),
	}, nil
}
