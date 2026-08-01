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
		return nil, kernel.New(constant.ErrCodeInvalidRoleType)
	}
	if scopeType != constant.ScopeTypeGlobal && scopeType != constant.ScopeTypeRegion && scopeType != constant.ScopeTypeCommunity {
		return nil, kernel.New(constant.ErrCodeInvalidScopeType)
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
		return kernel.New(constant.ErrCodeRoleCannotDeleteSystem)
	}
	return nil
}

func (r *Role) EnsureAssignable() error {
	if !r.Assignable {
		return kernel.New(constant.ErrCodeRoleNotAssignable)
	}
	return nil
}

func (r *Role) EnsureCustom() error {
	if r.IsSystem() {
		return kernel.New(constant.ErrCodeInvalidRoleType)
	}
	return nil
}

func (r *Role) EnsureAssignmentScopeMatch(scopeType constant.ScopeType) error {
	if r.ScopeType != scopeType {
		return kernel.New(constant.ErrCodeRoleScopeMismatch)
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
		return kernel.New(constant.ErrCodeInvalidRoleType)
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
