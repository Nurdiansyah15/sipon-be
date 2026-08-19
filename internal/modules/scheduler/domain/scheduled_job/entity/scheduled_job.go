package entity

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"

	"sipon-be/internal/modules/scheduler/domain/scheduled_job/constant"
)

type ScheduledJob struct {
	ID           uuid.UUID
	Type         string
	Payload      json.RawMessage
	ScheduleType constant.ScheduleType
	CronExpr     *string
	RunAt        *time.Time
	NextRunAt    time.Time
	LastRunAt    *time.Time
	Status       constant.Status
	RetryCount   int
	MaxRetry     int
	LastError    *string
	ReferenceID  *string
	LeaseUntil   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewOneOffJob(jobType string, payload json.RawMessage, runAt time.Time) *ScheduledJob {
	if payload == nil {
		payload = json.RawMessage("{}")
	}
	now := time.Now()
	return &ScheduledJob{
		ID:           uuid.New(),
		Type:         jobType,
		Payload:      payload,
		ScheduleType: constant.ScheduleTypeOneOff,
		RunAt:        &runAt,
		NextRunAt:    runAt,
		Status:       constant.StatusActive,
		MaxRetry:     3,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func NewRecurringJob(jobType string, payload json.RawMessage, cronExpr string, parser cron.Parser, loc *time.Location) (*ScheduledJob, error) {
	if payload == nil {
		payload = json.RawMessage("{}")
	}
	sched, err := parser.Parse(cronExpr)
	if err != nil {
		return nil, fmt.Errorf("cron expression tidak valid %q: %w", cronExpr, err)
	}
	now := time.Now()
	expr := cronExpr
	return &ScheduledJob{
		ID:           uuid.New(),
		Type:         jobType,
		Payload:      payload,
		ScheduleType: constant.ScheduleTypeRecurring,
		CronExpr:     &expr,
		NextRunAt:    sched.Next(now.In(loc)),
		Status:       constant.StatusActive,
		MaxRetry:     3,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (j *ScheduledJob) MarkFired(nextRunAt time.Time) {
	now := time.Now()
	j.LastRunAt = &now
	j.NextRunAt = nextRunAt
	j.Status = constant.StatusActive
	j.UpdatedAt = now
}

func (j *ScheduledJob) MarkCompleted() {
	now := time.Now()
	j.LastRunAt = &now
	j.Status = constant.StatusCompleted
	j.UpdatedAt = now
}

func (j *ScheduledJob) MarkFailed(errMsg string) {
	now := time.Now()
	j.LastRunAt = &now
	j.RetryCount++
	j.LastError = &errMsg
	if j.RetryCount >= j.MaxRetry {
		j.Status = constant.StatusFailed
	} else {
		j.Status = constant.StatusActive
	}
	j.UpdatedAt = now
}

func (j *ScheduledJob) Pause() {
	j.Status = constant.StatusPaused
	j.UpdatedAt = time.Now()
}

func (j *ScheduledJob) Resume() {
	j.Status = constant.StatusActive
	j.UpdatedAt = time.Now()
}

func (j *ScheduledJob) UpdateSchedule(cronExpr string, parser cron.Parser, loc *time.Location) error {
	sched, err := parser.Parse(cronExpr)
	if err != nil {
		return fmt.Errorf("cron expression tidak valid %q: %w", cronExpr, err)
	}
	expr := cronExpr
	j.CronExpr = &expr
	j.NextRunAt = sched.Next(time.Now().In(loc))
	j.UpdatedAt = time.Now()
	return nil
}
