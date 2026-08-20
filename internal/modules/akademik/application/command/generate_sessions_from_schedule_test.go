package command

import (
	"context"
	"testing"
	"time"

	"sipon-be/internal/modules/akademik/application/dto"
	schConst "sipon-be/internal/modules/akademik/domain/activity_schedule/constant"
	schEntity "sipon-be/internal/modules/akademik/domain/activity_schedule/entity"
	sesEntity "sipon-be/internal/modules/akademik/domain/activity_session/entity"
)

type fakeTransactor struct{}

func (fakeTransactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func timePtr(t time.Time) *time.Time { return &t }

func TestGenerateSessionsFromScheduleDaily(t *testing.T) {
	start := mustParseTime(t, "2026-08-10T00:00:00Z")
	end := mustParseTime(t, "2026-08-12T00:00:00Z")
	sch := &schEntity.ActivitySchedule{
		ID:        "sch-1",
		Type:      schConst.ActivityScheduleTypeDaily,
		StartDate: &start,
		EndDate:   &end,
		StartTime: "08:00:00",
		EndTime:   "09:00:00",
	}

	uc := NewGenerateSessionsFromScheduleUseCase(
		&fakeScheduleRepo{schedule: sch},
		&fakeSessionRepo{},
		fakeTransactor{},
		nil, // scheduleAutoOpenUC — skip auto-open scheduling in unit tests
	)

	resp, err := uc.Execute(context.Background(), "sch-1", dto.GenerateSessionsRequest{
		FromDate: "2026-08-01",
		ToDate:   "2026-08-31",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalCreated != 3 {
		t.Fatalf("expected 3 sessions created, got %d (skipped=%d)", resp.TotalCreated, resp.TotalSkipped)
	}
	if len(resp.Sessions) != 3 {
		t.Fatalf("expected 3 session responses, got %d", len(resp.Sessions))
	}
}

func TestGenerateSessionsFromScheduleSkipsExisting(t *testing.T) {
	start := mustParseTime(t, "2026-08-10T00:00:00Z")
	end := mustParseTime(t, "2026-08-12T00:00:00Z")
	sch := &schEntity.ActivitySchedule{
		ID:        "sch-1",
		Type:      schConst.ActivityScheduleTypeDaily,
		StartDate: &start,
		EndDate:   &end,
		StartTime: "08:00:00",
		EndTime:   "09:00:00",
	}

	// Sesi existing untuk 2026-08-10 08:00:00 (platform TZ fallback UTC di test).
	existingStarts := time.Date(2026, 8, 10, 8, 0, 0, 0, time.UTC)
	existing := &sesEntity.ActivitySession{
		ID:       "sesi-existing",
		StartsAt: existingStarts,
		EndsAt:   existingStarts.Add(time.Hour),
		Status:   "scheduled",
	}

	uc := NewGenerateSessionsFromScheduleUseCase(
		&fakeScheduleRepo{schedule: sch},
		&fakeSessionRepo{existing: []*sesEntity.ActivitySession{existing}},
		fakeTransactor{},
		nil,
	)

	resp, err := uc.Execute(context.Background(), "sch-1", dto.GenerateSessionsRequest{
		FromDate: "2026-08-01",
		ToDate:   "2026-08-31",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalCreated != 2 {
		t.Fatalf("expected 2 created (1 skipped), got created=%d skipped=%d", resp.TotalCreated, resp.TotalSkipped)
	}
	if resp.TotalSkipped != 1 {
		t.Fatalf("expected 1 skipped, got %d", resp.TotalSkipped)
	}
}

func TestGenerateSessionsFromScheduleWeekly(t *testing.T) {
	sch := &schEntity.ActivitySchedule{
		ID:        "sch-1",
		Type:      schConst.ActivityScheduleTypeWeekly,
		StartTime: "08:00:00",
		EndTime:   "09:00:00",
	}

	uc := NewGenerateSessionsFromScheduleUseCase(
		&fakeScheduleRepo{
			schedule: sch,
			weeklies: []schEntity.ActivityScheduleWeekly{
				{DayOfWeek: schConst.DayOfWeekMonday},
				{DayOfWeek: schConst.DayOfWeekFriday},
			},
		},
		&fakeSessionRepo{},
		fakeTransactor{},
		nil,
	)

	// 2026-08-01 .. 2026-08-31 → Senin & Jumat, total 9 hari.
	resp, err := uc.Execute(context.Background(), "sch-1", dto.GenerateSessionsRequest{
		FromDate: "2026-08-01",
		ToDate:   "2026-08-31",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalCreated != 9 {
		t.Fatalf("expected 9 sessions created, got %d", resp.TotalCreated)
	}
	if resp.TotalSkipped != 0 {
		t.Fatalf("expected 0 skipped, got %d", resp.TotalSkipped)
	}
}

func TestGenerateSessionsInvalidDateRange(t *testing.T) {
	sch := &schEntity.ActivitySchedule{
		ID:        "sch-1",
		Type:      schConst.ActivityScheduleTypeDaily,
		StartTime: "08:00:00",
		EndTime:   "09:00:00",
	}
	uc := NewGenerateSessionsFromScheduleUseCase(
		&fakeScheduleRepo{schedule: sch},
		&fakeSessionRepo{},
		fakeTransactor{},
		nil,
	)
	_, err := uc.Execute(context.Background(), "sch-1", dto.GenerateSessionsRequest{
		FromDate: "2026-08-31",
		ToDate:   "2026-08-01",
	})
	if err == nil {
		t.Fatal("expected error for inverted date range")
	}
}

func mustParseTime(t *testing.T, v string) time.Time {
	t.Helper()
	tt, err := time.Parse(time.RFC3339, v)
	if err != nil {
		t.Fatalf("parse %q: %v", v, err)
	}
	return tt
}
