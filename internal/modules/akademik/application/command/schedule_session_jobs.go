package command

import (
	"context"
	"encoding/json"
	"time"

	"github.com/robfig/cron/v3"

	sjEntity "sipon-be/internal/shared/scheduler/domain/scheduled_job/entity"
	sjRepo "sipon-be/internal/shared/scheduler/domain/scheduled_job/repository"
)

type ScheduleSessionJobsUseCase struct {
	scheduledJobRepo sjRepo.Repository
	parser           cron.Parser
	loc              *time.Location
}

func NewScheduleSessionJobsUseCase(
	scheduledJobRepo sjRepo.Repository,
	parser cron.Parser,
	loc *time.Location,
) *ScheduleSessionJobsUseCase {
	return &ScheduleSessionJobsUseCase{
		scheduledJobRepo: scheduledJobRepo,
		parser:           parser,
		loc:              loc,
	}
}

func (uc *ScheduleSessionJobsUseCase) Execute(ctx context.Context, sessionID string, endsAt time.Time) error {
	syncPayload, _ := json.Marshal(map[string]string{"session_id": sessionID})
	syncJob, err := sjEntity.NewRecurringJob(
		JobTypeFingerprintSync,
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
	closeJob := sjEntity.NewOneOffJob(JobTypeSessionAutoClose, closePayload, endsAt)
	closeJob.ReferenceID = &sessionID
	return uc.scheduledJobRepo.Save(ctx, closeJob)
}
