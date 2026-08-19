package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Status durable inbox message job.
type Status string

const (
	StatusPending   Status = "PENDING"
	StatusRunning   Status = "RUNNING"
	StatusRetryWait Status = "RETRY_WAIT"
	StatusSucceeded Status = "SUCCEEDED"
	StatusFailed    Status = "FAILED"
)

// MessageJob adalah durable inbox record. Dibuat/commit sebelum handler diproses
// agar redelivery dapat di-deduplicate dan lifecycle eksekusi dapat diaudit.
type MessageJob struct {
	ID            uuid.UUID
	RoutingKey    string
	Payload       json.RawMessage
	Version       int
	CorrelationID string
	Status        Status
	AttemptCount  int
	MaxAttempts   int
	NextAttemptAt time.Time
	RunningAt     *time.Time
	SucceededAt   *time.Time
	FailedAt      *time.Time
	LockedUntil   *time.Time
	LastError     *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewMessageJob(
	id uuid.UUID,
	routingKey string,
	payload json.RawMessage,
	version int,
	correlationID string,
	maxAttempts int,
) *MessageJob {
	now := time.Now().UTC()
	if len(payload) == 0 {
		payload = json.RawMessage("{}")
	}
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	return &MessageJob{
		ID:            id,
		RoutingKey:    routingKey,
		Payload:       payload,
		Version:       version,
		CorrelationID: correlationID,
		Status:        StatusPending,
		MaxAttempts:   maxAttempts,
		NextAttemptAt: now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// StartRun menandai job mulai diproses, menaikkan attempt, dan mengunci lease.
func (j *MessageJob) StartRun(now time.Time, leaseUntil time.Time) {
	j.Status = StatusRunning
	j.AttemptCount++
	j.RunningAt = &now
	j.LockedUntil = &leaseUntil
	j.UpdatedAt = now
}

// Succeed menandai job selesai sukses.
func (j *MessageJob) Succeed(now time.Time) {
	j.Status = StatusSucceeded
	j.SucceededAt = &now
	j.RunningAt = nil
	j.LockedUntil = nil
	j.UpdatedAt = now
}

// ScheduleRetry menandai job menunggu retry pada nextAttemptAt.
func (j *MessageJob) ScheduleRetry(errMsg string, nextAttemptAt time.Time, now time.Time) {
	j.Status = StatusRetryWait
	j.LastError = &errMsg
	j.NextAttemptAt = nextAttemptAt
	j.RunningAt = nil
	j.LockedUntil = nil
	j.UpdatedAt = now
}

// Fail menandai job gagal terminal (retry habis / fatal).
func (j *MessageJob) Fail(errMsg string, now time.Time) {
	j.Status = StatusFailed
	j.LastError = &errMsg
	j.FailedAt = &now
	j.RunningAt = nil
	j.LockedUntil = nil
	j.UpdatedAt = now
}

func (j *MessageJob) IsTerminal() bool {
	return j.Status == StatusSucceeded || j.Status == StatusFailed
}
