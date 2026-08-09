package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sipon-be/internal/modules/keuangan/domain/journal/constant"
	"sipon-be/internal/modules/keuangan/domain/journal/entity"
	"sipon-be/internal/modules/keuangan/domain/journal/repository"
	journalVO "sipon-be/internal/modules/keuangan/domain/journal/valueobject"
	"sipon-be/internal/shared/kernel"
)

const journalEntryColumns = `
	id, journal_number, entry_date, description, source_type, source_id,
	period_id, total_debit, total_credit, posted_by, posted_at, status,
	created_at, updated_at
`

type PostgresJournalRepository struct {
	db *sql.DB
}

func NewPostgresJournalRepository(db *sql.DB) *PostgresJournalRepository {
	return &PostgresJournalRepository{db: db}
}

func (r *PostgresJournalRepository) Save(ctx context.Context, entry *entity.JournalEntry) error {
	execer := execerFromContext(ctx, r.db)

	query := `INSERT INTO journal_entries (` + journalEntryColumns + `) VALUES (
		$1,$2,$3,$4,$5,$6,
		$7,$8,$9,$10,$11,$12,
		$13,$14
	)`

	_, err := execer.ExecContext(ctx, query,
		entry.ID, entry.JournalNumber, entry.EntryDate, entry.Description,
		nullStr((*string)(entry.SourceType)), nullStr(entry.SourceID),
		entry.PeriodID, entry.TotalDebit, entry.TotalCredit, entry.PostedBy,
		nullTimeVal(entry.PostedAt), string(entry.Status),
		entry.CreatedAt, entry.UpdatedAt,
	)
	if err != nil {
		return kernel.WrapMsg(constant.CodeJournalPersistenceFailed, "gagal menyimpan jurnal", err)
	}

	if len(entry.Lines) > 0 {
		if err := r.SaveLines(ctx, entry.ID, entry.Lines); err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresJournalRepository) Update(ctx context.Context, entry *entity.JournalEntry) error {
	execer := execerFromContext(ctx, r.db)

	query := `UPDATE journal_entries SET
		description=$1, total_debit=$2, total_credit=$3, posted_by=$4,
		posted_at=$5, status=$6, updated_at=$7
		WHERE id=$8`

	res, err := execer.ExecContext(ctx, query,
		entry.Description, entry.TotalDebit, entry.TotalCredit, entry.PostedBy,
		nullTimeVal(entry.PostedAt), string(entry.Status), entry.UpdatedAt,
		entry.ID,
	)
	if err != nil {
		return kernel.WrapMsg(constant.CodeJournalPersistenceFailed, "gagal memperbarui jurnal", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.WrapMsg(constant.CodeJournalNotFound, "Jurnal tidak ditemukan", nil)
	}
	return nil
}

func (r *PostgresJournalRepository) FindByID(ctx context.Context, id string) (*entity.JournalEntry, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+journalEntryColumns+` FROM journal_entries WHERE id=$1`, id)
	entry, err := r.scan(row)
	if err != nil {
		return nil, err
	}

	lines, err := r.FindLinesByEntryID(ctx, id)
	if err != nil {
		return nil, err
	}
	entry.Lines = lines
	return entry, nil
}

func (r *PostgresJournalRepository) NextJournalNumber(ctx context.Context) (journalVO.JournalNumber, error) {
	execer := execerFromContext(ctx, r.db)
	now := time.Now()
	year := now.Year()

	seq, err := nextNumberSeq(ctx, execer, "journal", year)
	if err != nil {
		return journalVO.JournalNumber{}, kernel.WrapMsg(constant.CodeJournalPersistenceFailed, "gagal membuat nomor jurnal", err)
	}

	return journalVO.NewJournalNumber(fmt.Sprintf("%d", year), fmt.Sprintf("%02d", int(now.Month())), seq), nil
}

func (r *PostgresJournalRepository) FindByNumber(ctx context.Context, number string) (*entity.JournalEntry, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+journalEntryColumns+` FROM journal_entries WHERE journal_number=$1`, number)
	entry, err := r.scan(row)
	if err != nil {
		return nil, err
	}

	lines, err := r.FindLinesByEntryID(ctx, entry.ID)
	if err != nil {
		return nil, err
	}
	entry.Lines = lines
	return entry, nil
}

func (r *PostgresJournalRepository) List(ctx context.Context, q repository.JournalListQuery) (*repository.JournalListResult, error) {
	execer := execerFromContext(ctx, r.db)

	where := `WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if q.PeriodID != nil && *q.PeriodID != "" {
		where += fmt.Sprintf(` AND period_id=$%d`, argIdx)
		args = append(args, *q.PeriodID)
		argIdx++
	}
	if q.Status != nil && *q.Status != "" {
		where += fmt.Sprintf(` AND status=$%d`, argIdx)
		args = append(args, *q.Status)
		argIdx++
	}
	if q.SourceType != nil && *q.SourceType != "" {
		where += fmt.Sprintf(` AND source_type=$%d`, argIdx)
		args = append(args, *q.SourceType)
		argIdx++
	}

	var total int64
	countRow := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM journal_entries `+where, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, kernel.WrapMsg(constant.CodeJournalQueryFailed, "gagal menghitung jumlah jurnal", err)
	}

	limit := q.Limit
	offset := (q.Page - 1) * q.Limit
	listArgs := append(append([]interface{}{}, args...), limit, offset)

	query := fmt.Sprintf(`SELECT %s FROM journal_entries %s ORDER BY entry_date DESC, created_at DESC LIMIT $%d OFFSET $%d`,
		journalEntryColumns, where, argIdx, argIdx+1)

	rows, err := execer.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, kernel.WrapMsg(constant.CodeJournalQueryFailed, "gagal mendaftar jurnal", err)
	}
	defer rows.Close()

	items := make([]*entity.JournalEntry, 0)
	for rows.Next() {
		entry, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(constant.CodeJournalQueryFailed, "gagal membaca data jurnal", err)
	}

	return &repository.JournalListResult{Items: items, Total: total}, nil
}

func (r *PostgresJournalRepository) FindBySource(ctx context.Context, sourceType string, sourceID string) (*entity.JournalEntry, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx,
		`SELECT `+journalEntryColumns+` FROM journal_entries WHERE source_type=$1 AND source_id=$2`,
		sourceType, sourceID,
	)
	entry, err := r.scan(row)
	if err != nil {
		return nil, err
	}

	lines, err := r.FindLinesByEntryID(ctx, entry.ID)
	if err != nil {
		return nil, err
	}
	entry.Lines = lines
	return entry, nil
}

func (r *PostgresJournalRepository) SaveLines(ctx context.Context, entryID string, lines []*entity.JournalEntryLine) error {
	execer := execerFromContext(ctx, r.db)

	query := `INSERT INTO journal_entry_lines (id, journal_entry_id, account_id, account_code, description, debit, credit) VALUES ($1,$2,$3,$4,$5,$6,$7)`

	for _, line := range lines {
		_, err := execer.ExecContext(ctx, query,
			line.ID, line.JournalEntryID, line.AccountID, line.AccountCode,
			nullStr(line.Description), line.Debit, line.Credit,
		)
		if err != nil {
			return kernel.WrapMsg(constant.CodeJournalPersistenceFailed, "gagal menyimpan baris jurnal", err)
		}
	}
	return nil
}

func (r *PostgresJournalRepository) FindLinesByEntryID(ctx context.Context, entryID string) ([]*entity.JournalEntryLine, error) {
	execer := execerFromContext(ctx, r.db)

	rows, err := execer.QueryContext(ctx,
		`SELECT id, journal_entry_id, account_id, account_code, description, debit, credit FROM journal_entry_lines WHERE journal_entry_id=$1 ORDER BY id ASC`,
		entryID,
	)
	if err != nil {
		return nil, kernel.WrapMsg(constant.CodeJournalQueryFailed, "gagal mencari baris jurnal", err)
	}
	defer rows.Close()

	items := make([]*entity.JournalEntryLine, 0)
	for rows.Next() {
		var (
			id, jeID, accountID, accountCode string
			description                      sql.NullString
			debit, credit                    float64
		)
		if err := rows.Scan(&id, &jeID, &accountID, &accountCode, &description, &debit, &credit); err != nil {
			return nil, kernel.WrapMsg(constant.CodeJournalQueryFailed, "gagal membaca data baris jurnal", err)
		}
		items = append(items, &entity.JournalEntryLine{
			ID:             id,
			JournalEntryID: jeID,
			AccountID:      accountID,
			AccountCode:    accountCode,
			Description:    strFromNull(description),
			Debit:          debit,
			Credit:         credit,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(constant.CodeJournalQueryFailed, "gagal membaca data baris jurnal", err)
	}
	return items, nil
}

func (r *PostgresJournalRepository) ComputeAccountBalances(ctx context.Context, periodID string) (map[string]repository.AccountBalance, error) {
	execer := execerFromContext(ctx, r.db)

	rows, err := execer.QueryContext(ctx,
		`SELECT jel.account_id, COALESCE(SUM(jel.debit), 0) AS total_debit, COALESCE(SUM(jel.credit), 0) AS total_credit
		 FROM journal_entry_lines jel
		 JOIN journal_entries je ON je.id = jel.journal_entry_id
		 WHERE je.period_id = $1 AND je.status = 'posted'
		 GROUP BY jel.account_id`,
		periodID,
	)
	if err != nil {
		return nil, kernel.WrapMsg(constant.CodeJournalQueryFailed, "gagal menghitung saldo akun", err)
	}
	defer rows.Close()

	balances := make(map[string]repository.AccountBalance)
	for rows.Next() {
		var (
			accountID                     string
			totalDebit, totalCredit       float64
		)
		if err := rows.Scan(&accountID, &totalDebit, &totalCredit); err != nil {
			return nil, kernel.WrapMsg(constant.CodeJournalQueryFailed, "gagal membaca saldo akun", err)
		}
		balances[accountID] = repository.AccountBalance{Debit: totalDebit, Credit: totalCredit}
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(constant.CodeJournalQueryFailed, "gagal membaca saldo akun", err)
	}
	return balances, nil
}

func (r *PostgresJournalRepository) scan(sc scanner) (*entity.JournalEntry, error) {
	var (
		id, journalNumber, description, periodID, postedBy, status              string
		sourceType, sourceID                                                    sql.NullString
		entryDate                                                               time.Time
		totalDebit, totalCredit                                                 float64
		postedAt                                                                sql.NullTime
		createdAt, updatedAt                                                    time.Time
	)

	err := sc.Scan(
		&id, &journalNumber, &entryDate, &description, &sourceType, &sourceID,
		&periodID, &totalDebit, &totalCredit, &postedBy, &postedAt, &status,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.WrapMsg(constant.CodeJournalNotFound, "Jurnal tidak ditemukan", nil)
		}
		return nil, kernel.WrapMsg(constant.CodeJournalQueryFailed, "gagal membaca data jurnal", err)
	}

	var st *constant.SourceType
	if sourceType.Valid {
		v := constant.SourceType(sourceType.String)
		st = &v
	}

	return &entity.JournalEntry{
		ID:            id,
		JournalNumber: journalNumber,
		EntryDate:     entryDate,
		Description:   description,
		SourceType:    st,
		SourceID:      strFromNull(sourceID),
		PeriodID:      periodID,
		TotalDebit:    totalDebit,
		TotalCredit:   totalCredit,
		PostedBy:      postedBy,
		PostedAt:      timeFromNull(postedAt),
		Status:        constant.JournalStatus(status),
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}, nil
}
