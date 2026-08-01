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
		return nil, kernel.New(constant.ErrCodeUserRoleIDRequired)
	}
	if userID == "" {
		return nil, kernel.New(constant.ErrCodeUserRoleUserIDRequired)
	}
	if roleID == "" {
		return nil, kernel.New(constant.ErrCodeUserRoleRoleIDRequired)
	}
	if scopeType != constant.ScopeTypeGlobal && scopeType != constant.ScopeTypeRegion && scopeType != constant.ScopeTypeCommunity {
		return nil, kernel.New(constant.ErrCodeInvalidScopeType)
	}
	if scopeType == constant.ScopeTypeGlobal && scopeID != nil {
		return nil, kernel.New(constant.ErrCodeUserRoleScopeIDEmpty)
	}
	if (scopeType == constant.ScopeTypeRegion || scopeType == constant.ScopeTypeCommunity) && (scopeID == nil || *scopeID == "") {
		return nil, kernel.New(constant.ErrCodeUserRoleScopeIDRequired)
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
		return kernel.New(constant.ErrCodeUserRoleNotActive)
	}
	now := time.Now()
	ur.IsActive = false
	ur.DeactivatedAt = &now
	return nil
}

func (ur *UserRole) Reactivate() error {
	if ur.IsActive {
		return kernel.New(constant.ErrCodeUserRoleNotActive)
	}
	if ur.IsExpired() {
		return kernel.New(constant.ErrCodeUserRoleExpired)
	}
	ur.IsActive = true
	ur.DeactivatedAt = nil
	return nil
}

func (ur *UserRole) UpdateExpiration(expiredAt *time.Time) {
	ur.ExpiredAt = expiredAt
}
