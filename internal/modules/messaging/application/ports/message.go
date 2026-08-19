package ports

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Message adalah envelope transport untuk event asynchronous. ID bersifat tetap
// dan dipakai sebagai idempotency key — ia tidak berubah ketika message di-retry.
type Message struct {
	ID            uuid.UUID       `json:"id"`
	Type          string          `json:"type"`
	Version       int             `json:"version"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Payload       json.RawMessage `json:"payload"`
	CorrelationID string          `json:"correlation_id"`
	CausationID   *uuid.UUID      `json:"causation_id,omitempty"`
}

// NewMessage membuat envelope baru. CorrelationID di-generate baru; untuk message
// yang berasal dari scheduler/HTTP, caller dapat meng-override setelah konstruksi.
func NewMessage(msgType string, payload json.RawMessage) (Message, error) {
	m := Message{
		ID:            uuid.New(),
		Type:          msgType,
		Version:       1,
		OccurredAt:    time.Now().UTC(),
		Payload:       payload,
		CorrelationID: uuid.NewString(),
	}
	if err := m.Validate(); err != nil {
		return Message{}, err
	}
	return m, nil
}

// Validate memastikan envelope memenuhi kontrak transport sebelum dipublish atau
// di-dispatch.
func (m Message) Validate() error {
	if m.ID == uuid.Nil {
		return errors.New("messaging: id wajib diisi")
	}
	if m.Type == "" {
		return errors.New("messaging: type wajib diisi")
	}
	if m.Version < 1 {
		return errors.New("messaging: version harus >= 1")
	}
	if m.OccurredAt.IsZero() {
		return errors.New("messaging: occurred_at wajib diisi")
	}
	if len(m.Payload) == 0 || !json.Valid(m.Payload) {
		return errors.New("messaging: payload harus JSON yang valid")
	}
	return nil
}
