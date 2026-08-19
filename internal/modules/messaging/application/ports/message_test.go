package ports

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func validMsg() Message {
	return Message{
		ID:         uuid.New(),
		Type:       "module.resource.action",
		Version:    1,
		OccurredAt: time.Now().UTC(),
		Payload:    json.RawMessage(`{"session_id":"s1"}`),
	}
}

func TestNewMessage_Defaults(t *testing.T) {
	msg, err := NewMessage("akademik.session.auto_close", json.RawMessage(`{"session_id":"s1"}`))
	if err != nil {
		t.Fatalf("NewMessage: %v", err)
	}
	if msg.ID == uuid.Nil {
		t.Fatal("id harus diisi")
	}
	if msg.Type != "akademik.session.auto_close" {
		t.Fatalf("type: got %q", msg.Type)
	}
	if msg.Version != 1 {
		t.Fatalf("version: got %d", msg.Version)
	}
	if msg.OccurredAt.IsZero() {
		t.Fatal("occurred_at harus diisi")
	}
	if msg.CorrelationID == "" {
		t.Fatal("correlation_id harus diisi")
	}
	if msg.CausationID != nil {
		t.Fatal("causation_id harus kosong pada message baru")
	}
}

func TestMessage_Validate(t *testing.T) {
	cases := []struct {
		name    string
		edit    func(*Message)
		wantErr bool
	}{
		{"valid", func(m *Message) {}, false},
		{"nil id", func(m *Message) { m.ID = uuid.Nil }, true},
		{"empty type", func(m *Message) { m.Type = "" }, true},
		{"zero version", func(m *Message) { m.Version = 0 }, true},
		{"zero occurred_at", func(m *Message) { m.OccurredAt = time.Time{} }, true},
		{"nil payload", func(m *Message) { m.Payload = nil }, true},
		{"invalid json payload", func(m *Message) { m.Payload = json.RawMessage(`{`) }, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := validMsg()
			tc.edit(&m)
			err := m.Validate()
			if tc.wantErr && err == nil {
				t.Fatal("harus error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("tidak boleh error: %v", err)
			}
		})
	}
}
