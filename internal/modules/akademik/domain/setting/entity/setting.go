package entity

import (
	"encoding/json"
	"time"

	"sipon-be/internal/modules/akademik/domain/setting/constant"
)

// AkademikSetting merepresentasikan single-row settings akademik.
// Settings disimpan sebagai JSONB; key hardcoded di domain/setting/constant.
type AkademikSetting struct {
	ID        string
	Settings  json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewAkademikSetting(id string, settings json.RawMessage) *AkademikSetting {
	now := time.Now()
	return &AkademikSetting{
		ID:        id,
		Settings:  settings,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// GetDefaultProgramID mengembalikan program default untuk santri, atau nil
// jika belum dikonfigurasi.
func (s *AkademikSetting) GetDefaultProgramID() (*string, error) {
	data := map[string]interface{}{}
	if len(s.Settings) > 0 {
		if err := json.Unmarshal(s.Settings, &data); err != nil {
			return nil, err
		}
	}
	val, ok := data[constant.KeyDefaultProgramID]
	if !ok || val == nil {
		return nil, nil
	}
	str, ok := val.(string)
	if !ok || str == "" {
		return nil, nil
	}
	return &str, nil
}

// SetDefaultProgramID menetapkan (atau menghapus, jika nil) program default
// dalam JSON settings.
func (s *AkademikSetting) SetDefaultProgramID(programID *string) error {
	data := map[string]interface{}{}
	if len(s.Settings) > 0 {
		if err := json.Unmarshal(s.Settings, &data); err != nil {
			return err
		}
	}
	if programID == nil {
		delete(data, constant.KeyDefaultProgramID)
	} else {
		data[constant.KeyDefaultProgramID] = *programID
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	s.Settings = b
	s.UpdatedAt = time.Now()
	return nil
}
