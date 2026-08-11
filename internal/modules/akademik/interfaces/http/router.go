package http

import (
	"github.com/gin-gonic/gin"

	"sipon-be/internal/shared/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, h *AkademikHandler, jwtAuth, principalLoad gin.HandlerFunc) {
	akademik := router.Group("/api/v1/web/akademik")
	akademik.Use(jwtAuth, principalLoad)
	{
		// Public: daftar program aktif untuk dipilih pendaftar PSB.
		akademik.GET("/programs/active", h.ListActivePrograms)

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
		}

		// Activity Session
		sessions := akademik.Group("/sessions")
		sessions.Use(middleware.RequirePermission("manage_akademik"))
		{
			sessions.GET("", h.ListSessions)
			sessions.GET("/:id", h.GetSession)
			sessions.POST("", h.CreateSession)
			sessions.POST("/:id/cancel", h.CancelSession)
			sessions.POST("/:id/complete", h.CompleteSession)
			sessions.GET("/:id/attendance", h.ListAttendance)
			sessions.POST("/:id/attendance", h.RecordAttendance)
			sessions.PUT("/:id/attendance/:santriId", h.UpdateAttendance)
		}
	}
}
