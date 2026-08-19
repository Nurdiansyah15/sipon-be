package application

import (
	"context"
	"encoding/json"
	"testing"

	"sipon-be/internal/modules/messaging/domain/message/valueobject"
	messagingerrors "sipon-be/internal/modules/messaging/domain/message_job/errors"
)

func TestRegistry_RegisterAndDispatch(t *testing.T) {
	reg := NewRegistry()
	called := false
	handler := func(ctx context.Context, msg valueobject.Message) error {
		called = true
		if msg.Type != "a.b" {
			t.Fatalf("type: got %q", msg.Type)
		}
		return nil
	}

	if err := reg.Register("a.b", handler); err != nil {
		t.Fatalf("Register: %v", err)
	}
	msg, _ := valueobject.NewMessage("a.b", json.RawMessage(`{}`))
	if err := reg.Dispatch(context.Background(), msg); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if !called {
		t.Fatal("handler tidak dipanggil")
	}
	if !reg.Has("a.b") {
		t.Fatal("Has harus true setelah register")
	}
}

func TestRegistry_RejectDuplicate(t *testing.T) {
	reg := NewRegistry()
	h := func(ctx context.Context, msg valueobject.Message) error { return nil }
	if err := reg.Register("a.b", h); err != nil {
		t.Fatalf("Register pertama: %v", err)
	}
	if err := reg.Register("a.b", h); err == nil {
		t.Fatal("duplicate routing key harus ditolak")
	}
}

func TestRegistry_RejectEmptyKey(t *testing.T) {
	reg := NewRegistry()
	h := func(ctx context.Context, msg valueobject.Message) error { return nil }
	if err := reg.Register("", h); err == nil {
		t.Fatal("routing key kosong harus ditolak")
	}
}

func TestRegistry_RejectNilHandler(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register("a.b", nil); err == nil {
		t.Fatal("handler nil harus ditolak")
	}
}

func TestRegistry_UnknownKeyIsFatal(t *testing.T) {
	reg := NewRegistry()
	msg, _ := valueobject.NewMessage("unknown.key", json.RawMessage(`{}`))
	err := reg.Dispatch(context.Background(), msg)
	if err == nil {
		t.Fatal("unknown routing key harus error")
	}
	if !messagingerrors.IsFatal(err) {
		t.Fatalf("unknown routing key harus fatal, got %T", err)
	}
}

func TestRegistry_DispatchPropagatesHandlerError(t *testing.T) {
	reg := NewRegistry()
	handler := func(ctx context.Context, msg valueobject.Message) error {
		return messagingerrors.NewRetryableError(nil)
	}
	if err := reg.Register("a.b", handler); err != nil {
		t.Fatalf("Register: %v", err)
	}
	msg, _ := valueobject.NewMessage("a.b", json.RawMessage(`{}`))
	err := reg.Dispatch(context.Background(), msg)
	if err == nil {
		t.Fatal("harus error dari handler")
	}
	if !messagingerrors.IsRetryable(err) {
		t.Fatalf("harus retryable, got %T", err)
	}
}
