package entity

import (
	"time"

	"sipon-be/internal/modules/identity/domain/role/constant"
	"sipon-be/internal/modules/identity/domain/role/valueobject"
	"sipon-be/internal/shared/kernel"
)

type RoleScope struct {
	ID         string
	RoleID     string
	ScopeType  valueobject.RoleScopeType
	ScopeValue string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewRoleScope(id, roleID string, scopeType valueobject.RoleScopeType, scopeValue string) (*RoleScope, error) {
	if id == "" {
		return nil, kernel.WrapMsg(constant.ErrCodeRoleScopeIDRequired, "ID role scope wajib diisi", nil)
	}
	if roleID == "" {
		return nil, kernel.WrapMsg(constant.ErrCodeRoleScopeRoleIDRequired, "ID role wajib diisi", nil)
	}

	normalizedValue, err := valueobject.NewRoleScopeValue(scopeType, scopeValue)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &RoleScope{
		ID:         id,
		RoleID:     roleID,
		ScopeType:  scopeType,
		ScopeValue: normalizedValue,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}
