package entity_test

import (
	"testing"
	"time"

	"sipon-be/internal/modules/akademik/domain/activity_session/constant"
	"sipon-be/internal/modules/akademik/domain/activity_session/entity"
)

func createTestSession() *entity.ActivitySession {
	start := time.Date(2026, 8, 10, 19, 30, 0, 0, time.UTC)
	end := time.Date(2026, 8, 10, 21, 0, 0, 0, time.UTC)
	s, _ := entity.NewActivitySession("ses-1", "sch-1", start, end)
	return s
}

func TestNewActivitySession(t *testing.T) {
	start := time.Date(2026, 8, 10, 19, 30, 0, 0, time.UTC)
	end := time.Date(2026, 8, 10, 21, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		id        string
		schedule  string
		startsAt  time.Time
		endsAt    time.Time
		wantErr   bool
	}{
		{"valid session", "ses-1", "sch-1", start, end, false},
		{"empty id", "", "sch-1", start, end, true},
		{"empty schedule", "ses-2", "", start, end, true},
		{"end before start", "ses-3", "sch-1", end, start, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := entity.NewActivitySession(tt.id, tt.schedule, tt.startsAt, tt.endsAt)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if s.Status != constant.ActivitySessionStatusScheduled {
				t.Errorf("expected scheduled status, got %s", s.Status)
			}
		})
	}
}

func TestActivitySessionLifecycle(t *testing.T) {
	s := createTestSession()

	if err := s.Open(); err != nil {
		t.Fatalf("open failed: %v", err)
	}
	if s.Status != constant.ActivitySessionStatusOpen {
		t.Errorf("expected open, got %s", s.Status)
	}

	if err := s.Complete(); err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	if s.Status != constant.ActivitySessionStatusCompleted {
		t.Errorf("expected completed, got %s", s.Status)
	}
	if err := s.Complete(); err == nil {
		t.Error("expected error when completing completed session")
	}
	if err := s.Cancel(); err == nil {
		t.Error("expected error when cancelling completed session")
	}
}

func TestActivitySessionCancel(t *testing.T) {
	s := createTestSession()
	if err := s.Cancel(); err != nil {
		t.Fatalf("cancel failed: %v", err)
	}
	if s.Status != constant.ActivitySessionStatusCancelled {
		t.Errorf("expected cancelled, got %s", s.Status)
	}
	if err := s.Cancel(); err == nil {
		t.Error("expected error when cancelling already-cancelled session")
	}
}
