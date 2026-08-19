package ports

import (
	"fmt"
	"time"
)

// RetryQueueName menamai retry queue TTL untuk kombinasi (queue role, routing key,
// delay). Routing key dipertahankan dalam nama agar topology bisa meng-set
// x-dead-letter-routing-key ke routing key asli, sehingga setelah TTL message
// kembali ke main exchange dengan routing key yang benar.
func RetryQueueName(queue, routingKey string, delay time.Duration) string {
	return fmt.Sprintf("%s.retry.%s.%d", queue, routingKey, int(delay.Seconds()))
}
