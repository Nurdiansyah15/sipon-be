package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	fconstant "sipon-be/internal/modules/feedback/domain/feedback/constant"
	fentity "sipon-be/internal/modules/feedback/domain/feedback/entity"
	frepo "sipon-be/internal/modules/feedback/domain/feedback/repository"
	"sipon-be/internal/shared/kernel"
)

const feedbackColumns = `
	id, user_id, title, body, category, is_takedown, takedown_reason, takedown_by, takedown_at,
	like_count, comment_count, created_at, updated_at, deleted_at
`

type PostgresFeedbackRepository struct {
	db *sql.DB
}

func NewPostgresFeedbackRepository(db *sql.DB) *PostgresFeedbackRepository {
	return &PostgresFeedbackRepository{db: db}
}

func (r *PostgresFeedbackRepository) Save(ctx context.Context, f *fentity.Feedback) error {
	execer := execerFromContext(ctx, r.db)
	query := `INSERT INTO feedbacks (` + feedbackColumns + `) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`
	_, err := execer.ExecContext(ctx, query,
		f.ID, f.UserID, f.Title, f.Body, string(f.Category), f.IsTakedown,
		nullStr(f.TakedownReason), nullStr(f.TakedownBy), nullTimeVal(f.TakedownAt),
		f.LikeCount, f.CommentCount, f.CreatedAt, f.UpdatedAt, nullTimeVal(f.DeletedAt),
	)
	if err != nil {
		return kernel.Wrap(fconstant.CodeFeedbackPersistenceFailed, fmt.Errorf("save feedback: %w", err))
	}
	return nil
}

func (r *PostgresFeedbackRepository) Update(ctx context.Context, f *fentity.Feedback) error {
	execer := execerFromContext(ctx, r.db)
	query := `UPDATE feedbacks SET
		title=$1, body=$2, category=$3, is_takedown=$4, takedown_reason=$5, takedown_by=$6, takedown_at=$7,
		like_count=$8, comment_count=$9, updated_at=$10, deleted_at=$11
		WHERE id=$12 AND deleted_at IS NULL`
	res, err := execer.ExecContext(ctx, query,
		f.Title, f.Body, string(f.Category), f.IsTakedown,
		nullStr(f.TakedownReason), nullStr(f.TakedownBy), nullTimeVal(f.TakedownAt),
		f.LikeCount, f.CommentCount, f.UpdatedAt, nullTimeVal(f.DeletedAt),
		f.ID,
	)
	if err != nil {
		return kernel.Wrap(fconstant.CodeFeedbackPersistenceFailed, fmt.Errorf("update feedback: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(fconstant.CodeFeedbackNotFound)
	}
	return nil
}

func (r *PostgresFeedbackRepository) FindByID(ctx context.Context, id string) (*fentity.Feedback, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+feedbackColumns+` FROM feedbacks WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanFeedback(row)
}

func (r *PostgresFeedbackRepository) List(ctx context.Context, q frepo.FeedbackListQuery) (*frepo.FeedbackListResult, error) {
	execer := execerFromContext(ctx, r.db)

	where := `WHERE deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 1
	if q.UserID != "" {
		where += fmt.Sprintf(` AND user_id=$%d`, argIdx)
		args = append(args, q.UserID)
		argIdx++
	}
	if !q.IncludeTakedown {
		where += ` AND is_takedown = false`
	}
	if q.Category != nil && *q.Category != "" {
		where += fmt.Sprintf(` AND category=$%d`, argIdx)
		args = append(args, *q.Category)
		argIdx++
	}
	if q.Search != "" {
		like := "%" + strings.ToLower(q.Search) + "%"
		where += fmt.Sprintf(` AND (LOWER(title) LIKE $%d OR LOWER(body) LIKE $%d)`, argIdx, argIdx+1)
		args = append(args, like, like)
		argIdx += 2
	}

	var total int64
	countRow := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM feedbacks `+where, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, kernel.Wrap(fconstant.CodeFeedbackQueryFailed, fmt.Errorf("count feedbacks: %w", err))
	}

	limit := q.Limit
	if limit < 1 {
		limit = 10
	}
	offset := (q.Page - 1) * q.Limit
	query := fmt.Sprintf(`SELECT %s FROM feedbacks %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		feedbackColumns, where, argIdx, argIdx+1)
	listArgs := append(append([]interface{}{}, args...), limit, offset)

	rows, err := execer.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, kernel.Wrap(fconstant.CodeFeedbackQueryFailed, fmt.Errorf("list feedbacks: %w", err))
	}
	defer rows.Close()

	items := make([]*fentity.Feedback, 0)
	for rows.Next() {
		f, err := scanFeedback(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, f)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(fconstant.CodeFeedbackQueryFailed, fmt.Errorf("iterate feedback rows: %w", err))
	}

	return &frepo.FeedbackListResult{Items: items, Total: total}, nil
}

func (r *PostgresFeedbackRepository) IncrementLikeCount(ctx context.Context, id string) error {
	return r.updateCount(ctx, id, "like_count", +1)
}

func (r *PostgresFeedbackRepository) DecrementLikeCount(ctx context.Context, id string) error {
	return r.updateCount(ctx, id, "like_count", -1)
}

func (r *PostgresFeedbackRepository) IncrementCommentCount(ctx context.Context, id string) error {
	return r.updateCount(ctx, id, "comment_count", +1)
}

func (r *PostgresFeedbackRepository) DecrementCommentCount(ctx context.Context, id string) error {
	return r.updateCount(ctx, id, "comment_count", -1)
}

func (r *PostgresFeedbackRepository) updateCount(ctx context.Context, id, column string, delta int) error {
	execer := execerFromContext(ctx, r.db)
	query := fmt.Sprintf(`UPDATE feedbacks SET %s = GREATEST(%s + $1, 0), updated_at = NOW() WHERE id=$2 AND deleted_at IS NULL`, column, column)
	res, err := execer.ExecContext(ctx, query, delta, id)
	if err != nil {
		return kernel.Wrap(fconstant.CodeFeedbackPersistenceFailed, fmt.Errorf("update feedback %s: %w", column, err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(fconstant.CodeFeedbackNotFound)
	}
	return nil
}

func scanFeedback(sc scanner) (*fentity.Feedback, error) {
	var (
		id, userID, title, body, category string
		isTakedown                        bool
		takedownReason, takedownBy        sql.NullString
		takedownAt                        sql.NullTime
		likeCount, commentCount           int
		createdAt, updatedAt              time.Time
		deletedAt                         sql.NullTime
	)

	err := sc.Scan(
		&id, &userID, &title, &body, &category, &isTakedown,
		&takedownReason, &takedownBy, &takedownAt,
		&likeCount, &commentCount, &createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(fconstant.CodeFeedbackNotFound)
		}
		return nil, kernel.Wrap(fconstant.CodeFeedbackQueryFailed, fmt.Errorf("scan feedback: %w", err))
	}

	return &fentity.Feedback{
		ID:             id,
		UserID:         userID,
		Title:          title,
		Body:           body,
		Category:       fconstant.FeedbackCategory(category),
		IsTakedown:     isTakedown,
		TakedownReason: strFromNull(takedownReason),
		TakedownBy:     strFromNull(takedownBy),
		TakedownAt:     timeFromNull(takedownAt),
		LikeCount:      likeCount,
		CommentCount:   commentCount,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		DeletedAt:      timeFromNull(deletedAt),
	}, nil
}
