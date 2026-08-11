package entity_test

import (
	"encoding/json"
	"testing"

	"sipon-be/internal/modules/akademik/domain/setting/constant"
	"sipon-be/internal/modules/akademik/domain/setting/entity"
)

func TestGetDefaultProgramID(t *testing.T) {
	programID := "prog-tahfidz-1"

	t.Run("empty settings returns nil", func(t *testing.T) {
		s := entity.NewAkademikSetting(constant.SettingsRowID, json.RawMessage(`{}`))
		got, err := s.GetDefaultProgramID()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Fatalf("expected nil, got %v", *got)
		}
	})

	t.Run("set then get returns the program id", func(t *testing.T) {
		s := entity.NewAkademikSetting(constant.SettingsRowID, json.RawMessage(`{}`))
		if err := s.SetDefaultProgramID(&programID); err != nil {
			t.Fatalf("set error: %v", err)
		}
		got, err := s.GetDefaultProgramID()
		if err != nil {
			t.Fatalf("get error: %v", err)
		}
		if got == nil || *got != programID {
			t.Fatalf("expected %q, got %v", programID, got)
		}
	})

	t.Run("set nil removes the key", func(t *testing.T) {
		s := entity.NewAkademikSetting(constant.SettingsRowID, json.RawMessage(`{}`))
		if err := s.SetDefaultProgramID(&programID); err != nil {
			t.Fatalf("set error: %v", err)
		}
		if err := s.SetDefaultProgramID(nil); err != nil {
			t.Fatalf("clear error: %v", err)
		}
		got, err := s.GetDefaultProgramID()
		if err != nil {
			t.Fatalf("get error: %v", err)
		}
		if got != nil {
			t.Fatalf("expected nil after clear, got %v", *got)
		}
	})

	t.Run("preserves unknown keys on update", func(t *testing.T) {
		s := entity.NewAkademikSetting(constant.SettingsRowID, json.RawMessage(`{"some_future_key":"x"}`))
		if err := s.SetDefaultProgramID(&programID); err != nil {
			t.Fatalf("set error: %v", err)
		}
		var data map[string]interface{}
		if err := json.Unmarshal(s.Settings, &data); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if data["some_future_key"] != "x" {
			t.Fatalf("unknown key not preserved: %v", data)
		}
	})
}
