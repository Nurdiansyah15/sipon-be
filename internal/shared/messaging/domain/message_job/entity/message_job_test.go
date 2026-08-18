package entity

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewMessageJob_Defaults(t *testing.T) {
	id := uuid.New()
	j := NewMessageJob(id, "a.b", json.RawMessage(`{}`), 1, "corr", 0)

	if j.ID != id {
		t.Fatal("id harus dipertahankan")
	}
	if j.Status != StatusPending {
		t.Fatalf("status: got %s", j.Status)
	}
	if j.MaxAttempts != 5 {
		t.Fatalf("max_attempts default harus 5, got %d", j.MaxAttempts)
	}
	if j.AttemptCount != 0 {
		t.Fatalf("attempt_count harus 0, got %d", j.AttemptCount)
	}
	if j.IsTerminal() {
		t.Fatal("PENDING tidak boleh terminal")
	}
}

func TestMessageJob_SucceedTransition(t *testing.T) {
	now := time.Now().UTC()
	lease := now.Add(time.Minute)

	j := NewMessageJob(uuid.New(), "a.b", json.RawMessage(`{}`), 1, "corr", 5)
	j.StartRun(now, lease)

	if j.Status != StatusRunning {
		t.Fatalf("status: got %s", j.Status)
	}
	if j.AttemptCount != 1 {
		t.Fatalf("attempt_count harus 1, got %d", j.AttemptCount)
	}
	if j.RunningAt == nil || !j.RunningAt.Equal(now) {
		t.Fatal("running_at harus di-set")
	}
	if j.LockedUntil == nil || !j.LockedUntil.Equal(lease) {
		t.Fatal("locked_until harus di-set")
	}

	j.Succeed(now.Add(time.Second))
	if j.Status != StatusSucceeded {
		t.Fatalf("status: got %s", j.Status)
	}
	if j.SucceededAt == nil {
		t.Fatal("succeeded_at harus di-set")
	}
	if j.LockedUntil != nil || j.RunningAt != nil {
		t.Fatal("lock/running harus di-reset setelah sukses")
	}
	if !j.IsTerminal() {
		t.Fatal("SUCCEEDED harus terminal")
	}
}

func TestMessageJob_RetryThenFail(t *testing.T) {
	now := time.Now().UTC()
	lease := now.Add(time.Minute)

	j := NewMessageJob(uuid.New(), "a.b", json.RawMessage(`{}`), 1, "corr", 5)
	j.StartRun(now, lease)
	j.ScheduleRetry("transient", now.Add(time.Minute), now)

	if j.Status != StatusRetryWait {
		t.Fatalf("status: got %s", j.Status)
	}
	if j.LastError == nil || *j.LastError != "transient" {
		t.Fatal("last_error harus tersimpan")
	}
	if j.IsTerminal() {
		t.Fatal("RETRY_WAIT tidak boleh terminal")
	}

	j.StartRun(now.Add(time.Minute), lease)
	if j.AttemptCount != 2 {
		t.Fatalf("attempt_count harus 2, got %d", j.AttemptCount)
	}
	if j.Status != StatusRunning {
		t.Fatalf("status: got %s", j.Status)
	}

	j.Fail("permanent", now.Add(2*time.Minute))
	if j.Status != StatusFailed {
		t.Fatalf("status: got %s", j.Status)
	}
	if j.FailedAt == nil {
		t.Fatal("failed_at harus di-set")
	}
	if j.LastError == nil || *j.LastError != "permanent" {
		t.Fatal("last_error harus diperbarui")
	}
	if !j.IsTerminal() {
		t.Fatal("FAILED harus terminal")
	}
}
