package akademik

import (
	"context"
	"database/sql"

	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/query"
	"sipon-be/internal/modules/akademik/infrastructure/kesantriangateway"
	"sipon-be/internal/modules/akademik/infrastructure/persistence"
	akademikHTTP "sipon-be/internal/modules/akademik/interfaces/http"
	"sipon-be/internal/modules/kesantrian"
	"sipon-be/internal/shared/config"
)

type Module struct {
	handler       *akademikHTTP.AkademikHandler
	jwtAuth       gin.HandlerFunc
	principalLoad gin.HandlerFunc

	getDefaultProgramIDUC    *query.GetAkademikSettingUseCase
	assignSantriProgramUC    *command.AssignSantriProgramUseCase
	getSantriProgramUC       *query.GetSantriProgramUseCase
}

func NewModule(
	db *sql.DB,
	cfg *config.Config,
	kesantrianContract kesantrian.Contract,
	jwtAuth gin.HandlerFunc,
	principalLoad gin.HandlerFunc,
) *Module {
	programRepo := persistence.NewPostgresProgramRepository(db)
	periodRepo := persistence.NewPostgresAcademicPeriodRepository(db)
	registrationRepo := persistence.NewPostgresSantriRegistrationRepository(db)
	activityRepo := persistence.NewPostgresActivityRepository(db)
	activityPeriodRepo := persistence.NewPostgresActivityPeriodRepository(db)
	activityPeriodProgramRepo := persistence.NewPostgresActivityPeriodProgramRepository(db)
	scheduleRepo := persistence.NewPostgresActivityScheduleRepository(db)
	sessionRepo := persistence.NewPostgresActivitySessionRepository(db)
	attendanceRepo := persistence.NewPostgresAttendanceRepository(db)
	settingRepo := persistence.NewPostgresAkademikSettingRepository(db)
	santriProgramRepo := persistence.NewPostgresSantriProgramRepository(db)
	transactor := persistence.NewPostgresTransactor(db)

	kesantrianGW := kesantriangateway.New(kesantrianContract)

	// program
	createProgramUC := command.NewCreateProgramUseCase(programRepo)
	updateProgramUC := command.NewUpdateProgramUseCase(programRepo)
	listProgramsUC := query.NewListProgramsUseCase(programRepo)
	getProgramUC := query.NewGetProgramUseCase(programRepo)

	// academic period
	createPeriodUC := command.NewCreateAcademicPeriodUseCase(periodRepo)
	updatePeriodUC := command.NewUpdateAcademicPeriodUseCase(periodRepo)
	openPeriodUC := command.NewOpenAcademicPeriodUseCase(periodRepo)
	closePeriodUC := command.NewCloseAcademicPeriodUseCase(periodRepo)
	listPeriodsUC := query.NewListAcademicPeriodsUseCase(periodRepo)
	getPeriodUC := query.NewGetAcademicPeriodUseCase(periodRepo)

	// santri registration
	registerSantriUC := command.NewRegisterSantriUseCase(registrationRepo, periodRepo, kesantrianGW)
	completeRegistrationUC := command.NewCompleteRegistrationUseCase(registrationRepo)
	cancelRegistrationUC := command.NewCancelRegistrationUseCase(registrationRepo)
	listRegistrationsUC := query.NewListSantriRegistrationsUseCase(registrationRepo, periodRepo, kesantrianGW)
	getRegistrationUC := query.NewGetSantriRegistrationUseCase(registrationRepo, periodRepo)

	// activity
	createActivityUC := command.NewCreateActivityUseCase(activityRepo)
	updateActivityUC := command.NewUpdateActivityUseCase(activityRepo)
	listActivitiesUC := query.NewListActivitiesUseCase(activityRepo)
	getActivityUC := query.NewGetActivityUseCase(activityRepo)

	// activity period
	createActivityPeriodUC := command.NewCreateActivityPeriodUseCase(activityPeriodRepo, activityRepo, periodRepo)
	activatePeriodUC := command.NewActivateActivityPeriodUseCase(activityPeriodRepo)
	deactivatePeriodUC := command.NewDeactivateActivityPeriodUseCase(activityPeriodRepo)
	listActivityPeriodsUC := query.NewListActivityPeriodsUseCase(activityPeriodRepo, activityRepo, periodRepo)

	// activity period program
	assignProgramUC := command.NewAssignProgramUseCase(programRepo, activityPeriodRepo, activityPeriodProgramRepo)
	removeProgramUC := command.NewRemoveProgramUseCase(programRepo, activityPeriodRepo, activityPeriodProgramRepo)
	listPeriodProgramsUC := query.NewListActivityPeriodProgramsUseCase(programRepo, activityPeriodProgramRepo)

	// activity schedule
	createScheduleUC := command.NewCreateScheduleUseCase(scheduleRepo, activityPeriodRepo, transactor)
	updateScheduleUC := command.NewUpdateScheduleUseCase(scheduleRepo, transactor)
	deleteScheduleUC := command.NewDeleteScheduleUseCase(scheduleRepo)
	listSchedulesUC := query.NewListActivitySchedulesUseCase(scheduleRepo, activityPeriodRepo, activityRepo)
	getScheduleUC := query.NewGetActivityScheduleUseCase(scheduleRepo, activityPeriodRepo, activityRepo)

	// activity session
	createSessionUC := command.NewCreateSessionUseCase(sessionRepo, scheduleRepo)
	cancelSessionUC := command.NewCancelSessionUseCase(sessionRepo)
	completeSessionUC := command.NewCompleteSessionUseCase(sessionRepo)
	listSessionsUC := query.NewListActivitySessionsUseCase(sessionRepo, scheduleRepo, activityPeriodRepo, activityRepo)
	getSessionUC := query.NewGetActivitySessionUseCase(sessionRepo, scheduleRepo, activityPeriodRepo, activityRepo, attendanceRepo)

	// attendance
	recordAttendanceUC := command.NewRecordAttendanceUseCase(attendanceRepo, sessionRepo, kesantrianGW)
	updateAttendanceUC := command.NewUpdateAttendanceUseCase(attendanceRepo)
	listAttendanceUC := query.NewListAttendanceUseCase(attendanceRepo, sessionRepo, kesantrianGW)

	// setting
	updateSettingUC := command.NewUpdateAkademikSettingUseCase(settingRepo, programRepo)
	getSettingUC := query.NewGetAkademikSettingUseCase(settingRepo, programRepo)

	// santri program
	assignSantriProgramUC := command.NewAssignSantriProgramUseCase(santriProgramRepo, programRepo)
	getSantriProgramUC := query.NewGetSantriProgramUseCase(santriProgramRepo, programRepo)

	handler := akademikHTTP.NewAkademikHandler(
		createProgramUC, updateProgramUC, listProgramsUC, getProgramUC,
		createPeriodUC, updatePeriodUC, openPeriodUC, closePeriodUC, listPeriodsUC, getPeriodUC,
		registerSantriUC, completeRegistrationUC, cancelRegistrationUC, listRegistrationsUC, getRegistrationUC,
		createActivityUC, updateActivityUC, listActivitiesUC, getActivityUC,
		createActivityPeriodUC, activatePeriodUC, deactivatePeriodUC, listActivityPeriodsUC,
		assignProgramUC, removeProgramUC, listPeriodProgramsUC,
		createScheduleUC, updateScheduleUC, deleteScheduleUC, listSchedulesUC, getScheduleUC,
		createSessionUC, cancelSessionUC, completeSessionUC, listSessionsUC, getSessionUC,
		recordAttendanceUC, updateAttendanceUC, listAttendanceUC,
		updateSettingUC, getSettingUC,
	)

	return &Module{
		handler:                 handler,
		jwtAuth:                 jwtAuth,
		principalLoad:           principalLoad,
		getDefaultProgramIDUC:   getSettingUC,
		assignSantriProgramUC:   assignSantriProgramUC,
		getSantriProgramUC:      getSantriProgramUC,
	}
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	grp := router.Group("/")
	akademikHTTP.RegisterRoutes(grp, m.handler, m.jwtAuth, m.principalLoad)
}

func (m *Module) GetDefaultProgramID(ctx context.Context) (*string, error) {
	setting, err := m.getDefaultProgramIDUC.Execute(ctx)
	if err != nil {
		return nil, err
	}
	return setting.DefaultProgramID, nil
}

func (m *Module) AssignSantriProgram(ctx context.Context, santriID, programID string) error {
	return m.assignSantriProgramUC.Execute(ctx, santriID, programID)
}

func (m *Module) GetSantriProgram(ctx context.Context, santriID string) (*SantriProgramInfo, error) {
	info, err := m.getSantriProgramUC.Execute(ctx, santriID)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, nil
	}
	return &SantriProgramInfo{
		SantriID:    info.SantriID,
		ProgramID:   info.ProgramID,
		ProgramCode: info.ProgramCode,
		ProgramName: info.ProgramName,
		IsActive:    info.IsActive,
	}, nil
}
