package entity

import (
	"time"

	"sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	"sipon-be/internal/shared/kernel"
)

type FeeComponent struct {
	ID                  string
	Code                string
	Name                string
	RevenueAccountID    string
	ReceivableAccountID string
	Amount              float64
	IsPeriodic          bool
	PeriodType          *constant.PeriodType
	Description         *string
	IsActive            bool
	CreatedBy           string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           *time.Time
}

func NewFeeComponent(id, code, name string, revenueAccountID, receivableAccountID string, amount float64, createdBy string) (*FeeComponent, error) {
	if id == "" || code == "" || name == "" || createdBy == "" || revenueAccountID == "" || receivableAccountID == "" {
		return nil, kernel.WrapMsg(constant.CodeFeeComponentNotFound, "Data komponen biaya tidak lengkap", nil)
	}
	now := time.Now()
	return &FeeComponent{
		ID:                  id,
		Code:                code,
		Name:                name,
		RevenueAccountID:    revenueAccountID,
		ReceivableAccountID: receivableAccountID,
		Amount:              amount,
		IsActive:            true,
		CreatedBy:           createdBy,
		CreatedAt:           now,
		UpdatedAt:           now,
	}, nil
}

func (f *FeeComponent) Update(revenueAccountID, receivableAccountID, name string, amount float64, isPeriodic bool, periodType *constant.PeriodType, description *string) {
	f.RevenueAccountID = revenueAccountID
	f.ReceivableAccountID = receivableAccountID
	f.Name = name
	f.Amount = amount
	f.IsPeriodic = isPeriodic
	f.PeriodType = periodType
	f.Description = description
	f.UpdatedAt = time.Now()
}

func (f *FeeComponent) Deactivate() {
	f.IsActive = false
	f.UpdatedAt = time.Now()
}

func (f *FeeComponent) Activate() {
	f.IsActive = true
	f.UpdatedAt = time.Now()
}

func (f *FeeComponent) SoftDelete() {
	now := time.Now()
	f.DeletedAt = &now
	f.UpdatedAt = now
}
