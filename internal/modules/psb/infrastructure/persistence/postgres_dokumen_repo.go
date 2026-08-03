package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	dconstant "sipon-be/internal/modules/psb/domain/dokumen/constant"
	dentity "sipon-be/internal/modules/psb/domain/dokumen/entity"
	"sipon-be/internal/shared/kernel"
)

const dokumenColumns = `
	id, pendaftar_id, stage, kind, key, status,
	original_filename, mime_type, size, notes,
	verified_by, verified_at,
	created_at, updated_at, deleted_at
`

type PostgresDokumenRepository struct {
	db *sql.DB
}

func NewPostgresDokumenRepository(db *sql.DB) *PostgresDokumenRepository {
	return &PostgresDokumenRepository{db: db}
}

func (r *PostgresDokumenRepository) Save(ctx context.Context, d *dentity.PendaftarDokumen) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx,
		`INSERT INTO pendaftar_dokumen (`+dokumenColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		d.ID, d.PendaftarID, string(d.Stage), string(d.Kind), d.Key, string(d.Status),
		nullStr(d.OriginalFilename), nullStr(d.MimeType), nullInt64(d.Size), nullStr(d.Notes),
		nullStr(d.VerifiedBy), nullTimeVal(d.VerifiedAt),
		d.CreatedAt, d.UpdatedAt, nullTimeVal(d.DeletedAt),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.New(dconstant.CodeDokumenPersistenceFailed)
		}
		return kernel.Wrap(dconstant.CodeDokumenPersistenceFailed, fmt.Errorf("save dokumen: %w", err))
	}
	return nil
}

func (r *PostgresDokumenRepository) Update(ctx context.Context, d *dentity.PendaftarDokumen) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx,
		`UPDATE pendaftar_dokumen SET stage=$1, kind=$2, key=$3, status=$4,
		 original_filename=$5, mime_type=$6, size=$7, notes=$8,
		 verified_by=$9, verified_at=$10,
		 updated_at=$11, deleted_at=$12 WHERE id=$13 AND deleted_at IS NULL`,
		string(d.Stage), string(d.Kind), d.Key, string(d.Status),
		nullStr(d.OriginalFilename), nullStr(d.MimeType), nullInt64(d.Size), nullStr(d.Notes),
		nullStr(d.VerifiedBy), nullTimeVal(d.VerifiedAt),
		d.UpdatedAt, nullTimeVal(d.DeletedAt), d.ID,
	)
	if err != nil {
		return kernel.Wrap(dconstant.CodeDokumenPersistenceFailed, fmt.Errorf("update dokumen: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(dconstant.CodeDokumenNotFound)
	}
	return nil
}

func (r *PostgresDokumenRepository) FindByID(ctx context.Context, id string) (*dentity.PendaftarDokumen, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+dokumenColumns+` FROM pendaftar_dokumen WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanDokumen(row)
}

func (r *PostgresDokumenRepository) FindByPendaftarID(ctx context.Context, pendaftarID string) ([]*dentity.PendaftarDokumen, error) {
	execer := execerFromContext(ctx, r.db)
	rows, err := execer.QueryContext(ctx, `SELECT `+dokumenColumns+` FROM pendaftar_dokumen WHERE pendaftar_id=$1 AND deleted_at IS NULL`, pendaftarID)
	if err != nil {
		return nil, kernel.Wrap(dconstant.CodeDokumenQueryFailed, fmt.Errorf("list dokumen: %w", err))
	}
	defer rows.Close()

	return scanDokumenRows(rows)
}

func (r *PostgresDokumenRepository) FindByPendaftarIDAndStage(ctx context.Context, pendaftarID string, stage dconstant.DokumenStage) ([]*dentity.PendaftarDokumen, error) {
	execer := execerFromContext(ctx, r.db)
	rows, err := execer.QueryContext(ctx, `SELECT `+dokumenColumns+` FROM pendaftar_dokumen WHERE pendaftar_id=$1 AND stage=$2 AND deleted_at IS NULL`, pendaftarID, string(stage))
	if err != nil {
		return nil, kernel.Wrap(dconstant.CodeDokumenQueryFailed, fmt.Errorf("list dokumen by stage: %w", err))
	}
	defer rows.Close()

	return scanDokumenRows(rows)
}

func (r *PostgresDokumenRepository) HardDeleteByPendaftarID(ctx context.Context, pendaftarID string) (int64, error) {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx, `DELETE FROM pendaftar_dokumen WHERE pendaftar_id=$1`, pendaftarID)
	if err != nil {
		return 0, kernel.Wrap(dconstant.CodeDokumenPersistenceFailed, fmt.Errorf("hard delete dokumen: %w", err))
	}
	return res.RowsAffected()
}

func scanDokumenRows(rows *sql.Rows) ([]*dentity.PendaftarDokumen, error) {
	var items []*dentity.PendaftarDokumen
	for rows.Next() {
		d, err := scanDokumen(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

func scanDokumen(sc scanner) (*dentity.PendaftarDokumen, error) {
	var (
		id, pendaftarID, stage, kind, key, status         string
		originalFilename, mimeType, notes, verifiedBy     sql.NullString
		size                                              sql.NullInt64
		verifiedAt                                        sql.NullTime
		createdAt, updatedAt                              time.Time
		deletedAt                                         sql.NullTime
	)
	err := sc.Scan(&id, &pendaftarID, &stage, &kind, &key, &status,
		&originalFilename, &mimeType, &size, &notes,
		&verifiedBy, &verifiedAt,
		&createdAt, &updatedAt, &deletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(dconstant.CodeDokumenNotFound)
		}
		return nil, kernel.Wrap(dconstant.CodeDokumenQueryFailed, fmt.Errorf("scan dokumen: %w", err))
	}

	return &dentity.PendaftarDokumen{
		ID:               id,
		PendaftarID:      pendaftarID,
		Stage:            dconstant.DokumenStage(stage),
		Kind:             dconstant.DokumenKind(kind),
		Key:              key,
		Status:           dconstant.DokumenStatus(status),
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
