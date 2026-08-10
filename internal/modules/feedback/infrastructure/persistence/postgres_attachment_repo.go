package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	acconstant "sipon-be/internal/modules/feedback/domain/attachment/constant"
	aentity "sipon-be/internal/modules/feedback/domain/attachment/entity"
	"sipon-be/internal/shared/kernel"
)

const attachmentColumns = `
	id, feedback_id, key, original_filename, mime_type, size, sort_order,
	created_at, updated_at, deleted_at
`

type PostgresAttachmentRepository struct {
	db *sql.DB
}

func NewPostgresAttachmentRepository(db *sql.DB) *PostgresAttachmentRepository {
	return &PostgresAttachmentRepository{db: db}
}

func (r *PostgresAttachmentRepository) Save(ctx context.Context, a *aentity.Attachment) error {
	execer := execerFromContext(ctx, r.db)
	query := `INSERT INTO feedback_attachments (` + attachmentColumns + `) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := execer.ExecContext(ctx, query,
		a.ID, a.FeedbackID, a.Key, nullStr(a.OriginalFilename), nullStr(a.MimeType), nullInt64(a.Size), a.SortOrder,
		a.CreatedAt, a.UpdatedAt, nullTimeVal(a.DeletedAt),
	)
	if err != nil {
		return kernel.Wrap(acconstant.CodeAttachmentPersistenceFailed, fmt.Errorf("save attachment: %w", err))
	}
	return nil
}

func (r *PostgresAttachmentRepository) FindByID(ctx context.Context, id string) (*aentity.Attachment, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+attachmentColumns+` FROM feedback_attachments WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanAttachment(row)
}

func (r *PostgresAttachmentRepository) ListByFeedbackID(ctx context.Context, feedbackID string) ([]*aentity.Attachment, error) {
	execer := execerFromContext(ctx, r.db)
	rows, err := execer.QueryContext(ctx, `SELECT `+attachmentColumns+` FROM feedback_attachments WHERE feedback_id=$1 AND deleted_at IS NULL ORDER BY sort_order ASC`, feedbackID)
	if err != nil {
		return nil, kernel.Wrap(acconstant.CodeAttachmentQueryFailed, fmt.Errorf("list attachments: %w", err))
	}
	defer rows.Close()

	items := make([]*aentity.Attachment, 0)
	for rows.Next() {
		a, err := scanAttachment(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(acconstant.CodeAttachmentQueryFailed, fmt.Errorf("iterate attachment rows: %w", err))
	}
	return items, nil
}

func (r *PostgresAttachmentRepository) CountByFeedbackID(ctx context.Context, feedbackID string) (int64, error) {
	execer := execerFromContext(ctx, r.db)
	var count int64
	err := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM feedback_attachments WHERE feedback_id=$1 AND deleted_at IS NULL`, feedbackID).Scan(&count)
	if err != nil {
		return 0, kernel.Wrap(acconstant.CodeAttachmentQueryFailed, fmt.Errorf("count attachments: %w", err))
	}
	return count, nil
}

func (r *PostgresAttachmentRepository) CountByFeedbackIDs(ctx context.Context, feedbackIDs []string) (map[string]int64, error) {
	execer := execerFromContext(ctx, r.db)
	if len(feedbackIDs) == 0 {
		return map[string]int64{}, nil
	}

	result := make(map[string]int64, len(feedbackIDs))
	rows, err := execer.QueryContext(ctx, `SELECT feedback_id, COUNT(*) FROM feedback_attachments WHERE feedback_id = ANY($1) AND deleted_at IS NULL GROUP BY feedback_id`, toUUIDSlice(feedbackIDs))
	if err != nil {
		return nil, kernel.Wrap(acconstant.CodeAttachmentQueryFailed, fmt.Errorf("count attachments by feedback ids: %w", err))
	}
	defer rows.Close()

	for rows.Next() {
		var feedbackID string
		var count int64
		if err := rows.Scan(&feedbackID, &count); err != nil {
			return nil, kernel.Wrap(acconstant.CodeAttachmentQueryFailed, fmt.Errorf("scan attachment count: %w", err))
		}
		result[feedbackID] = count
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(acconstant.CodeAttachmentQueryFailed, fmt.Errorf("iterate attachment count rows: %w", err))
	}

	return result, nil
}

func (r *PostgresAttachmentRepository) SoftDelete(ctx context.Context, id string) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx, `UPDATE feedback_attachments SET deleted_at = NOW(), updated_at = NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return kernel.Wrap(acconstant.CodeAttachmentPersistenceFailed, fmt.Errorf("soft delete attachment: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(acconstant.CodeAttachmentNotFound)
	}
	return nil
}

func (r *PostgresAttachmentRepository) MaxSortOrder(ctx context.Context, feedbackID string) (int, error) {
	execer := execerFromContext(ctx, r.db)
	var max int
	err := execer.QueryRowContext(ctx, `SELECT COALESCE(MAX(sort_order), 0) FROM feedback_attachments WHERE feedback_id=$1 AND deleted_at IS NULL`, feedbackID).Scan(&max)
	if err != nil {
		return 0, kernel.Wrap(acconstant.CodeAttachmentQueryFailed, fmt.Errorf("max sort order attachments: %w", err))
	}
	return max, nil
}

func scanAttachment(sc scanner) (*aentity.Attachment, error) {
	var (
		id, feedbackID, key        string
		originalFilename, mimeType sql.NullString
		size                       sql.NullInt64
		sortOrder                  int
		createdAt, updatedAt       time.Time
		deletedAt                  sql.NullTime
	)

	err := sc.Scan(
		&id, &feedbackID, &key, &originalFilename, &mimeType, &size, &sortOrder,
		&createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(acconstant.CodeAttachmentNotFound)
		}
		return nil, kernel.Wrap(acconstant.CodeAttachmentQueryFailed, fmt.Errorf("scan attachment: %w", err))
	}

	return &aentity.Attachment{
		ID:               id,
		FeedbackID:       feedbackID,
		Key:              key,
		OriginalFilename: strFromNull(originalFilename),
		MimeType:         strFromNull(mimeType),
		Size:             int64FromNull(size),
		SortOrder:        sortOrder,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		DeletedAt:        timeFromNull(deletedAt),
	}, nil
}
