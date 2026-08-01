package constant

import "sipon-be/internal/shared/kernel"

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

const (
	ErrCodeRoleNotFound            kernel.Code = "ROLE_NOT_FOUND"
	ErrCodeRoleNotAssignable       kernel.Code = "ROLE_NOT_ASSIGNABLE"
	ErrCodeRoleCannotDeleteSystem  kernel.Code = "ROLE_CANNOT_DELETE_SYSTEM"
	ErrCodeRoleScopeMismatch       kernel.Code = "ROLE_SCOPE_MISMATCH"
	ErrCodeUserRoleAlreadyAssigned kernel.Code = "USER_ROLE_ALREADY_ASSIGNED"
	ErrCodeUserRoleExpired         kernel.Code = "USER_ROLE_EXPIRED"
	ErrCodeUserRoleNotActive       kernel.Code = "USER_ROLE_NOT_ACTIVE"
	ErrCodeInvalidScopeType        kernel.Code = "INVALID_SCOPE_TYPE"
	ErrCodeInvalidRoleType         kernel.Code = "INVALID_ROLE_TYPE"
	ErrCodeUserRoleIDRequired      kernel.Code = "USER_ROLE_ID_REQUIRED"
	ErrCodeUserRoleUserIDRequired  kernel.Code = "USER_ROLE_USER_ID_REQUIRED"
	ErrCodeUserRoleRoleIDRequired  kernel.Code = "USER_ROLE_ROLE_ID_REQUIRED"
	ErrCodeUserRoleScopeIDRequired kernel.Code = "USER_ROLE_SCOPE_ID_REQUIRED"
	ErrCodeUserRoleScopeIDEmpty    kernel.Code = "USER_ROLE_SCOPE_ID_EMPTY"
	ErrCodeRoleScopeIDRequired     kernel.Code = "ROLE_SCOPE_ID_REQUIRED"
	ErrCodeRoleScopeRoleIDRequired kernel.Code = "ROLE_SCOPE_ROLE_ID_REQUIRED"
)

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
