package messaging

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

// RetryDelayFor memilih delay TTL bertingkat berdasarkan attempt count.
func RetryDelayFor(attempt int, delays []time.Duration) time.Duration {
	if len(delays) == 0 {
		return time.Minute
	}
	idx := attempt - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(delays) {
		idx = len(delays) - 1
	}
	return delays[idx]
}
