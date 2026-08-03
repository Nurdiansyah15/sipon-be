package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	rconstant "sipon-be/internal/modules/psb/domain/review/constant"
	rentity "sipon-be/internal/modules/psb/domain/review/entity"
	"sipon-be/internal/shared/kernel"
)

const reviewColumns = `
	id, pendaftar_id, stage, action, notes, reviewed_by, created_at
`

type PostgresReviewRepository struct {
	db *sql.DB
}

func NewPostgresReviewRepository(db *sql.DB) *PostgresReviewRepository {
	return &PostgresReviewRepository{db: db}
}

func (r *PostgresReviewRepository) Save(ctx context.Context, rev *rentity.PendaftarReview) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx,
		`INSERT INTO pendaftar_reviews (`+reviewColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		rev.ID, rev.PendaftarID, string(rev.Stage), string(rev.Action), nullStr(rev.Notes), rev.ReviewedBy, rev.CreatedAt,
	)
	if err != nil {
		return kernel.Wrap(rconstant.ErrCodeInvalidReview, fmt.Errorf("save review: %w", err))
	}
	return nil
}

func (r *PostgresReviewRepository) FindByPendaftarID(ctx context.Context, pendaftarID string) ([]*rentity.PendaftarReview, error) {
	execer := execerFromContext(ctx, r.db)
	rows, err := execer.QueryContext(ctx, `SELECT `+reviewColumns+` FROM pendaftar_reviews WHERE pendaftar_id=$1 ORDER BY created_at ASC`, pendaftarID)
	if err != nil {
		return nil, kernel.Wrap(rconstant.ErrCodeInvalidReview, fmt.Errorf("list reviews: %w", err))
	}
	defer rows.Close()

	var items []*rentity.PendaftarReview
	for rows.Next() {
		rev, err := scanReview(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, rev)
	}
	return items, rows.Err()
}

func (r *PostgresReviewRepository) HardDeleteByPendaftarID(ctx context.Context, pendaftarID string) (int64, error) {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx, `DELETE FROM pendaftar_reviews WHERE pendaftar_id=$1`, pendaftarID)
	if err != nil {
		return 0, kernel.Wrap(rconstant.ErrCodeInvalidReview, fmt.Errorf("hard delete reviews: %w", err))
	}
	return res.RowsAffected()
}

func scanReview(sc scanner) (*rentity.PendaftarReview, error) {
	var (
		id, pendaftarID, stage, action, reviewedBy string
		notes                                      sql.NullString
		createdAt                                  time.Time
	)
	err := sc.Scan(&id, &pendaftarID, &stage, &action, &notes, &reviewedBy, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(rconstant.ErrCodeInvalidReview)
		}
		return nil, kernel.Wrap(rconstant.ErrCodeInvalidReview, fmt.Errorf("scan review: %w", err))
	}

	return &rentity.PendaftarReview{
		ID:          id,
		PendaftarID: pendaftarID,
		Stage:       rconstant.ReviewStage(stage),
		Action:      rconstant.ReviewAction(action),
		Notes:       strFromNull(notes),
		ReviewedBy:  reviewedBy,
		CreatedAt:   createdAt,
	}, nil
}
