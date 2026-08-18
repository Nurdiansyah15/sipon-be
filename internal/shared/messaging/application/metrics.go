package application

import "sync/atomic"

// Metrics adalah koleksi counter ringan (tanpa dependency) untuk observability
// pipeline messaging. Increment dilakukan di OutboxRelay dan MessageConsumer.
type Metrics struct {
	OutboxPublished   atomic.Int64
	OutboxPublishFail atomic.Int64
	OutboxRecovered   atomic.Int64

	Handled   atomic.Int64
	Succeeded atomic.Int64
	Failed    atomic.Int64
	Retried   atomic.Int64
	Duplicate atomic.Int64
}

type MetricsSnapshot struct {
	OutboxPublished   int64 `json:"outbox_published"`
	OutboxPublishFail int64 `json:"outbox_publish_fail"`
	OutboxRecovered   int64 `json:"outbox_recovered"`
	MessagesHandled   int64 `json:"messages_handled"`
	MessagesSucceeded int64 `json:"messages_succeeded"`
	MessagesFailed    int64 `json:"messages_failed"`
	MessagesRetried   int64 `json:"messages_retried"`
	MessagesDuplicate int64 `json:"messages_duplicate"`
}

func (m *Metrics) Snapshot() MetricsSnapshot {
	if m == nil {
		return MetricsSnapshot{}
	}
	return MetricsSnapshot{
		OutboxPublished:   m.OutboxPublished.Load(),
		OutboxPublishFail: m.OutboxPublishFail.Load(),
		OutboxRecovered:   m.OutboxRecovered.Load(),
		MessagesHandled:   m.Handled.Load(),
		MessagesSucceeded: m.Succeeded.Load(),
		MessagesFailed:    m.Failed.Load(),
		MessagesRetried:   m.Retried.Load(),
		MessagesDuplicate: m.Duplicate.Load(),
	}
}
