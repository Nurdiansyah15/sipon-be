package repository

import (
	"context"

	"sipon-be/internal/modules/identity/domain/role/constant"
	"sipon-be/internal/modules/identity/domain/role/entity"
)

type RoleRepository interface {
	Save(ctx context.Context, role *entity.Role) error
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*entity.Role, error)
	FindByName(ctx context.Context, name constant.RoleName) (*entity.Role, error)
	ListByType(ctx context.Context, roleType constant.RoleType) ([]*entity.Role, error)
}

type UserRoleRepository interface {
	Save(ctx context.Context, userRole *entity.UserRole) error
	Update(ctx context.Context, userRole *entity.UserRole) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*entity.UserRole, error)
	FindActiveByUserID(ctx context.Context, userID string) ([]*entity.UserRole, error)
	ListActiveUserIDsByRoleName(ctx context.Context, roleName constant.RoleName) ([]string, error)
}

type RolePermissionRepository interface {
	Save(ctx context.Context, rp *entity.RolePermission) error
	Delete(ctx context.Context, roleID string, permissionKey constant.PermissionKey) error
	ListByRoleID(ctx context.Context, roleID string) ([]*entity.RolePermission, error)
}

type RoleScopeRepository interface {
	Save(ctx context.Context, scope *entity.RoleScope) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*entity.RoleScope, error)
	FindByRoleID(ctx context.Context, roleID string) ([]*entity.RoleScope, error)
}
