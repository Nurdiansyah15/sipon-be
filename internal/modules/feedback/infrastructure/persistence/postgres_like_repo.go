package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	lconstant "sipon-be/internal/modules/feedback/domain/like/constant"
	lentity "sipon-be/internal/modules/feedback/domain/like/entity"
	"sipon-be/internal/shared/kernel"
)

type PostgresLikeRepository struct {
	db *sql.DB
}

func NewPostgresLikeRepository(db *sql.DB) *PostgresLikeRepository {
	return &PostgresLikeRepository{db: db}
}

func (r *PostgresLikeRepository) Save(ctx context.Context, l *lentity.Like) error {
	execer := execerFromContext(ctx, r.db)
	query := `INSERT INTO feedback_likes (id, user_id, target_type, target_id, created_at) VALUES ($1,$2,$3,$4,$5)`
	_, err := execer.ExecContext(ctx, query, l.ID, l.UserID, string(l.TargetType), l.TargetID, l.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.New(lconstant.CodeLikePersistenceFailed)
		}
		return kernel.Wrap(lconstant.CodeLikePersistenceFailed, fmt.Errorf("save like: %w", err))
	}
	return nil
}

func (r *PostgresLikeRepository) Delete(ctx context.Context, userID string, targetType lconstant.LikeTargetType, targetID string) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx, `DELETE FROM feedback_likes WHERE user_id=$1 AND target_type=$2 AND target_id=$3`, userID, string(targetType), targetID)
	if err != nil {
		return kernel.Wrap(lconstant.CodeLikePersistenceFailed, fmt.Errorf("delete like: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(lconstant.CodeLikeNotFound)
	}
	return nil
}

func (r *PostgresLikeRepository) Exists(ctx context.Context, userID string, targetType lconstant.LikeTargetType, targetID string) (bool, error) {
	execer := execerFromContext(ctx, r.db)
	var count int
	err := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM feedback_likes WHERE user_id=$1 AND target_type=$2 AND target_id=$3`, userID, string(targetType), targetID).Scan(&count)
	if err != nil {
		return false, kernel.Wrap(lconstant.CodeLikeQueryFailed, fmt.Errorf("exists like: %w", err))
	}
	return count > 0, nil
}

func (r *PostgresLikeRepository) ListLikedTargetIDs(ctx context.Context, userID string, targetType lconstant.LikeTargetType, targetIDs []string) (map[string]bool, error) {
	execer := execerFromContext(ctx, r.db)
	if len(targetIDs) == 0 {
		return map[string]bool{}, nil
	}

	result := make(map[string]bool, len(targetIDs))
	for _, id := range targetIDs {
		result[id] = false
	}

	rows, err := execer.QueryContext(ctx, `SELECT target_id FROM feedback_likes WHERE user_id=$1 AND target_type=$2 AND target_id = ANY($3)`, userID, string(targetType), toUUIDSlice(targetIDs))
	if err != nil {
		return nil, kernel.Wrap(lconstant.CodeLikeQueryFailed, fmt.Errorf("list liked target ids: %w", err))
	}
	defer rows.Close()

	for rows.Next() {
		var targetID string
		if err := rows.Scan(&targetID); err != nil {
			return nil, kernel.Wrap(lconstant.CodeLikeQueryFailed, fmt.Errorf("scan liked target id: %w", err))
		}
		result[targetID] = true
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(lconstant.CodeLikeQueryFailed, fmt.Errorf("iterate liked target rows: %w", err))
	}

	return result, nil
}

func toUUIDSlice(ids []string) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		parsed, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		out = append(out, parsed)
	}
	return out
}
