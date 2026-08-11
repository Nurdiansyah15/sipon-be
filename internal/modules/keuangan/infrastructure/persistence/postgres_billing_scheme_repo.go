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
			return kernel.WrapMsg(constant.CodeBillingSchemeDuplicate, "Skema tagihan dengan nama yang sama sudah ada", err)
		}
		return kernel.WrapMsg(constant.CodeBillingSchemePersistenceFailed, "gagal menyimpan skema tagihan", err)
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
			return kernel.WrapMsg(constant.CodeBillingSchemeDuplicate, "Skema tagihan dengan nama yang sama sudah ada", err)
		}
		return kernel.WrapMsg(constant.CodeBillingSchemePersistenceFailed, "gagal memperbarui skema tagihan", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.WrapMsg(constant.CodeBillingSchemeNotFound, "Skema tagihan tidak ditemukan", nil)
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
		return nil, kernel.WrapMsg(constant.CodeBillingSchemeQueryFailed, "gagal menghitung jumlah skema tagihan", err)
	}

	limit := q.Limit
	offset := (q.Page - 1) * q.Limit
	listArgs := append(append([]interface{}{}, args...), limit, offset)

	query := fmt.Sprintf(`SELECT %s FROM billing_schemes %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		billingSchemeColumns, where, argIdx, argIdx+1)

	rows, err := execer.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, kernel.WrapMsg(constant.CodeBillingSchemeQueryFailed, "gagal mendaftar skema tagihan", err)
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
		return nil, kernel.WrapMsg(constant.CodeBillingSchemeQueryFailed, "gagal membaca data skema tagihan", err)
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
				return kernel.WrapMsg(constant.CodeSchemeItemDuplicate, "Item skema tagihan duplikat", err)
			}
			return kernel.WrapMsg(constant.CodeBillingSchemePersistenceFailed, "gagal menambahkan item skema tagihan", err)
		}
	}
	return nil
}

func (r *PostgresBillingSchemeRepository) RemoveItem(ctx context.Context, schemeID string, itemID string) error {
	execer := execerFromContext(ctx, r.db)

	res, err := execer.ExecContext(ctx, `DELETE FROM billing_scheme_items WHERE id=$1 AND billing_scheme_id=$2`, itemID, schemeID)
	if err != nil {
		return kernel.WrapMsg(constant.CodeBillingSchemePersistenceFailed, "gagal menghapus item skema tagihan", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.WrapMsg(constant.CodeSchemeItemNotFound, "Item skema tagihan tidak ditemukan", nil)
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
		return nil, kernel.WrapMsg(constant.CodeBillingSchemeQueryFailed, "gagal mencari item skema tagihan", err)
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
			return nil, kernel.WrapMsg(constant.CodeBillingSchemeQueryFailed, "gagal membaca data item skema tagihan", err)
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
		return nil, kernel.WrapMsg(constant.CodeBillingSchemeQueryFailed, "gagal membaca data item skema tagihan", err)
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
			return nil, kernel.WrapMsg(constant.CodeBillingSchemeNotFound, "Skema tagihan tidak ditemukan", nil)
		}
		return nil, kernel.WrapMsg(constant.CodeBillingSchemeQueryFailed, "gagal membaca data skema tagihan", err)
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
			return kernel.WrapMsg(constant.CodeSchemeAssignmentExists, "Penugasan skema tagihan sudah ada", err)
		}
		return kernel.WrapMsg(constant.CodeBillingSchemePersistenceFailed, "gagal menyimpan penugasan skema tagihan", err)
	}
	return nil
}

func (r *PostgresSantriBillingAssignmentRepository) FindActiveBySantriID(ctx context.Context, santriID string) (*entity.SantriBillingAssignment, error) {
	return r.FindActiveBySantriIDAt(ctx, santriID, time.Now())
}

// FindActiveBySantriIDAt mengambil assignment terbaru yang masih berlaku pada
// tanggal tertentu (effective_from <= atDate < effective_until). Dipakai untuk
// generate tagihan batch supaya masa berlaku skema dievaluasi terhadap
// issued_date yang diinput, bukan selalu waktu server.
func (r *PostgresSantriBillingAssignmentRepository) FindActiveBySantriIDAt(ctx context.Context, santriID string, atDate time.Time) (*entity.SantriBillingAssignment, error) {
	execer := execerFromContext(ctx, r.db)

	row := execer.QueryRowContext(ctx,
		`SELECT `+assignmentColumns+` FROM santri_billing_assignments
		 WHERE santri_id=$1 AND effective_from <= $2 AND (effective_until IS NULL OR effective_until > $2)
		 ORDER BY effective_from DESC LIMIT 1`,
		santriID, atDate,
	)
	return r.scanAssignment(row)
}

func (r *PostgresSantriBillingAssignmentRepository) FindByID(ctx context.Context, id string) (*entity.SantriBillingAssignment, error) {
	execer := execerFromContext(ctx, r.db)

	row := execer.QueryRowContext(ctx,
		`SELECT `+assignmentColumns+` FROM santri_billing_assignments WHERE id=$1`,
		id,
	)
	return r.scanAssignment(row)
}

func (r *PostgresSantriBillingAssignmentRepository) Update(ctx context.Context, a *entity.SantriBillingAssignment) error {
	execer := execerFromContext(ctx, r.db)

	res, err := execer.ExecContext(ctx,
		`UPDATE santri_billing_assignments
		 SET billing_scheme_id=$2, effective_from=$3, effective_until=$4
		 WHERE id=$1`,
		a.ID, a.BillingSchemeID, a.EffectiveFrom, nullTimeVal(a.EffectiveUntil),
	)
	if err != nil {
		return kernel.WrapMsg(constant.CodeBillingSchemePersistenceFailed, "gagal memperbarui penugasan skema tagihan", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.WrapMsg(constant.CodeBillingSchemeNotFound, "Skema tagihan tidak ditemukan", nil)
	}
	return nil
}

func (r *PostgresSantriBillingAssignmentRepository) EndAssignment(ctx context.Context, id string, effectiveUntil time.Time) error {
	execer := execerFromContext(ctx, r.db)

	res, err := execer.ExecContext(ctx, `UPDATE santri_billing_assignments SET effective_until=$1 WHERE id=$2`, effectiveUntil, id)
	if err != nil {
		return kernel.WrapMsg(constant.CodeBillingSchemePersistenceFailed, "gagal mengakhiri penugasan skema tagihan", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return kernel.WrapMsg(constant.CodeBillingSchemeNotFound, "Skema tagihan tidak ditemukan", nil)
	}
	return nil
}

// ListBySantriID mengambil seluruh penugasan (aktif maupun riwayat) milik satu
// santri, diurutkan dari tanggal berlaku terbaru ke terlama.
func (r *PostgresSantriBillingAssignmentRepository) ListBySantriID(ctx context.Context, santriID string) ([]*entity.SantriBillingAssignment, error) {
	execer := execerFromContext(ctx, r.db)

	rows, err := execer.QueryContext(ctx,
		`SELECT `+assignmentColumns+` FROM santri_billing_assignments
		 WHERE santri_id=$1
		 ORDER BY effective_from DESC`,
		santriID,
	)
	if err != nil {
		return nil, kernel.WrapMsg(constant.CodeBillingSchemeQueryFailed, "gagal membaca riwayat penugasan skema tagihan", err)
	}
	defer rows.Close()

	results := make([]*entity.SantriBillingAssignment, 0)
	for rows.Next() {
		a, err := r.scanAssignment(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, a)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(constant.CodeBillingSchemeQueryFailed, "gagal membaca riwayat penugasan skema tagihan", err)
	}
	return results, nil
}

// HasOverlappingAssignment melaporkan apakah sudah ada penugasan lain untuk
// santri yang sama yang rentang berlakunya tumpang-tindih dengan rentang
// [from, until) yang diberikan. excludeID dipakai untuk mengecualikan
// assignment itu sendiri saat edit. until nil berarti rentang terbuka.
func (r *PostgresSantriBillingAssignmentRepository) HasOverlappingAssignment(ctx context.Context, santriID string, from time.Time, until *time.Time, excludeID string) (bool, error) {
	execer := execerFromContext(ctx, r.db)

	var exists bool
	err := execer.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM santri_billing_assignments
			WHERE santri_id=$1
			  AND ($4::text = '' OR id != $4::uuid)
			  AND effective_from < COALESCE($2::date, 'infinity'::date)
			  AND (effective_until IS NULL OR effective_until > $3)
		)`,
		santriID, from, nullTimeVal(until), excludeID,
	).Scan(&exists)
	if err != nil {
		return false, kernel.WrapMsg(constant.CodeBillingSchemeQueryFailed, "gagal memeriksa tumpang-tindih penugasan skema tagihan", err)
	}
	return exists, nil
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
			return nil, kernel.WrapMsg(constant.CodeBillingSchemeNotFound, "Skema tagihan tidak ditemukan", nil)
		}
		return nil, kernel.WrapMsg(constant.CodeBillingSchemeQueryFailed, "gagal membaca data penugasan skema tagihan", err)
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
