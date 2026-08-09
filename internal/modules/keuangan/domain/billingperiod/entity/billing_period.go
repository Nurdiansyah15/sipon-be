package entity

import (
	"time"

	"sipon-be/internal/modules/keuangan/domain/billingperiod/constant"
	feeConst "sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	"sipon-be/internal/shared/kernel"
)

type BillingPeriod struct {
	ID         string
	Name       string
	PeriodType feeConst.PeriodType
	StartDate  time.Time
	EndDate    time.Time
	Status     constant.BillingPeriodStatus
	CreatedBy  string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewBillingPeriod(id, name string, periodType feeConst.PeriodType, startDate, endDate time.Time, createdBy string) (*BillingPeriod, error) {
	if id == "" || name == "" || createdBy == "" {
		return nil, kernel.WrapMsg(constant.CodeBillingPeriodNotFound, "Data periode tagihan tidak lengkap", nil)
	}
	if endDate.Before(startDate) {
		return nil, kernel.WrapMsg(constant.CodeBillingPeriodNotFound, "Tanggal akhir tidak boleh sebelum tanggal mulai", nil)
	}
	now := time.Now()
	return &BillingPeriod{
		ID:         id,
		Name:       name,
		PeriodType: periodType,
		StartDate:  startDate,
		EndDate:    endDate,
		Status:     constant.BillingPeriodDraft,
		CreatedBy:  createdBy,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (p *BillingPeriod) Open() error {
	if p.Status != constant.BillingPeriodDraft {
		return kernel.WrapMsg(constant.CodeBillingPeriodInvalidStatus, "Status periode tagihan tidak memungkinkan dibuka", nil)
	}
	p.Status = constant.BillingPeriodOpen
	p.UpdatedAt = time.Now()
	return nil
}

func (p *BillingPeriod) Close() error {
	if p.Status != constant.BillingPeriodOpen {
		return kernel.WrapMsg(constant.CodeBillingPeriodInvalidStatus, "Status periode tagihan tidak memungkinkan ditutup", nil)
	}
	p.Status = constant.BillingPeriodClosed
	p.UpdatedAt = time.Now()
	return nil
}

func (p *BillingPeriod) IsOpen() bool {
	return p.Status == constant.BillingPeriodOpen
}
