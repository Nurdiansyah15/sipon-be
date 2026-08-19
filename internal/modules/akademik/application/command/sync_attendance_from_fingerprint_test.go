package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/ports"
	"sipon-be/internal/modules/akademik/application/resolver"
	apEntity "sipon-be/internal/modules/akademik/domain/activity_period/entity"
	appEntity "sipon-be/internal/modules/akademik/domain/activity_period_program/entity"
	schEntity "sipon-be/internal/modules/akademik/domain/activity_schedule/entity"
	sesConst "sipon-be/internal/modules/akademik/domain/activity_session/constant"
	sesEntity "sipon-be/internal/modules/akademik/domain/activity_session/entity"
	attEntity "sipon-be/internal/modules/akademik/domain/attendance/entity"
	"sipon-be/internal/shared/kernel"
)

// fakeKesantrianByNIS mengembalikan info santri berbeda per NIS, supaya satu
// sesi sync bisa memproses beberapa pin sekaligus.
type fakeKesantrianByNIS struct {
	byNIS    map[string]*ports.SantriBasicInfo
	errByNIS map[string]error
}

func (f *fakeKesantrianByNIS) GetSantriByID(ctx context.Context, santriID string) (*ports.SantriBasicInfo, error) {
	return nil, nil
}
func (f *fakeKesantrianByNIS) GetSantriByUserID(ctx context.Context, userID string) (*ports.SantriBasicInfo, error) {
	return nil, nil
}
func (f *fakeKesantrianByNIS) GetSantriByNIS(ctx context.Context, nis string) (*ports.SantriBasicInfo, error) {
	if err := f.errByNIS[nis]; err != nil {
		return nil, err
	}
	return f.byNIS[nis], nil
}
func (f *fakeKesantrianByNIS) ListActiveSantriWithUserID(ctx context.Context) ([]ports.SantriBasicInfo, error) {
	return nil, nil
}

type fakeFingerprintReader struct {
	pins []ports.FingerprintScanPin
	err  error

	gotFrom time.Time
	gotTo   time.Time
}

func (f *fakeFingerprintReader) ListDistinctPinInRange(ctx context.Context, from, to time.Time) ([]ports.FingerprintScanPin, error) {
	f.gotFrom = from
	f.gotTo = to
	return f.pins, f.err
}

func newSyncUC(
	session *sesEntity.ActivitySession,
	kesantrian ports.KesantrianReader,
	registration *fakeRegistrationRepo,
	attendance *fakeAttendanceRepo,
	santriProgram *fakeSantriProgramRepo,
	appProgram *fakeAPProgramRepo,
	program *fakeProgramRepo,
	reader ports.FingerprintReader,
) *SyncAttendanceFromFingerprintUseCase {
	sessionRepo := &fakeSessionRepo{session: session}
	scheduleRepo := &fakeScheduleRepo{schedule: &schEntity.ActivitySchedule{ID: session.ActivityScheduleID, ActivityPeriodID: "ap-1"}}
	apRepo := &fakeAPRepo{period: &apEntity.ActivityPeriod{ID: "ap-1", AcademicPeriodID: "period-1"}}
	periodResolver := resolver.NewSessionPeriodResolver(sessionRepo, scheduleRepo, apRepo)
	programResolver := resolver.NewSessionProgramResolver(sessionRepo, scheduleRepo, appProgram, program)
	checkinUC := NewCheckinByNISUseCase(sessionRepo, kesantrian, periodResolver, registration, attendance, santriProgram, programResolver)
	return NewSyncAttendanceFromFingerprintUseCase(sessionRepo, reader, checkinUC)
}

func openSessionWithRange(start, end time.Time) *sesEntity.ActivitySession {
	return &sesEntity.ActivitySession{
		ID:                 "session-1",
		ActivityScheduleID: "schedule-1",
		StartsAt:           start,
		EndsAt:             end,
		Status:             sesConst.ActivitySessionStatusOpen,
	}
}

func validKesantrianByNIS() *fakeKesantrianByNIS {
	return &fakeKesantrianByNIS{
		byNIS: map[string]*ports.SantriBasicInfo{
			"NIS001": {SantriID: "santri-1", Status: "SANTRI"},
			"NIS002": {SantriID: "santri-2", Status: "SANTRI"},
			"NIS003": {SantriID: "santri-3", Status: "SANTRI"},
		},
		errByNIS: map[string]error{},
	}
}

func TestSyncRecordsAllScans(t *testing.T) {
	start := time.Now().Add(-time.Hour)
	end := start.Add(2 * time.Hour)
	uc := newSyncUC(
		openSessionWithRange(start, end),
		validKesantrianByNIS(),
		&fakeRegistrationRepo{regBySantriAndPeriod: validRegistration()},
		&fakeAttendanceRepo{},
		&fakeSantriProgramRepo{activeBySantri: activeProgram("program-1")},
		&fakeAPProgramRepo{links: []*appEntity.ActivityPeriodProgram{{ID: "link-1", ActivityPeriodID: "ap-1", ProgramID: "program-1"}}},
		&fakeProgramRepo{},
		&fakeFingerprintReader{pins: []ports.FingerprintScanPin{
			{PIN: "NIS001", FirstScanAt: start.Add(time.Minute)},
			{PIN: "NIS002", FirstScanAt: start.Add(2 * time.Minute)},
		}},
	)

	resp, err := uc.Execute(context.Background(), "session-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalScans != 2 || resp.Recorded != 2 || resp.Skipped != 0 || len(resp.Errors) != 0 {
		t.Fatalf("expected 2 recorded, got %+v", resp)
	}
}

func TestSyncSkipsAlreadyRecorded(t *testing.T) {
	start := time.Now().Add(-time.Hour)
	end := start.Add(2 * time.Hour)
	existing := &attEntity.Attendance{ID: "att-1", ActivitySessionID: "session-1", SantriID: "santri-1"}
	uc := newSyncUC(
		openSessionWithRange(start, end),
		validKesantrianByNIS(),
		&fakeRegistrationRepo{regBySantriAndPeriod: validRegistration()},
		&fakeAttendanceRepo{findByResult: existing},
		&fakeSantriProgramRepo{activeBySantri: activeProgram("program-1")},
		&fakeAPProgramRepo{links: []*appEntity.ActivityPeriodProgram{{ID: "link-1", ActivityPeriodID: "ap-1", ProgramID: "program-1"}}},
		&fakeProgramRepo{},
		&fakeFingerprintReader{pins: []ports.FingerprintScanPin{{PIN: "NIS001"}}},
	)

	resp, err := uc.Execute(context.Background(), "session-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Recorded != 0 || resp.Skipped != 1 || len(resp.Errors) != 0 {
		t.Fatalf("expected 1 skipped, got %+v", resp)
	}
}

func TestSyncCollectsInvalidPinAsError(t *testing.T) {
	start := time.Now().Add(-time.Hour)
	end := start.Add(2 * time.Hour)
	kes := validKesantrianByNIS()
	kes.errByNIS["NIS999"] = kernel.New(application.ErrCodeNotFound)
	uc := newSyncUC(
		openSessionWithRange(start, end),
		kes,
		&fakeRegistrationRepo{regBySantriAndPeriod: validRegistration()},
		&fakeAttendanceRepo{},
		&fakeSantriProgramRepo{activeBySantri: activeProgram("program-1")},
		&fakeAPProgramRepo{links: []*appEntity.ActivityPeriodProgram{{ID: "link-1", ActivityPeriodID: "ap-1", ProgramID: "program-1"}}},
		&fakeProgramRepo{},
		&fakeFingerprintReader{pins: []ports.FingerprintScanPin{{PIN: "NIS999"}}},
	)

	resp, err := uc.Execute(context.Background(), "session-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Errors) != 1 || resp.Errors[0].PIN != "NIS999" {
		t.Fatalf("expected 1 error entry, got %+v", resp)
	}
}

func TestSyncRejectsNonOpenSession(t *testing.T) {
	session := openSessionWithRange(time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	session.Status = sesConst.ActivitySessionStatusScheduled
	uc := newSyncUC(
		session,
		validKesantrianByNIS(),
		&fakeRegistrationRepo{regBySantriAndPeriod: validRegistration()},
		&fakeAttendanceRepo{},
		&fakeSantriProgramRepo{activeBySantri: activeProgram("program-1")},
		&fakeAPProgramRepo{},
		&fakeProgramRepo{},
		&fakeFingerprintReader{},
	)

	_, err := uc.Execute(context.Background(), "session-1")
	if err == nil {
		t.Fatal("expected error for non-open session")
	}
	var ke *kernel.AppError
	if !errors.As(err, &ke) || ke.Code != application.ErrCodeUnprocessableEntity {
		t.Fatalf("expected unprocessable entity error, got %v", err)
	}
}

func TestSyncUsesSessionRange(t *testing.T) {
	start := time.Now().Add(-3 * time.Hour)
	end := start.Add(2 * time.Hour) // ends in the past → to harus di-clamp ke EndsAt
	reader := &fakeFingerprintReader{}
	uc := newSyncUC(
		openSessionWithRange(start, end),
		validKesantrianByNIS(),
		&fakeRegistrationRepo{regBySantriAndPeriod: validRegistration()},
		&fakeAttendanceRepo{},
		&fakeSantriProgramRepo{activeBySantri: activeProgram("program-1")},
		&fakeAPProgramRepo{},
		&fakeProgramRepo{},
		reader,
	)

	if _, err := uc.Execute(context.Background(), "session-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reader.gotFrom.Equal(start) {
		t.Fatalf("expected from = session.StartsAt, got %v", reader.gotFrom)
	}
	if !reader.gotTo.Equal(end) {
		t.Fatalf("expected to clamped to session.EndsAt (past), got %v", reader.gotTo)
	}
}
