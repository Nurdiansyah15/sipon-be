package persistence

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	scopeconstant "sipon-be/internal/modules/system/domain/scope/constant"
	scopesentity "sipon-be/internal/modules/system/domain/scope/entity"
	scopesvo "sipon-be/internal/modules/system/domain/scope/valueobject"
	"sipon-be/internal/shared/kernel"
)

type PostgresScopeRepository struct {
	db *sql.DB
}

func NewPostgresScopeRepository(db *sql.DB) *PostgresScopeRepository {
	return &PostgresScopeRepository{db: db}
}

func (r *PostgresScopeRepository) Save(ctx context.Context, scope *scopesentity.Scope) error {
	query := `INSERT INTO scopes (id, scope_type, code, name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, query,
		scope.ID, string(scope.ScopeType), scope.Code, scope.Name,
		nullStr(scope.Description), scope.IsActive,
		scope.CreatedAt, scope.UpdatedAt,
	)
	if err != nil {
		return kernel.WrapMsg(scopeconstant.ErrCodeScopeInternal, "gagal menyimpan scope", err)
	}
	return nil
}

func (r *PostgresScopeRepository) Update(ctx context.Context, scope *scopesentity.Scope) error {
	query := `UPDATE scopes SET name = $1, description = $2, is_active = $3, updated_at = $4 WHERE id = $5`
	_, err := r.db.ExecContext(ctx, query,
		scope.Name, nullStr(scope.Description), scope.IsActive, scope.UpdatedAt, scope.ID,
	)
	if err != nil {
		return kernel.WrapMsg(scopeconstant.ErrCodeScopeInternal, "gagal memperbarui scope", err)
	}
	return nil
}

func (r *PostgresScopeRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM scopes WHERE id = $1`, id)
	if err != nil {
		return kernel.WrapMsg(scopeconstant.ErrCodeScopeInternal, "gagal menghapus scope", err)
	}
	return nil
}

func (r *PostgresScopeRepository) FindByID(ctx context.Context, id string) (*scopesentity.Scope, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, scope_type, code, name, description, is_active, created_at, updated_at FROM scopes WHERE id = $1`, id)

	var m ScopeModel
	if err := row.Scan(&m.ID, &m.ScopeType, &m.Code, &m.Name, &m.Description, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.WrapMsg(scopeconstant.ErrCodeScopeNotFound, "Scope tidak ditemukan", nil)
		}
		return nil, kernel.WrapMsg(scopeconstant.ErrCodeScopeInternal, "gagal mencari scope berdasarkan ID", err)
	}
	return scopeFromModel(m), nil
}

func (r *PostgresScopeRepository) FindByTypeAndCode(ctx context.Context, scopeType, code string) (*scopesentity.Scope, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, scope_type, code, name, description, is_active, created_at, updated_at FROM scopes WHERE scope_type = $1 AND code = $2`, strings.ToLower(scopeType), strings.ToLower(code))

	var m ScopeModel
	if err := row.Scan(&m.ID, &m.ScopeType, &m.Code, &m.Name, &m.Description, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, kernel.WrapMsg(scopeconstant.ErrCodeScopeNotFound, "Scope tidak ditemukan", nil)
		}
		return nil, kernel.WrapMsg(scopeconstant.ErrCodeScopeInternal, "gagal mencari scope berdasarkan jenis dan kode", err)
	}
	return scopeFromModel(m), nil
}

func (r *PostgresScopeRepository) ListByType(ctx context.Context, scopeType string, includeInactive bool) ([]*scopesentity.Scope, error) {
	query := `SELECT id, scope_type, code, name, description, is_active, created_at, updated_at FROM scopes WHERE scope_type = $1`
	args := []any{strings.ToLower(scopeType)}
	if !includeInactive {
		query += ` AND is_active = true`
	}
	query += ` ORDER BY code`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, kernel.WrapMsg(scopeconstant.ErrCodeScopeInternal, "gagal mendaftar scope", err)
	}
	defer rows.Close()
	return scanScopes(rows)
}

func (r *PostgresScopeRepository) ListAll(ctx context.Context, includeInactive bool) ([]*scopesentity.Scope, error) {
	query := `SELECT id, scope_type, code, name, description, is_active, created_at, updated_at FROM scopes`
	if !includeInactive {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY scope_type, code`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, kernel.WrapMsg(scopeconstant.ErrCodeScopeInternal, "gagal mendaftar scope", err)
	}
	defer rows.Close()
	return scanScopes(rows)
}

func (r *PostgresScopeRepository) ListActiveCodesByType(ctx context.Context, scopeType string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT code FROM scopes WHERE scope_type = $1 AND is_active = true ORDER BY code`, strings.ToLower(scopeType))
	if err != nil {
		return nil, kernel.WrapMsg(scopeconstant.ErrCodeScopeInternal, "gagal membaca kode scope master", err)
	}
	defer rows.Close()

	codes := make([]string, 0)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, kernel.WrapMsg(scopeconstant.ErrCodeScopeInternal, "gagal membaca kode scope", err)
		}
		codes = append(codes, code)
	}
	return codes, rows.Err()
}

func scopeFromModel(m ScopeModel) *scopesentity.Scope {
	var desc *string
	if m.Description.Valid {
		desc = &m.Description.String
	}
	return &scopesentity.Scope{
		ID:          m.ID,
		ScopeType:   scopesvo.ScopeType(m.ScopeType),
		Code:        m.Code,
		Name:        m.Name,
		Description: desc,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func scanScopes(rows *sql.Rows) ([]*scopesentity.Scope, error) {
	var scopes []*scopesentity.Scope
	for rows.Next() {
		var m ScopeModel
		if err := rows.Scan(&m.ID, &m.ScopeType, &m.Code, &m.Name, &m.Description, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, kernel.WrapMsg(scopeconstant.ErrCodeScopeInternal, "gagal membaca data scope", err)
		}
		scopes = append(scopes, scopeFromModel(m))
	}
	return scopes, rows.Err()
}

func nullStr(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}
