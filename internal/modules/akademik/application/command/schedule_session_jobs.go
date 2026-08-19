package command

import (
	"context"
	"encoding/json"
	"time"

	"sipon-be/internal/modules/scheduler"
)

// ScheduledJobTypes memuat routing key yang dipakai command saat membuat job.
// Routing key disuntikkan dari composition root (interfaces/mq), sehingga command
// layer tidak bergantung langsung pada adapter transport.
type ScheduledJobTypes struct {
	FingerprintSync  string
	SessionAutoClose string
}

type ScheduleSessionJobsUseCase struct {
	schedulerContract scheduler.Contract
	jobTypes          ScheduledJobTypes
}

func NewScheduleSessionJobsUseCase(
	schedulerContract scheduler.Contract,
	jobTypes ScheduledJobTypes,
) *ScheduleSessionJobsUseCase {
	return &ScheduleSessionJobsUseCase{
		schedulerContract: schedulerContract,
		jobTypes:          jobTypes,
	}
}

func (uc *ScheduleSessionJobsUseCase) Execute(ctx context.Context, sessionID string, endsAt time.Time) error {
	syncPayload, _ := json.Marshal(map[string]string{"session_id": sessionID})
	if err := uc.schedulerContract.ScheduleRecurring(ctx, scheduler.ScheduleRecurringInput{
		JobType:     uc.jobTypes.FingerprintSync,
		Payload:     syncPayload,
		CronExpr:    "*/1 * * * *",
		ReferenceID: sessionID,
	}); err != nil {
		return err
	}

	closePayload, _ := json.Marshal(map[string]string{"session_id": sessionID})
	return uc.schedulerContract.ScheduleOneOff(ctx, scheduler.ScheduleOneOffInput{
		JobType:     uc.jobTypes.SessionAutoClose,
		Payload:     closePayload,
		RunAt:       endsAt,
		ReferenceID: sessionID,
	})
}
