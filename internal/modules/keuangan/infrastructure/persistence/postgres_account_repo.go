package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sipon-be/internal/modules/keuangan/domain/account/constant"
	"sipon-be/internal/modules/keuangan/domain/account/entity"
	"sipon-be/internal/modules/keuangan/domain/account/repository"
	"sipon-be/internal/shared/kernel"
)

const accountColumns = `
	id, code, name, type, parent_id, level, is_postable, normal_balance, description,
	is_active, is_system, created_by, created_at, updated_at, deleted_at
`

type PostgresAccountRepository struct {
	db *sql.DB
}

func NewPostgresAccountRepository(db *sql.DB) *PostgresAccountRepository {
	return &PostgresAccountRepository{db: db}
}

func (r *PostgresAccountRepository) Save(ctx context.Context, acc *entity.Account) error {
	execer := execerFromContext(ctx, r.db)

	query := `INSERT INTO accounts (` + accountColumns + `) VALUES (
		$1,$2,$3,$4,$5,$6,$7,$8,$9,
		$10,$11,$12,$13,$14,$15
	)`

	_, err := execer.ExecContext(ctx, query,
		acc.ID, acc.Code, acc.Name, string(acc.Type),
		nullStr(acc.ParentID), acc.Level, acc.IsPostable, string(acc.NormalBalance),
		nullStr(acc.Description), acc.IsActive, acc.IsSystem,
		acc.CreatedBy, acc.CreatedAt, acc.UpdatedAt, nullTimeVal(acc.DeletedAt),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeAccountDuplicate, err)
		}
		return kernel.Wrap(constant.CodeAccountPersistenceFailed, fmt.Errorf("save account: %w", err))
	}
	return nil
}

func (r *PostgresAccountRepository) Update(ctx context.Context, acc *entity.Account) error {
	execer := execerFromContext(ctx, r.db)

	query := `UPDATE accounts SET
		name=$1, description=$2, is_postable=$3, is_active=$4,
		updated_at=$5, deleted_at=$6
		WHERE id=$7 AND deleted_at IS NULL`

	res, err := execer.ExecContext(ctx, query,
		acc.Name, nullStr(acc.Description), acc.IsPostable, acc.IsActive,
		acc.UpdatedAt, nullTimeVal(acc.DeletedAt),
		acc.ID,
	)
	if err != nil {
		return kernel.Wrap(constant.CodeAccountPersistenceFailed, fmt.Errorf("update account: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeAccountNotFound)
	}
	return nil
}

func (r *PostgresAccountRepository) FindByID(ctx context.Context, id string) (*entity.Account, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+accountColumns+` FROM accounts WHERE id=$1 AND deleted_at IS NULL`, id)
	return r.scan(row)
}

func (r *PostgresAccountRepository) FindByCode(ctx context.Context, code string) (*entity.Account, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+accountColumns+` FROM accounts WHERE code=$1 AND deleted_at IS NULL`, code)
	return r.scan(row)
}

func (r *PostgresAccountRepository) List(ctx context.Context, q repository.AccountListQuery) (*repository.AccountListResult, error) {
	execer := execerFromContext(ctx, r.db)

	where := `WHERE deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 1
	if q.Type != nil && *q.Type != "" {
		where += fmt.Sprintf(` AND type=$%d`, argIdx)
		args = append(args, *q.Type)
		argIdx++
	}
	if q.Active != nil {
		where += fmt.Sprintf(` AND is_active=$%d`, argIdx)
		args = append(args, *q.Active)
		argIdx++
	}

	var total int64
	countRow := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM accounts `+where, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, kernel.Wrap(constant.CodeAccountQueryFailed, fmt.Errorf("count accounts: %w", err))
	}

	limit := q.Limit
	offset := (q.Page - 1) * q.Limit
	listArgs := append(append([]interface{}{}, args...), limit, offset)

	query := fmt.Sprintf(`SELECT %s FROM accounts %s ORDER BY code ASC LIMIT $%d OFFSET $%d`,
		accountColumns, where, argIdx, argIdx+1)

	rows, err := execer.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeAccountQueryFailed, fmt.Errorf("list accounts: %w", err))
	}
	defer rows.Close()

	items := make([]*entity.Account, 0)
	for rows.Next() {
		acc, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, acc)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(constant.CodeAccountQueryFailed, fmt.Errorf("iterate account rows: %w", err))
	}

	return &repository.AccountListResult{Items: items, Total: total}, nil
}

func (r *PostgresAccountRepository) ListAll(ctx context.Context) ([]*entity.Account, error) {
	execer := execerFromContext(ctx, r.db)

	rows, err := execer.QueryContext(ctx, `SELECT `+accountColumns+` FROM accounts WHERE deleted_at IS NULL ORDER BY code ASC`)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeAccountQueryFailed, fmt.Errorf("list all accounts: %w", err))
	}
	defer rows.Close()

	items := make([]*entity.Account, 0)
	for rows.Next() {
		acc, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, acc)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(constant.CodeAccountQueryFailed, fmt.Errorf("iterate accounts: %w", err))
	}
	return items, nil
}

func (r *PostgresAccountRepository) ListPostable(ctx context.Context) ([]*entity.Account, error) {
	execer := execerFromContext(ctx, r.db)

	rows, err := execer.QueryContext(ctx,
		`SELECT `+accountColumns+` FROM accounts WHERE is_postable=true AND is_active=true AND deleted_at IS NULL ORDER BY code ASC`,
	)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeAccountQueryFailed, fmt.Errorf("list postable accounts: %w", err))
	}
	defer rows.Close()

	items := make([]*entity.Account, 0)
	for rows.Next() {
		acc, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, acc)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(constant.CodeAccountQueryFailed, fmt.Errorf("iterate postable accounts: %w", err))
	}
	return items, nil
}

func (r *PostgresAccountRepository) FindChildren(ctx context.Context, parentID string) ([]*entity.Account, error) {
	execer := execerFromContext(ctx, r.db)

	rows, err := execer.QueryContext(ctx,
		`SELECT `+accountColumns+` FROM accounts WHERE parent_id=$1 AND deleted_at IS NULL ORDER BY code ASC`,
		parentID,
	)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeAccountQueryFailed, fmt.Errorf("find children: %w", err))
	}
	defer rows.Close()

	items := make([]*entity.Account, 0)
	for rows.Next() {
		acc, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, acc)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(constant.CodeAccountQueryFailed, fmt.Errorf("iterate children: %w", err))
	}
	return items, nil
}

func (r *PostgresAccountRepository) HasJournalEntries(ctx context.Context, accountID string) (bool, error) {
	execer := execerFromContext(ctx, r.db)

	var exists bool
	err := execer.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM journal_entry_lines WHERE account_id=$1)`,
		accountID,
	).Scan(&exists)
	if err != nil {
		return false, kernel.Wrap(constant.CodeAccountQueryFailed, fmt.Errorf("has journal entries: %w", err))
	}
	return exists, nil
}

func (r *PostgresAccountRepository) ExistsByCode(ctx context.Context, code string, excludeID string) (bool, error) {
	execer := execerFromContext(ctx, r.db)

	var exists bool
	var err error
	if excludeID == "" {
		query := `SELECT EXISTS(SELECT 1 FROM accounts WHERE code=$1 AND deleted_at IS NULL)`
		err = execer.QueryRowContext(ctx, query, code).Scan(&exists)
	} else {
		query := `SELECT EXISTS(SELECT 1 FROM accounts WHERE code=$1 AND deleted_at IS NULL AND id!=$2)`
		err = execer.QueryRowContext(ctx, query, code, excludeID).Scan(&exists)
	}
	if err != nil {
		return false, kernel.Wrap(constant.CodeAccountQueryFailed, fmt.Errorf("exists by code: %w", err))
	}
	return exists, nil
}

func (r *PostgresAccountRepository) scan(sc scanner) (*entity.Account, error) {
	var (
		id, code, name, accType, createdBy                                  string
		parentID, description                                               sql.NullString
		level                                                               int
		isPostable, isActive, isSystem                                      bool
		normalBalance                                                       string
		createdAt, updatedAt                                                time.Time
		deletedAt                                                           sql.NullTime
	)

	err := sc.Scan(
		&id, &code, &name, &accType, &parentID, &level, &isPostable, &normalBalance, &description,
		&isActive, &isSystem, &createdBy, &createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeAccountNotFound)
		}
		return nil, kernel.Wrap(constant.CodeAccountQueryFailed, fmt.Errorf("scan account: %w", err))
	}

	return &entity.Account{
		ID:            id,
		Code:          code,
		Name:          name,
		Type:          constant.AccountType(accType),
		ParentID:      strFromNull(parentID),
		Level:         level,
		IsPostable:    isPostable,
		NormalBalance: constant.NormalBalance(normalBalance),
		Description:   strFromNull(description),
		IsActive:      isActive,
		IsSystem:      isSystem,
		CreatedBy:     createdBy,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		DeletedAt:     timeFromNull(deletedAt),
	}, nil
}
