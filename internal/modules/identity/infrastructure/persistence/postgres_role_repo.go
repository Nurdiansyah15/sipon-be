package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	roleconstant "sipon-be/internal/modules/identity/domain/role/constant"
	roleentity "sipon-be/internal/modules/identity/domain/role/entity"
	rolevo "sipon-be/internal/modules/identity/domain/role/valueobject"
	"sipon-be/internal/shared/kernel"
)

type PostgresRoleRepository struct {
	db *sql.DB
}

func NewPostgresRoleRepository(db *sql.DB) *PostgresRoleRepository {
	return &PostgresRoleRepository{db: db}
}

func (r *PostgresRoleRepository) Save(ctx context.Context, role *roleentity.Role) error {
	query := `INSERT INTO roles (id, name, display_name, description, role_type, scope_type, assignable, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query,
		role.ID, string(role.Name), role.DisplayName, descStr(role.Description),
		string(role.RoleType), string(role.ScopeType), role.Assignable,
		role.CreatedAt, role.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save role: %w", err)
	}
	return nil
}

func (r *PostgresRoleRepository) Update(ctx context.Context, role *roleentity.Role) error {
	query := `UPDATE roles SET display_name = $1, description = $2, role_type = $3, scope_type = $4, assignable = $5, updated_at = $6 WHERE id = $7`

	_, err := r.db.ExecContext(ctx, query,
		role.DisplayName, descStr(role.Description),
		string(role.RoleType), string(role.ScopeType), role.Assignable,
		role.UpdatedAt, role.ID,
	)
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}
	return nil
}

func (r *PostgresRoleRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM roles WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete role: %w", err)
	}
	return nil
}

func (r *PostgresRoleRepository) FindByID(ctx context.Context, id string) (*roleentity.Role, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, name, display_name, description, role_type, scope_type, assignable, created_at, updated_at FROM roles WHERE id = $1`, id)

	var m RoleModel
	if err := row.Scan(&m.ID, &m.Name, &m.DisplayName, &m.Description, &m.RoleType, &m.ScopeType, &m.Assignable, &m.CreatedAt, &m.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, kernel.New(roleconstant.ErrCodeRoleNotFound)
		}
		return nil, fmt.Errorf("find role by id: %w", err)
	}

	return roleFromModel(m), nil
}

func (r *PostgresRoleRepository) FindByName(ctx context.Context, name roleconstant.RoleName) (*roleentity.Role, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, name, display_name, description, role_type, scope_type, assignable, created_at, updated_at FROM roles WHERE name = $1`, string(name))

	var m RoleModel
	if err := row.Scan(&m.ID, &m.Name, &m.DisplayName, &m.Description, &m.RoleType, &m.ScopeType, &m.Assignable, &m.CreatedAt, &m.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, kernel.New(roleconstant.ErrCodeRoleNotFound)
		}
		return nil, fmt.Errorf("find role by name: %w", err)
	}

	return roleFromModel(m), nil
}

func (r *PostgresRoleRepository) ListByType(ctx context.Context, roleType roleconstant.RoleType) ([]*roleentity.Role, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, display_name, description, role_type, scope_type, assignable, created_at, updated_at FROM roles WHERE role_type = $1 ORDER BY name`, string(roleType))
	if err != nil {
		return nil, fmt.Errorf("list roles by type: %w", err)
	}
	defer rows.Close()

	return scanRoles(rows)
}

func roleFromModel(m RoleModel) *roleentity.Role {
	var desc *string
	if m.Description.Valid {
		desc = &m.Description.String
	}

	return &roleentity.Role{
		ID:          m.ID,
		Name:        roleconstant.RoleName(m.Name),
		DisplayName: m.DisplayName,
		Description: desc,
		RoleType:    roleconstant.RoleType(m.RoleType),
		ScopeType:   roleconstant.ScopeType(m.ScopeType),
		Assignable:  m.Assignable,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func scanRoles(rows *sql.Rows) ([]*roleentity.Role, error) {
	var roles []*roleentity.Role
	for rows.Next() {
		var m RoleModel
		if err := rows.Scan(&m.ID, &m.Name, &m.DisplayName, &m.Description, &m.RoleType, &m.ScopeType, &m.Assignable, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, roleFromModel(m))
	}
	return roles, rows.Err()
}

type PostgresUserRoleRepository struct {
	db *sql.DB
}

func NewPostgresUserRoleRepository(db *sql.DB) *PostgresUserRoleRepository {
	return &PostgresUserRoleRepository{db: db}
}

func (r *PostgresUserRoleRepository) Save(ctx context.Context, userRole *roleentity.UserRole) error {
	query := `INSERT INTO user_roles (id, user_id, role_id, scope_type, scope_id, assigned_at, assigned_by, expired_at, is_active, notes, deactivated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.ExecContext(ctx, query,
		userRole.ID, userRole.UserID, userRole.RoleID,
		string(userRole.ScopeType), strPtr(userRole.ScopeID),
		userRole.AssignedAt, userRole.AssignedBy,
		timePtrNull(userRole.ExpiredAt), userRole.IsActive,
		strPtr(userRole.Notes), timePtrNull(userRole.DeactivatedAt),
	)
	if err != nil {
		return fmt.Errorf("save user role: %w", err)
	}
	return nil
}

func (r *PostgresUserRoleRepository) Update(ctx context.Context, userRole *roleentity.UserRole) error {
	query := `UPDATE user_roles SET scope_type = $1, scope_id = $2, expired_at = $3, is_active = $4, notes = $5, deactivated_at = $6 WHERE id = $7`

	_, err := r.db.ExecContext(ctx, query,
		string(userRole.ScopeType), strPtr(userRole.ScopeID),
		timePtrNull(userRole.ExpiredAt), userRole.IsActive,
		strPtr(userRole.Notes), timePtrNull(userRole.DeactivatedAt),
		userRole.ID,
	)
	if err != nil {
		return fmt.Errorf("update user role: %w", err)
	}
	return nil
}

func (r *PostgresUserRoleRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM user_roles WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete user role: %w", err)
	}
	return nil
}

func (r *PostgresUserRoleRepository) FindByID(ctx context.Context, id string) (*roleentity.UserRole, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, user_id, role_id, scope_type, scope_id, assigned_at, assigned_by, expired_at, is_active, notes, deactivated_at FROM user_roles WHERE id = $1`, id)

	var m UserRoleModel
	if err := row.Scan(&m.ID, &m.UserID, &m.RoleID, &m.ScopeType, &m.ScopeID, &m.AssignedAt, &m.AssignedBy, &m.ExpiredAt, &m.IsActive, &m.Notes, &m.DeactivatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, kernel.New(roleconstant.ErrCodeUserRoleNotActive)
		}
		return nil, fmt.Errorf("find user role by id: %w", err)
	}

	return userRoleFromModel(m), nil
}

func (r *PostgresUserRoleRepository) FindActiveByUserID(ctx context.Context, userID string) ([]*roleentity.UserRole, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, user_id, role_id, scope_type, scope_id, assigned_at, assigned_by, expired_at, is_active, notes, deactivated_at FROM user_roles WHERE user_id = $1 AND is_active = true`, userID)
	if err != nil {
		return nil, fmt.Errorf("find active user roles: %w", err)
	}
	defer rows.Close()

	return scanUserRoles(rows)
}

func (r *PostgresUserRoleRepository) ListActiveUserIDsByRoleName(ctx context.Context, roleName roleconstant.RoleName) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT ur.user_id FROM user_roles ur INNER JOIN roles r ON r.id = ur.role_id WHERE r.name = $1 AND ur.is_active = true`, string(roleName))
	if err != nil {
		return nil, fmt.Errorf("list active user ids by role name: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan user id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func userRoleFromModel(m UserRoleModel) *roleentity.UserRole {
	var scopeID *string
	if m.ScopeID.Valid {
		scopeID = &m.ScopeID.String
	}

	var notes *string
	if m.Notes.Valid {
		notes = &m.Notes.String
	}

	return &roleentity.UserRole{
		ID:            m.ID,
		UserID:        m.UserID,
		RoleID:        m.RoleID,
		ScopeType:     roleconstant.ScopeType(m.ScopeType),
		ScopeID:       scopeID,
		AssignedAt:    m.AssignedAt,
		AssignedBy:    m.AssignedBy,
		ExpiredAt:     timePtr(m.ExpiredAt),
		IsActive:      m.IsActive,
		Notes:         notes,
		DeactivatedAt: timePtr(m.DeactivatedAt),
	}
}

func scanUserRoles(rows *sql.Rows) ([]*roleentity.UserRole, error) {
	var roles []*roleentity.UserRole
	for rows.Next() {
		var m UserRoleModel
		if err := rows.Scan(&m.ID, &m.UserID, &m.RoleID, &m.ScopeType, &m.ScopeID, &m.AssignedAt, &m.AssignedBy, &m.ExpiredAt, &m.IsActive, &m.Notes, &m.DeactivatedAt); err != nil {
			return nil, fmt.Errorf("scan user role: %w", err)
		}
		roles = append(roles, userRoleFromModel(m))
	}
	return roles, rows.Err()
}

type PostgresRolePermissionRepository struct {
	db *sql.DB
}

func NewPostgresRolePermissionRepository(db *sql.DB) *PostgresRolePermissionRepository {
	return &PostgresRolePermissionRepository{db: db}
}

func (r *PostgresRolePermissionRepository) Save(ctx context.Context, rp *roleentity.RolePermission) error {
	query := `INSERT INTO role_permissions (id, role_id, permission_key, assigned_at, assigned_by, notes)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		rp.ID, rp.RoleID, string(rp.PermissionKey),
		rp.AssignedAt, rp.AssignedBy, strPtr(rp.Notes),
	)
	if err != nil {
		return fmt.Errorf("save role permission: %w", err)
	}
	return nil
}

func (r *PostgresRolePermissionRepository) Delete(ctx context.Context, roleID string, permissionKey roleconstant.PermissionKey) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = $1 AND permission_key = $2`, roleID, string(permissionKey))
	if err != nil {
		return fmt.Errorf("delete role permission: %w", err)
	}
	return nil
}

func (r *PostgresRolePermissionRepository) ListByRoleID(ctx context.Context, roleID string) ([]*roleentity.RolePermission, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, role_id, permission_key, assigned_at, assigned_by, notes FROM role_permissions WHERE role_id = $1`, roleID)
	if err != nil {
		return nil, fmt.Errorf("list role permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*roleentity.RolePermission
	for rows.Next() {
		var m RolePermissionModel
		if err := rows.Scan(&m.ID, &m.RoleID, &m.PermissionKey, &m.AssignedAt, &m.AssignedBy, &m.Notes); err != nil {
			return nil, fmt.Errorf("scan role permission: %w", err)
		}
		var notes *string
		if m.Notes.Valid {
			notes = &m.Notes.String
		}
		permissions = append(permissions, &roleentity.RolePermission{
			ID:            m.ID,
			RoleID:        m.RoleID,
			PermissionKey: roleconstant.PermissionKey(m.PermissionKey),
			AssignedAt:    m.AssignedAt,
			AssignedBy:    m.AssignedBy,
			Notes:         notes,
		})
	}
	return permissions, rows.Err()
}

type PostgresRoleScopeRepository struct {
	db *sql.DB
}

func NewPostgresRoleScopeRepository(db *sql.DB) *PostgresRoleScopeRepository {
	return &PostgresRoleScopeRepository{db: db}
}

func (r *PostgresRoleScopeRepository) Save(ctx context.Context, scope *roleentity.RoleScope) error {
	query := `INSERT INTO role_scopes (id, role_id, scope_type, scope_value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		scope.ID, scope.RoleID, string(scope.ScopeType), scope.ScopeValue,
		scope.CreatedAt, scope.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save role scope: %w", err)
	}
	return nil
}

func (r *PostgresRoleScopeRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM role_scopes WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete role scope: %w", err)
	}
	return nil
}

func (r *PostgresRoleScopeRepository) FindByID(ctx context.Context, id string) (*roleentity.RoleScope, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, role_id, scope_type, scope_value, created_at, updated_at FROM role_scopes WHERE id = $1`, id)

	var m RoleScopeModel
	if err := row.Scan(&m.ID, &m.RoleID, &m.ScopeType, &m.ScopeValue, &m.CreatedAt, &m.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, kernel.New(rolevo.ErrCodeInvalidScopeType)
		}
		return nil, fmt.Errorf("find role scope by id: %w", err)
	}

	return &roleentity.RoleScope{
		ID:         m.ID,
		RoleID:     m.RoleID,
		ScopeType:  rolevo.RoleScopeType(m.ScopeType),
		ScopeValue: m.ScopeValue,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}, nil
}

func (r *PostgresRoleScopeRepository) FindByRoleID(ctx context.Context, roleID string) ([]*roleentity.RoleScope, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, role_id, scope_type, scope_value, created_at, updated_at FROM role_scopes WHERE role_id = $1`, roleID)
	if err != nil {
		return nil, fmt.Errorf("find role scopes: %w", err)
	}
	defer rows.Close()

	var scopes []*roleentity.RoleScope
	for rows.Next() {
		var m RoleScopeModel
		if err := rows.Scan(&m.ID, &m.RoleID, &m.ScopeType, &m.ScopeValue, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan role scope: %w", err)
		}
		scopes = append(scopes, &roleentity.RoleScope{
			ID:         m.ID,
			RoleID:     m.RoleID,
			ScopeType:  rolevo.RoleScopeType(m.ScopeType),
			ScopeValue: m.ScopeValue,
			CreatedAt:  m.CreatedAt,
			UpdatedAt:  m.UpdatedAt,
		})
	}
	return scopes, rows.Err()
}

func descStr(desc *string) interface{} {
	if desc == nil {
		return nil
	}
	return *desc
}

func strPtr(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

func timePtrNull(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}
