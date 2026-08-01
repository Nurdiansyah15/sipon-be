package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"sipon-be/internal/modules/identity/domain"
)

type PostgresQueryRepo struct {
	db *sql.DB
}

func NewPostgresQueryRepo(db *sql.DB) *PostgresQueryRepo {
	return &PostgresQueryRepo{db: db}
}

func (r *PostgresQueryRepo) List(ctx context.Context, status string, roleID string, search string, sortBy string, sortType string, page, limit int) ([]*domain.User, int64, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("u.status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}

	if search != "" {
		conditions = append(conditions, fmt.Sprintf("(u.username ILIKE $%d OR u.email ILIKE $%d OR u.fullname ILIKE $%d)", argIdx, argIdx+1, argIdx+2))
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
		argIdx += 3
	}

	conditions = append(conditions, "u.deleted_at IS NULL")

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	var joins string
	if roleID != "" {
		joins = fmt.Sprintf(" INNER JOIN user_roles ur ON ur.user_id = u.id AND ur.role_id = $%d AND ur.is_active = true", argIdx)
		args = append(args, roleID)
		argIdx++
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(DISTINCT u.id) FROM users u%s %s", joins, whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	allowedSortColumns := map[string]string{
		"username":   "u.username",
		"email":      "u.email",
		"status":     "u.status",
		"created_at": "u.created_at",
	}
	orderColumn, ok := allowedSortColumns[sortBy]
	if !ok {
		orderColumn = "u.created_at"
	}
	orderDir := "DESC"
	if sortType == "asc" || sortType == "ASC" {
		orderDir = "ASC"
	}

	offset := (page - 1) * limit
	dataQuery := fmt.Sprintf(`SELECT DISTINCT u.id, u.username, u.fullname, u.email, u.phone, u.avatar_key, u.status, u.created_at, u.updated_at, u.last_login_at, u.deleted_at, u.failed_login_attempts, u.locked_until FROM users u%s %s ORDER BY %s %s LIMIT $%d OFFSET $%d`, joins, whereClause, orderColumn, orderDir, argIdx, argIdx+1)
	dataArgs := append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var m UserModel
		if err := rows.Scan(&m.ID, &m.Username, &m.Fullname, &m.Email, &m.Phone, &m.AvatarKey, &m.Status, &m.CreatedAt, &m.UpdatedAt, &m.LastLoginAt, &m.DeletedAt, &m.FailedLoginAttempts, &m.LockedUntil); err != nil {
			return nil, 0, fmt.Errorf("scan user row: %w", err)
		}

		user, err := userFromModel(m)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, total, rows.Err()
}

func (r *PostgresQueryRepo) FindByIDWithRoles(ctx context.Context, userID string) (*domain.User, []string, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, username, fullname, email, phone, avatar_key, status, created_at, updated_at, last_login_at, deleted_at, failed_login_attempts, locked_until FROM users WHERE id = $1 AND deleted_at IS NULL`, userID)

	var m UserModel
	if err := scanUser(row, &m); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, fmt.Errorf("user not found")
		}
		return nil, nil, fmt.Errorf("find user by id: %w", err)
	}

	user, err := userFromModel(m)
	if err != nil {
		return nil, nil, err
	}

	roleRows, err := r.db.QueryContext(ctx, `SELECT r.name FROM roles r INNER JOIN user_roles ur ON ur.role_id = r.id WHERE ur.user_id = $1 AND ur.is_active = true`, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("find user roles: %w", err)
	}
	defer roleRows.Close()

	var roles []string
	for roleRows.Next() {
		var name string
		if err := roleRows.Scan(&name); err != nil {
			return nil, nil, fmt.Errorf("scan role name: %w", err)
		}
		roles = append(roles, name)
	}

	return user, roles, roleRows.Err()
}

type PostgresRoleListRepo struct {
	db *sql.DB
}

func NewPostgresRoleListRepo(db *sql.DB) *PostgresRoleListRepo {
	return &PostgresRoleListRepo{db: db}
}

func (r *PostgresRoleListRepo) List(ctx context.Context, roleType string, scopeType string, assignable *bool, sortBy string, sortType string, page, limit int) ([]*domain.Role, int64, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if roleType != "" {
		conditions = append(conditions, fmt.Sprintf("role_type = $%d", argIdx))
		args = append(args, roleType)
		argIdx++
	}

	if scopeType != "" {
		conditions = append(conditions, fmt.Sprintf("scope_type = $%d", argIdx))
		args = append(args, scopeType)
		argIdx++
	}

	if assignable != nil {
		conditions = append(conditions, fmt.Sprintf("assignable = $%d", argIdx))
		args = append(args, *assignable)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM roles %s", whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count roles: %w", err)
	}

	roleAllowedSort := map[string]string{
		"name":       "name",
		"role_type":  "role_type",
		"scope_type": "scope_type",
		"created_at": "created_at",
	}
	roleOrderCol, ok := roleAllowedSort[sortBy]
	if !ok {
		roleOrderCol = "created_at"
	}
	roleOrderDir := "DESC"
	if sortType == "asc" || sortType == "ASC" {
		roleOrderDir = "ASC"
	}

	offset := (page - 1) * limit
	dataQuery := fmt.Sprintf(`SELECT id, name, display_name, description, role_type, scope_type, assignable, created_at, updated_at FROM roles %s ORDER BY %s %s LIMIT $%d OFFSET $%d`, whereClause, roleOrderCol, roleOrderDir, argIdx, argIdx+1)
	dataArgs := append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list roles: %w", err)
	}
	defer rows.Close()

	roles, err := scanRoles(rows)
	if err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

type PostgresUserRoleListRepo struct {
	db *sql.DB
}

func NewPostgresUserRoleListRepo(db *sql.DB) *PostgresUserRoleListRepo {
	return &PostgresUserRoleListRepo{db: db}
}

func (r *PostgresUserRoleListRepo) List(ctx context.Context, userID, roleID, scopeType, scopeID string, isActive *bool, sortBy string, sortType string, page, limit int) ([]*domain.UserRole, int64, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if userID != "" {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, userID)
		argIdx++
	}

	if roleID != "" {
		conditions = append(conditions, fmt.Sprintf("role_id = $%d", argIdx))
		args = append(args, roleID)
		argIdx++
	}

	if scopeType != "" {
		conditions = append(conditions, fmt.Sprintf("scope_type = $%d", argIdx))
		args = append(args, scopeType)
		argIdx++
	}

	if scopeID != "" {
		conditions = append(conditions, fmt.Sprintf("scope_id = $%d", argIdx))
		args = append(args, scopeID)
		argIdx++
	}

	if isActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *isActive)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM user_roles %s", whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count user roles: %w", err)
	}

	urAllowedSort := map[string]string{
		"assigned_at": "assigned_at",
		"user_id":     "user_id",
		"role_id":     "role_id",
	}
	urOrderCol, ok := urAllowedSort[sortBy]
	if !ok {
		urOrderCol = "assigned_at"
	}
	urOrderDir := "DESC"
	if sortType == "asc" || sortType == "ASC" {
		urOrderDir = "ASC"
	}

	offset := (page - 1) * limit
	dataQuery := fmt.Sprintf(`SELECT id, user_id, role_id, scope_type, scope_id, assigned_at, assigned_by, expired_at, is_active, notes, deactivated_at FROM user_roles %s ORDER BY %s %s LIMIT $%d OFFSET $%d`, whereClause, urOrderCol, urOrderDir, argIdx, argIdx+1)
	dataArgs := append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list user roles: %w", err)
	}
	defer rows.Close()

	userRoles, err := scanUserRoles(rows)
	if err != nil {
		return nil, 0, err
	}

	return userRoles, total, nil
}
