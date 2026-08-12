package http

import (
	"github.com/gin-gonic/gin"

	"sipon-be/internal/shared/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, h *AkademikHandler, jwtAuth, principalLoad gin.HandlerFunc) {
	// Presensi — tanpa JWT (check-in santri via NIS dari halaman presensi).
	presensi := router.Group("/api/v1/web/akademik/presensi")
	{
		presensi.GET("/:sessionId", h.PresensiSessionInfo)
		presensi.POST("/:sessionId/checkin", h.Checkin)
		presensi.GET("/:sessionId/attendance", h.ListPresensiAttendance)
	}

	akademik := router.Group("/api/v1/web/akademik")
	akademik.Use(jwtAuth, principalLoad)
	{
		// Public: daftar program aktif untuk dipilih pendaftar PSB.
		akademik.GET("/programs/active", h.ListActivePrograms)

		// Santri portal (non-admin) — hanya butuh JWT, tanpa permission.
		akademik.GET("/my/summary", h.MySummary)
		akademik.GET("/my/program", h.MyProgram)
		akademik.GET("/my/kegiatan", h.MyActivities)
		akademik.GET("/my/jadwal", h.MySchedules)
		akademik.GET("/my/absensi", h.MyAttendance)
		akademik.GET("/my/absensi/periods", h.MyAttendancePeriods)
		akademik.POST("/my/program-transfer-requests", h.RequestProgramTransfer)
		akademik.GET("/my/program-transfer-requests", h.ListMyProgramTransferRequests)
		akademik.POST("/my/herregistrasi", h.ApplyHerregistrasi)
		akademik.POST("/my/herregistrasi/submit", h.SubmitHerregistrasi)
		akademik.GET("/my/herregistrasi", h.MyHerregistrasiDetail)
		akademik.POST("/my/herregistrasi/dokumen/presign", h.PresignMyHerregistrasiDocument)
		akademik.POST("/my/herregistrasi/dokumen/confirm", h.ConfirmMyHerregistrasiDocument)
		akademik.DELETE("/my/herregistrasi/dokumen/:id", h.DeleteMyHerregistrasiDocument)
		akademik.GET("/my/herregistrasi/dokumen/:id/download", h.DownloadMyHerregistrasiDocument)

		// Settings
		settings := akademik.Group("/settings")
		settings.Use(middleware.RequirePermission("manage_akademik"))
		{
			settings.GET("", h.GetAkademikSetting)
			settings.PUT("", h.UpdateAkademikSetting)
		}

		// Program
		programs := akademik.Group("/programs")
		programs.Use(middleware.RequirePermission("manage_akademik"))
		{
			programs.GET("", h.ListPrograms)
			programs.GET("/:id", h.GetProgram)
			programs.POST("", h.CreateProgram)
			programs.PUT("/:id", h.UpdateProgram)
		}

		// Academic Period
		periods := akademik.Group("/periods")
		periods.Use(middleware.RequirePermission("manage_akademik"))
		{
			periods.GET("", h.ListPeriods)
			periods.GET("/:id", h.GetPeriod)
			periods.POST("", h.CreatePeriod)
			periods.PUT("/:id", h.UpdatePeriod)
			periods.POST("/:id/open", h.OpenPeriod)
			periods.POST("/:id/close", h.ClosePeriod)
			periods.GET("/:id/dokumen-requirements", h.ListPeriodDocumentRequirements)
			periods.POST("/:id/dokumen-requirements", h.CreatePeriodDocumentRequirement)
			periods.PUT("/dokumen-requirements/:id", h.UpdatePeriodDocumentRequirement)
			periods.DELETE("/dokumen-requirements/:id", h.DeletePeriodDocumentRequirement)
		}

		// Santri Registration
		registrations := akademik.Group("/registrations")
		registrations.Use(middleware.RequirePermission("manage_akademik"))
		{
			registrations.GET("", h.ListRegistrations)
			registrations.GET("/:id", h.GetRegistration)
			registrations.POST("", h.RegisterSantri)
			registrations.POST("/:id/complete", h.CompleteRegistration)
			registrations.POST("/:id/cancel", h.CancelRegistration)
			registrations.POST("/:id/revision", h.RequestRevision)
			registrations.GET("/:id/dokumen", h.ListRegistrationDocuments)
			registrations.POST("/:id/dokumen/:dokumenId/verify", h.VerifyRegistrationDocument)
			registrations.POST("/:id/dokumen/:dokumenId/reject", h.RejectRegistrationDocument)
		}

		// Activity
		activities := akademik.Group("/activities")
		activities.Use(middleware.RequirePermission("manage_akademik"))
		{
			activities.GET("", h.ListActivities)
			activities.GET("/:id", h.GetActivity)
			activities.POST("", h.CreateActivity)
			activities.PUT("/:id", h.UpdateActivity)
		}

		// Activity Period
		activityPeriods := akademik.Group("/activity-periods")
		activityPeriods.Use(middleware.RequirePermission("manage_akademik"))
		{
			activityPeriods.GET("", h.ListActivityPeriods)
			activityPeriods.POST("", h.CreateActivityPeriod)
			activityPeriods.POST("/:id/activate", h.ActivateActivityPeriod)
			activityPeriods.POST("/:id/deactivate", h.DeactivateActivityPeriod)
			activityPeriods.GET("/:id/programs", h.ListPeriodPrograms)
			activityPeriods.POST("/:id/programs", h.AssignProgram)
			activityPeriods.DELETE("/:id/programs/:programId", h.RemoveProgram)
			activityPeriods.GET("/:id/schedules", h.ListSchedules)
		}

		// Activity Schedule
		schedules := akademik.Group("/schedules")
		schedules.Use(middleware.RequirePermission("manage_akademik"))
		{
			schedules.GET("/:id", h.GetSchedule)
			schedules.POST("", h.CreateSchedule)
			schedules.PUT("/:id", h.UpdateSchedule)
			schedules.DELETE("/:id", h.DeleteSchedule)
			schedules.POST("/:id/generate-sessions", h.GenerateSessionsFromSchedule)
		}

		// Schedule Calendar
		akademik.GET("/calendar", h.GetScheduleCalendar)

		// Activity Session
		sessions := akademik.Group("/sessions")
		sessions.Use(middleware.RequirePermission("manage_akademik"))
		{
			sessions.GET("", h.ListSessions)
			sessions.GET("/:id", h.GetSession)
			sessions.POST("", h.CreateSession)
			sessions.POST("/:id/open", h.OpenSession)
			sessions.POST("/:id/cancel", h.CancelSession)
			sessions.POST("/:id/complete", h.CompleteSession)
			sessions.GET("/:id/eligible-santri", h.ListEligibleSessionSantri)
			sessions.GET("/:id/attendance", h.ListAttendance)
			sessions.POST("/:id/attendance", h.RecordAttendance)
			sessions.PUT("/:id/attendance/:santriId", h.UpdateAttendance)
		}

		// Santri Program Management (Admin)
		santriPrograms := akademik.Group("/admin/santri")
		santriPrograms.Use(middleware.RequirePermission("manage_akademik"))
		{
			santriPrograms.GET("", h.ListSantriPrograms)
			santriPrograms.GET("/:santriId/program", h.GetSantriProgramAdmin)
			santriPrograms.PUT("/:santriId/program", h.AssignSantriProgramAdmin)
		}

		// Program Transfer Requests (Admin)
		programTransfers := akademik.Group("/admin/program-transfer-requests")
		programTransfers.Use(middleware.RequirePermission("manage_akademik"))
		{
			programTransfers.GET("", h.ListProgramTransferRequests)
			programTransfers.GET("/:id", h.GetProgramTransferRequest)
			programTransfers.POST("/:id/approve", h.ApproveProgramTransferRequest)
			programTransfers.POST("/:id/reject", h.RejectProgramTransferRequest)
		}
	}
}
