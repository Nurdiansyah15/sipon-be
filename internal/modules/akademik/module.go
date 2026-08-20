package akademik

import (
	"context"
	"database/sql"

	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/query"
	"sipon-be/internal/modules/akademik/application/resolver"
	"sipon-be/internal/modules/akademik/infrastructure/external"
	"sipon-be/internal/modules/akademik/infrastructure/fingerprintgateway"
	"sipon-be/internal/modules/akademik/infrastructure/kesantriangateway"
	"sipon-be/internal/modules/akademik/infrastructure/persistence"
	"sipon-be/internal/modules/akademik/infrastructure/schedulergateway"
	akademikHTTP "sipon-be/internal/modules/akademik/interfaces/http"
	akademikMQ "sipon-be/internal/modules/akademik/interfaces/mq"
	"sipon-be/internal/modules/fingerprint"
	"sipon-be/internal/modules/kesantrian"
	"sipon-be/internal/modules/scheduler"
	"sipon-be/internal/shared/config"
	messaging "sipon-be/internal/modules/messaging"
)

type Module struct {
	handler       *akademikHTTP.AkademikHandler
	jwtAuth       gin.HandlerFunc
	principalLoad gin.HandlerFunc

	getDefaultProgramIDUC *query.GetAkademikSettingUseCase
	assignSantriProgramUC *command.AssignSantriProgramUseCase
	getSantriProgramUC    *query.GetSantriProgramUseCase

	syncFingerprintUC  *command.SyncAttendanceFromFingerprintUseCase
	autoCloseSessionUC *command.AutoCloseSessionUseCase
	autoOpenSessionUC  *command.AutoOpenSessionUseCase
}

func NewModule(
	db *sql.DB,
	cfg *config.Config,
	kesantrianContract kesantrian.Contract,
	fingerprintContract fingerprint.Contract,
	schedulerContract scheduler.Contract,
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
	programTransferRequestRepo := persistence.NewPostgresProgramTransferRequestRepository(db)
	docRequirementRepo := persistence.NewPostgresHerregistrasiDocumentRequirementRepository(db)
	docRepo := persistence.NewPostgresHerregistrasiDocumentRepository(db)
	transactor := persistence.NewPostgresTransactor(db)

	kesantrianGW := kesantriangateway.New(kesantrianContract)
	fingerprintGW := fingerprintgateway.New(fingerprintContract)
	schedulerGW := schedulergateway.New(schedulerContract)
	periodResolver := resolver.NewSessionPeriodResolver(sessionRepo, scheduleRepo, activityPeriodRepo)
	programResolver := resolver.NewSessionProgramResolver(sessionRepo, scheduleRepo, activityPeriodProgramRepo, programRepo)

	fileUploader, _ := external.NewMinioFileUploader(
		cfg.Minio.Endpoint,
		cfg.Minio.PublicEndpoint,
		cfg.Minio.AccessKey,
		cfg.Minio.SecretKey,
		cfg.Minio.Bucket,
		cfg.Minio.PrivateBucket,
		cfg.Minio.UseSSL,
		cfg.Minio.InternalUseSSL,
	)

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
	completeRegistrationUC := command.NewCompleteRegistrationUseCase(registrationRepo, docRequirementRepo, docRepo)
	cancelRegistrationUC := command.NewCancelRegistrationUseCase(registrationRepo)
	listRegistrationsUC := query.NewListSantriRegistrationsUseCase(registrationRepo, periodRepo, kesantrianGW)
	getRegistrationUC := query.NewGetSantriRegistrationUseCase(registrationRepo, periodRepo)

	// herregistrasi dokumen & revisi
	createDocRequirementUC := command.NewCreateHerregistrasiDocumentRequirementUseCase(docRequirementRepo, periodRepo)
	updateDocRequirementUC := command.NewUpdateHerregistrasiDocumentRequirementUseCase(docRequirementRepo)
	deleteDocRequirementUC := command.NewDeleteHerregistrasiDocumentRequirementUseCase(docRequirementRepo)
	listDocRequirementsUC := query.NewListHerregistrasiDocumentRequirementsUseCase(docRequirementRepo)
	requestRevisionUC := command.NewRequestRevisionUseCase(registrationRepo)
	listRegistrationDocsUC := query.NewListHerregistrasiDocumentsUseCase(registrationRepo, docRequirementRepo, docRepo)
	verifyDocUC := command.NewVerifyHerregistrasiDocumentUseCase(docRepo)
	rejectDocUC := command.NewRejectHerregistrasiDocumentUseCase(docRepo)
	myHerregDetailUC := query.NewGetMyHerregistrasiDetailUseCase(kesantrianGW, periodRepo, registrationRepo, docRequirementRepo, docRepo)
	presignDocUC := command.NewPresignHerregistrasiDocumentUseCase(fileUploader)
	confirmDocUC := command.NewConfirmHerregistrasiDocumentUseCase(kesantrianGW, periodRepo, registrationRepo, docRequirementRepo, docRepo, fileUploader)
	deleteDocUC := command.NewDeleteHerregistrasiDocumentUseCase(kesantrianGW, registrationRepo, docRepo, fileUploader)
	downloadDocUC := query.NewDownloadHerregistrasiDocumentUseCase(kesantrianGW, registrationRepo, docRepo, fileUploader)

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
	getCalendarUC := query.NewGetScheduleCalendarUseCase(scheduleRepo, activityPeriodRepo, activityRepo)

	// activity session
	scheduleJobsUC := command.NewScheduleSessionJobsUseCase(
		schedulerGW,
		command.ScheduledJobTypes{
			FingerprintSync:  akademikMQ.RoutingFingerprintSync,
			SessionAutoClose: akademikMQ.RoutingSessionAutoClose,
		},
	)
	scheduleAutoOpenUC := command.NewScheduleAutoOpenUseCase(schedulerGW, akademikMQ.RoutingSessionAutoOpen)
	createSessionUC := command.NewCreateSessionUseCase(sessionRepo, scheduleRepo, scheduleAutoOpenUC)
	generateSessionsUC := command.NewGenerateSessionsFromScheduleUseCase(scheduleRepo, sessionRepo, transactor, scheduleAutoOpenUC)
	openSessionUC := command.NewOpenSessionUseCase(sessionRepo, scheduleJobsUC)
	cancelSessionUC := command.NewCancelSessionUseCase(sessionRepo)
	completeSessionUC := command.NewCompleteSessionUseCase(sessionRepo, attendanceRepo, santriProgramRepo, programResolver)
	autoCloseSessionUC := command.NewAutoCloseSessionUseCase(completeSessionUC, schedulerGW, akademikMQ.RoutingFingerprintSync)
	autoOpenSessionUC := command.NewAutoOpenSessionUseCase(sessionRepo, openSessionUC)
	listSessionsUC := query.NewListActivitySessionsUseCase(sessionRepo, scheduleRepo, activityPeriodRepo, activityRepo)
	getSessionUC := query.NewGetActivitySessionUseCase(sessionRepo, scheduleRepo, activityPeriodRepo, activityRepo, attendanceRepo, registrationRepo)

	// attendance
	recordAttendanceUC := command.NewRecordAttendanceUseCase(attendanceRepo, sessionRepo, registrationRepo, periodResolver, kesantrianGW)
	updateAttendanceUC := command.NewUpdateAttendanceUseCase(attendanceRepo, sessionRepo)
	listAttendanceUC := query.NewListAttendanceUseCase(attendanceRepo, sessionRepo, kesantrianGW)
	listEligibleSantriUC := query.NewListEligibleSessionSantriUseCase(registrationRepo, periodResolver, kesantrianGW)

	// setting
	updateSettingUC := command.NewUpdateAkademikSettingUseCase(settingRepo, programRepo)
	getSettingUC := query.NewGetAkademikSettingUseCase(settingRepo, programRepo)

	// santri program
	assignSantriProgramUC := command.NewAssignSantriProgramUseCase(santriProgramRepo, programRepo)
	getSantriProgramUC := query.NewGetSantriProgramUseCase(santriProgramRepo, programRepo)

	// santri program management (admin)
	assignSantriProgramAdminUC := command.NewAssignSantriProgramAdminUseCase(santriProgramRepo, programRepo, transactor)
	getSantriProgramAdminUC := query.NewGetSantriProgramAdminUseCase(santriProgramRepo, programRepo)
	listSantriProgramsUC := query.NewListSantriProgramsUseCase(kesantrianGW, santriProgramRepo, programRepo)

	// program transfer requests
	requestProgramTransferUC := command.NewRequestProgramTransferUseCase(programTransferRequestRepo, santriProgramRepo, programRepo, kesantrianGW)
	approveProgramTransferUC := command.NewApproveProgramTransferUseCase(programTransferRequestRepo, santriProgramRepo, programRepo, transactor)
	rejectProgramTransferUC := command.NewRejectProgramTransferUseCase(programTransferRequestRepo, programRepo)
	listProgramTransferRequestsUC := query.NewListProgramTransferRequestsUseCase(programTransferRequestRepo, programRepo, kesantrianGW)
	getProgramTransferRequestUC := query.NewGetProgramTransferRequestUseCase(programTransferRequestRepo, programRepo, kesantrianGW)
	listMyProgramTransferReqUC := query.NewListMyProgramTransferRequestsUseCase(programTransferRequestRepo, programRepo, kesantrianGW)

	// santri portal (non-admin)
	getMySummaryUC := query.NewGetMySummaryUseCase(kesantrianGW, periodRepo, registrationRepo, santriProgramRepo, programRepo)
	getMyProgramUC := query.NewGetMyProgramUseCase(kesantrianGW, santriProgramRepo, programRepo)
	listMyActivitiesUC := query.NewListMyActivitiesUseCase(kesantrianGW, periodRepo, santriProgramRepo, activityPeriodRepo, activityRepo, scheduleRepo)
	listMySchedulesUC := query.NewListMySchedulesUseCase(kesantrianGW, periodRepo, santriProgramRepo, activityPeriodRepo, activityRepo, scheduleRepo)
	applyHerregistrasiUC := command.NewApplyHerregistrasiUseCase(kesantrianGW, periodRepo, registrationRepo, santriProgramRepo)
	submitHerregistrasiUC := command.NewSubmitHerregistrasiUseCase(kesantrianGW, periodRepo, registrationRepo, docRequirementRepo, docRepo)

	// presensi & riwayat absensi
	checkinUC := command.NewCheckinByNISUseCase(sessionRepo, kesantrianGW, periodResolver, registrationRepo, attendanceRepo, santriProgramRepo, programResolver)
	syncFingerprintUC := command.NewSyncAttendanceFromFingerprintUseCase(sessionRepo, fingerprintGW, checkinUC)
	presensiInfoUC := query.NewGetPresensiSessionInfoUseCase(sessionRepo, scheduleRepo, activityPeriodRepo, activityRepo, periodRepo, registrationRepo, attendanceRepo)
	presensiListUC := query.NewListPresensiAttendanceUseCase(attendanceRepo, kesantrianGW)
	myAttendanceUC := query.NewGetMyAttendanceUseCase(kesantrianGW, periodRepo, santriProgramRepo, activityPeriodRepo, activityRepo, scheduleRepo, sessionRepo, attendanceRepo)

	handler := akademikHTTP.NewAkademikHandler(
		createProgramUC, updateProgramUC, listProgramsUC, getProgramUC,
		createPeriodUC, updatePeriodUC, openPeriodUC, closePeriodUC, listPeriodsUC, getPeriodUC,
		registerSantriUC, completeRegistrationUC, cancelRegistrationUC, listRegistrationsUC, getRegistrationUC,
		createActivityUC, updateActivityUC, listActivitiesUC, getActivityUC,
		createActivityPeriodUC, activatePeriodUC, deactivatePeriodUC, listActivityPeriodsUC,
		assignProgramUC, removeProgramUC, listPeriodProgramsUC,
		createScheduleUC, updateScheduleUC, deleteScheduleUC, listSchedulesUC, getScheduleUC, getCalendarUC,
		createSessionUC, generateSessionsUC, openSessionUC, cancelSessionUC, completeSessionUC, listSessionsUC, getSessionUC,
		recordAttendanceUC, updateAttendanceUC, listAttendanceUC, listEligibleSantriUC,
		updateSettingUC, getSettingUC,
		getMySummaryUC, getMyProgramUC, listMyActivitiesUC, listMySchedulesUC, applyHerregistrasiUC,
		submitHerregistrasiUC, myAttendanceUC,
		presensiInfoUC, presensiListUC, checkinUC, syncFingerprintUC,
		createDocRequirementUC, updateDocRequirementUC, deleteDocRequirementUC, listDocRequirementsUC,
		requestRevisionUC, listRegistrationDocsUC, verifyDocUC, rejectDocUC,
		myHerregDetailUC, presignDocUC, confirmDocUC, deleteDocUC, downloadDocUC,
		assignSantriProgramAdminUC, getSantriProgramAdminUC, listSantriProgramsUC,
		requestProgramTransferUC, approveProgramTransferUC, rejectProgramTransferUC,
		listProgramTransferRequestsUC, getProgramTransferRequestUC, listMyProgramTransferReqUC,
	)

	return &Module{
		handler:               handler,
		jwtAuth:               jwtAuth,
		principalLoad:         principalLoad,
		getDefaultProgramIDUC: getSettingUC,
		assignSantriProgramUC: assignSantriProgramUC,
		getSantriProgramUC:    getSantriProgramUC,
		syncFingerprintUC:     syncFingerprintUC,
		autoCloseSessionUC:    autoCloseSessionUC,
		autoOpenSessionUC:     autoOpenSessionUC,
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

// RegisterMessageHandlers mendaftarkan handler asynchronous module akademik ke
// shared messaging registry dan mengembalikan binding queue. Layer transport
// (scheduler worker / consumer MQ) memakai facade ini sebagai satu-satunya pintu
// registrasi — handler bisnis hidup di interfaces/mq, bukan di module.go.
func (m *Module) RegisterMessageHandlers(registry messaging.Contract) ([]messaging.Binding, error) {
	return akademikMQ.RegisterHandlers(registry, akademikMQ.Dependencies{
		FingerprintSync:  m.syncFingerprintUC,
		SessionAutoClose: m.autoCloseSessionUC,
		SessionAutoOpen:  m.autoOpenSessionUC,
	})
}
