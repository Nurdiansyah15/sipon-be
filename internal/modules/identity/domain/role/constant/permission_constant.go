package constant

import "sipon-be/internal/shared/kernel"

const (
	ErrCodeInvalidPermissionKey kernel.Code = "INVALID_PERMISSION_KEY"
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
	PermissionManageSantri          PermissionKey = "manage_santri"
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
	{Key: PermissionManageSantri, DisplayName: "Manage Santri", Description: "Allows managing santri profiles, requests, and documents"},
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
		PermissionManageSantri,
	},
	SuperAdminRoleName: {
		PermissionAssignRole,
		PermissionManageUsers,
		PermissionResetUserPassword,
		PermissionDeactivateUser,
		PermissionManageRoles,
		PermissionManageRolePermissions,
		PermissionManageSantri,
	},
	AdminRoleName: {
		PermissionManageUsers,
		PermissionResetUserPassword,
		PermissionDeactivateUser,
		PermissionManageSantri,
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
	PermissionManageSantri: {
		Key:         PermissionManageSantri,
		DisplayName: "Manage Santri",
		Description: "Allows managing santri profiles, requests, and documents",
	},
}
