package command

import (
	"context"
	"encoding/json"
	"time"

	"sipon-be/internal/modules/akademik/application/ports"
)

// ScheduleReminderUseCase menjadwalkan job one-off reminder untuk satu sesi.
// Job fires pada waktu (startsAt - reminderMinutes) sebelum sesi dimulai, lalu
// dikonsumsi handler akademik untuk mengirim notifikasi ke user yang sudah
// herregistrasi pada periode tersebut.
type ScheduleReminderUseCase struct {
	scheduler ports.Scheduler
	jobType   string
}

func NewScheduleReminderUseCase(
	scheduler ports.Scheduler,
	jobType string,
) *ScheduleReminderUseCase {
	return &ScheduleReminderUseCase{
		scheduler: scheduler,
		jobType:   jobType,
	}
}

// Execute menjadwalkan reminder. Jika reminderMinutes <= 0, tidak ada yang
// dijadwalkan (reminder nonaktif).
func (uc *ScheduleReminderUseCase) Execute(ctx context.Context, sessionID string, startsAt time.Time, reminderMinutes int) error {
	if reminderMinutes <= 0 {
		return nil
	}
	runAt := startsAt.Add(-time.Duration(reminderMinutes) * time.Minute)
	payload, _ := json.Marshal(map[string]string{"session_id": sessionID})
	return uc.scheduler.ScheduleOneOff(ctx, ports.ScheduleOneOffInput{
		JobType:     uc.jobType,
		Payload:     payload,
		RunAt:       runAt,
		ReferenceID: sessionID,
	})
}
