package entity

import (
	"time"

	"sipon-be/internal/modules/identity/domain/role/constant"
	"sipon-be/internal/shared/kernel"
)

type RolePermission struct {
	ID            string
	RoleID        string
	PermissionKey constant.PermissionKey
	AssignedAt    time.Time
	AssignedBy    string
	Notes         *string
}

func NewRolePermission(id, roleID string, permissionKey constant.PermissionKey, assignedBy string, notes *string) (*RolePermission, error) {
	if !constant.IsValidPermissionKey(permissionKey) {
		return nil, kernel.WrapMsg(constant.ErrCodeInvalidPermissionKey, "Permission key tidak valid", nil)
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
