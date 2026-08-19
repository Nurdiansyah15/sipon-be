package command

import (
	"context"
	"encoding/json"
	"time"

	"sipon-be/internal/modules/akademik/application/ports"
)

// ScheduledJobTypes memuat routing key yang dipakai command saat membuat job.
// Routing key disuntikkan dari composition root (interfaces/mq), sehingga command
// layer tidak bergantung langsung pada adapter transport.
type ScheduledJobTypes struct {
	FingerprintSync  string
	SessionAutoClose string
}

type ScheduleSessionJobsUseCase struct {
	scheduler ports.Scheduler
	jobTypes  ScheduledJobTypes
}

func NewScheduleSessionJobsUseCase(
	scheduler ports.Scheduler,
	jobTypes ScheduledJobTypes,
) *ScheduleSessionJobsUseCase {
	return &ScheduleSessionJobsUseCase{
		scheduler: scheduler,
		jobTypes:  jobTypes,
	}
}

func (uc *ScheduleSessionJobsUseCase) Execute(ctx context.Context, sessionID string, endsAt time.Time) error {
	syncPayload, _ := json.Marshal(map[string]string{"session_id": sessionID})
	if err := uc.scheduler.ScheduleRecurring(ctx, ports.ScheduleRecurringInput{
		JobType:     uc.jobTypes.FingerprintSync,
		Payload:     syncPayload,
		CronExpr:    "*/1 * * * *",
		ReferenceID: sessionID,
	}); err != nil {
		return err
	}

	closePayload, _ := json.Marshal(map[string]string{"session_id": sessionID})
	return uc.scheduler.ScheduleOneOff(ctx, ports.ScheduleOneOffInput{
		JobType:     uc.jobTypes.SessionAutoClose,
		Payload:     closePayload,
		RunAt:       endsAt,
		ReferenceID: sessionID,
	})
}
