package constant

type ScheduleType string

const (
	ScheduleTypeOneOff    ScheduleType = "ONE_OFF"
	ScheduleTypeRecurring ScheduleType = "RECURRING"
)

type Status string

const (
	StatusActive     Status = "ACTIVE"
	StatusProcessing Status = "PROCESSING"
	StatusPaused     Status = "PAUSED"
	StatusCompleted  Status = "COMPLETED"
	StatusFailed     Status = "FAILED"
)
