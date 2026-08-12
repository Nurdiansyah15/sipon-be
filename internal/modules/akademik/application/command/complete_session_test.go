package command

import (
	"context"
	"testing"
	"time"

	appEntity "sipon-be/internal/modules/akademik/domain/activity_period_program/entity"
	schEntity "sipon-be/internal/modules/akademik/domain/activity_schedule/entity"
	sesConst "sipon-be/internal/modules/akademik/domain/activity_session/constant"
	sesEntity "sipon-be/internal/modules/akademik/domain/activity_session/entity"
	attConst "sipon-be/internal/modules/akademik/domain/attendance/constant"
	"sipon-be/internal/modules/akademik/application"
)

func newCompleteUC(
	session *sesEntity.ActivitySession,
	attendance *fakeAttendanceRepo,
	santriProgram *fakeSantriProgramRepo,
	appProgram *fakeAPProgramRepo,
	program *fakeProgramRepo,
) *CompleteSessionUseCase {
	sessionRepo := &fakeSessionRepo{session: session}
	scheduleRepo := &fakeScheduleRepo{schedule: &schEntity.ActivitySchedule{ID: session.ActivityScheduleID, ActivityPeriodID: "ap-1"}}
	resolver := application.NewSessionProgramResolver(sessionRepo, scheduleRepo, appProgram, program)
	return NewCompleteSessionUseCase(sessionRepo, attendance, santriProgram, resolver)
}

func newSession(status sesConst.ActivitySessionStatus) *sesEntity.ActivitySession {
	now := time.Now()
	return &sesEntity.ActivitySession{
		ID:                 "session-1",
		ActivityScheduleID: "schedule-1",
		StartsAt:           now,
		EndsAt:             now.Add(time.Hour),
		Status:             status,
	}
}

func TestCompleteSessionAutoAbsent(t *testing.T) {
	attendance := &fakeAttendanceRepo{recordedSantriIDs: []string{"santri-2"}}
	santriProgram := &fakeSantriProgramRepo{
		santriByProgram: map[string][]string{
			"program-1": {"santri-1", "santri-2", "santri-3"},
		},
	}
	appProgram := &fakeAPProgramRepo{
		links: []*appEntity.ActivityPeriodProgram{
			{ID: "link-1", ActivityPeriodID: "ap-1", ProgramID: "program-1"},
		},
	}
	program := &fakeProgramRepo{}

	uc := newCompleteUC(newSession(sesConst.ActivitySessionStatusOpen), attendance, santriProgram, appProgram, program)
	_, err := uc.Execute(context.Background(), "session-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// santri-1 & santri-3 belum tercatat → absent; santri-2 sudah hadir → skip.
	if len(attendance.saved) != 2 {
		t.Fatalf("expected 2 absent records, got %d", len(attendance.saved))
	}
	statuses := map[string]attConst.AttendanceStatus{}
	for _, a := range attendance.saved {
		statuses[a.SantriID] = a.Status
	}
	if statuses["santri-1"] != attConst.AttendanceStatusAbsent {
		t.Fatalf("expected santri-1 absent, got %s", statuses["santri-1"])
	}
	if statuses["santri-3"] != attConst.AttendanceStatusAbsent {
		t.Fatalf("expected santri-3 absent, got %s", statuses["santri-3"])
	}
	if _, ok := statuses["santri-2"]; ok {
		t.Fatalf("santri-2 already recorded, should not be saved")
	}
}

func TestCompleteSessionAutoAbsentNoProgramScope(t *testing.T) {
	// Tanpa activity_period_programs → kegiatan berlaku untuk semua program aktif.
	attendance := &fakeAttendanceRepo{}
	santriProgram := &fakeSantriProgramRepo{
		santriByProgram: map[string][]string{
			"program-1": {"santri-1", "santri-2"},
			"program-2": {"santri-3"},
		},
	}
	appProgram := &fakeAPProgramRepo{} // no links → no scope
	program := &fakeProgramRepo{activeIDs: []string{"program-1", "program-2"}}

	uc := newCompleteUC(newSession(sesConst.ActivitySessionStatusScheduled), attendance, santriProgram, appProgram, program)
	_, err := uc.Execute(context.Background(), "session-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(attendance.saved) != 3 {
		t.Fatalf("expected 3 absent records (all active programs), got %d", len(attendance.saved))
	}
}

func TestCompleteSessionAutoAbsentSkipsOtherProgram(t *testing.T) {
	attendance := &fakeAttendanceRepo{}
	santriProgram := &fakeSantriProgramRepo{
		santriByProgram: map[string][]string{
			"program-1": {"santri-1"},
			"program-2": {"santri-9"},
		},
	}
	appProgram := &fakeAPProgramRepo{
		links: []*appEntity.ActivityPeriodProgram{
			{ID: "link-1", ActivityPeriodID: "ap-1", ProgramID: "program-1"},
		},
	}
	program := &fakeProgramRepo{}

	uc := newCompleteUC(newSession(sesConst.ActivitySessionStatusOpen), attendance, santriProgram, appProgram, program)
	_, err := uc.Execute(context.Background(), "session-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(attendance.saved) != 1 {
		t.Fatalf("expected 1 absent record, got %d", len(attendance.saved))
	}
	if attendance.saved[0].SantriID != "santri-1" {
		t.Fatalf("expected absent for santri-1, got %s", attendance.saved[0].SantriID)
	}
}

func TestCompleteSessionInvalidStatus(t *testing.T) {
	// Sudah completed → Complete() menolak, tidak ada auto-absen.
	attendance := &fakeAttendanceRepo{}
	santriProgram := &fakeSantriProgramRepo{
		santriByProgram: map[string][]string{"program-1": {"santri-1"}},
	}
	appProgram := &fakeAPProgramRepo{
		links: []*appEntity.ActivityPeriodProgram{
			{ID: "link-1", ActivityPeriodID: "ap-1", ProgramID: "program-1"},
		},
	}
	program := &fakeProgramRepo{}

	uc := newCompleteUC(newSession(sesConst.ActivitySessionStatusCompleted), attendance, santriProgram, appProgram, program)
	_, err := uc.Execute(context.Background(), "session-1")
	if err == nil {
		t.Fatal("expected error for already completed session")
	}
	if len(attendance.saved) != 0 {
		t.Fatalf("expected no auto-absent, got %d records", len(attendance.saved))
	}
}
