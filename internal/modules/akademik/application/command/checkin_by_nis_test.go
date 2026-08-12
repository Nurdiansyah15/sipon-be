package command

import (
	"context"
	"testing"
	"time"

	appEntity "sipon-be/internal/modules/akademik/domain/activity_period_program/entity"
	apEntity "sipon-be/internal/modules/akademik/domain/activity_period/entity"
	schEntity "sipon-be/internal/modules/akademik/domain/activity_schedule/entity"
	sesConst "sipon-be/internal/modules/akademik/domain/activity_session/constant"
	sesEntity "sipon-be/internal/modules/akademik/domain/activity_session/entity"
	attConst "sipon-be/internal/modules/akademik/domain/attendance/constant"
	regConst "sipon-be/internal/modules/akademik/domain/santri_registration/constant"
	regEntity "sipon-be/internal/modules/akademik/domain/santri_registration/entity"
	spEntity "sipon-be/internal/modules/akademik/domain/santri_program/entity"
	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/ports"
)

func newCheckinUC(
	session *sesEntity.ActivitySession,
	kesantrian *fakeKesantrian,
	registration *fakeRegistrationRepo,
	attendance *fakeAttendanceRepo,
	santriProgram *fakeSantriProgramRepo,
	appProgram *fakeAPProgramRepo,
	program *fakeProgramRepo,
) *CheckinByNISUseCase {
	sessionRepo := &fakeSessionRepo{session: session}
	scheduleRepo := &fakeScheduleRepo{schedule: &schEntity.ActivitySchedule{ID: session.ActivityScheduleID, ActivityPeriodID: "ap-1"}}
	apRepo := &fakeAPRepo{period: &apEntity.ActivityPeriod{ID: "ap-1", AcademicPeriodID: "period-1"}}
	periodResolver := application.NewSessionPeriodResolver(sessionRepo, scheduleRepo, apRepo)
	programResolver := application.NewSessionProgramResolver(sessionRepo, scheduleRepo, appProgram, program)
	return NewCheckinByNISUseCase(sessionRepo, kesantrian, periodResolver, registration, attendance, santriProgram, programResolver)
}

func openSessionFixture() *sesEntity.ActivitySession {
	now := time.Now()
	return &sesEntity.ActivitySession{
		ID:                 "session-1",
		ActivityScheduleID: "schedule-1",
		StartsAt:           now,
		EndsAt:             now.Add(time.Hour),
		Status:             sesConst.ActivitySessionStatusOpen,
	}
}

func validRegistration() *regEntity.SantriRegistration {
	return &regEntity.SantriRegistration{ID: "reg-1", SantriID: "santri-1", AcademicPeriodID: "period-1", Status: regConst.SantriRegistrationStatusCompleted}
}

func activeProgram(programID string) *spEntity.SantriProgram {
	return &spEntity.SantriProgram{ID: "sp-1", SantriID: "santri-1", ProgramID: programID, IsActive: true}
}

func TestCheckinAcceptedFromMatchingProgram(t *testing.T) {
	fx := openSessionFixture()
	uc := newCheckinUC(
		fx,
		&fakeKesantrian{byNIS: &ports.SantriBasicInfo{SantriID: "santri-1", Status: "SANTRI", Fullname: strPtr("Ahmad")}},
		&fakeRegistrationRepo{regBySantriAndPeriod: validRegistration()},
		&fakeAttendanceRepo{},
		&fakeSantriProgramRepo{activeBySantri: activeProgram("program-1")},
		&fakeAPProgramRepo{links: []*appEntity.ActivityPeriodProgram{{ID: "link-1", ActivityPeriodID: "ap-1", ProgramID: "program-1"}}},
		&fakeProgramRepo{},
	)
	resp, err := uc.Execute(context.Background(), "session-1", "NIS001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.Attendance.SantriID != "santri-1" {
		t.Fatalf("expected successful checkin for santri-1, got %+v", resp)
	}
}

func TestCheckinRejectedFromOtherProgram(t *testing.T) {
	fx := openSessionFixture()
	uc := newCheckinUC(
		fx,
		&fakeKesantrian{byNIS: &ports.SantriBasicInfo{SantriID: "santri-1", Status: "SANTRI", Fullname: strPtr("Ahmad")}},
		&fakeRegistrationRepo{regBySantriAndPeriod: validRegistration()},
		&fakeAttendanceRepo{},
		&fakeSantriProgramRepo{activeBySantri: activeProgram("program-9")},
		&fakeAPProgramRepo{links: []*appEntity.ActivityPeriodProgram{{ID: "link-1", ActivityPeriodID: "ap-1", ProgramID: "program-1"}}},
		&fakeProgramRepo{},
	)
	_, err := uc.Execute(context.Background(), "session-1", "NIS001")
	if err == nil {
		t.Fatal("expected rejection for santri from other program")
	}
}

func TestCheckinRejectedWithoutProgram(t *testing.T) {
	fx := openSessionFixture()
	uc := newCheckinUC(
		fx,
		&fakeKesantrian{byNIS: &ports.SantriBasicInfo{SantriID: "santri-1", Status: "SANTRI", Fullname: strPtr("Ahmad")}},
		&fakeRegistrationRepo{regBySantriAndPeriod: validRegistration()},
		&fakeAttendanceRepo{},
		&fakeSantriProgramRepo{activeBySantri: nil},
		&fakeAPProgramRepo{links: []*appEntity.ActivityPeriodProgram{{ID: "link-1", ActivityPeriodID: "ap-1", ProgramID: "program-1"}}},
		&fakeProgramRepo{},
	)
	_, err := uc.Execute(context.Background(), "session-1", "NIS001")
	if err == nil {
		t.Fatal("expected rejection for santri without program")
	}
}

func TestCheckinRejectedNotHerreg(t *testing.T) {
	fx := openSessionFixture()
	uc := newCheckinUC(
		fx,
		&fakeKesantrian{byNIS: &ports.SantriBasicInfo{SantriID: "santri-1", Status: "SANTRI", Fullname: strPtr("Ahmad")}},
		&fakeRegistrationRepo{regBySantriAndPeriod: &regEntity.SantriRegistration{ID: "reg-1", SantriID: "santri-1", AcademicPeriodID: "period-1", Status: "pending"}},
		&fakeAttendanceRepo{},
		&fakeSantriProgramRepo{activeBySantri: activeProgram("program-1")},
		&fakeAPProgramRepo{links: []*appEntity.ActivityPeriodProgram{{ID: "link-1", ActivityPeriodID: "ap-1", ProgramID: "program-1"}}},
		&fakeProgramRepo{},
	)
	_, err := uc.Execute(context.Background(), "session-1", "NIS001")
	if err == nil {
		t.Fatal("expected rejection for santri who has not completed herreg")
	}
}

func TestCheckinAcceptedAnyProgramWhenNoScope(t *testing.T) {
	fx := openSessionFixture()
	uc := newCheckinUC(
		fx,
		&fakeKesantrian{byNIS: &ports.SantriBasicInfo{SantriID: "santri-1", Status: "SANTRI", Fullname: strPtr("Ahmad")}},
		&fakeRegistrationRepo{regBySantriAndPeriod: validRegistration()},
		&fakeAttendanceRepo{},
		&fakeSantriProgramRepo{activeBySantri: activeProgram("program-2")},
		&fakeAPProgramRepo{}, // no scope
		&fakeProgramRepo{activeIDs: []string{"program-1", "program-2"}},
	)
	resp, err := uc.Execute(context.Background(), "session-1", "NIS001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.Attendance.Status != string(attConst.AttendanceStatusPresent) {
		t.Fatalf("expected successful checkin, got %+v", resp)
	}
}

func strPtr(s string) *string { return &s }
