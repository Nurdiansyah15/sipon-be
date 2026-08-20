package command

import (
	"context"
	"encoding/json"
	"time"

	"sipon-be/internal/modules/akademik/application/ports"
)

// ScheduleAutoOpenUseCase menjadwalkan job one-off auto-open untuk satu sesi.
// Job akan fires pada waktu StartsAt sesi. Jika sesi sudah dibuka manual
// sebelum job fires, handler auto-open akan skip (idempotent).
type ScheduleAutoOpenUseCase struct {
	scheduler ports.Scheduler
	jobType   string
}

func NewScheduleAutoOpenUseCase(
	scheduler ports.Scheduler,
	jobType string,
) *ScheduleAutoOpenUseCase {
	return &ScheduleAutoOpenUseCase{
		scheduler: scheduler,
		jobType:   jobType,
	}
}

func (uc *ScheduleAutoOpenUseCase) Execute(ctx context.Context, sessionID string, startsAt time.Time) error {
	payload, _ := json.Marshal(map[string]string{"session_id": sessionID})
	return uc.scheduler.ScheduleOneOff(ctx, ports.ScheduleOneOffInput{
		JobType:     uc.jobType,
		Payload:     payload,
		RunAt:       startsAt,
		ReferenceID: sessionID,
	})
}
