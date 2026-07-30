package domain

import "context"

type UserRepository interface {
	Save(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByLoginIdentifier(ctx context.Context, identifier LoginIdentifier) (*User, error)
	FindByIdentity(ctx context.Context, kind LoginIdentifierKind, value string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByLoginIdentity(ctx context.Context, kind LoginIdentifierKind, value string) (bool, error)
	UpdateUsername(ctx context.Context, userID, newUsername string) error
}

type VerificationRepository interface {
	Save(ctx context.Context, code *VerificationCode) error
	FindLatestByUserAndPurpose(ctx context.Context, userID string, purpose CodePurpose) (*VerificationCode, error)
	Update(ctx context.Context, code *VerificationCode) error
}

type RoleRepository interface {
	Save(ctx context.Context, role *Role) error
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*Role, error)
	FindByName(ctx context.Context, name RoleName) (*Role, error)
	ListByType(ctx context.Context, roleType RoleType) ([]*Role, error)
}

type UserRoleRepository interface {
	Save(ctx context.Context, userRole *UserRole) error
	Update(ctx context.Context, userRole *UserRole) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*UserRole, error)
	FindActiveByUserID(ctx context.Context, userID string) ([]*UserRole, error)
	ListActiveUserIDsByRoleName(ctx context.Context, roleName RoleName) ([]string, error)
}

type RolePermissionRepository interface {
	Save(ctx context.Context, rp *RolePermission) error
	Delete(ctx context.Context, roleID string, permissionKey PermissionKey) error
	ListByRoleID(ctx context.Context, roleID string) ([]*RolePermission, error)
}

type RoleScopeRepository interface {
	Save(ctx context.Context, scope *RoleScope) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*RoleScope, error)
	FindByRoleID(ctx context.Context, roleID string) ([]*RoleScope, error)
}
