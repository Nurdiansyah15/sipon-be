package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Status outbox event.
type Status string

const (
	StatusPending    Status = "PENDING"
	StatusPublishing Status = "PUBLISHING"
	StatusPublished  Status = "PUBLISHED"
	StatusFailed     Status = "FAILED"
)

// OutboxEntry adalah buffer publish transactional. Row dibuat dalam transaksi yang
// sama dengan perubahan bisnis, lalu dibawa ke RabbitMQ oleh Outbox Relay.
type OutboxEntry struct {
	ID            uuid.UUID
	RoutingKey    string
	Payload       json.RawMessage
	Version       int
	CorrelationID string
	CausationID   *uuid.UUID
	Status        Status
	AttemptCount  int
	NextAttemptAt time.Time
	LockedAt      *time.Time
	PublishedAt   *time.Time
	LastError     *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewOutboxEntry(routingKey string, payload json.RawMessage, correlationID string) *OutboxEntry {
	now := time.Now().UTC()
	if len(payload) == 0 {
		payload = json.RawMessage("{}")
	}
	if correlationID == "" {
		correlationID = uuid.NewString()
	}
	return &OutboxEntry{
		ID:            uuid.New(),
		RoutingKey:    routingKey,
		Payload:       payload,
		Version:       1,
		CorrelationID: correlationID,
		Status:        StatusPending,
		NextAttemptAt: now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// MarkPublishing menandai entry diklaim oleh relay untuk dipublish.
func (e *OutboxEntry) MarkPublishing(now time.Time) {
	e.Status = StatusPublishing
	lock := now
	e.LockedAt = &lock
	e.UpdatedAt = now
}

// MarkPublished menandai entry berhasil dipublish dan di-confirm broker.
func (e *OutboxEntry) MarkPublished(now time.Time) {
	e.Status = StatusPublished
	e.PublishedAt = &now
	e.LockedAt = nil
	e.UpdatedAt = now
}

// MarkFailed menandai publish gagal dan menjadwalkan attempt berikutnya.
func (e *OutboxEntry) MarkFailed(errMsg string, nextAttemptAt time.Time, now time.Time) {
	e.Status = StatusFailed
	e.LastError = &errMsg
	e.NextAttemptAt = nextAttemptAt
	e.AttemptCount++
	e.LockedAt = nil
	e.UpdatedAt = now
}
