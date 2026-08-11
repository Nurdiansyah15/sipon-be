package entity_test

import (
	"testing"

	"sipon-be/internal/modules/akademik/domain/activity_schedule/constant"
	"sipon-be/internal/modules/akademik/domain/activity_schedule/entity"
)

func TestNewActivitySchedule(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		periodID  string
		typ       constant.ActivityScheduleType
		startTime string
		endTime   string
		wantErr   bool
	}{
		{"valid weekly", "sch-1", "ap-1", constant.ActivityScheduleTypeWeekly, "19:30:00", "21:00:00", false},
		{"valid once", "sch-2", "ap-1", constant.ActivityScheduleTypeOnce, "08:00:00", "12:00:00", false},
		{"empty id", "", "ap-1", constant.ActivityScheduleTypeWeekly, "19:30:00", "21:00:00", true},
		{"empty period", "sch-3", "", constant.ActivityScheduleTypeWeekly, "19:30:00", "21:00:00", true},
		{"invalid type", "sch-4", "ap-1", "hourly", "19:30:00", "21:00:00", true},
		{"malformed time", "sch-5", "ap-1", constant.ActivityScheduleTypeWeekly, "19:30", "21:00:00", true},
		{"end before start", "sch-6", "ap-1", constant.ActivityScheduleTypeWeekly, "21:00:00", "19:30:00", true},
		{"equal times", "sch-7", "ap-1", constant.ActivityScheduleTypeWeekly, "19:30:00", "19:30:00", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := entity.NewActivitySchedule(tt.id, tt.periodID, tt.typ, tt.startTime, tt.endTime, nil, nil)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if s.Type != tt.typ {
				t.Errorf("expected type %s, got %s", tt.typ, s.Type)
			}
		})
	}
}

func TestActivityScheduleUpdate(t *testing.T) {
	s, _ := entity.NewActivitySchedule("sch-1", "ap-1", constant.ActivityScheduleTypeWeekly, "19:30:00", "21:00:00", nil, nil)

	if err := s.Update("20:00:00", "21:30:00", nil, nil); err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if s.StartTime != "20:00:00" || s.EndTime != "21:30:00" {
		t.Errorf("expected updated times, got %s-%s", s.StartTime, s.EndTime)
	}

	if err := s.Update("21:00:00", "20:30:00", nil, nil); err == nil {
		t.Error("expected error when end before start")
	}

	if err := s.Update("19:00", "21:00:00", nil, nil); err == nil {
		t.Error("expected error on malformed time")
	}
}

func TestIsValidDayOfWeek(t *testing.T) {
	if !constant.IsValidDayOfWeek("monday") {
		t.Error("monday should be valid")
	}
	if !constant.IsValidDayOfWeek("sunday") {
		t.Error("sunday should be valid")
	}
	if constant.IsValidDayOfWeek("funday") {
		t.Error("funday should be invalid")
	}
	if constant.IsValidDayOfWeek("") {
		t.Error("empty should be invalid")
	}
}
