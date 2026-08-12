package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	ptrConst "sipon-be/internal/modules/akademik/domain/program_transfer_request/constant"
	ptrEntity "sipon-be/internal/modules/akademik/domain/program_transfer_request/entity"
	ptrRepo "sipon-be/internal/modules/akademik/domain/program_transfer_request/repository"
	"sipon-be/internal/shared/kernel"
)

const ptrColumns = `
	id, santri_id, from_program_id, to_program_id, status, notes, admin_notes,
	reviewed_by, reviewed_at, created_at, updated_at, deleted_at
`

type PostgresProgramTransferRequestRepository struct {
	db *sql.DB
}

func NewPostgresProgramTransferRequestRepository(db *sql.DB) *PostgresProgramTransferRequestRepository {
	return &PostgresProgramTransferRequestRepository{db: db}
}

func (r *PostgresProgramTransferRequestRepository) Save(ctx context.Context, req *ptrEntity.ProgramTransferRequest) error {
	execer := execerFromContext(ctx, r.db)
	_, err := execer.ExecContext(ctx,
		`INSERT INTO program_transfer_requests (`+ptrColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		req.ID, req.SantriID, req.FromProgramID, req.ToProgramID, string(req.Status),
		nullStr(req.Notes), nullStr(req.AdminNotes), nullStr(req.ReviewedBy), nullTimeVal(req.ReviewedAt),
		req.CreatedAt, req.UpdatedAt, nullTimeVal(req.DeletedAt),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(ptrConst.CodeProgramTransferRequestDuplicate, err)
		}
		return kernel.Wrap(ptrConst.CodeProgramTransferRequestPersistenceFailed, fmt.Errorf("save program transfer request: %w", err))
	}
	return nil
}

func (r *PostgresProgramTransferRequestRepository) Update(ctx context.Context, req *ptrEntity.ProgramTransferRequest) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx,
		`UPDATE program_transfer_requests SET status=$1, notes=$2, admin_notes=$3, reviewed_by=$4,
		        reviewed_at=$5, updated_at=$6, deleted_at=$7 WHERE id=$8 AND deleted_at IS NULL`,
		string(req.Status), nullStr(req.Notes), nullStr(req.AdminNotes), nullStr(req.ReviewedBy),
		nullTimeVal(req.ReviewedAt), req.UpdatedAt, nullTimeVal(req.DeletedAt), req.ID,
	)
	if err != nil {
		return kernel.Wrap(ptrConst.CodeProgramTransferRequestPersistenceFailed, fmt.Errorf("update program transfer request: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(ptrConst.CodeProgramTransferRequestNotFound)
	}
	return nil
}

func (r *PostgresProgramTransferRequestRepository) FindByID(ctx context.Context, id string) (*ptrEntity.ProgramTransferRequest, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx,
		`SELECT `+ptrColumns+` FROM program_transfer_requests WHERE id=$1 AND deleted_at IS NULL`, id)
	return scanPtr(row)
}

func (r *PostgresProgramTransferRequestRepository) FindPendingBySantriID(ctx context.Context, santriID string) (*ptrEntity.ProgramTransferRequest, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx,
		`SELECT `+ptrColumns+` FROM program_transfer_requests
		 WHERE santri_id=$1 AND status='pending' AND deleted_at IS NULL
		 ORDER BY created_at DESC LIMIT 1`, santriID)
	return scanPtr(row)
}

func (r *PostgresProgramTransferRequestRepository) List(ctx context.Context, q ptrRepo.ProgramTransferRequestListQuery) (*ptrRepo.ProgramTransferRequestListResult, error) {
	execer := execerFromContext(ctx, r.db)

	where := `deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 1
	if q.SantriID != nil && *q.SantriID != "" {
		where += fmt.Sprintf(` AND santri_id=$%d`, argIdx)
		args = append(args, *q.SantriID)
		argIdx++
	}
	if q.Status != nil && *q.Status != "" {
		where += fmt.Sprintf(` AND status=$%d`, argIdx)
		args = append(args, *q.Status)
		argIdx++
	}

	var total int64
	if err := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM program_transfer_requests WHERE `+where, args...).Scan(&total); err != nil {
		return nil, kernel.Wrap(ptrConst.CodeProgramTransferRequestQueryFailed, err)
	}

	limit := q.Limit
	offset := (q.Page - 1) * q.Limit
	listArgs := append(append([]interface{}{}, args...), limit, offset)
	rows, err := execer.QueryContext(ctx,
		fmt.Sprintf(`SELECT %s FROM program_transfer_requests WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
			ptrColumns, where, argIdx, argIdx+1), listArgs...)
	if err != nil {
		return nil, kernel.Wrap(ptrConst.CodeProgramTransferRequestQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*ptrEntity.ProgramTransferRequest, 0)
	for rows.Next() {
		req, err := scanPtr(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, req)
	}
	return &ptrRepo.ProgramTransferRequestListResult{Items: items, Total: total}, rows.Err()
}

func scanPtr(sc scanner) (*ptrEntity.ProgramTransferRequest, error) {
	var (
		id, santriID, fromProgramID, toProgramID, status string
		notes, adminNotes, reviewedBy                    sql.NullString
		reviewedAt, deletedAt                            sql.NullTime
		createdAt, updatedAt                             time.Time
	)
	err := sc.Scan(
		&id, &santriID, &fromProgramID, &toProgramID, &status,
		&notes, &adminNotes, &reviewedBy, &reviewedAt,
		&createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(ptrConst.CodeProgramTransferRequestNotFound)
		}
		return nil, kernel.Wrap(ptrConst.CodeProgramTransferRequestQueryFailed, fmt.Errorf("scan program transfer request: %w", err))
	}
	return &ptrEntity.ProgramTransferRequest{
		ID:            id,
		SantriID:      santriID,
		FromProgramID: fromProgramID,
		ToProgramID:   toProgramID,
		Status:        ptrConst.ProgramTransferRequestStatus(status),
		Notes:         strFromNull(notes),
		AdminNotes:    strFromNull(adminNotes),
		ReviewedBy:    strFromNull(reviewedBy),
		ReviewedAt:    timeFromNull(reviewedAt),
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		DeletedAt:     timeFromNull(deletedAt),
	}, nil
}
