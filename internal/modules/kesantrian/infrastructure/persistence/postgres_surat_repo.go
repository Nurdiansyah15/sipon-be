package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	suratconstant "sipon-be/internal/modules/kesantrian/domain/surat/constant"
	"sipon-be/internal/modules/kesantrian/domain/surat/entity"
	"sipon-be/internal/modules/kesantrian/domain/surat/repository"
	"sipon-be/internal/shared/kernel"
)

const suratColumns = `id, nomor, seq, tipe_surat_id, keterangan, tanggal, created_by, scope_id, created_at, updated_at`

var suratSortColumns = map[string]string{
	"nomor":      "nomor",
	"tanggal":    "tanggal",
	"created_at": "created_at",
}

type PostgresSuratRepository struct {
	db *sql.DB
}

func NewPostgresSuratRepository(db *sql.DB) *PostgresSuratRepository {
	return &PostgresSuratRepository{db: db}
}

func (r *PostgresSuratRepository) Save(ctx context.Context, s *entity.Surat) error {
	execer := execerFromContext(ctx, r.db)
	query := `INSERT INTO surat (` + suratColumns + `) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := execer.ExecContext(ctx, query,
		s.ID, s.Nomor, s.Seq, s.TipeSuratID, nullStr(s.Keterangan), s.Tanggal, s.CreatedBy, nullStr(s.ScopeID), s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(suratconstant.CodeSuratPersistenceFailed, err)
		}
		return kernel.Wrap(suratconstant.CodeSuratPersistenceFailed, fmt.Errorf("save surat: %w", err))
	}
	return nil
}

func (r *PostgresSuratRepository) Delete(ctx context.Context, id string) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx, `DELETE FROM surat WHERE id=$1`, id)
	if err != nil {
		return kernel.Wrap(suratconstant.CodeSuratPersistenceFailed, fmt.Errorf("delete surat: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(suratconstant.CodeSuratNotFound)
	}
	return nil
}

func (r *PostgresSuratRepository) FindByID(ctx context.Context, id string) (*entity.Surat, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+suratColumns+` FROM surat WHERE id=$1`, id)
	return scanSurat(row)
}

func (r *PostgresSuratRepository) List(ctx context.Context, q repository.SuratListQuery) (*repository.SuratListResult, error) {
	execer := execerFromContext(ctx, r.db)

	where := `WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if q.TipeSuratID != nil && *q.TipeSuratID != "" {
		where += fmt.Sprintf(` AND tipe_surat_id = $%d`, argIdx)
		args = append(args, *q.TipeSuratID)
		argIdx++
	}
	if q.Bulan != nil {
		where += fmt.Sprintf(` AND EXTRACT(MONTH FROM tanggal) = $%d`, argIdx)
		args = append(args, *q.Bulan)
		argIdx++
	}
	if q.Tahun != nil {
		where += fmt.Sprintf(` AND EXTRACT(YEAR FROM tanggal) = $%d`, argIdx)
		args = append(args, *q.Tahun)
		argIdx++
	}
	if q.Search != nil && *q.Search != "" {
		where += fmt.Sprintf(` AND (nomor ILIKE $%d OR keterangan ILIKE $%d)`, argIdx, argIdx)
		args = append(args, "%"+*q.Search+"%")
		argIdx++
	}

	// Scope: batasi surat sesuai akses scope user.
	// - IsDenied (restricted kosong) -> tidak ada surat yang boleh diakses.
	// - IsRestricted -> hanya surat dengan scope_id yang diizinkan atau global
	//   (scope_id IS NULL dianggap publik).
	if q.Scope.IsDenied() {
		return &repository.SuratListResult{Items: []*entity.Surat{}, Total: 0}, nil
	}
	if q.Scope.IsRestricted() {
		where += fmt.Sprintf(` AND (scope_id IS NULL OR scope_id = ANY($%d))`, argIdx)
		args = append(args, q.Scope.AllowedOptions())
		argIdx++
	}

	var total int64
	if err := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM surat `+where, args...).Scan(&total); err != nil {
		return nil, kernel.Wrap(suratconstant.CodeSuratQueryFailed, fmt.Errorf("count surat: %w", err))
	}

	sortCol, ok := suratSortColumns[q.SortBy]
	if !ok {
		sortCol = "created_at"
	}
	sortDir := "DESC"
	if q.SortType == "asc" {
		sortDir = "ASC"
	}

	listArgs := append(append([]interface{}{}, args...), q.Limit, (q.Page-1)*q.Limit)
	query := fmt.Sprintf(`SELECT %s FROM surat %s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		suratColumns, where, sortCol, sortDir, argIdx, argIdx+1)

	rows, err := execer.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, kernel.Wrap(suratconstant.CodeSuratQueryFailed, fmt.Errorf("list surat: %w", err))
	}
	defer rows.Close()

	items := make([]*entity.Surat, 0)
	for rows.Next() {
		item, err := scanSurat(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(suratconstant.CodeSuratQueryFailed, fmt.Errorf("iterate surat rows: %w", err))
	}

	return &repository.SuratListResult{Items: items, Total: total}, nil
}

func (r *PostgresSuratRepository) FindMaxSeqByMonthYear(ctx context.Context, scopeID *string, bulan, tahun int) (int, error) {
	execer := execerFromContext(ctx, r.db)

	// Advisory lock must be taken and held on the same tx connection that will
	// read the max seq and insert the surat. database/sql + pgx extended
	// protocol reject multi-statement queries in a prepared statement, so the
	// lock is executed as its own statement first (pg_advisory_xact_lock is
	// transaction-scoped, so it stays held until commit/rollback).
	// Sequence dibagi per scope (scope_id NULL untuk surat global/publik).
	if _, err := execer.ExecContext(ctx,
		`SELECT pg_advisory_xact_lock(hashtext('persuratan_seq_' || COALESCE($1::text, 'global') || '_' || $2::text || '_' || $3::text))`,
		strPtrVal(scopeID), strconv.Itoa(bulan), strconv.Itoa(tahun),
	); err != nil {
		return 0, kernel.Wrap(suratconstant.CodeSuratNomorFailed, fmt.Errorf("acquire seq lock: %w", err))
	}

	var maxSeq sql.NullInt64
	err := execer.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(seq), 0) FROM surat WHERE EXTRACT(MONTH FROM tanggal) = $1 AND EXTRACT(YEAR FROM tanggal) = $2 AND scope_id IS NOT DISTINCT FROM $3`,
		bulan, tahun, nullStr(scopeID),
	).Scan(&maxSeq)
	if err != nil {
		return 0, kernel.Wrap(suratconstant.CodeSuratNomorFailed, fmt.Errorf("find max seq: %w", err))
	}
	return int(maxSeq.Int64), nil
}

func (r *PostgresSuratRepository) SaveDokumenLink(ctx context.Context, link *entity.SuratDokumenAset) error {
	execer := execerFromContext(ctx, r.db)
	query := `INSERT INTO surat_dokumen_aset (id, surat_id, dokumen_aset_id, created_at) VALUES ($1,$2,$3,$4)`
	_, err := execer.ExecContext(ctx, query, link.ID, link.SuratID, link.DokumenAsetID, link.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(suratconstant.CodeSuratDokumenExists, err)
		}
		return kernel.Wrap(suratconstant.CodeSuratPersistenceFailed, fmt.Errorf("save dokumen link: %w", err))
	}
	return nil
}

func (r *PostgresSuratRepository) DeleteDokumenLink(ctx context.Context, suratID, dokumenAsetID string) error {
	execer := execerFromContext(ctx, r.db)
	res, err := execer.ExecContext(ctx, `DELETE FROM surat_dokumen_aset WHERE surat_id=$1 AND dokumen_aset_id=$2`, suratID, dokumenAsetID)
	if err != nil {
		return kernel.Wrap(suratconstant.CodeSuratPersistenceFailed, fmt.Errorf("delete dokumen link: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(suratconstant.CodeSuratDokumenNotFound)
	}
	return nil
}

func (r *PostgresSuratRepository) FindDokumenAsetIDsBySuratID(ctx context.Context, suratID string) ([]string, error) {
	execer := execerFromContext(ctx, r.db)
	rows, err := execer.QueryContext(ctx, `SELECT dokumen_aset_id FROM surat_dokumen_aset WHERE surat_id=$1 ORDER BY created_at`, suratID)
	if err != nil {
		return nil, kernel.Wrap(suratconstant.CodeSuratQueryFailed, fmt.Errorf("list dokumen links: %w", err))
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, kernel.Wrap(suratconstant.CodeSuratQueryFailed, fmt.Errorf("scan dokumen link: %w", err))
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *PostgresSuratRepository) FindDetail(ctx context.Context, id string) (*repository.SuratDetail, error) {
	s, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	dokIDs, err := r.FindDokumenAsetIDsBySuratID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &repository.SuratDetail{Surat: s, DokumenAsetIDs: dokIDs}, nil
}

func scanSurat(sc scanner) (*entity.Surat, error) {
	var (
		id, nomor, tipeSuratID, createdBy string
		keterangan                        sql.NullString
		scopeID                           sql.NullString
		seq                               int
		tanggal                           time.Time
		createdAt, updatedAt              time.Time
	)
	err := sc.Scan(&id, &nomor, &seq, &tipeSuratID, &keterangan, &tanggal, &createdBy, &scopeID, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(suratconstant.CodeSuratNotFound)
		}
		return nil, kernel.Wrap(suratconstant.CodeSuratQueryFailed, fmt.Errorf("scan surat: %w", err))
	}
	return &entity.Surat{
		ID:          id,
		Nomor:       nomor,
		Seq:         seq,
		TipeSuratID: tipeSuratID,
		Keterangan:  strFromNull(keterangan),
		Tanggal:     tanggal,
		CreatedBy:   createdBy,
		ScopeID:     strPtrFromNull(scopeID),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}
