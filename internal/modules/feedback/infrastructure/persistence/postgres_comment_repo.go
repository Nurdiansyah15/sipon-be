package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	cconstant "sipon-be/internal/modules/feedback/domain/comment/constant"
	centity "sipon-be/internal/modules/feedback/domain/comment/entity"
	crepo "sipon-be/internal/modules/feedback/domain/comment/repository"
	"sipon-be/internal/shared/kernel"
)

const commentColumns = `
	id, feedback_id, user_id, body, reply_to_id, is_takedown, takedown_reason, takedown_by, takedown_at,
	like_count, created_at, updated_at, deleted_at
`

type PostgresCommentRepository struct {
	db *sql.DB
}

func NewPostgresCommentRepository(db *sql.DB) *PostgresCommentRepository {
	return &PostgresCommentRepository{db: db}
}

func (r *PostgresCommentRepository) Save(ctx context.Context, c *centity.Comment) error {
	execer := execerFromContext(ctx, r.db)
	query := `INSERT INTO feedback_comments (` + commentColumns + `) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`
	_, err := execer.ExecContext(ctx, query,
		c.ID, c.FeedbackID, c.UserID, c.Body, nullStr(c.ReplyToID), c.IsTakedown,
		nullStr(c.TakedownReason), nullStr(c.TakedownBy), nullTimeVal(c.TakedownAt),
		c.LikeCount, c.CreatedAt, c.UpdatedAt, nullTimeVal(c.DeletedAt),
	)
	if err != nil {
		return kernel.Wrap(cconstant.CodeCommentPersistenceFailed, fmt.Errorf("save comment: %w", err))
	}
	return nil
}

func (r *PostgresCommentRepository) Update(ctx context.Context, c *centity.Comment) error {
	execer := execerFromContext(ctx, r.db)
	query := `UPDATE feedback_comments SET
		body=$1, reply_to_id=$2, is_takedown=$3, takedown_reason=$4, takedown_by=$5, takedown_at=$6,
		like_count=$7, updated_at=$8, deleted_at=$9
		WHERE id=$10 AND deleted_at IS NULL`
	res, err := execer.ExecContext(ctx, query,
		c.Body, nullStr(c.ReplyToID), c.IsTakedown,
		nullStr(c.TakedownReason), nullStr(c.TakedownBy), nullTimeVal(c.TakedownAt),
		c.LikeCount, c.UpdatedAt, nullTimeVal(c.DeletedAt),
		c.ID,
	)
	if err != nil {
		return kernel.Wrap(cconstant.CodeCommentPersistenceFailed, fmt.Errorf("update comment: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(cconstant.CodeCommentNotFound)
	}
	return nil
}

func (r *PostgresCommentRepository) FindByID(ctx context.Context, id string) (*centity.Comment, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+commentColumns+` FROM feedback_comments WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanComment(row)
}

func (r *PostgresCommentRepository) List(ctx context.Context, q crepo.CommentListQuery) (*crepo.CommentListResult, error) {
	execer := execerFromContext(ctx, r.db)

	where := `WHERE feedback_id=$1 AND deleted_at IS NULL`
	args := []interface{}{q.FeedbackID}
	argIdx := 2
	if !q.IncludeTakedown {
		where += ` AND is_takedown = false`
	}

	var total int64
	countRow := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM feedback_comments `+where, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, kernel.Wrap(cconstant.CodeCommentQueryFailed, fmt.Errorf("count comments: %w", err))
	}

	limit := q.Limit
	if limit < 1 {
		limit = 10
	}
	offset := (q.Page - 1) * q.Limit
	query := fmt.Sprintf(`SELECT %s FROM feedback_comments %s ORDER BY created_at ASC LIMIT $%d OFFSET $%d`,
		commentColumns, where, argIdx, argIdx+1)
	listArgs := append(append([]interface{}{}, args...), limit, offset)

	rows, err := execer.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, kernel.Wrap(cconstant.CodeCommentQueryFailed, fmt.Errorf("list comments: %w", err))
	}
	defer rows.Close()

	items := make([]*centity.Comment, 0)
	for rows.Next() {
		c, err := scanComment(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(cconstant.CodeCommentQueryFailed, fmt.Errorf("iterate comment rows: %w", err))
	}

	return &crepo.CommentListResult{Items: items, Total: total}, nil
}

func (r *PostgresCommentRepository) IncrementLikeCount(ctx context.Context, id string) error {
	return r.updateCount(ctx, id, +1)
}

func (r *PostgresCommentRepository) DecrementLikeCount(ctx context.Context, id string) error {
	return r.updateCount(ctx, id, -1)
}

func (r *PostgresCommentRepository) updateCount(ctx context.Context, id string, delta int) error {
	execer := execerFromContext(ctx, r.db)
	query := `UPDATE feedback_comments SET like_count = GREATEST(like_count + $1, 0), updated_at = NOW() WHERE id=$2 AND deleted_at IS NULL`
	res, err := execer.ExecContext(ctx, query, delta, id)
	if err != nil {
		return kernel.Wrap(cconstant.CodeCommentPersistenceFailed, fmt.Errorf("update comment like_count: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(cconstant.CodeCommentNotFound)
	}
	return nil
}

func scanComment(sc scanner) (*centity.Comment, error) {
	var (
		id, feedbackID, userID, body string
		replyToID                    sql.NullString
		isTakedown                   bool
		takedownReason, takedownBy   sql.NullString
		takedownAt                   sql.NullTime
		likeCount                    int
		createdAt, updatedAt         time.Time
		deletedAt                    sql.NullTime
	)

	err := sc.Scan(
		&id, &feedbackID, &userID, &body, &replyToID, &isTakedown,
		&takedownReason, &takedownBy, &takedownAt,
		&likeCount, &createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(cconstant.CodeCommentNotFound)
		}
		return nil, kernel.Wrap(cconstant.CodeCommentQueryFailed, fmt.Errorf("scan comment: %w", err))
	}

	return &centity.Comment{
		ID:             id,
		FeedbackID:     feedbackID,
		UserID:         userID,
		Body:           body,
		ReplyToID:      strFromNull(replyToID),
		IsTakedown:     isTakedown,
		TakedownReason: strFromNull(takedownReason),
		TakedownBy:     strFromNull(takedownBy),
		TakedownAt:     timeFromNull(takedownAt),
		LikeCount:      likeCount,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		DeletedAt:      timeFromNull(deletedAt),
	}, nil
}
