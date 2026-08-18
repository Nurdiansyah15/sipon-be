package entity

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewOutboxEntry_Defaults(t *testing.T) {
	e := NewOutboxEntry("a.b", json.RawMessage(`{"x":1}`), "")

	if e.ID == uuid.Nil {
		t.Fatal("id wajib diisi")
	}
	if e.RoutingKey != "a.b" {
		t.Fatalf("routing_key: got %q", e.RoutingKey)
	}
	if e.Version != 1 {
		t.Fatalf("version: got %d", e.Version)
	}
	if e.Status != StatusPending {
		t.Fatalf("status: got %s", e.Status)
	}
	if e.AttemptCount != 0 {
		t.Fatalf("attempt_count harus 0, got %d", e.AttemptCount)
	}
	if e.CorrelationID == "" {
		t.Fatal("correlation_id harus digenerate bila kosong")
	}
	if e.NextAttemptAt.IsZero() {
		t.Fatal("next_attempt_at harus diisi")
	}
	if string(e.Payload) != `{"x":1}` {
		t.Fatalf("payload berubah: %s", e.Payload)
	}
}

func TestNewOutboxEntry_EmptyPayloadDefaultsToObject(t *testing.T) {
	e := NewOutboxEntry("a.b", nil, "corr")
	if string(e.Payload) != "{}" {
		t.Fatalf("payload kosong harus jadi {}, got %s", e.Payload)
	}
	if e.CorrelationID != "corr" {
		t.Fatalf("correlation_id harus dipertahankan, got %q", e.CorrelationID)
	}
}

func TestOutboxEntry_PublishTransition(t *testing.T) {
	now := time.Now().UTC()
	e := NewOutboxEntry("a.b", json.RawMessage(`{}`), "corr")

	e.MarkPublishing(now)
	if e.Status != StatusPublishing {
		t.Fatalf("status: got %s", e.Status)
	}
	if e.LockedAt == nil || !e.LockedAt.Equal(now) {
		t.Fatal("locked_at harus di-set saat publishing")
	}

	e.MarkPublished(now.Add(time.Minute))
	if e.Status != StatusPublished {
		t.Fatalf("status: got %s", e.Status)
	}
	if e.PublishedAt == nil {
		t.Fatal("published_at harus di-set")
	}
	if e.LockedAt != nil {
		t.Fatal("locked_at harus di-reset setelah published")
	}
}

func TestOutboxEntry_FailTransition(t *testing.T) {
	now := time.Now().UTC()
	e := NewOutboxEntry("a.b", json.RawMessage(`{}`), "corr")
	e.MarkPublishing(now)

	next := now.Add(30 * time.Second)
	e.MarkFailed("boom", next, now)

	if e.Status != StatusFailed {
		t.Fatalf("status: got %s", e.Status)
	}
	if e.AttemptCount != 1 {
		t.Fatalf("attempt_count harus 1, got %d", e.AttemptCount)
	}
	if e.LastError == nil || *e.LastError != "boom" {
		t.Fatalf("last_error tidak tersimpan")
	}
	if !e.NextAttemptAt.Equal(next) {
		t.Fatal("next_attempt_at harus maju")
	}
	if e.LockedAt != nil {
		t.Fatal("locked_at harus di-reset setelah failed")
	}
}
