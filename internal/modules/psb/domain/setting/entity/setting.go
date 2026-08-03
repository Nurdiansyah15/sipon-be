package entity

import (
	"encoding/json"
	"time"

	"sipon-be/internal/modules/psb/domain/setting/constant"
	"sipon-be/internal/shared/kernel"
)

type PsbSetting struct {
	ID           string
	Name         string
	StartPeriod  time.Time
	EndPeriod    time.Time
	Status       constant.SettingStatus
	Quota        json.RawMessage
	RegFee       float64
	BankAccounts json.RawMessage
	DataPurgedAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func NewPsbSetting(id, name string, startPeriod, endPeriod time.Time, quota, bankAccounts json.RawMessage, regFee float64) (*PsbSetting, error) {
	if id == "" || name == "" {
		return nil, kernel.New(constant.ErrCodeInvalidSetting)
	}
	now := time.Now()
	return &PsbSetting{
		ID:           id,
		Name:         name,
		StartPeriod:  startPeriod,
		EndPeriod:    endPeriod,
		Status:       constant.SettingStatusActive,
		Quota:        quota,
		RegFee:       regFee,
		BankAccounts: bankAccounts,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (s *PsbSetting) Close() error {
	if s.Status != constant.SettingStatusActive {
		return kernel.New(constant.ErrCodeInvalidSetting)
	}
	s.Status = constant.SettingStatusClosed
	s.UpdatedAt = time.Now()
	return nil
}

func (s *PsbSetting) IsActive() bool {
	return s.Status == constant.SettingStatusActive
}

func (s *PsbSetting) CanPurge() bool {
	return s.Status == constant.SettingStatusClosed && s.DataPurgedAt == nil
}

func (s *PsbSetting) MarkPurged() {
	now := time.Now()
	s.DataPurgedAt = &now
	s.UpdatedAt = now
}

func (s *PsbSetting) SoftDelete() {
	now := time.Now()
	s.DeletedAt = &now
	s.UpdatedAt = now
}
