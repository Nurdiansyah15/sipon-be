package policy

import (
	"math"
	"math/rand/v2"
	"time"
)

// RetryPolicy menentukan max attempt per routing key dengan default global.
type RetryPolicy struct {
	defaultMax int
	overrides  map[string]int
}

func NewRetryPolicy(defaultMax int) *RetryPolicy {
	return &RetryPolicy{
		defaultMax: defaultMax,
		overrides:  make(map[string]int),
	}
}

func (p *RetryPolicy) Register(routingKey string, maxRetry int) {
	p.overrides[routingKey] = maxRetry
}

func (p *RetryPolicy) MaxRetryFor(routingKey string) int {
	if max, ok := p.overrides[routingKey]; ok {
		return max
	}
	return p.defaultMax
}

func (p *RetryPolicy) IsRetryable(attempt, max int) bool {
	return attempt < max
}

// CalculateRetryDelay menghitung exponential backoff dengan jitter 80%-120%.
// Delay selalu naik terhadap attempt sebelumnya (sebelum mencapai cap maksimal).
func CalculateRetryDelay(attempt int, base, max time.Duration) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	if base <= 0 {
		base = time.Second
	}
	jitter := rand.Float64()*0.4 + 0.8
	delay := float64(base) * math.Pow(2, float64(attempt)) * jitter
	if delay > float64(max) {
		delay = float64(max)
	}
	return time.Duration(delay)
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
