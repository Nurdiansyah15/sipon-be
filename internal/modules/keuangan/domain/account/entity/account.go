package entity

import (
	"time"

	"sipon-be/internal/modules/keuangan/domain/account/constant"
	"sipon-be/internal/shared/kernel"
)

type Account struct {
	ID            string
	Code          string
	Name          string
	Type          constant.AccountType
	SubType       *constant.AccountSubType
	ParentID      *string
	Level         int
	IsPostable    bool
	NormalBalance constant.NormalBalance
	Description   *string
	IsActive      bool
	IsSystem      bool
	CreatedBy     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

func NewAccount(id, code, name string, accType constant.AccountType, parentID *string, level int, normalBalance constant.NormalBalance, subType *constant.AccountSubType, createdBy string) (*Account, error) {
	if id == "" || code == "" || name == "" || createdBy == "" {
		return nil, kernel.WrapMsg(constant.CodeAccountNotFound, "Data akun tidak lengkap", nil)
	}
	if subType != nil && !constant.IsValidSubTypeForType(accType, *subType) {
		return nil, kernel.WrapMsg(constant.CodeAccountInvalidSubType, "Sub-tipe akun tidak valid untuk tipe akun ini", nil)
	}
	if level > 0 && subType == nil {
		return nil, kernel.WrapMsg(constant.CodeAccountSubTypeRequired, "Akun postable wajib memiliki sub-tipe", nil)
	}
	now := time.Now()
	return &Account{
		ID:            id,
		Code:          code,
		Name:          name,
		Type:          accType,
		SubType:       subType,
		ParentID:      parentID,
		Level:         level,
		IsPostable:    level > 0,
		NormalBalance: normalBalance,
		IsActive:      true,
		IsSystem:      false,
		CreatedBy:     createdBy,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func (a *Account) Update(name string, description *string, isPostable bool, subType *constant.AccountSubType) error {
	if a.IsSystem {
		return kernel.WrapMsg(constant.CodeAccountIsSystem, "Akun sistem tidak dapat diubah", nil)
	}
	if subType != nil && !constant.IsValidSubTypeForType(a.Type, *subType) {
		return kernel.WrapMsg(constant.CodeAccountInvalidSubType, "Sub-tipe akun tidak valid untuk tipe akun ini", nil)
	}
	if isPostable && subType == nil {
		return kernel.WrapMsg(constant.CodeAccountSubTypeRequired, "Akun postable wajib memiliki sub-tipe", nil)
	}
	a.Name = name
	a.Description = description
	a.IsPostable = isPostable
	a.SubType = subType
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Account) Deactivate() error {
	if a.IsSystem {
		return kernel.WrapMsg(constant.CodeAccountIsSystem, "Akun sistem tidak dapat dinonaktifkan", nil)
	}
	a.IsActive = false
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Account) Activate() {
	a.IsActive = true
	a.UpdatedAt = time.Now()
}

func (a *Account) SoftDelete() error {
	if a.IsSystem {
		return kernel.WrapMsg(constant.CodeAccountIsSystem, "Akun sistem tidak dapat dihapus", nil)
	}
	now := time.Now()
	a.DeletedAt = &now
	a.UpdatedAt = now
	return nil
}

func (a *Account) EnsurePostable() error {
	if !a.IsPostable || !a.IsActive {
		return kernel.WrapMsg(constant.CodeAccountNotPostable, "Akun tidak dapat diposting", nil)
	}
	return nil
}
