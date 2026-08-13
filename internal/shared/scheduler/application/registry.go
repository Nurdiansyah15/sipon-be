package application

import (
	"context"
	"encoding/json"
	"time"
)

type HandlerFunc func(ctx context.Context, payload json.RawMessage) error

type Registry struct {
	handlers map[string]HandlerFunc
}

func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]HandlerFunc)}
}

func (r *Registry) Register(jobType string, handler HandlerFunc) {
	r.handlers[jobType] = handler
}

func (r *Registry) Dispatch(ctx context.Context, jobType string, payload json.RawMessage) error {
	h, ok := r.handlers[jobType]
	if !ok {
		return ErrHandlerNotFound
	}
	return h(ctx, payload)
}

func (r *Registry) Has(jobType string) bool {
	_, ok := r.handlers[jobType]
	return ok
}

type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string { return e.Err.Error() }
func (e *RetryableError) Unwrap() error { return e.Err }

type FatalError struct {
	Err error
}

func (e *FatalError) Error() string { return e.Err.Error() }
func (e *FatalError) Unwrap() error { return e.Err }

func IsFatal(err error) bool {
	_, ok := err.(*FatalError)
	return ok
}

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

func (p *RetryPolicy) Register(jobType string, maxRetry int) {
	p.overrides[jobType] = maxRetry
}

func (p *RetryPolicy) MaxRetryFor(jobType string) int {
	if v, ok := p.overrides[jobType]; ok {
		return v
	}
	return p.defaultMax
}

func CalculateRetryDelay(retryCount int) time.Duration {
	base := 30 * time.Second
	delay := base
	for i := 0; i < retryCount; i++ {
		delay *= 2
	}
	if delay > 30*time.Minute {
		delay = 30 * time.Minute
	}
	return delay
}
