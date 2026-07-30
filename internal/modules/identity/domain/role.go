package domain

import (
	"sipon-be/internal/shared/kernel"
	"time"
)

type RoleName string

const (
	UserGodRoleName    RoleName = "usergod"
	SuperAdminRoleName RoleName = "superadmin"
	AdminRoleName      RoleName = "admin"
	MemberRoleName     RoleName = "member"
)

type RoleType string

const (
	RoleTypeSystem RoleType = "system"
	RoleTypeCustom RoleType = "custom"
)

type ScopeType string

const (
	ScopeTypeGlobal    ScopeType = "global"
	ScopeTypeRegion    ScopeType = "region"
	ScopeTypeCommunity ScopeType = "community"
)

type PermissionKey string

const (
	PermissionManageSystemSettings  PermissionKey = "manage_system_settings"
	PermissionAssignRole            PermissionKey = "assign_role"
	PermissionManageUsers           PermissionKey = "manage_users"
	PermissionResetUserPassword     PermissionKey = "reset_user_password"
	PermissionDeactivateUser        PermissionKey = "deactivate_user"
	PermissionManageRoles           PermissionKey = "manage_roles"
	PermissionManageRolePermissions PermissionKey = "manage_role_permissions"
)

const (
	ErrCodeRoleNotFound            kernel.Code = "ROLE_NOT_FOUND"
	ErrCodeRoleNotAssignable       kernel.Code = "ROLE_NOT_ASSIGNABLE"
	ErrCodeRoleCannotDeleteSystem  kernel.Code = "ROLE_CANNOT_DELETE_SYSTEM"
	ErrCodeRoleScopeMismatch       kernel.Code = "ROLE_SCOPE_MISMATCH"
	ErrCodeUserRoleAlreadyAssigned kernel.Code = "USER_ROLE_ALREADY_ASSIGNED"
	ErrCodeUserRoleExpired         kernel.Code = "USER_ROLE_EXPIRED"
	ErrCodeUserRoleNotActive       kernel.Code = "USER_ROLE_NOT_ACTIVE"
	ErrCodeInvalidPermissionKey    kernel.Code = "INVALID_PERMISSION_KEY"
	ErrCodeInvalidScopeType        kernel.Code = "INVALID_SCOPE_TYPE"
	ErrCodeInvalidRoleType         kernel.Code = "INVALID_ROLE_TYPE"
)

type PermissionDefinition struct {
	Key         PermissionKey
	DisplayName string
	Description string
}

var AllPermissionDefinitions = []PermissionDefinition{
	{Key: PermissionManageSystemSettings, DisplayName: "Manage System Settings", Description: "Allows managing system-wide settings"},
	{Key: PermissionAssignRole, DisplayName: "Assign Role", Description: "Allows assigning roles to users"},
	{Key: PermissionManageUsers, DisplayName: "Manage Users", Description: "Allows managing user accounts"},
	{Key: PermissionResetUserPassword, DisplayName: "Reset User Password", Description: "Allows resetting user passwords"},
	{Key: PermissionDeactivateUser, DisplayName: "Deactivate User", Description: "Allows deactivating user accounts"},
	{Key: PermissionManageRoles, DisplayName: "Manage Roles", Description: "Allows managing role definitions"},
	{Key: PermissionManageRolePermissions, DisplayName: "Manage Role Permissions", Description: "Allows managing permissions assigned to roles"},
}

func AllPermissionKeys() []PermissionKey {
	keys := make([]PermissionKey, len(AllPermissionDefinitions))
	for i, p := range AllPermissionDefinitions {
		keys[i] = p.Key
	}
	return keys
}

func IsValidPermissionKey(key PermissionKey) bool {
	for _, p := range AllPermissionDefinitions {
		if p.Key == key {
			return true
		}
	}
	return false
}

var RolePermissions = map[RoleName][]PermissionKey{
	UserGodRoleName: {
		PermissionManageSystemSettings,
		PermissionAssignRole,
		PermissionManageUsers,
		PermissionResetUserPassword,
		PermissionDeactivateUser,
		PermissionManageRoles,
		PermissionManageRolePermissions,
	},
	SuperAdminRoleName: {
		PermissionAssignRole,
		PermissionManageUsers,
		PermissionResetUserPassword,
		PermissionDeactivateUser,
		PermissionManageRoles,
		PermissionManageRolePermissions,
	},
	AdminRoleName: {
		PermissionManageUsers,
		PermissionResetUserPassword,
		PermissionDeactivateUser,
	},
	MemberRoleName: {},
}

func PermissionsForRole(name RoleName) []PermissionKey {
	return RolePermissions[name]
}

func RoleHasPermission(name RoleName, key PermissionKey) bool {
	for _, p := range RolePermissions[name] {
		if p == key {
			return true
		}
	}
	return false
}

type DefaultRoleDef struct {
	Name        RoleName
	DisplayName string
	Description string
	RoleType    RoleType
	ScopeType   ScopeType
	Assignable  bool
}

var DefaultRolesInit = map[RoleName]DefaultRoleDef{
	UserGodRoleName: {
		Name:        UserGodRoleName,
		DisplayName: "User God",
		Description: "Superuser with full system access and all permissions",
		RoleType:    RoleTypeSystem,
		ScopeType:   ScopeTypeGlobal,
		Assignable:  false,
	},
	SuperAdminRoleName: {
		Name:        SuperAdminRoleName,
		DisplayName: "Super Admin",
		Description: "Administrator with broad system management capabilities",
		RoleType:    RoleTypeSystem,
		ScopeType:   ScopeTypeGlobal,
		Assignable:  false,
	},
	AdminRoleName: {
		Name:        AdminRoleName,
		DisplayName: "Admin",
		Description: "Administrator with user management capabilities",
		RoleType:    RoleTypeSystem,
		ScopeType:   ScopeTypeGlobal,
		Assignable:  true,
	},
	MemberRoleName: {
		Name:        MemberRoleName,
		DisplayName: "Member",
		Description: "Regular member with basic access",
		RoleType:    RoleTypeSystem,
		ScopeType:   ScopeTypeGlobal,
		Assignable:  true,
	},
}

var DefaultPermissionsInit = map[PermissionKey]PermissionDefinition{
	PermissionManageSystemSettings: {
		Key:         PermissionManageSystemSettings,
		DisplayName: "Manage System Settings",
		Description: "Allows managing system-wide settings",
	},
	PermissionAssignRole: {
		Key:         PermissionAssignRole,
		DisplayName: "Assign Role",
		Description: "Allows assigning roles to users",
	},
	PermissionManageUsers: {
		Key:         PermissionManageUsers,
		DisplayName: "Manage Users",
		Description: "Allows managing user accounts",
	},
	PermissionResetUserPassword: {
		Key:         PermissionResetUserPassword,
		DisplayName: "Reset User Password",
		Description: "Allows resetting user passwords",
	},
	PermissionDeactivateUser: {
		Key:         PermissionDeactivateUser,
		DisplayName: "Deactivate User",
		Description: "Allows deactivating user accounts",
	},
	PermissionManageRoles: {
		Key:         PermissionManageRoles,
		DisplayName: "Manage Roles",
		Description: "Allows managing role definitions",
	},
	PermissionManageRolePermissions: {
		Key:         PermissionManageRolePermissions,
		DisplayName: "Manage Role Permissions",
		Description: "Allows managing permissions assigned to roles",
	},
}

type Role struct {
	ID          string
	Name        RoleName
	DisplayName string
	Description *string
	RoleType    RoleType
	ScopeType   ScopeType
	Assignable  bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewRole(id string, name RoleName, displayName string, description *string, roleType RoleType, scopeType ScopeType, assignable bool) (*Role, error) {
	if roleType != RoleTypeSystem && roleType != RoleTypeCustom {
		return nil, kernel.New(ErrCodeInvalidRoleType)
	}
	if scopeType != ScopeTypeGlobal && scopeType != ScopeTypeRegion && scopeType != ScopeTypeCommunity {
		return nil, kernel.New(ErrCodeInvalidScopeType)
	}

	now := time.Now()
	return &Role{
		ID:          id,
		Name:        name,
		DisplayName: displayName,
		Description: description,
		RoleType:    roleType,
		ScopeType:   scopeType,
		Assignable:  assignable,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (r *Role) IsSystem() bool {
	return r.RoleType == RoleTypeSystem
}

func (r *Role) EnsureCanDelete() error {
	if r.IsSystem() {
		return kernel.New(ErrCodeRoleCannotDeleteSystem)
	}
	return nil
}

func (r *Role) EnsureAssignable() error {
	if !r.Assignable {
		return kernel.New(ErrCodeRoleNotAssignable)
	}
	return nil
}

func (r *Role) EnsureCustom() error {
	if r.IsSystem() {
		return kernel.New(ErrCodeInvalidRoleType)
	}
	return nil
}

func (r *Role) EnsureAssignmentScopeMatch(scopeType ScopeType) error {
	if r.ScopeType != scopeType {
		return kernel.New(ErrCodeRoleScopeMismatch)
	}
	return nil
}

func (r *Role) HasPermission(key PermissionKey) bool {
	return RoleHasPermission(r.Name, key)
}

func (r *Role) Touch() {
	r.UpdatedAt = time.Now()
}

func (r *Role) UpdateDetails(displayName string, description *string) {
	r.DisplayName = displayName
	r.Description = description
	r.Touch()
}

type UserRole struct {
	ID            string
	UserID        string
	RoleID        string
	ScopeType     ScopeType
	ScopeID       *string
	AssignedAt    time.Time
	AssignedBy    string
	ExpiredAt     *time.Time
	IsActive      bool
	Notes         *string
	DeactivatedAt *time.Time
}

func NewUserRole(id, userID, roleID string, scopeType ScopeType, scopeID *string, assignedBy string, expiredAt *time.Time, notes *string) (*UserRole, error) {
	if scopeType != ScopeTypeGlobal && scopeType != ScopeTypeRegion && scopeType != ScopeTypeCommunity {
		return nil, kernel.New(ErrCodeInvalidScopeType)
	}

	now := time.Now()
	return &UserRole{
		ID:         id,
		UserID:     userID,
		RoleID:     roleID,
		ScopeType:  scopeType,
		ScopeID:    scopeID,
		AssignedAt: now,
		AssignedBy: assignedBy,
		ExpiredAt:  expiredAt,
		IsActive:   true,
		Notes:      notes,
	}, nil
}

func (ur *UserRole) IsExpired() bool {
	if ur.ExpiredAt == nil {
		return false
	}
	return time.Now().After(*ur.ExpiredAt)
}

func (ur *UserRole) IsUsable() bool {
	return ur.IsActive && !ur.IsExpired()
}

func (ur *UserRole) Deactivate() error {
	if !ur.IsActive {
		return kernel.New(ErrCodeUserRoleNotActive)
	}
	now := time.Now()
	ur.IsActive = false
	ur.DeactivatedAt = &now
	return nil
}

func (ur *UserRole) Reactivate() error {
	if ur.IsActive {
		return kernel.New(ErrCodeUserRoleNotActive)
	}
	if ur.IsExpired() {
		return kernel.New(ErrCodeUserRoleExpired)
	}
	ur.IsActive = true
	ur.DeactivatedAt = nil
	return nil
}

func (ur *UserRole) UpdateExpiration(expiredAt *time.Time) {
	ur.ExpiredAt = expiredAt
}

type RolePermission struct {
	ID            string
	RoleID        string
	PermissionKey PermissionKey
	AssignedAt    time.Time
	AssignedBy    string
	Notes         *string
}

func NewRolePermission(id, roleID string, permissionKey PermissionKey, assignedBy string, notes *string) (*RolePermission, error) {
	if !IsValidPermissionKey(permissionKey) {
		return nil, kernel.New(ErrCodeInvalidPermissionKey)
	}

	return &RolePermission{
		ID:            id,
		RoleID:        roleID,
		PermissionKey: permissionKey,
		AssignedAt:    time.Now(),
		AssignedBy:    assignedBy,
		Notes:         notes,
	}, nil
}

type RoleScope struct {
	ID         string
	RoleID     string
	ScopeType  ScopeType
	ScopeValue string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewRoleScope(id, roleID string, scopeType ScopeType, scopeValue string) (*RoleScope, error) {
	if scopeType != ScopeTypeGlobal && scopeType != ScopeTypeRegion && scopeType != ScopeTypeCommunity {
		return nil, kernel.New(ErrCodeInvalidScopeType)
	}
	if scopeValue == "" {
		return nil, kernel.New(ErrCodeInvalidScopeType)
	}

	now := time.Now()
	return &RoleScope{
		ID:         id,
		RoleID:     roleID,
		ScopeType:  scopeType,
		ScopeValue: scopeValue,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}
