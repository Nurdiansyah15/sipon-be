package entity

import (
	"time"

	"sipon-be/internal/modules/identity/domain/role/constant"
	"sipon-be/internal/shared/kernel"
)

type Role struct {
	ID          string
	Name        constant.RoleName
	DisplayName string
	Description *string
	RoleType    constant.RoleType
	ScopeType   constant.ScopeType
	Assignable  bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewRole(id string, name constant.RoleName, displayName string, description *string, roleType constant.RoleType, scopeType constant.ScopeType, assignable bool) (*Role, error) {
	if roleType != constant.RoleTypeSystem && roleType != constant.RoleTypeCustom {
		return nil, kernel.WrapMsg(constant.ErrCodeInvalidRoleType, "Jenis role tidak valid", nil)
	}
	if scopeType != constant.ScopeTypeGlobal && scopeType != constant.ScopeTypeRegion && scopeType != constant.ScopeTypeCommunity {
		return nil, kernel.WrapMsg(constant.ErrCodeInvalidScopeType, "Jenis scope tidak valid", nil)
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
	return r.RoleType == constant.RoleTypeSystem
}

func (r *Role) EnsureCanDelete() error {
	if r.IsSystem() {
		return kernel.WrapMsg(constant.ErrCodeRoleCannotDeleteSystem, "Role sistem tidak dapat dihapus", nil)
	}
	return nil
}

func (r *Role) EnsureAssignable() error {
	if !r.Assignable {
		return kernel.WrapMsg(constant.ErrCodeRoleNotAssignable, "Role tidak dapat ditugaskan", nil)
	}
	return nil
}

func (r *Role) EnsureCustom() error {
	if r.IsSystem() {
		return kernel.WrapMsg(constant.ErrCodeInvalidRoleType, "Role harus bertipe custom", nil)
	}
	return nil
}

func (r *Role) EnsureAssignmentScopeMatch(scopeType constant.ScopeType) error {
	if r.ScopeType != scopeType {
		return kernel.WrapMsg(constant.ErrCodeRoleScopeMismatch, "Scope role tidak sesuai", nil)
	}
	return nil
}

func (r *Role) HasPermission(key constant.PermissionKey) bool {
	return constant.RoleHasPermission(r.Name, key)
}

func (r *Role) Touch() {
	r.UpdatedAt = time.Now()
}

func (r *Role) UpdateDetails(displayName string, description string, assignable *bool) error {
	if r.IsSystem() {
		return kernel.WrapMsg(constant.ErrCodeInvalidRoleType, "Role sistem tidak dapat diubah", nil)
	}
	if displayName != "" {
		r.DisplayName = displayName
	}
	if description != "" {
		r.Description = &description
	}
	if assignable != nil {
		r.Assignable = *assignable
	}
	r.Touch()
	return nil
}
