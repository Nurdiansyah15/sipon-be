package command

import (
	"context"
	"encoding/json"
	"time"

	"github.com/robfig/cron/v3"

	sjEntity "sipon-be/internal/shared/scheduler/domain/scheduled_job/entity"
	sjRepo "sipon-be/internal/shared/scheduler/domain/scheduled_job/repository"
)

// ScheduledJobTypes memuat routing key yang dipakai command saat membuat job.
// Routing key disuntikkan dari composition root (interfaces/mq), sehingga command
// layer tidak bergantung langsung pada adapter transport.
type ScheduledJobTypes struct {
	FingerprintSync  string
	SessionAutoClose string
}

type ScheduleSessionJobsUseCase struct {
	scheduledJobRepo sjRepo.Repository
	parser           cron.Parser
	loc              *time.Location
	jobTypes         ScheduledJobTypes
}

func NewScheduleSessionJobsUseCase(
	scheduledJobRepo sjRepo.Repository,
	parser cron.Parser,
	loc *time.Location,
	jobTypes ScheduledJobTypes,
) *ScheduleSessionJobsUseCase {
	return &ScheduleSessionJobsUseCase{
		scheduledJobRepo: scheduledJobRepo,
		parser:           parser,
		loc:              loc,
		jobTypes:         jobTypes,
	}
}

func (uc *ScheduleSessionJobsUseCase) Execute(ctx context.Context, sessionID string, endsAt time.Time) error {
	syncPayload, _ := json.Marshal(map[string]string{"session_id": sessionID})
	syncJob, err := sjEntity.NewRecurringJob(
		uc.jobTypes.FingerprintSync,
		syncPayload,
		"*/1 * * * *",
		uc.parser,
		uc.loc,
	)
	if err != nil {
		return err
	}
	syncJob.ReferenceID = &sessionID
	if err := uc.scheduledJobRepo.Save(ctx, syncJob); err != nil {
		return err
	}

	closePayload, _ := json.Marshal(map[string]string{"session_id": sessionID})
	closeJob := sjEntity.NewOneOffJob(uc.jobTypes.SessionAutoClose, closePayload, endsAt)
	closeJob.ReferenceID = &sessionID
	return uc.scheduledJobRepo.Save(ctx, closeJob)
}
