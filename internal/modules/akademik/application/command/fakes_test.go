package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application/ports"
	appEntity "sipon-be/internal/modules/akademik/domain/activity_period_program/entity"
	apRepo "sipon-be/internal/modules/akademik/domain/activity_period/repository"
	apEntity "sipon-be/internal/modules/akademik/domain/activity_period/entity"
	schConst "sipon-be/internal/modules/akademik/domain/activity_schedule/constant"
	schEntity "sipon-be/internal/modules/akademik/domain/activity_schedule/entity"
	sesEntity "sipon-be/internal/modules/akademik/domain/activity_session/entity"
	sesRepo "sipon-be/internal/modules/akademik/domain/activity_session/repository"
	attEntity "sipon-be/internal/modules/akademik/domain/attendance/entity"
	attRepo "sipon-be/internal/modules/akademik/domain/attendance/repository"
	progEntity "sipon-be/internal/modules/akademik/domain/program/entity"
	progRepo "sipon-be/internal/modules/akademik/domain/program/repository"
	ptrEntity "sipon-be/internal/modules/akademik/domain/program_transfer_request/entity"
	ptrRepo "sipon-be/internal/modules/akademik/domain/program_transfer_request/repository"
	regEntity "sipon-be/internal/modules/akademik/domain/santri_registration/entity"
	regRepo "sipon-be/internal/modules/akademik/domain/santri_registration/repository"
	spEntity "sipon-be/internal/modules/akademik/domain/santri_program/entity"
)

// --- Kesantrian reader ---

type fakeKesantrian struct {
	byNIS     *ports.SantriBasicInfo
	byUserID  *ports.SantriBasicInfo
	byID      *ports.SantriBasicInfo
}

func (f *fakeKesantrian) GetSantriByID(ctx context.Context, santriID string) (*ports.SantriBasicInfo, error) {
	return f.byID, nil
}
func (f *fakeKesantrian) GetSantriByUserID(ctx context.Context, userID string) (*ports.SantriBasicInfo, error) {
	return f.byUserID, nil
}
func (f *fakeKesantrian) GetSantriByNIS(ctx context.Context, nis string) (*ports.SantriBasicInfo, error) {
	return f.byNIS, nil
}
func (f *fakeKesantrian) ListActiveSantriWithUserID(ctx context.Context) ([]ports.SantriBasicInfo, error) {
	return nil, nil
}

// --- Session repo ---

type fakeSessionRepo struct {
	session *sesEntity.ActivitySession
	updated *sesEntity.ActivitySession
	saved   []*sesEntity.ActivitySession
	existing []*sesEntity.ActivitySession
}

func (f *fakeSessionRepo) Save(ctx context.Context, s *sesEntity.ActivitySession) error {
	f.saved = append(f.saved, s)
	return nil
}
func (f *fakeSessionRepo) Update(ctx context.Context, s *sesEntity.ActivitySession) error {
	f.updated = s
	return nil
}
func (f *fakeSessionRepo) FindByID(ctx context.Context, id string) (*sesEntity.ActivitySession, error) {
	return f.session, nil
}
func (f *fakeSessionRepo) ListByScheduleIDs(ctx context.Context, scheduleIDs []string) ([]*sesEntity.ActivitySession, error) {
	return f.existing, nil
}
func (f *fakeSessionRepo) List(ctx context.Context, q sesRepo.ActivitySessionListQuery) (*sesRepo.ActivitySessionListResult, error) {
	return &sesRepo.ActivitySessionListResult{Items: f.existing, Total: int64(len(f.existing))}, nil
}

// --- Attendance repo ---

type fakeAttendanceRepo struct {
	recordedSantriIDs []string
	saved             []*attEntity.Attendance
	findByResult      *attEntity.Attendance
}

func (f *fakeAttendanceRepo) Save(ctx context.Context, a *attEntity.Attendance) error {
	f.saved = append(f.saved, a)
	return nil
}
func (f *fakeAttendanceRepo) Update(ctx context.Context, a *attEntity.Attendance) error { return nil }
func (f *fakeAttendanceRepo) FindByID(ctx context.Context, id string) (*attEntity.Attendance, error) {
	return nil, nil
}
func (f *fakeAttendanceRepo) FindBySessionAndSantri(ctx context.Context, sessionID, santriID string) (*attEntity.Attendance, error) {
	return f.findByResult, nil
}
func (f *fakeAttendanceRepo) ListBySession(ctx context.Context, sessionID string) ([]*attEntity.Attendance, error) {
	return nil, nil
}
func (f *fakeAttendanceRepo) ListSantriIDsBySession(ctx context.Context, sessionID string) ([]string, error) {
	return f.recordedSantriIDs, nil
}
func (f *fakeAttendanceRepo) ListBySantriAndPeriod(ctx context.Context, santriID, academicPeriodID string) ([]*attRepo.AttendanceWithSession, error) {
	return nil, nil
}

// --- Santri program repo ---

type fakeSantriProgramRepo struct {
	santriByProgram map[string][]string
	activeBySantri  *spEntity.SantriProgram
	deactivated     bool
	saved           []*spEntity.SantriProgram
}

func (f *fakeSantriProgramRepo) Save(ctx context.Context, sp *spEntity.SantriProgram) error {
	f.saved = append(f.saved, sp)
	return nil
}
func (f *fakeSantriProgramRepo) FindActiveBySantriID(ctx context.Context, santriID string) (*spEntity.SantriProgram, error) {
	return f.activeBySantri, nil
}
func (f *fakeSantriProgramRepo) FindBySantriID(ctx context.Context, santriID string) ([]*spEntity.SantriProgram, error) {
	return nil, nil
}
func (f *fakeSantriProgramRepo) ListActiveSantriIDsByProgramID(ctx context.Context, programID string) ([]string, error) {
	return f.santriByProgram[programID], nil
}
func (f *fakeSantriProgramRepo) ListActive(ctx context.Context) ([]*spEntity.SantriProgram, error) {
	return nil, nil
}
func (f *fakeSantriProgramRepo) DeactivateAllBySantriID(ctx context.Context, santriID string) error {
	f.deactivated = true
	return nil
}

// --- Schedule repo ---

type fakeScheduleRepo struct {
	schedule *schEntity.ActivitySchedule
	weeklies []schEntity.ActivityScheduleWeekly
	month    []schEntity.ActivityScheduleMonthly
	year     []schEntity.ActivityScheduleYearly
}

func (f *fakeScheduleRepo) Save(ctx context.Context, s *schEntity.ActivitySchedule) error { return nil }
func (f *fakeScheduleRepo) Update(ctx context.Context, s *schEntity.ActivitySchedule) error {
	return nil
}
func (f *fakeScheduleRepo) FindByID(ctx context.Context, id string) (*schEntity.ActivitySchedule, error) {
	return f.schedule, nil
}
func (f *fakeScheduleRepo) FindByIDs(ctx context.Context, ids []string) ([]*schEntity.ActivitySchedule, error) {
	return []*schEntity.ActivitySchedule{f.schedule}, nil
}
func (f *fakeScheduleRepo) ListByActivityPeriod(ctx context.Context, activityPeriodID string) ([]*schEntity.ActivitySchedule, error) {
	return nil, nil
}
func (f *fakeScheduleRepo) ListByActivityPeriodIDs(ctx context.Context, activityPeriodIDs []string) ([]*schEntity.ActivitySchedule, error) {
	return nil, nil
}
func (f *fakeScheduleRepo) ReplaceWeeklies(ctx context.Context, id string, days []schConst.DayOfWeek) error {
	return nil
}
func (f *fakeScheduleRepo) ReplaceMonthlies(ctx context.Context, id string, days []int) error { return nil }
func (f *fakeScheduleRepo) ReplaceYearlies(ctx context.Context, id string, dates []schEntity.YearlyDate) error {
	return nil
}
func (f *fakeScheduleRepo) ListWeeklies(ctx context.Context, scheduleID string) ([]schEntity.ActivityScheduleWeekly, error) {
	return f.weeklies, nil
}
func (f *fakeScheduleRepo) ListMonthlies(ctx context.Context, scheduleID string) ([]schEntity.ActivityScheduleMonthly, error) {
	return f.month, nil
}
func (f *fakeScheduleRepo) ListYearlies(ctx context.Context, scheduleID string) ([]schEntity.ActivityScheduleYearly, error) {
	return f.year, nil
}

// --- Activity period program repo ---

type fakeAPProgramRepo struct {
	links []*appEntity.ActivityPeriodProgram
}

func (f *fakeAPProgramRepo) Save(ctx context.Context, p *appEntity.ActivityPeriodProgram) error {
	return nil
}
func (f *fakeAPProgramRepo) Delete(ctx context.Context, id string) error { return nil }
func (f *fakeAPProgramRepo) FindByActivityPeriodAndProgram(ctx context.Context, activityPeriodID, programID string) (*appEntity.ActivityPeriodProgram, error) {
	return nil, nil
}
func (f *fakeAPProgramRepo) ListByActivityPeriod(ctx context.Context, activityPeriodID string) ([]*appEntity.ActivityPeriodProgram, error) {
	return f.links, nil
}

// --- Program repo ---

type fakeProgramRepo struct {
	activeIDs   []string
	programByID *progEntity.Program
	programByID2 *progEntity.Program
}

func (f *fakeProgramRepo) Save(ctx context.Context, p *progEntity.Program) error { return nil }
func (f *fakeProgramRepo) Update(ctx context.Context, p *progEntity.Program) error {
	return nil
}
func (f *fakeProgramRepo) FindByID(ctx context.Context, id string) (*progEntity.Program, error) {
	if f.programByID2 != nil && f.programByID2.ID == id {
		return f.programByID2, nil
	}
	return f.programByID, nil
}
func (f *fakeProgramRepo) FindByCode(ctx context.Context, code string) (*progEntity.Program, error) {
	return nil, nil
}
func (f *fakeProgramRepo) FindByIDs(ctx context.Context, ids []string) ([]*progEntity.Program, error) {
	return nil, nil
}
func (f *fakeProgramRepo) ListActiveIDs(ctx context.Context) ([]string, error) {
	return f.activeIDs, nil
}
func (f *fakeProgramRepo) List(ctx context.Context, q progRepo.ProgramListQuery) (*progRepo.ProgramListResult, error) {
	return &progRepo.ProgramListResult{Items: nil, Total: 0}, nil
}

// --- Activity period repo (for period resolver in checkin tests) ---

type fakeAPRepo struct {
	period *apEntity.ActivityPeriod
}

func (f *fakeAPRepo) Save(ctx context.Context, p *apEntity.ActivityPeriod) error { return nil }
func (f *fakeAPRepo) Update(ctx context.Context, p *apEntity.ActivityPeriod) error {
	return nil
}
func (f *fakeAPRepo) FindByID(ctx context.Context, id string) (*apEntity.ActivityPeriod, error) {
	return f.period, nil
}
func (f *fakeAPRepo) FindByActivityAndPeriod(ctx context.Context, activityID, academicPeriodID string) (*apEntity.ActivityPeriod, error) {
	return nil, nil
}
func (f *fakeAPRepo) FindByIDs(ctx context.Context, ids []string) ([]*apEntity.ActivityPeriod, error) {
	return nil, nil
}
func (f *fakeAPRepo) ListByPeriodAndProgram(ctx context.Context, periodID, programID string) ([]*apEntity.ActivityPeriod, error) {
	return nil, nil
}
func (f *fakeAPRepo) List(ctx context.Context, q apRepo.ActivityPeriodListQuery) (*apRepo.ActivityPeriodListResult, error) {
	return &apRepo.ActivityPeriodListResult{Items: nil, Total: 0}, nil
}

// --- Registration repo ---

type fakeRegistrationRepo struct {
	regBySantriAndPeriod *regEntity.SantriRegistration
}

func (f *fakeRegistrationRepo) Save(ctx context.Context, r *regEntity.SantriRegistration) error {
	return nil
}
func (f *fakeRegistrationRepo) Update(ctx context.Context, r *regEntity.SantriRegistration) error {
	return nil
}
func (f *fakeRegistrationRepo) FindByID(ctx context.Context, id string) (*regEntity.SantriRegistration, error) {
	return nil, nil
}
func (f *fakeRegistrationRepo) FindBySantriAndPeriod(ctx context.Context, santriID, academicPeriodID string) (*regEntity.SantriRegistration, error) {
	return f.regBySantriAndPeriod, nil
}
func (f *fakeRegistrationRepo) List(ctx context.Context, q regRepo.SantriRegistrationListQuery) (*regRepo.SantriRegistrationListResult, error) {
	return &regRepo.SantriRegistrationListResult{Items: nil, Total: 0}, nil
}
func (f *fakeRegistrationRepo) ListCompletedByAcademicPeriod(ctx context.Context, academicPeriodID string) ([]*regEntity.SantriRegistration, error) {
	return nil, nil
}

// --- Program transfer request repo ---

type fakePtrRepo struct {
	saved    []*ptrEntity.ProgramTransferRequest
	byID     *ptrEntity.ProgramTransferRequest
	pending  *ptrEntity.ProgramTransferRequest
	updated  *ptrEntity.ProgramTransferRequest
	list     []*ptrEntity.ProgramTransferRequest
}

func (f *fakePtrRepo) Save(ctx context.Context, req *ptrEntity.ProgramTransferRequest) error {
	f.saved = append(f.saved, req)
	return nil
}
func (f *fakePtrRepo) Update(ctx context.Context, req *ptrEntity.ProgramTransferRequest) error {
	f.updated = req
	return nil
}
func (f *fakePtrRepo) FindByID(ctx context.Context, id string) (*ptrEntity.ProgramTransferRequest, error) {
	return f.byID, nil
}
func (f *fakePtrRepo) FindPendingBySantriID(ctx context.Context, santriID string) (*ptrEntity.ProgramTransferRequest, error) {
	return f.pending, nil
}
func (f *fakePtrRepo) List(ctx context.Context, q ptrRepo.ProgramTransferRequestListQuery) (*ptrRepo.ProgramTransferRequestListResult, error) {
	return &ptrRepo.ProgramTransferRequestListResult{Items: f.list, Total: int64(len(f.list))}, nil
}

// helper untuk membangun request transfer pending
func newPtrFixture(id, santriID, fromProg, toProg string) *ptrEntity.ProgramTransferRequest {
	req, _ := ptrEntity.NewProgramTransferRequest(id, santriID, fromProg, toProg, nil)
	return req
}
