package ports

import (
	"context"

	"sipon-be/internal/modules/messaging/domain/message/valueobject"
)

// Publisher mengirim message ke exchange utama memakai msg.Type sebagai routing
// key. Implementasi: RabbitMQPublisher (dengan publisher confirm).
type Publisher interface {
	Publish(ctx context.Context, msg valueobject.Message) error
}

// QueuePublisher mengirim message langsung ke sebuah queue (biasanya retry queue)
// lewat default exchange, tanpa melalui routing key topic.
type QueuePublisher interface {
	PublishToQueue(ctx context.Context, queue string, msg valueobject.Message) error
}

// Delivery adalah satu message dari broker yang belum di-ack. Handler bertanggung
// jawab memanggil Ack/Nack.
type Delivery interface {
	Body() []byte
	Ack() error
	Nack(requeue bool) error
}

// ConsumerHandler memproses satu delivery. Implementasi diharapkan memanggil
// Ack/Nack pada Delivery.
type ConsumerHandler func(ctx context.Context, d Delivery) error

// Consumer mengonsumsi message dari sebuah queue dan menyerahkan tiap delivery ke
// handler. Implementasi: RabbitMQConsumer (manual acknowledgement).
type Consumer interface {
	Consume(ctx context.Context, queue string, handler ConsumerHandler) error
}
