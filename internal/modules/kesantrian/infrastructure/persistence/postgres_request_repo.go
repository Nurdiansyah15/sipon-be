package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sipon-be/internal/modules/kesantrian/domain/request/constant"
	"sipon-be/internal/modules/kesantrian/domain/request/entity"
	"sipon-be/internal/modules/kesantrian/domain/request/repository"
	"sipon-be/internal/shared/kernel"
)

const requestColumns = `
	id, user_id, nis, status, notes, reviewed_by, reviewed_at, created_at, updated_at, deleted_at
`

var requestSortColumns = map[string]string{
	"created_at": "created_at",
	"status":     "status",
}

type PostgresSantriRequestRepository struct {
	db *sql.DB
}

func NewPostgresSantriRequestRepository(db *sql.DB) *PostgresSantriRequestRepository {
	return &PostgresSantriRequestRepository{db: db}
}

func (r *PostgresSantriRequestRepository) Save(ctx context.Context, req *entity.SantriRequest) error {
	execer := execerFromContext(ctx, r.db)
	query := `INSERT INTO santri_requests (` + requestColumns + `) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := execer.ExecContext(ctx, query,
		req.ID, req.UserID, nullStr(req.NIS), string(req.Status), nullStr(req.Notes),
		nullStr(req.ReviewedBy), nullTimeVal(req.ReviewedAt), req.CreatedAt, req.UpdatedAt, nullTimeVal(req.DeletedAt),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeSantriRequestAlreadyExists, err)
		}
		return kernel.Wrap(constant.CodeSantriRequestPersistenceFailed, fmt.Errorf("save santri request: %w", err))
	}
	return nil
}

func (r *PostgresSantriRequestRepository) Update(ctx context.Context, req *entity.SantriRequest) error {
	execer := execerFromContext(ctx, r.db)
	query := `UPDATE santri_requests SET
		nis=$1, status=$2, notes=$3, reviewed_by=$4, reviewed_at=$5, updated_at=$6, deleted_at=$7
		WHERE id=$8 AND deleted_at IS NULL`
	res, err := execer.ExecContext(ctx, query,
		nullStr(req.NIS), string(req.Status), nullStr(req.Notes),
		nullStr(req.ReviewedBy), nullTimeVal(req.ReviewedAt), req.UpdatedAt, nullTimeVal(req.DeletedAt),
		req.ID,
	)
	if err != nil {
		return kernel.Wrap(constant.CodeSantriRequestPersistenceFailed, fmt.Errorf("update santri request: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeSantriRequestNotFound)
	}
	return nil
}

func (r *PostgresSantriRequestRepository) FindByID(ctx context.Context, id string) (*entity.SantriRequest, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+requestColumns+` FROM santri_requests WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanRequest(row)
}

func (r *PostgresSantriRequestRepository) FindPendingByUserID(ctx context.Context, userID string) (*entity.SantriRequest, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+requestColumns+` FROM santri_requests WHERE user_id=$1 AND status='pending' AND deleted_at IS NULL`, userID)
	return scanRequest(row)
}

func (r *PostgresSantriRequestRepository) List(ctx context.Context, q repository.SantriRequestListQuery) (*repository.SantriRequestListResult, error) {
	execer := execerFromContext(ctx, r.db)

	where := `WHERE deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 1
	if q.Status != nil && *q.Status != "" {
		where += fmt.Sprintf(` AND status = $%d`, argIdx)
		args = append(args, string(*q.Status))
		argIdx++
	}

	sortCol, ok := requestSortColumns[q.SortBy]
	if !ok {
		sortCol = "created_at"
	}
	sortDir := "DESC"
	if q.SortType == "asc" {
		sortDir = "ASC"
	}

	var total int64
	if err := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM santri_requests `+where, args...).Scan(&total); err != nil {
		return nil, kernel.Wrap(constant.CodeSantriRequestQueryFailed, fmt.Errorf("count santri requests: %w", err))
	}

	listArgs := append(append([]interface{}{}, args...), q.Limit, (q.Page-1)*q.Limit)
	query := fmt.Sprintf(`SELECT %s FROM santri_requests %s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		requestColumns, where, sortCol, sortDir, argIdx, argIdx+1)

	rows, err := execer.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeSantriRequestQueryFailed, fmt.Errorf("list santri requests: %w", err))
	}
	defer rows.Close()

	items := make([]*entity.SantriRequest, 0)
	for rows.Next() {
		item, err := scanRequest(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(constant.CodeSantriRequestQueryFailed, fmt.Errorf("iterate santri request rows: %w", err))
	}

	return &repository.SantriRequestListResult{Items: items, Total: total}, nil
}

func scanRequest(sc scanner) (*entity.SantriRequest, error) {
	var (
		id, userID, status     string
		nis, notes, reviewedBy sql.NullString
		reviewedAt, deletedAt  sql.NullTime
		createdAt, updatedAt   time.Time
	)

	err := sc.Scan(&id, &userID, &nis, &status, &notes, &reviewedBy, &reviewedAt, &createdAt, &updatedAt, &deletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeSantriRequestNotFound)
		}
		return nil, kernel.Wrap(constant.CodeSantriRequestQueryFailed, fmt.Errorf("scan santri request: %w", err))
	}

	return &entity.SantriRequest{
		ID:         id,
		UserID:     userID,
		NIS:        strFromNull(nis),
		Status:     constant.SantriRequestStatus(status),
		Notes:      strFromNull(notes),
		ReviewedBy: strFromNull(reviewedBy),
		ReviewedAt: timeFromNull(reviewedAt),
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		DeletedAt:  timeFromNull(deletedAt),
	}, nil
}
