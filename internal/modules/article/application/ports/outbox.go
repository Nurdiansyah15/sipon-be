package ports

import (
	"context"
	"encoding/json"
)

type OutboxWriter interface {
	Save(ctx context.Context, routingKey string, payload json.RawMessage) error
}
