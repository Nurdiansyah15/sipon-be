package entity

import (
	"time"

	"sipon-be/internal/modules/keuangan/domain/period/constant"
	"sipon-be/internal/shared/kernel"
)

type AccountingPeriod struct {
	ID        string
	Name      string
	StartDate time.Time
	EndDate   time.Time
	Status    constant.PeriodStatus
	ClosedBy  *string
	ClosedAt  *time.Time
	CreatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewAccountingPeriod(id, name string, startDate, endDate time.Time, createdBy string) (*AccountingPeriod, error) {
	if id == "" || name == "" || createdBy == "" {
		return nil, kernel.New(constant.CodePeriodNotFound)
	}
	if endDate.Before(startDate) {
		return nil, kernel.New(constant.CodePeriodNotFound)
	}
	now := time.Now()
	return &AccountingPeriod{
		ID:        id,
		Name:      name,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    constant.PeriodOpen,
		CreatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (p *AccountingPeriod) Close(closedBy string) error {
	if p.Status != constant.PeriodOpen {
		return kernel.New(constant.CodePeriodInvalidStatus)
	}
	now := time.Now()
	p.Status = constant.PeriodClosed
	p.ClosedBy = &closedBy
	p.ClosedAt = &now
	p.UpdatedAt = now
	return nil
}

func (p *AccountingPeriod) Reopen() error {
	if p.Status != constant.PeriodClosed {
		return kernel.New(constant.CodePeriodInvalidStatus)
	}
	p.Status = constant.PeriodOpen
	p.ClosedBy = nil
	p.ClosedAt = nil
	p.UpdatedAt = time.Now()
	return nil
}

func (p *AccountingPeriod) Lock(closedBy string) error {
	if p.Status != constant.PeriodClosed {
		return kernel.New(constant.CodePeriodInvalidStatus)
	}
	now := time.Now()
	p.Status = constant.PeriodLocked
	p.ClosedBy = &closedBy
	p.ClosedAt = &now
	p.UpdatedAt = now
	return nil
}

func (p *AccountingPeriod) StartClosing() error {
	if p.Status != constant.PeriodOpen {
		return kernel.New(constant.CodePeriodInvalidStatus)
	}
	p.Status = constant.PeriodClosing
	p.UpdatedAt = time.Now()
	return nil
}

func (p *AccountingPeriod) CanPost() bool {
	return p.Status == constant.PeriodOpen
}

func (p *AccountingPeriod) IsOpen() bool {
	return p.Status == constant.PeriodOpen
}
