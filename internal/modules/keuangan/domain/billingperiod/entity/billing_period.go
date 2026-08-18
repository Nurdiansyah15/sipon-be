package entity

import (
	"time"

	"sipon-be/internal/modules/keuangan/domain/billingperiod/constant"
	feeConst "sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	"sipon-be/internal/shared/kernel"
)

type BillingPeriod struct {
	ID                 string
	Name               string
	PeriodType         feeConst.PeriodType
	AccountingPeriodID string
	StartDate          time.Time
	EndDate            time.Time
	Status             constant.BillingPeriodStatus
	CreatedBy          string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// NewBillingPeriod membuat periode tagihan dengan validasi domain:
// rentang tanggal periode tagihan harus berada di dalam rentang tanggal
// periode akuntansi (accountingPeriodStart/End) yang menjadi induknya.
func NewBillingPeriod(id, name, accountingPeriodID string, periodType feeConst.PeriodType, startDate, endDate, accountingPeriodStart, accountingPeriodEnd time.Time, createdBy string) (*BillingPeriod, error) {
	if id == "" || name == "" || accountingPeriodID == "" || createdBy == "" {
		return nil, kernel.WrapMsg(constant.CodeBillingPeriodNotFound, "Data periode tagihan tidak lengkap", nil)
	}
	if endDate.Before(startDate) {
		return nil, kernel.WrapMsg(constant.CodeBillingPeriodNotFound, "Tanggal akhir tidak boleh sebelum tanggal mulai", nil)
	}
	if startDate.Before(accountingPeriodStart) || endDate.After(accountingPeriodEnd) {
		return nil, kernel.WrapMsg(constant.CodeBillingPeriodInvalidDateRange,
			"Rentang tanggal periode tagihan harus berada dalam rentang tanggal periode akuntansi yang dipilih", nil)
	}
	now := time.Now()
	return &BillingPeriod{
		ID:                 id,
		Name:               name,
		PeriodType:         periodType,
		AccountingPeriodID: accountingPeriodID,
		StartDate:          startDate,
		EndDate:            endDate,
		Status:             constant.BillingPeriodDraft,
		CreatedBy:          createdBy,
		CreatedAt:          now,
		UpdatedAt:          now,
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

// Update mengubah data periode tagihan. Hanya diperbolehkan selama status
// masih draft; rentang tanggal baru tetap harus berada dalam rentang tanggal
// periode akuntansi induknya.
func (p *BillingPeriod) Update(name string, periodType feeConst.PeriodType, startDate, endDate, accountingPeriodStart, accountingPeriodEnd time.Time) error {
	if p.Status != constant.BillingPeriodDraft {
		return kernel.WrapMsg(constant.CodeBillingPeriodInvalidStatus, "Hanya periode tagihan berstatus draft yang dapat diedit", nil)
	}
	if endDate.Before(startDate) {
		return kernel.WrapMsg(constant.CodeBillingPeriodInvalidDateRange, "Tanggal akhir tidak boleh sebelum tanggal mulai", nil)
	}
	if startDate.Before(accountingPeriodStart) || endDate.After(accountingPeriodEnd) {
		return kernel.WrapMsg(constant.CodeBillingPeriodInvalidDateRange,
			"Rentang tanggal periode tagihan harus berada dalam rentang tanggal periode akuntansi yang dipilih", nil)
	}
	p.Name = name
	p.PeriodType = periodType
	p.StartDate = startDate
	p.EndDate = endDate
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
