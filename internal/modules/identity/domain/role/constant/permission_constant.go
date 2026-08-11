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
	PermissionManagePSB             PermissionKey = "manage_psb"
	PermissionManagePSBSettings     PermissionKey = "manage_psb_settings"
	PermissionManageDokumen         PermissionKey = "manage_dokumen"
	PermissionCreateArticle         PermissionKey = "create_article"
	PermissionEditArticle           PermissionKey = "edit_article"
	PermissionPublishArticle        PermissionKey = "publish_article"
	PermissionManageArticleCategory PermissionKey = "manage_article_category"
	PermissionManageArticleSources  PermissionKey = "manage_article_sources"
	PermissionManageKeuangan        PermissionKey = "manage_keuangan"
	PermissionVerifyPayment         PermissionKey = "verify_payment"
	PermissionViewKeuanganReports   PermissionKey = "view_keuangan_reports"
	PermissionManageAccounts        PermissionKey = "manage_accounts"
	PermissionManageJournal         PermissionKey = "manage_journal"
	PermissionClosePeriod           PermissionKey = "close_period"
	PermissionManageFeedback        PermissionKey = "manage_feedback"
	PermissionManagePersuratan      PermissionKey = "manage_persuratan"
	PermissionManageAkademik        PermissionKey = "manage_akademik"
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
	{Key: PermissionManagePSB, DisplayName: "Manage PSB", Description: "Allows managing PSB registration: verifying documents, accept/reject, verify re-registration, generate NIS"},
	{Key: PermissionManagePSBSettings, DisplayName: "Manage PSB Settings", Description: "Allows managing PSB periods, quotas, fees, bank accounts, and purging period data"},
	{Key: PermissionManageDokumen, DisplayName: "Manage Dokumen", Description: "Allows managing administrative document assets for public/private download"},
	{Key: PermissionCreateArticle, DisplayName: "Create Article", Description: "Allows creating new articles"},
	{Key: PermissionEditArticle, DisplayName: "Edit Article", Description: "Allows editing and deleting articles"},
	{Key: PermissionPublishArticle, DisplayName: "Publish Article", Description: "Allows publishing and archiving articles"},
	{Key: PermissionManageArticleCategory, DisplayName: "Manage Article Category", Description: "Allows managing article categories"},
	{Key: PermissionManageArticleSources, DisplayName: "Manage Article Sources", Description: "Allows managing article RSS sources, category mapping, and triggering scrapes"},
	{Key: PermissionManageKeuangan, DisplayName: "Manage Keuangan", Description: "Allows managing fee components, billing schemes, invoices, and payments"},
	{Key: PermissionVerifyPayment, DisplayName: "Verify Payment", Description: "Allows verifying and rejecting manual payments"},
	{Key: PermissionViewKeuanganReports, DisplayName: "View Keuangan Reports", Description: "Allows accessing financial reports: summary, outstanding, ledger, trial balance, balance sheet, income statement"},
	{Key: PermissionManageAccounts, DisplayName: "Manage Accounts", Description: "Allows managing chart of accounts (COA)"},
	{Key: PermissionManageJournal, DisplayName: "Manage Journal", Description: "Allows creating and cancelling manual journal entries"},
	{Key: PermissionClosePeriod, DisplayName: "Close Period", Description: "Allows creating, closing, reopening, and locking accounting periods"},
	{Key: PermissionManageFeedback, DisplayName: "Manage Feedback", Description: "Allows moderating feedback: takedown/restore feedback and comments that are inappropriate"},
	{Key: PermissionManagePersuratan, DisplayName: "Manage Persuratan", Description: "Allows managing correspondence: letter types, outgoing letters, and linked documents"},
	{Key: PermissionManageAkademik, DisplayName: "Manage Akademik", Description: "Allows managing academic domain: programs, periods, registrations, activities, schedules, sessions, and attendance"},
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
		PermissionManagePSB,
		PermissionManagePSBSettings,
		PermissionManageDokumen,
		PermissionCreateArticle,
		PermissionEditArticle,
		PermissionPublishArticle,
		PermissionManageArticleCategory,
		PermissionManageArticleSources,
		PermissionManageKeuangan,
		PermissionVerifyPayment,
		PermissionViewKeuanganReports,
		PermissionManageAccounts,
		PermissionManageJournal,
		PermissionClosePeriod,
		PermissionManageFeedback,
		PermissionManagePersuratan,
		PermissionManageAkademik,
	},
	SuperAdminRoleName: {
		PermissionAssignRole,
		PermissionManageUsers,
		PermissionResetUserPassword,
		PermissionDeactivateUser,
		PermissionManageRoles,
		PermissionManageRolePermissions,
		PermissionManageSantri,
		PermissionManagePSB,
		PermissionManagePSBSettings,
		PermissionManageDokumen,
		PermissionCreateArticle,
		PermissionEditArticle,
		PermissionPublishArticle,
		PermissionManageArticleCategory,
		PermissionManageArticleSources,
		PermissionManageKeuangan,
		PermissionVerifyPayment,
		PermissionViewKeuanganReports,
		PermissionManageAccounts,
		PermissionManageJournal,
		PermissionClosePeriod,
		PermissionManageFeedback,
		PermissionManagePersuratan,
		PermissionManageAkademik,
	},
	AdminRoleName: {
		PermissionManageUsers,
		PermissionResetUserPassword,
		PermissionDeactivateUser,
		PermissionManageSantri,
		PermissionManagePSB,
		PermissionManagePSBSettings,
		PermissionManageDokumen,
		PermissionCreateArticle,
		PermissionEditArticle,
		PermissionPublishArticle,
		PermissionManageArticleCategory,
		PermissionManageArticleSources,
		PermissionManageKeuangan,
		PermissionVerifyPayment,
		PermissionViewKeuanganReports,
		PermissionManageAccounts,
		PermissionManageJournal,
		PermissionClosePeriod,
		PermissionManageFeedback,
		PermissionManagePersuratan,
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
	PermissionManagePSB: {
		Key:         PermissionManagePSB,
		DisplayName: "Manage PSB",
		Description: "Allows managing PSB registration: verifying documents, accept/reject, verify re-registration, generate NIS",
	},
	PermissionManagePSBSettings: {
		Key:         PermissionManagePSBSettings,
		DisplayName: "Manage PSB Settings",
		Description: "Allows managing PSB periods, quotas, fees, bank accounts, and purging period data",
	},
	PermissionManageDokumen: {
		Key:         PermissionManageDokumen,
		DisplayName: "Manage Dokumen",
		Description: "Allows managing administrative document assets for public/private download",
	},
	PermissionCreateArticle: {
		Key:         PermissionCreateArticle,
		DisplayName: "Create Article",
		Description: "Allows creating new articles",
	},
	PermissionEditArticle: {
		Key:         PermissionEditArticle,
		DisplayName: "Edit Article",
		Description: "Allows editing and deleting articles",
	},
	PermissionPublishArticle: {
		Key:         PermissionPublishArticle,
		DisplayName: "Publish Article",
		Description: "Allows publishing and archiving articles",
	},
	PermissionManageArticleCategory: {
		Key:         PermissionManageArticleCategory,
		DisplayName: "Manage Article Category",
		Description: "Allows managing article categories",
	},
	PermissionManageArticleSources: {
		Key:         PermissionManageArticleSources,
		DisplayName: "Manage Article Sources",
		Description: "Allows managing article RSS sources, category mapping, and triggering scrapes",
	},
	PermissionManageKeuangan: {
		Key:         PermissionManageKeuangan,
		DisplayName: "Manage Keuangan",
		Description: "Allows managing fee components, billing schemes, invoices, and payments",
	},
	PermissionVerifyPayment: {
		Key:         PermissionVerifyPayment,
		DisplayName: "Verify Payment",
		Description: "Allows verifying and rejecting manual payments",
	},
	PermissionViewKeuanganReports: {
		Key:         PermissionViewKeuanganReports,
		DisplayName: "View Keuangan Reports",
		Description: "Allows accessing financial reports: summary, outstanding, ledger, trial balance, balance sheet, income statement",
	},
	PermissionManageAccounts: {
		Key:         PermissionManageAccounts,
		DisplayName: "Manage Accounts",
		Description: "Allows managing chart of accounts (COA)",
	},
	PermissionManageJournal: {
		Key:         PermissionManageJournal,
		DisplayName: "Manage Journal",
		Description: "Allows creating and cancelling manual journal entries",
	},
	PermissionClosePeriod: {
		Key:         PermissionClosePeriod,
		DisplayName: "Close Period",
		Description: "Allows creating, closing, reopening, and locking accounting periods",
	},
	PermissionManageFeedback: {
		Key:         PermissionManageFeedback,
		DisplayName: "Manage Feedback",
		Description: "Allows moderating feedback: takedown/restore feedback and comments that are inappropriate",
	},
	PermissionManagePersuratan: {
		Key:         PermissionManagePersuratan,
		DisplayName: "Manage Persuratan",
		Description: "Allows managing correspondence: letter types, outgoing letters, and linked documents",
	},
	PermissionManageAkademik: {
		Key:         PermissionManageAkademik,
		DisplayName: "Manage Akademik",
		Description: "Allows managing academic domain: programs, periods, registrations, activities, schedules, sessions, and attendance",
	},
}
