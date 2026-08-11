package http

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/query"
	"sipon-be/internal/modules/akademik/domain/activity_schedule/constant"
	"sipon-be/internal/shared/httperror"
	"sipon-be/internal/shared/kernel"
	"sipon-be/internal/shared/respond"
)

type AkademikHandler struct {
	// program
	createProgramUC *command.CreateProgramUseCase
	updateProgramUC *command.UpdateProgramUseCase
	listProgramsUC  *query.ListProgramsUseCase
	getProgramUC    *query.GetProgramUseCase

	// academic period
	createPeriodUC *command.CreateAcademicPeriodUseCase
	updatePeriodUC *command.UpdateAcademicPeriodUseCase
	openPeriodUC   *command.OpenAcademicPeriodUseCase
	closePeriodUC  *command.CloseAcademicPeriodUseCase
	listPeriodsUC  *query.ListAcademicPeriodsUseCase
	getPeriodUC    *query.GetAcademicPeriodUseCase

	// santri registration
	registerSantriUC       *command.RegisterSantriUseCase
	completeRegistrationUC *command.CompleteRegistrationUseCase
	cancelRegistrationUC   *command.CancelRegistrationUseCase
	listRegistrationsUC    *query.ListSantriRegistrationsUseCase
	getRegistrationUC      *query.GetSantriRegistrationUseCase

	// activity
	createActivityUC *command.CreateActivityUseCase
	updateActivityUC *command.UpdateActivityUseCase
	listActivitiesUC *query.ListActivitiesUseCase
	getActivityUC    *query.GetActivityUseCase

	// activity period
	createActivityPeriodUC *command.CreateActivityPeriodUseCase
	activatePeriodUC       *command.ActivateActivityPeriodUseCase
	deactivatePeriodUC     *command.DeactivateActivityPeriodUseCase
	listActivityPeriodsUC  *query.ListActivityPeriodsUseCase

	// activity period program
	assignProgramUC    *command.AssignProgramUseCase
	removeProgramUC    *command.RemoveProgramUseCase
	listPeriodPrograms *query.ListActivityPeriodProgramsUseCase

	// activity schedule
	createScheduleUC *command.CreateScheduleUseCase
	updateScheduleUC *command.UpdateScheduleUseCase
	deleteScheduleUC *command.DeleteScheduleUseCase
	listSchedulesUC  *query.ListActivitySchedulesUseCase
	getScheduleUC    *query.GetActivityScheduleUseCase
	getCalendarUC    *query.GetScheduleCalendarUseCase

	// activity session
	createSessionUC   *command.CreateSessionUseCase
	openSessionUC     *command.OpenSessionUseCase
	cancelSessionUC   *command.CancelSessionUseCase
	completeSessionUC *command.CompleteSessionUseCase
	listSessionsUC    *query.ListActivitySessionsUseCase
	getSessionUC      *query.GetActivitySessionUseCase

	// attendance
	recordAttendanceUC   *command.RecordAttendanceUseCase
	updateAttendanceUC   *command.UpdateAttendanceUseCase
	listAttendanceUC     *query.ListAttendanceUseCase
	listEligibleSantriUC *query.ListEligibleSessionSantriUseCase

	// setting
	updateSettingUC *command.UpdateAkademikSettingUseCase
	getSettingUC    *query.GetAkademikSettingUseCase
}

func NewAkademikHandler(
	createProgramUC *command.CreateProgramUseCase,
	updateProgramUC *command.UpdateProgramUseCase,
	listProgramsUC *query.ListProgramsUseCase,
	getProgramUC *query.GetProgramUseCase,
	createPeriodUC *command.CreateAcademicPeriodUseCase,
	updatePeriodUC *command.UpdateAcademicPeriodUseCase,
	openPeriodUC *command.OpenAcademicPeriodUseCase,
	closePeriodUC *command.CloseAcademicPeriodUseCase,
	listPeriodsUC *query.ListAcademicPeriodsUseCase,
	getPeriodUC *query.GetAcademicPeriodUseCase,
	registerSantriUC *command.RegisterSantriUseCase,
	completeRegistrationUC *command.CompleteRegistrationUseCase,
	cancelRegistrationUC *command.CancelRegistrationUseCase,
	listRegistrationsUC *query.ListSantriRegistrationsUseCase,
	getRegistrationUC *query.GetSantriRegistrationUseCase,
	createActivityUC *command.CreateActivityUseCase,
	updateActivityUC *command.UpdateActivityUseCase,
	listActivitiesUC *query.ListActivitiesUseCase,
	getActivityUC *query.GetActivityUseCase,
	createActivityPeriodUC *command.CreateActivityPeriodUseCase,
	activatePeriodUC *command.ActivateActivityPeriodUseCase,
	deactivatePeriodUC *command.DeactivateActivityPeriodUseCase,
	listActivityPeriodsUC *query.ListActivityPeriodsUseCase,
	assignProgramUC *command.AssignProgramUseCase,
	removeProgramUC *command.RemoveProgramUseCase,
	listPeriodPrograms *query.ListActivityPeriodProgramsUseCase,
	createScheduleUC *command.CreateScheduleUseCase,
	updateScheduleUC *command.UpdateScheduleUseCase,
	deleteScheduleUC *command.DeleteScheduleUseCase,
	listSchedulesUC *query.ListActivitySchedulesUseCase,
	getScheduleUC *query.GetActivityScheduleUseCase,
	getCalendarUC *query.GetScheduleCalendarUseCase,
	createSessionUC *command.CreateSessionUseCase,
	openSessionUC *command.OpenSessionUseCase,
	cancelSessionUC *command.CancelSessionUseCase,
	completeSessionUC *command.CompleteSessionUseCase,
	listSessionsUC *query.ListActivitySessionsUseCase,
	getSessionUC *query.GetActivitySessionUseCase,
	recordAttendanceUC *command.RecordAttendanceUseCase,
	updateAttendanceUC *command.UpdateAttendanceUseCase,
	listAttendanceUC *query.ListAttendanceUseCase,
	listEligibleSantriUC *query.ListEligibleSessionSantriUseCase,
	updateSettingUC *command.UpdateAkademikSettingUseCase,
	getSettingUC *query.GetAkademikSettingUseCase,
) *AkademikHandler {
	return &AkademikHandler{
		createProgramUC:        createProgramUC,
		updateProgramUC:        updateProgramUC,
		listProgramsUC:         listProgramsUC,
		getProgramUC:           getProgramUC,
		createPeriodUC:         createPeriodUC,
		updatePeriodUC:         updatePeriodUC,
		openPeriodUC:           openPeriodUC,
		closePeriodUC:          closePeriodUC,
		listPeriodsUC:          listPeriodsUC,
		getPeriodUC:            getPeriodUC,
		registerSantriUC:       registerSantriUC,
		completeRegistrationUC: completeRegistrationUC,
		cancelRegistrationUC:   cancelRegistrationUC,
		listRegistrationsUC:    listRegistrationsUC,
		getRegistrationUC:      getRegistrationUC,
		createActivityUC:       createActivityUC,
		updateActivityUC:       updateActivityUC,
		listActivitiesUC:       listActivitiesUC,
		getActivityUC:          getActivityUC,
		createActivityPeriodUC: createActivityPeriodUC,
		activatePeriodUC:       activatePeriodUC,
		deactivatePeriodUC:     deactivatePeriodUC,
		listActivityPeriodsUC:  listActivityPeriodsUC,
		assignProgramUC:        assignProgramUC,
		removeProgramUC:        removeProgramUC,
		listPeriodPrograms:     listPeriodPrograms,
		createScheduleUC:       createScheduleUC,
		updateScheduleUC:       updateScheduleUC,
		deleteScheduleUC:       deleteScheduleUC,
		listSchedulesUC:        listSchedulesUC,
		getScheduleUC:          getScheduleUC,
		getCalendarUC:          getCalendarUC,
		createSessionUC:        createSessionUC,
		openSessionUC:          openSessionUC,
		cancelSessionUC:        cancelSessionUC,
		completeSessionUC:      completeSessionUC,
		listSessionsUC:         listSessionsUC,
		getSessionUC:           getSessionUC,
		recordAttendanceUC:     recordAttendanceUC,
		updateAttendanceUC:     updateAttendanceUC,
		listAttendanceUC:       listAttendanceUC,
		listEligibleSantriUC:   listEligibleSantriUC,
		updateSettingUC:        updateSettingUC,
		getSettingUC:           getSettingUC,
	}
}

// --- Program ---

func (h *AkademikHandler) ListPrograms(c *gin.Context) {
	var q dto.ProgramListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listProgramsUC.Execute(c.Request.Context(), q)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar program berhasil diambil", items, meta)
}

func (h *AkademikHandler) GetProgram(c *gin.Context) {
	resp, err := h.getProgramUC.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "program berhasil diambil", resp)
}

func (h *AkademikHandler) ListActivePrograms(c *gin.Context) {
	active := "active"
	items, _, err := h.listProgramsUC.Execute(c.Request.Context(), dto.ProgramListQuery{
		Status: &active,
		Page:   1,
		Limit:  100,
	})
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "daftar program aktif berhasil diambil", items)
}

func (h *AkademikHandler) CreateProgram(c *gin.Context) {
	var req dto.CreateProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createProgramUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "program berhasil dibuat", resp)
}

func (h *AkademikHandler) UpdateProgram(c *gin.Context) {
	var req dto.UpdateProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updateProgramUC.Execute(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "program berhasil diupdate", resp)
}

// --- Academic Period ---

func (h *AkademikHandler) ListPeriods(c *gin.Context) {
	var q dto.AcademicPeriodListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listPeriodsUC.Execute(c.Request.Context(), q)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar periode akademik berhasil diambil", items, meta)
}

func (h *AkademikHandler) GetPeriod(c *gin.Context) {
	resp, err := h.getPeriodUC.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "periode akademik berhasil diambil", resp)
}

func (h *AkademikHandler) CreatePeriod(c *gin.Context) {
	var req dto.CreateAcademicPeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createPeriodUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "periode akademik berhasil dibuat", resp)
}

func (h *AkademikHandler) UpdatePeriod(c *gin.Context) {
	var req dto.UpdateAcademicPeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updatePeriodUC.Execute(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "periode akademik berhasil diupdate", resp)
}

func (h *AkademikHandler) OpenPeriod(c *gin.Context) {
	resp, err := h.openPeriodUC.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "periode akademik berhasil dibuka", resp)
}

func (h *AkademikHandler) ClosePeriod(c *gin.Context) {
	resp, err := h.closePeriodUC.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "periode akademik berhasil ditutup", resp)
}

// --- Santri Registration ---

func (h *AkademikHandler) ListRegistrations(c *gin.Context) {
	var q dto.SantriRegistrationListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listRegistrationsUC.Execute(c.Request.Context(), q)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar registrasi berhasil diambil", items, meta)
}

func (h *AkademikHandler) GetRegistration(c *gin.Context) {
	resp, err := h.getRegistrationUC.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "registrasi berhasil diambil", resp)
}

func (h *AkademikHandler) RegisterSantri(c *gin.Context) {
	var req dto.CreateSantriRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.registerSantriUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "registrasi santri berhasil dibuat", resp)
}

func (h *AkademikHandler) CompleteRegistration(c *gin.Context) {
	resp, err := h.completeRegistrationUC.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "registrasi berhasil diselesaikan", resp)
}

func (h *AkademikHandler) CancelRegistration(c *gin.Context) {
	resp, err := h.cancelRegistrationUC.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "registrasi berhasil dibatalkan", resp)
}

// --- Activity ---

func (h *AkademikHandler) ListActivities(c *gin.Context) {
	var q dto.ActivityListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listActivitiesUC.Execute(c.Request.Context(), q)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar kegiatan berhasil diambil", items, meta)
}

func (h *AkademikHandler) GetActivity(c *gin.Context) {
	resp, err := h.getActivityUC.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "kegiatan berhasil diambil", resp)
}

func (h *AkademikHandler) CreateActivity(c *gin.Context) {
	var req dto.CreateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createActivityUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "kegiatan berhasil dibuat", resp)
}

func (h *AkademikHandler) UpdateActivity(c *gin.Context) {
	var req dto.UpdateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updateActivityUC.Execute(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "kegiatan berhasil diupdate", resp)
}

// --- Activity Period ---

func (h *AkademikHandler) ListActivityPeriods(c *gin.Context) {
	var q dto.ActivityPeriodListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listActivityPeriodsUC.Execute(c.Request.Context(), q)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar aktivasi kegiatan berhasil diambil", items, meta)
}

func (h *AkademikHandler) CreateActivityPeriod(c *gin.Context) {
	var req dto.CreateActivityPeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createActivityPeriodUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "aktivasi kegiatan berhasil dibuat", resp)
}

func (h *AkademikHandler) ActivateActivityPeriod(c *gin.Context) {
	resp, err := h.activatePeriodUC.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "aktivasi kegiatan berhasil diaktifkan", resp)
}

func (h *AkademikHandler) DeactivateActivityPeriod(c *gin.Context) {
	resp, err := h.deactivatePeriodUC.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "aktivasi kegiatan berhasil dinonaktifkan", resp)
}

// --- Activity Period Program ---

func (h *AkademikHandler) ListPeriodPrograms(c *gin.Context) {
	items, err := h.listPeriodPrograms.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "daftar program kegiatan berhasil diambil", items)
}

func (h *AkademikHandler) AssignProgram(c *gin.Context) {
	var req dto.AssignProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.assignProgramUC.Execute(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "program berhasil ditambahkan", resp)
}

func (h *AkademikHandler) RemoveProgram(c *gin.Context) {
	if err := h.removeProgramUC.Execute(c.Request.Context(), c.Param("id"), c.Param("programId")); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "program berhasil dihapus", nil)
}

// --- Activity Schedule ---

func (h *AkademikHandler) ListSchedules(c *gin.Context) {
	items, err := h.listSchedulesUC.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "daftar jadwal kegiatan berhasil diambil", items)
}

func (h *AkademikHandler) GetSchedule(c *gin.Context) {
	resp, err := h.getScheduleUC.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "jadwal kegiatan berhasil diambil", resp)
}

func (h *AkademikHandler) CreateSchedule(c *gin.Context) {
	var req dto.CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createScheduleUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "jadwal kegiatan berhasil dibuat", resp)
}

func (h *AkademikHandler) UpdateSchedule(c *gin.Context) {
	var req dto.UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updateScheduleUC.Execute(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "jadwal kegiatan berhasil diupdate", resp)
}

func (h *AkademikHandler) DeleteSchedule(c *gin.Context) {
	if err := h.deleteScheduleUC.Execute(c.Request.Context(), c.Param("id")); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "jadwal kegiatan berhasil dihapus", nil)
}

func (h *AkademikHandler) GetScheduleCalendar(c *gin.Context) {
	var q dto.ScheduleCalendarQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httperror.Handle(c, err)
		return
	}
	from, err := time.Parse("2006-01-02", q.From)
	if err != nil {
		httperror.Handle(c, kernel.New(application.ErrCodeBadRequest))
		return
	}
	to, err := time.Parse("2006-01-02", q.To)
	if err != nil {
		httperror.Handle(c, kernel.New(application.ErrCodeBadRequest))
		return
	}
	types := parseScheduleTypes(q.Types)
	resp, err := h.getCalendarUC.Execute(c.Request.Context(), from, to, q.AcademicPeriodID, types)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "kalender kegiatan berhasil diambil", resp)
}

func parseScheduleTypes(raw string) []constant.ActivityScheduleType {
	if raw == "" {
		return nil
	}
	seen := map[constant.ActivityScheduleType]struct{}{}
	parts := strings.Split(raw, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		t := constant.ActivityScheduleType(p)
		if t != constant.ActivityScheduleTypeOnce &&
			t != constant.ActivityScheduleTypeDaily &&
			t != constant.ActivityScheduleTypeWeekly &&
			t != constant.ActivityScheduleTypeMonthly &&
			t != constant.ActivityScheduleTypeYearly {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
	}
	out := make([]constant.ActivityScheduleType, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	return out
}

// --- Activity Session ---

func (h *AkademikHandler) ListSessions(c *gin.Context) {
	var q dto.ActivitySessionListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listSessionsUC.Execute(c.Request.Context(), q)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "daftar sesi kegiatan berhasil diambil", items, meta)
}

func (h *AkademikHandler) GetSession(c *gin.Context) {
	resp, err := h.getSessionUC.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "sesi kegiatan berhasil diambil", resp)
}

func (h *AkademikHandler) CreateSession(c *gin.Context) {
	var req dto.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createSessionUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "sesi kegiatan berhasil dibuat", resp)
}

func (h *AkademikHandler) CancelSession(c *gin.Context) {
	resp, err := h.cancelSessionUC.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "sesi kegiatan berhasil dibatalkan", resp)
}

func (h *AkademikHandler) OpenSession(c *gin.Context) {
	resp, err := h.openSessionUC.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "sesi kegiatan berhasil dibuka", resp)
}

func (h *AkademikHandler) CompleteSession(c *gin.Context) {
	resp, err := h.completeSessionUC.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "sesi kegiatan berhasil diselesaikan", resp)
}

// --- Attendance ---

func (h *AkademikHandler) ListAttendance(c *gin.Context) {
	items, err := h.listAttendanceUC.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "daftar absensi berhasil diambil", items)
}

func (h *AkademikHandler) ListEligibleSessionSantri(c *gin.Context) {
	items, err := h.listEligibleSantriUC.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "daftar santri yang berhak absen berhasil diambil", items)
}

func (h *AkademikHandler) RecordAttendance(c *gin.Context) {
	var req dto.RecordAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, err := h.recordAttendanceUC.Execute(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "absensi berhasil dicatat", items)
}

func (h *AkademikHandler) UpdateAttendance(c *gin.Context) {
	var req dto.UpdateAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updateAttendanceUC.Execute(c.Request.Context(), c.Param("id"), c.Param("santriId"), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "absensi berhasil diupdate", resp)
}

// --- Akademik Settings ---

func (h *AkademikHandler) GetAkademikSetting(c *gin.Context) {
	resp, err := h.getSettingUC.Execute(c.Request.Context())
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "pengaturan akademik berhasil diambil", resp)
}

func (h *AkademikHandler) UpdateAkademikSetting(c *gin.Context) {
	var req dto.UpdateAkademikSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updateSettingUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "pengaturan akademik berhasil diperbarui", resp)
}
