package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sipon-be/internal/modules/kesantrian/domain/dokumen/constant"
	"sipon-be/internal/modules/kesantrian/domain/dokumen/entity"
	"sipon-be/internal/shared/kernel"
)

const dokumenColumns = `
	id, santri_id, kind, key, status, original_filename, mime_type, size, notes,
	verified_by, verified_at, created_at, updated_at, deleted_at
`

type PostgresSantriDokumenRepository struct {
	db *sql.DB
}

func NewPostgresSantriDokumenRepository(db *sql.DB) *PostgresSantriDokumenRepository {
	return &PostgresSantriDokumenRepository{db: db}
}

func (r *PostgresSantriDokumenRepository) Save(ctx context.Context, d *entity.SantriDokumen) error {
	execer := execerFromContext(ctx, r.db)
	query := `INSERT INTO santri_dokumen (` + dokumenColumns + `) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`
	_, err := execer.ExecContext(ctx, query,
		d.ID, d.SantriID, string(d.Kind), d.Key, string(d.Status), nullStr(d.OriginalFilename), nullStr(d.MimeType), nullInt64(d.Size), nullStr(d.Notes),
		nullStr(d.VerifiedBy), nullTimeVal(d.VerifiedAt), d.CreatedAt, d.UpdatedAt, nullTimeVal(d.DeletedAt),
	)
	if err != nil {
		return kernel.Wrap(constant.CodeDokumenPersistenceFailed, fmt.Errorf("save dokumen: %w", err))
	}
	return nil
}

func (r *PostgresSantriDokumenRepository) Update(ctx context.Context, d *entity.SantriDokumen) error {
	execer := execerFromContext(ctx, r.db)
	query := `UPDATE santri_dokumen SET
		status=$1, original_filename=$2, mime_type=$3, size=$4, notes=$5,
		verified_by=$6, verified_at=$7, updated_at=$8, deleted_at=$9
		WHERE id=$10 AND deleted_at IS NULL`
	res, err := execer.ExecContext(ctx, query,
		string(d.Status), nullStr(d.OriginalFilename), nullStr(d.MimeType), nullInt64(d.Size), nullStr(d.Notes),
		nullStr(d.VerifiedBy), nullTimeVal(d.VerifiedAt), d.UpdatedAt, nullTimeVal(d.DeletedAt),
		d.ID,
	)
	if err != nil {
		return kernel.Wrap(constant.CodeDokumenPersistenceFailed, fmt.Errorf("update dokumen: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeDokumenNotFound)
	}
	return nil
}

func (r *PostgresSantriDokumenRepository) FindByID(ctx context.Context, id string) (*entity.SantriDokumen, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+dokumenColumns+` FROM santri_dokumen WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanDokumen(row)
}

func (r *PostgresSantriDokumenRepository) FindBySantriID(ctx context.Context, santriID string) ([]*entity.SantriDokumen, error) {
	execer := execerFromContext(ctx, r.db)
	rows, err := execer.QueryContext(ctx, `SELECT `+dokumenColumns+` FROM santri_dokumen WHERE santri_id=$1 AND deleted_at IS NULL ORDER BY created_at DESC`, santriID)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeDokumenQueryFailed, fmt.Errorf("list dokumen: %w", err))
	}
	defer rows.Close()

	items := make([]*entity.SantriDokumen, 0)
	for rows.Next() {
		d, err := scanDokumen(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(constant.CodeDokumenQueryFailed, fmt.Errorf("iterate dokumen rows: %w", err))
	}
	return items, nil
}

func (r *PostgresSantriDokumenRepository) FindBySantriIDAndKind(ctx context.Context, santriID string, kind constant.DokumenKind) (*entity.SantriDokumen, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+dokumenColumns+` FROM santri_dokumen WHERE santri_id=$1 AND kind=$2 AND deleted_at IS NULL`, santriID, string(kind))
	return scanDokumen(row)
}

func scanDokumen(sc scanner) (*entity.SantriDokumen, error) {
	var (
		id, santriID, kind, key, status               string
		originalFilename, mimeType, notes, verifiedBy sql.NullString
		size                                          sql.NullInt64
		verifiedAt, deletedAt                         sql.NullTime
		createdAt, updatedAt                          time.Time
	)

	err := sc.Scan(
		&id, &santriID, &kind, &key, &status, &originalFilename, &mimeType, &size, &notes,
		&verifiedBy, &verifiedAt, &createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeDokumenNotFound)
		}
		return nil, kernel.Wrap(constant.CodeDokumenQueryFailed, fmt.Errorf("scan dokumen: %w", err))
	}

	return &entity.SantriDokumen{
		ID:               id,
		SantriID:         santriID,
		Kind:             constant.DokumenKind(kind),
		Key:              key,
		Status:           constant.DokumenStatus(status),
		OriginalFilename: strFromNull(originalFilename),
		MimeType:         strFromNull(mimeType),
		Size:             int64FromNull(size),
		Notes:            strFromNull(notes),
		VerifiedBy:       strFromNull(verifiedBy),
		VerifiedAt:       timeFromNull(verifiedAt),
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		DeletedAt:        timeFromNull(deletedAt),
	}, nil
}
