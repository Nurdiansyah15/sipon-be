package schedulergateway

import (
	"context"

	"sipon-be/internal/modules/akademik/application/ports"
	"sipon-be/internal/modules/scheduler"
)

// Gateway mengadaptasi modul scheduler (Contract) ke port Scheduler milik
// modul akademik, sehingga lapisan application/command tidak bergantung
// langsung pada modul scheduler.
type Gateway struct {
	contract scheduler.Contract
}

func New(contract scheduler.Contract) *Gateway {
	return &Gateway{contract: contract}
}

func (g *Gateway) ScheduleRecurring(ctx context.Context, in ports.ScheduleRecurringInput) error {
	return g.contract.ScheduleRecurring(ctx, scheduler.ScheduleRecurringInput{
		JobType:     in.JobType,
		Payload:     in.Payload,
		CronExpr:    in.CronExpr,
		ReferenceID: in.ReferenceID,
	})
}

func (g *Gateway) ScheduleOneOff(ctx context.Context, in ports.ScheduleOneOffInput) error {
	return g.contract.ScheduleOneOff(ctx, scheduler.ScheduleOneOffInput{
		JobType:     in.JobType,
		Payload:     in.Payload,
		RunAt:       in.RunAt,
		ReferenceID: in.ReferenceID,
	})
}

func (g *Gateway) PauseByTypeAndReferenceID(ctx context.Context, jobType, referenceID string) error {
	return g.contract.PauseByTypeAndReferenceID(ctx, jobType, referenceID)
}
