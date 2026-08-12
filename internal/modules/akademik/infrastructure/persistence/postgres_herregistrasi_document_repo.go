package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sipon-be/internal/modules/akademik/domain/herregistrasi_document/constant"
	"sipon-be/internal/modules/akademik/domain/herregistrasi_document/entity"
	"sipon-be/internal/shared/kernel"
)

const herregDocumentColumns = `
	id, santri_registration_id, kind, key, status, original_filename, mime_type, size,
	notes, verified_by, verified_at, created_at, updated_at, deleted_at
`

type PostgresHerregistrasiDocumentRepository struct {
	db *sql.DB
}

func NewPostgresHerregistrasiDocumentRepository(db *sql.DB) *PostgresHerregistrasiDocumentRepository {
	return &PostgresHerregistrasiDocumentRepository{db: db}
}

func (r *PostgresHerregistrasiDocumentRepository) Save(ctx context.Context, doc *entity.HerregistrasiDocument) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx,
		`INSERT INTO herregistrasi_documents (`+herregDocumentColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		doc.ID, doc.SantriRegistrationID, doc.Kind, doc.Key, string(doc.Status),
		nullStr(doc.OriginalFilename), nullStr(doc.MimeType), nullInt64(doc.Size),
		nullStr(doc.Notes), nullStr(doc.VerifiedBy), nullTimeVal(doc.VerifiedAt),
		doc.CreatedAt, doc.UpdatedAt, nullTimeVal(doc.DeletedAt),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeHerregistrasiDocumentDuplicate, err)
		}
		return kernel.Wrap(constant.CodeHerregistrasiDocumentPersistenceFailed, fmt.Errorf("save herreg document: %w", err))
	}
	return nil
}

func (r *PostgresHerregistrasiDocumentRepository) Update(ctx context.Context, doc *entity.HerregistrasiDocument) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx,
		`UPDATE herregistrasi_documents SET status=$1, notes=$2, verified_by=$3, verified_at=$4, updated_at=$5, deleted_at=$6 WHERE id=$7 AND deleted_at IS NULL`,
		string(doc.Status), nullStr(doc.Notes), nullStr(doc.VerifiedBy), nullTimeVal(doc.VerifiedAt),
		doc.UpdatedAt, nullTimeVal(doc.DeletedAt), doc.ID,
	)
	if err != nil {
		return kernel.Wrap(constant.CodeHerregistrasiDocumentPersistenceFailed, fmt.Errorf("update herreg document: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeHerregistrasiDocumentNotFound)
	}
	return nil
}

func (r *PostgresHerregistrasiDocumentRepository) FindByID(ctx context.Context, id string) (*entity.HerregistrasiDocument, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx,
		`SELECT `+herregDocumentColumns+` FROM herregistrasi_documents WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanHerregDocument(row)
}

func (r *PostgresHerregistrasiDocumentRepository) FindByRegistration(ctx context.Context, registrationID string) ([]*entity.HerregistrasiDocument, error) {
	execer := execerFromContext(ctx, r.db)
	rows, err := execer.QueryContext(ctx,
		`SELECT `+herregDocumentColumns+` FROM herregistrasi_documents
		 WHERE santri_registration_id=$1 AND deleted_at IS NULL ORDER BY created_at ASC`, registrationID)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeHerregistrasiDocumentQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*entity.HerregistrasiDocument, 0)
	for rows.Next() {
		doc, err := scanHerregDocument(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, doc)
	}
	return items, rows.Err()
}

func (r *PostgresHerregistrasiDocumentRepository) FindByRegistrationAndKind(ctx context.Context, registrationID, kind string) (*entity.HerregistrasiDocument, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx,
		`SELECT `+herregDocumentColumns+` FROM herregistrasi_documents WHERE santri_registration_id=$1 AND kind=$2 AND deleted_at IS NULL`,
		registrationID, kind)
	return scanHerregDocument(row)
}

func (r *PostgresHerregistrasiDocumentRepository) Delete(ctx context.Context, id string) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx,
		`UPDATE herregistrasi_documents SET deleted_at=$1, updated_at=$2 WHERE id=$3 AND deleted_at IS NULL`,
		time.Now(), time.Now(), id)
	if err != nil {
		return kernel.Wrap(constant.CodeHerregistrasiDocumentPersistenceFailed, fmt.Errorf("delete herreg document: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeHerregistrasiDocumentNotFound)
	}
	return nil
}

func scanHerregDocument(sc scanner) (*entity.HerregistrasiDocument, error) {
	var (
		id, registrationID, kind, key, status string
		originalFilename, mimeType            sql.NullString
		size                                  sql.NullInt64
		notes, verifiedBy                     sql.NullString
		verifiedAt                            sql.NullTime
		createdAt, updatedAt                  time.Time
		deletedAt                             sql.NullTime
	)
	err := sc.Scan(&id, &registrationID, &kind, &key, &status,
		&originalFilename, &mimeType, &size,
		&notes, &verifiedBy, &verifiedAt, &createdAt, &updatedAt, &deletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeHerregistrasiDocumentNotFound)
		}
		return nil, kernel.Wrap(constant.CodeHerregistrasiDocumentQueryFailed, fmt.Errorf("scan herreg document: %w", err))
	}
	return &entity.HerregistrasiDocument{
		ID:                   id,
		SantriRegistrationID: registrationID,
		Kind:                 kind,
		Key:                  key,
		Status:               constant.HerregistrasiDocumentStatus(status),
		OriginalFilename:     strFromNull(originalFilename),
		MimeType:             strFromNull(mimeType),
		Size:                 int64FromNull(size),
		Notes:                strFromNull(notes),
		VerifiedBy:           strFromNull(verifiedBy),
		VerifiedAt:           timeFromNull(verifiedAt),
		CreatedAt:            createdAt,
		UpdatedAt:            updatedAt,
		DeletedAt:            timeFromNull(deletedAt),
	}, nil
}
