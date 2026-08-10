package entity

import (
	"encoding/json"
	"time"

	"sipon-be/internal/modules/keuangan/domain/setting/constant"
)

// KeuanganSetting merepresentasikan single-row settings keuangan.
// Settings disimpan sebagai JSONB; key hardcoded di domain/setting/constant.
type KeuanganSetting struct {
	ID        string
	Settings  json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewKeuanganSetting(id string, settings json.RawMessage) *KeuanganSetting {
	now := time.Now()
	return &KeuanganSetting{
		ID:        id,
		Settings:  settings,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// GetDefaultPaymentDebitAccountID mengembalikan account kas/bank default untuk
// pembayaran, atau nil jika belum dikonfigurasi.
func (s *KeuanganSetting) GetDefaultPaymentDebitAccountID() (*string, error) {
	data := map[string]interface{}{}
	if len(s.Settings) > 0 {
		if err := json.Unmarshal(s.Settings, &data); err != nil {
			return nil, err
		}
	}
	val, ok := data[constant.KeyDefaultPaymentDebitAccountID]
	if !ok || val == nil {
		return nil, nil
	}
	str, ok := val.(string)
	if !ok || str == "" {
		return nil, nil
	}
	return &str, nil
}

// SetDefaultPaymentDebitAccountID menetapkan (atau menghapus, jika nil) account
// default pembayaran dalam JSON settings.
func (s *KeuanganSetting) SetDefaultPaymentDebitAccountID(accountID *string) error {
	data := map[string]interface{}{}
	if len(s.Settings) > 0 {
		if err := json.Unmarshal(s.Settings, &data); err != nil {
			return err
		}
	}
	if accountID == nil {
		delete(data, constant.KeyDefaultPaymentDebitAccountID)
	} else {
		data[constant.KeyDefaultPaymentDebitAccountID] = *accountID
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	s.Settings = b
	s.UpdatedAt = time.Now()
	return nil
}
