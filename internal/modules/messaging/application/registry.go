package application

import (
	"context"
	"fmt"

	"sipon-be/internal/modules/messaging/domain/message/valueobject"
	messagingerrors "sipon-be/internal/modules/messaging/domain/message_job/errors"
)

// HandlerFunc menerima Message lengkap (bukan hanya payload) sehingga log/usecase
// dapat memakai message ID dan correlation ID.
type HandlerFunc func(ctx context.Context, msg valueobject.Message) error

// Registry memetakan routing key ke handler. Analog dengan HTTP router: dipakai
// oleh direct-dispatch bridge maupun RabbitMQ consumer.
type Registry struct {
	handlers map[string]HandlerFunc
}

func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]HandlerFunc)}
}

// Register menolak routing key kosong, handler nil, dan routing key duplikat —
// duplikat tidak boleh diam-diam meng-overwrite handler lama.
func (r *Registry) Register(routingKey string, handler HandlerFunc) error {
	if routingKey == "" {
		return fmt.Errorf("messaging: routing key kosong")
	}
	if handler == nil {
		return fmt.Errorf("messaging: handler untuk %q nil", routingKey)
	}
	if _, exists := r.handlers[routingKey]; exists {
		return fmt.Errorf("messaging: routing key %q sudah terdaftar", routingKey)
	}
	r.handlers[routingKey] = handler
	return nil
}

func (r *Registry) Dispatch(ctx context.Context, msg valueobject.Message) error {
	h, ok := r.handlers[msg.Type]
	if !ok {
		return messagingerrors.NewFatalError(fmt.Errorf("messaging: tidak ada handler untuk routing key %q", msg.Type))
	}
	return h(ctx, msg)
}

func (r *Registry) Has(routingKey string) bool {
	_, ok := r.handlers[routingKey]
	return ok
}
