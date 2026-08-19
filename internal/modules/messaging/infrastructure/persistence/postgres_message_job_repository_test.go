package persistence

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"sipon-be/internal/modules/messaging/domain/message_job/entity"
)

func messageJobRow(id uuid.UUID, routingKey string, status entity.Status, attempt int) []driver.Value {
	now := time.Now().UTC()
	return []driver.Value{
		id.String(), routingKey, []byte(`{}`), 1, "corr", string(status),
		int64(attempt), 5, now, nil, nil, nil, nil, nil, now, now,
	}
}

func TestPostgresMessageJobRepository_Save(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgresMessageJobRepository(db)
	job := entity.NewMessageJob(uuid.New(), "a.b", nil, 1, "corr", 5)

	mock.ExpectExec(`(?s)INSERT INTO message_jobs`).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.Save(context.Background(), job); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPostgresMessageJobRepository_ClaimPending(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgresMessageJobRepository(db)
	now := time.Now().UTC()
	leaseUntil := now.Add(time.Minute)
	id := uuid.New()

	rows := sqlmock.NewRows([]string{
		"id", "routing_key", "payload", "version", "correlation_id", "status",
		"attempt_count", "max_attempts", "next_attempt_at", "running_at", "succeeded_at",
		"failed_at", "locked_until", "last_error", "created_at", "updated_at",
	}).AddRow(messageJobRow(id, "a.b", entity.StatusPending, 0)...)

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)SELECT .* FROM message_jobs`).
		WithArgs(now, 10).
		WillReturnRows(rows)
	mock.ExpectExec(`(?s)UPDATE message_jobs SET status = 'RUNNING'`).
		WithArgs(now, leaseUntil, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	jobs, err := repo.ClaimPending(context.Background(), now, 10, leaseUntil)
	if err != nil {
		t.Fatalf("ClaimPending: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("harus 1 job, got %d", len(jobs))
	}
	if jobs[0].Status != entity.StatusRunning {
		t.Fatalf("job harus RUNNING, got %s", jobs[0].Status)
	}
	if jobs[0].AttemptCount != 1 {
		t.Fatalf("attempt_count harus 1, got %d", jobs[0].AttemptCount)
	}
	if jobs[0].LockedUntil == nil {
		t.Fatal("job harus ter-lease")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPostgresMessageJobRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgresMessageJobRepository(db)
	job := entity.NewMessageJob(uuid.New(), "a.b", nil, 1, "corr", 5)
	now := time.Now().UTC()
	job.Succeed(now)

	mock.ExpectExec(`(?s)UPDATE message_jobs SET`).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.Update(context.Background(), job); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPostgresMessageJobRepository_RecoverStuckRunning(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgresMessageJobRepository(db)

	mock.ExpectExec(`(?s)UPDATE message_jobs SET status = 'PENDING'`).
		WillReturnResult(sqlmock.NewResult(0, 3))
	n, err := repo.RecoverStuckRunning(context.Background(), time.Now().Add(-time.Minute))
	if err != nil {
		t.Fatalf("RecoverStuckRunning: %v", err)
	}
	if n != 3 {
		t.Fatalf("harus recover 3 row, got %d", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPostgresMessageJobRepository_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgresMessageJobRepository(db)
	id := uuid.New()

	rows := sqlmock.NewRows([]string{
		"id", "routing_key", "payload", "version", "correlation_id", "status",
		"attempt_count", "max_attempts", "next_attempt_at", "running_at", "succeeded_at",
		"failed_at", "locked_until", "last_error", "created_at", "updated_at",
	}).AddRow(messageJobRow(id, "a.b", entity.StatusSucceeded, 1)...)

	mock.ExpectQuery(`(?s)SELECT .* FROM message_jobs WHERE id =`).
		WithArgs(id).
		WillReturnRows(rows)

	job, err := repo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if job.ID != id {
		t.Fatalf("id: got %s", job.ID)
	}
	if job.Status != entity.StatusSucceeded {
		t.Fatalf("status: got %s", job.Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
