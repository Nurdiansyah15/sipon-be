package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"sipon-be/internal/modules/messaging/domain/event_outbox/entity"
)

func TestPostgresTransactor_Commit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	tr := NewPostgresTransactor(db)

	mock.ExpectBegin()
	mock.ExpectCommit()

	if err := tr.WithTx(context.Background(), func(ctx context.Context) error {
		return nil
	}); err != nil {
		t.Fatalf("WithTx: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPostgresTransactor_RollbackOnError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	tr := NewPostgresTransactor(db)
	sentinel := errors.New("boom")

	mock.ExpectBegin()
	mock.ExpectRollback()

	err = tr.WithTx(context.Background(), func(ctx context.Context) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("harus mengembalikan error fn, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

// TestTransactor_OutboxWithinBusinessTx membuktikan business write (diwakili
// transactor) dan INSERT event_outbox berjalan dalam satu DB transaction yang
// sama: INSERT harus terjadi di antara BEGIN dan COMMIT.
func TestTransactor_OutboxWithinBusinessTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	tr := NewPostgresTransactor(db)
	repo := NewPostgresOutboxRepository(db)
	entry := entity.NewOutboxEntry("a.b", json.RawMessage(`{}`), "corr")

	mock.ExpectBegin()
	mock.ExpectExec(`(?s)INSERT INTO event_outbox`).
		WithArgs(entry.ID, entry.RoutingKey, sqlmock.AnyArg(), entry.Version,
			entry.CorrelationID, nil, string(entity.StatusPending), entry.AttemptCount,
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = tr.WithTx(context.Background(), func(txCtx context.Context) error {
		return repo.Save(txCtx, entry)
	})
	if err != nil {
		t.Fatalf("WithTx: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

// TestTransactor_RollbackDropsOutbox membuktikan ketika fn error, INSERT outbox
// ikut di-rollback (tidak ada commit).
func TestTransactor_RollbackDropsOutbox(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	tr := NewPostgresTransactor(db)
	repo := NewPostgresOutboxRepository(db)
	entry := entity.NewOutboxEntry("a.b", json.RawMessage(`{}`), "corr")
	sentinel := errors.New("business failed")

	mock.ExpectBegin()
	mock.ExpectExec(`(?s)INSERT INTO event_outbox`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectRollback()

	err = tr.WithTx(context.Background(), func(txCtx context.Context) error {
		if err := repo.Save(txCtx, entry); err != nil {
			return err
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("harus mengembalikan error fn, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
