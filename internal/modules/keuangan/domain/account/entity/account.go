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

func NewAccount(id, code, name string, accType constant.AccountType, parentID *string, level int, normalBalance constant.NormalBalance, createdBy string) (*Account, error) {
	if id == "" || code == "" || name == "" || createdBy == "" {
		return nil, kernel.New(constant.CodeAccountNotFound)
	}
	now := time.Now()
	return &Account{
		ID:            id,
		Code:          code,
		Name:          name,
		Type:          accType,
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

func (a *Account) Update(name string, description *string, isPostable bool) error {
	if a.IsSystem {
		return kernel.New(constant.CodeAccountIsSystem)
	}
	a.Name = name
	a.Description = description
	a.IsPostable = isPostable
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Account) Deactivate() error {
	if a.IsSystem {
		return kernel.New(constant.CodeAccountIsSystem)
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
		return kernel.New(constant.CodeAccountIsSystem)
	}
	now := time.Now()
	a.DeletedAt = &now
	a.UpdatedAt = now
	return nil
}

func (a *Account) EnsurePostable() error {
	if !a.IsPostable || !a.IsActive {
		return kernel.New(constant.CodeAccountNotPostable)
	}
	return nil
}
