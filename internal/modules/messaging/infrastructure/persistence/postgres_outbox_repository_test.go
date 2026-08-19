package persistence

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"sipon-be/internal/modules/messaging/domain/event_outbox/entity"
)

func outboxRow(id uuid.UUID, routingKey string, status entity.Status, attempt int) []driver.Value {
	return []driver.Value{
		id.String(), routingKey, []byte(`{}`), 1, "corr", nil, string(status),
		int64(attempt), time.Now().UTC(), nil, nil, nil, time.Now().UTC(), time.Now().UTC(),
	}
}

func TestPostgresOutboxRepository_Save(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgresOutboxRepository(db)
	entry := entity.NewOutboxEntry("a.b", json.RawMessage(`{}`), "corr")

	mock.ExpectExec(`(?s)INSERT INTO event_outbox`).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.Save(context.Background(), entry); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPostgresOutboxRepository_ClaimDue(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgresOutboxRepository(db)
	now := time.Now().UTC()
	id := uuid.New()

	rows := sqlmock.NewRows([]string{
		"id", "routing_key", "payload", "version", "correlation_id", "causation_id", "status",
		"attempt_count", "next_attempt_at", "locked_at", "published_at", "last_error",
		"created_at", "updated_at",
	}).AddRow(outboxRow(id, "a.b", entity.StatusPending, 0)...)

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)SELECT .* FROM event_outbox`).
		WithArgs(now, 10).
		WillReturnRows(rows)
	mock.ExpectExec(`(?s)UPDATE event_outbox SET status = 'PUBLISHING'`).
		WithArgs(now, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	entries, err := repo.ClaimDue(context.Background(), now, 10)
	if err != nil {
		t.Fatalf("ClaimDue: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("harus 1 entry, got %d", len(entries))
	}
	if entries[0].Status != entity.StatusPublishing {
		t.Fatalf("entry harus PUBLISHING, got %s", entries[0].Status)
	}
	if entries[0].LockedAt == nil {
		t.Fatal("entry harus ter-lease")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPostgresOutboxRepository_ClaimDue_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgresOutboxRepository(db)
	now := time.Now().UTC()

	rows := sqlmock.NewRows([]string{
		"id", "routing_key", "payload", "version", "correlation_id", "causation_id", "status",
		"attempt_count", "next_attempt_at", "locked_at", "published_at", "last_error",
		"created_at", "updated_at",
	})

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)SELECT .* FROM event_outbox`).
		WithArgs(now, 10).
		WillReturnRows(rows)
	mock.ExpectCommit()

	entries, err := repo.ClaimDue(context.Background(), now, 10)
	if err != nil {
		t.Fatalf("ClaimDue: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("tidak ada row, got %d", len(entries))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPostgresOutboxRepository_MarkPublished(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgresOutboxRepository(db)
	id := uuid.New()

	mock.ExpectExec(`(?s)UPDATE event_outbox SET status = 'PUBLISHED'`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.MarkPublished(context.Background(), id, time.Now().UTC()); err != nil {
		t.Fatalf("MarkPublished: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPostgresOutboxRepository_MarkFailed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgresOutboxRepository(db)
	id := uuid.New()

	mock.ExpectExec(`(?s)UPDATE event_outbox SET status = 'FAILED'`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.MarkFailed(context.Background(), id, "boom", time.Now().Add(30*time.Second)); err != nil {
		t.Fatalf("MarkFailed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPostgresOutboxRepository_RecoverStuckPublishing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgresOutboxRepository(db)

	mock.ExpectExec(`(?s)UPDATE event_outbox SET status = 'PENDING'`).
		WillReturnResult(sqlmock.NewResult(0, 2))
	n, err := repo.RecoverStuckPublishing(context.Background(), time.Now().Add(-time.Minute))
	if err != nil {
		t.Fatalf("RecoverStuckPublishing: %v", err)
	}
	if n != 2 {
		t.Fatalf("harus recover 2 row, got %d", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPostgresOutboxRepository_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgresOutboxRepository(db)
	id := uuid.New()

	rows := sqlmock.NewRows([]string{
		"id", "routing_key", "payload", "version", "correlation_id", "causation_id", "status",
		"attempt_count", "next_attempt_at", "locked_at", "published_at", "last_error",
		"created_at", "updated_at",
	}).AddRow(outboxRow(id, "a.b", entity.StatusPublished, 0)...)

	mock.ExpectQuery(`(?s)SELECT .* FROM event_outbox WHERE id =`).
		WithArgs(id).
		WillReturnRows(rows)

	e, err := repo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if e.ID != id {
		t.Fatalf("id: got %s", e.ID)
	}
	if e.Status != entity.StatusPublished {
		t.Fatalf("status: got %s", e.Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
