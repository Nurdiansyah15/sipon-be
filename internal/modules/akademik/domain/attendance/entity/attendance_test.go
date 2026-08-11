package entity_test

import (
	"testing"

	"sipon-be/internal/modules/akademik/domain/attendance/constant"
	"sipon-be/internal/modules/akademik/domain/attendance/entity"
)

func TestNewAttendance(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		sessionID string
		santriID  string
		status    constant.AttendanceStatus
		wantErr   bool
	}{
		{"valid present", "att-1", "ses-1", "san-1", constant.AttendanceStatusPresent, false},
		{"valid absent", "att-2", "ses-1", "san-1", constant.AttendanceStatusAbsent, false},
		{"valid excused", "att-3", "ses-1", "san-1", constant.AttendanceStatusExcused, false},
		{"empty id", "", "ses-1", "san-1", constant.AttendanceStatusPresent, true},
		{"empty session", "att-4", "", "san-1", constant.AttendanceStatusPresent, true},
		{"empty santri", "att-5", "ses-1", "", constant.AttendanceStatusPresent, true},
		{"invalid status", "att-6", "ses-1", "san-1", "unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := entity.NewAttendance(tt.id, tt.sessionID, tt.santriID, tt.status)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if a.Status != tt.status {
				t.Errorf("expected status %s, got %s", tt.status, a.Status)
			}
		})
	}
}

func TestAttendanceUpdateStatus(t *testing.T) {
	a, _ := entity.NewAttendance("att-1", "ses-1", "san-1", constant.AttendanceStatusPresent)

	if err := a.UpdateStatus(constant.AttendanceStatusExcused); err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if a.Status != constant.AttendanceStatusExcused {
		t.Errorf("expected excused, got %s", a.Status)
	}

	if err := a.UpdateStatus("unknown"); err == nil {
		t.Error("expected error on invalid status")
	}
}
