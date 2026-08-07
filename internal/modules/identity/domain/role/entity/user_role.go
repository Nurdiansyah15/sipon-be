package entity

import (
	"time"

	"sipon-be/internal/modules/identity/domain/role/constant"
	"sipon-be/internal/shared/kernel"
)

type UserRole struct {
	ID            string
	UserID        string
	RoleID        string
	ScopeType     constant.ScopeType
	ScopeID       *string
	AssignedAt    time.Time
	AssignedBy    string
	ExpiredAt     *time.Time
	IsActive      bool
	Notes         *string
	DeactivatedAt *time.Time
}

func NewUserRole(id, userID, roleID string, scopeType constant.ScopeType, scopeID *string, assignedBy string, expiredAt *time.Time, notes *string) (*UserRole, error) {
	if id == "" {
		return nil, kernel.WrapMsg(constant.ErrCodeUserRoleIDRequired, "ID user role wajib diisi", nil)
	}
	if userID == "" {
		return nil, kernel.WrapMsg(constant.ErrCodeUserRoleUserIDRequired, "ID pengguna wajib diisi", nil)
	}
	if roleID == "" {
		return nil, kernel.WrapMsg(constant.ErrCodeUserRoleRoleIDRequired, "ID role wajib diisi", nil)
	}
	if scopeType != constant.ScopeTypeGlobal && scopeType != constant.ScopeTypeRegion && scopeType != constant.ScopeTypeCommunity {
		return nil, kernel.WrapMsg(constant.ErrCodeInvalidScopeType, "Jenis scope tidak valid", nil)
	}
	if scopeType == constant.ScopeTypeGlobal && scopeID != nil {
		return nil, kernel.WrapMsg(constant.ErrCodeUserRoleScopeIDEmpty, "Scope ID tidak boleh diisi untuk scope global", nil)
	}
	if (scopeType == constant.ScopeTypeRegion || scopeType == constant.ScopeTypeCommunity) && (scopeID == nil || *scopeID == "") {
		return nil, kernel.WrapMsg(constant.ErrCodeUserRoleScopeIDRequired, "Scope ID wajib diisi untuk scope regional atau komunitas", nil)
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
		return kernel.WrapMsg(constant.ErrCodeUserRoleNotActive, "User role tidak aktif", nil)
	}
	now := time.Now()
	ur.IsActive = false
	ur.DeactivatedAt = &now
	return nil
}

func (ur *UserRole) Reactivate() error {
	if ur.IsActive {
		return kernel.WrapMsg(constant.ErrCodeUserRoleNotActive, "User role sudah aktif", nil)
	}
	if ur.IsExpired() {
		return kernel.WrapMsg(constant.ErrCodeUserRoleExpired, "User role telah kedaluwarsa", nil)
	}
	ur.IsActive = true
	ur.DeactivatedAt = nil
	return nil
}

func (ur *UserRole) UpdateExpiration(expiredAt *time.Time) {
	ur.ExpiredAt = expiredAt
}
