package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sipon-be/internal/modules/keuangan/domain/billingscheme/constant"
	"sipon-be/internal/modules/keuangan/domain/billingscheme/entity"
	"sipon-be/internal/modules/keuangan/domain/billingscheme/repository"
	"sipon-be/internal/shared/kernel"
)

const billingSchemeColumns = `
	id, name, description, is_active, created_by, created_at, updated_at
`

type PostgresBillingSchemeRepository struct {
	db *sql.DB
}

func NewPostgresBillingSchemeRepository(db *sql.DB) *PostgresBillingSchemeRepository {
	return &PostgresBillingSchemeRepository{db: db}
}

func (r *PostgresBillingSchemeRepository) Save(ctx context.Context, scheme *entity.BillingScheme) error {
	execer := execerFromContext(ctx, r.db)

	query := `INSERT INTO billing_schemes (` + billingSchemeColumns + `) VALUES (
		$1,$2,$3,$4,$5,$6,$7
	)`

	_, err := execer.ExecContext(ctx, query,
		scheme.ID, scheme.Name, nullStr(scheme.Description), scheme.IsActive,
		scheme.CreatedBy, scheme.CreatedAt, scheme.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeBillingSchemeDuplicate, err)
		}
		return kernel.Wrap(constant.CodeBillingSchemePersistenceFailed, fmt.Errorf("save billing scheme: %w", err))
	}
	return nil
}

func (r *PostgresBillingSchemeRepository) Update(ctx context.Context, scheme *entity.BillingScheme) error {
	execer := execerFromContext(ctx, r.db)

	query := `UPDATE billing_schemes SET
		name=$1, description=$2, is_active=$3, updated_at=$4
		WHERE id=$5`

	res, err := execer.ExecContext(ctx, query,
		scheme.Name, nullStr(scheme.Description), scheme.IsActive,
		scheme.UpdatedAt, scheme.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeBillingSchemeDuplicate, err)
		}
		return kernel.Wrap(constant.CodeBillingSchemePersistenceFailed, fmt.Errorf("update billing scheme: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeBillingSchemeNotFound)
	}
	return nil
}

func (r *PostgresBillingSchemeRepository) FindByID(ctx context.Context, id string) (*entity.BillingScheme, error) {
	execer := execerFromContext(ctx, r.db)
	row := execer.QueryRowContext(ctx, `SELECT `+billingSchemeColumns+` FROM billing_schemes WHERE id=$1`, id)
	scheme, err := r.scan(row)
	if err != nil {
		return nil, err
	}

	items, err := r.findItemsBySchemeID(ctx, id)
	if err != nil {
		return nil, err
	}
	scheme.Items = items
	return scheme, nil
}

func (r *PostgresBillingSchemeRepository) List(ctx context.Context, q repository.BillingSchemeListQuery) (*repository.BillingSchemeListResult, error) {
	execer := execerFromContext(ctx, r.db)

	where := `WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if q.Active != nil {
		where += fmt.Sprintf(` AND is_active=$%d`, argIdx)
		args = append(args, *q.Active)
		argIdx++
	}

	var total int64
	countRow := execer.QueryRowContext(ctx, `SELECT COUNT(*) FROM billing_schemes `+where, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, kernel.Wrap(constant.CodeBillingSchemeQueryFailed, fmt.Errorf("count billing schemes: %w", err))
	}

	limit := q.Limit
	offset := (q.Page - 1) * q.Limit
	listArgs := append(append([]interface{}{}, args...), limit, offset)

	query := fmt.Sprintf(`SELECT %s FROM billing_schemes %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		billingSchemeColumns, where, argIdx, argIdx+1)

	rows, err := execer.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeBillingSchemeQueryFailed, fmt.Errorf("list billing schemes: %w", err))
	}
	defer rows.Close()

	items := make([]*entity.BillingScheme, 0)
	for rows.Next() {
		scheme, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, scheme)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(constant.CodeBillingSchemeQueryFailed, fmt.Errorf("iterate billing scheme rows: %w", err))
	}

	for _, scheme := range items {
		schemeItems, err := r.findItemsBySchemeID(ctx, scheme.ID)
		if err != nil {
			return nil, err
		}
		scheme.Items = schemeItems
	}

	return &repository.BillingSchemeListResult{Items: items, Total: total}, nil
}

func (r *PostgresBillingSchemeRepository) AddItems(ctx context.Context, schemeID string, items []*entity.BillingSchemeItem) error {
	execer := execerFromContext(ctx, r.db)

	query := `INSERT INTO billing_scheme_items (id, billing_scheme_id, fee_component_id, amount_override, is_required, sort_order, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`

	for _, item := range items {
		_, err := execer.ExecContext(ctx, query,
			item.ID, item.BillingSchemeID, item.FeeComponentID,
			nullFloat64(item.AmountOverride), item.IsRequired, item.SortOrder, item.CreatedAt,
		)
		if err != nil {
			if isUniqueViolation(err) {
				return kernel.Wrap(constant.CodeSchemeItemDuplicate, err)
			}
			return kernel.Wrap(constant.CodeBillingSchemePersistenceFailed, fmt.Errorf("add scheme item: %w", err))
		}
	}
	return nil
}

func (r *PostgresBillingSchemeRepository) RemoveItem(ctx context.Context, schemeID string, itemID string) error {
	execer := execerFromContext(ctx, r.db)

	res, err := execer.ExecContext(ctx, `DELETE FROM billing_scheme_items WHERE id=$1 AND billing_scheme_id=$2`, itemID, schemeID)
	if err != nil {
		return kernel.Wrap(constant.CodeBillingSchemePersistenceFailed, fmt.Errorf("remove scheme item: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeSchemeItemNotFound)
	}
	return nil
}

func (r *PostgresBillingSchemeRepository) findItemsBySchemeID(ctx context.Context, schemeID string) ([]*entity.BillingSchemeItem, error) {
	execer := execerFromContext(ctx, r.db)

	rows, err := execer.QueryContext(ctx,
		`SELECT id, billing_scheme_id, fee_component_id, amount_override, is_required, sort_order, created_at FROM billing_scheme_items WHERE billing_scheme_id=$1 ORDER BY sort_order ASC`,
		schemeID,
	)
	if err != nil {
		return nil, kernel.Wrap(constant.CodeBillingSchemeQueryFailed, fmt.Errorf("find items by scheme id: %w", err))
	}
	defer rows.Close()

	items := make([]*entity.BillingSchemeItem, 0)
	for rows.Next() {
		var (
			id, bsID, fcID    string
			amountOverride    sql.NullFloat64
			isRequired        bool
			sortOrder         int
			createdAt         time.Time
		)
		if err := rows.Scan(&id, &bsID, &fcID, &amountOverride, &isRequired, &sortOrder, &createdAt); err != nil {
			return nil, kernel.Wrap(constant.CodeBillingSchemeQueryFailed, fmt.Errorf("scan scheme item: %w", err))
		}
		items = append(items, &entity.BillingSchemeItem{
			ID:              id,
			BillingSchemeID: bsID,
			FeeComponentID:  fcID,
			AmountOverride:  float64FromNull(amountOverride),
			IsRequired:      isRequired,
			SortOrder:       sortOrder,
			CreatedAt:       createdAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(constant.CodeBillingSchemeQueryFailed, fmt.Errorf("iterate scheme items: %w", err))
	}
	return items, nil
}

func (r *PostgresBillingSchemeRepository) scan(sc scanner) (*entity.BillingScheme, error) {
	var (
		id, name, createdBy                                             string
		description                                                     sql.NullString
		isActive                                                        bool
		createdAt, updatedAt                                            time.Time
	)

	err := sc.Scan(&id, &name, &description, &isActive, &createdBy, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeBillingSchemeNotFound)
		}
		return nil, kernel.Wrap(constant.CodeBillingSchemeQueryFailed, fmt.Errorf("scan billing scheme: %w", err))
	}

	return &entity.BillingScheme{
		ID:          id,
		Name:        name,
		Description: strFromNull(description),
		IsActive:    isActive,
		CreatedBy:   createdBy,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

// ——— SantriBillingAssignmentRepository ———

const assignmentColumns = `
	id, santri_id, billing_scheme_id, effective_from, effective_until, assigned_by, created_at
`

type PostgresSantriBillingAssignmentRepository struct {
	db *sql.DB
}

func NewPostgresSantriBillingAssignmentRepository(db *sql.DB) *PostgresSantriBillingAssignmentRepository {
	return &PostgresSantriBillingAssignmentRepository{db: db}
}

func (r *PostgresSantriBillingAssignmentRepository) Save(ctx context.Context, a *entity.SantriBillingAssignment) error {
	execer := execerFromContext(ctx, r.db)

	query := `INSERT INTO santri_billing_assignments (` + assignmentColumns + `) VALUES (
		$1,$2,$3,$4,$5,$6,$7
	)`

	_, err := execer.ExecContext(ctx, query,
		a.ID, a.SantriID, a.BillingSchemeID, a.EffectiveFrom,
		nullTimeVal(a.EffectiveUntil), a.AssignedBy, a.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return kernel.Wrap(constant.CodeSchemeAssignmentExists, err)
		}
		return kernel.Wrap(constant.CodeBillingSchemePersistenceFailed, fmt.Errorf("save assignment: %w", err))
	}
	return nil
}

func (r *PostgresSantriBillingAssignmentRepository) FindActiveBySantriID(ctx context.Context, santriID string) (*entity.SantriBillingAssignment, error) {
	execer := execerFromContext(ctx, r.db)

	row := execer.QueryRowContext(ctx,
		`SELECT `+assignmentColumns+` FROM santri_billing_assignments WHERE santri_id=$1 AND (effective_until IS NULL OR effective_until > NOW()) ORDER BY effective_from DESC LIMIT 1`,
		santriID,
	)
	return r.scanAssignment(row)
}

func (r *PostgresSantriBillingAssignmentRepository) EndAssignment(ctx context.Context, id string, effectiveUntil time.Time) error {
	execer := execerFromContext(ctx, r.db)

	res, err := execer.ExecContext(ctx, `UPDATE santri_billing_assignments SET effective_until=$1 WHERE id=$2`, effectiveUntil, id)
	if err != nil {
		return kernel.Wrap(constant.CodeBillingSchemePersistenceFailed, fmt.Errorf("end assignment: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.New(constant.CodeBillingSchemeNotFound)
	}
	return nil
}

func (r *PostgresSantriBillingAssignmentRepository) scanAssignment(sc scanner) (*entity.SantriBillingAssignment, error) {
	var (
		id, santriID, billingSchemeID, assignedBy string
		effectiveFrom                              time.Time
		effectiveUntil                             sql.NullTime
		createdAt                                  time.Time
	)

	err := sc.Scan(&id, &santriID, &billingSchemeID, &effectiveFrom, &effectiveUntil, &assignedBy, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.New(constant.CodeBillingSchemeNotFound)
		}
		return nil, kernel.Wrap(constant.CodeBillingSchemeQueryFailed, fmt.Errorf("scan assignment: %w", err))
	}

	return &entity.SantriBillingAssignment{
		ID:              id,
		SantriID:        santriID,
		BillingSchemeID: billingSchemeID,
		EffectiveFrom:   effectiveFrom,
		EffectiveUntil:  timeFromNull(effectiveUntil),
		AssignedBy:      assignedBy,
		CreatedAt:       createdAt,
	}, nil
}
