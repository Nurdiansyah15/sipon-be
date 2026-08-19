package messaging

import (
	"context"
	"errors"

	"sipon-be/internal/modules/messaging/application/ports"
)

// Module adalah facade komposisi pipeline messaging: registry handler, retry
// policy, dan publisher (opsional). Dibuat konsisten dengan pola modul
// (module.go + contract.go) pada internal/modules/*.
type Module struct {
	registry  *ports.Registry
	policy    *ports.RetryPolicy
	publisher ports.Publisher
}

// NewModule membuat facade messaging. defaultMaxRetry dipakai sebagai batas
// global retry di RetryPolicy.
func NewModule(defaultMaxRetry int) *Module {
	return &Module{
		registry: ports.NewRegistry(),
		policy:   ports.NewRetryPolicy(defaultMaxRetry),
	}
}

// WithPublisher memasang publisher (mis. RabbitMQ) untuk operasi Publish.
func (m *Module) WithPublisher(p ports.Publisher) *Module {
	m.publisher = p
	return m
}

func (m *Module) Registry() *ports.Registry {
	return m.registry
}

func (m *Module) RetryPolicy() *ports.RetryPolicy {
	return m.policy
}

// Register mendaftarkan handler untuk sebuah routing key.
func (m *Module) Register(routingKey string, handler ports.HandlerFunc) error {
	return m.registry.Register(routingKey, handler)
}

// Dispatch mengirim message ke handler yang terdaftar (mode direct).
func (m *Module) Dispatch(ctx context.Context, msg ports.Message) error {
	return m.registry.Dispatch(ctx, msg)
}

// Publish mengirim message ke broker via publisher yang telah dipasang.
func (m *Module) Publish(ctx context.Context, msg ports.Message) error {
	if m.publisher == nil {
		return errors.New("messaging: publisher belum dipasang")
	}
	return m.publisher.Publish(ctx, msg)
}
