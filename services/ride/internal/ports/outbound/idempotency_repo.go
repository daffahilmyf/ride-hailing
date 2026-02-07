package outbound

import "context"

type IdempotencyRepo interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Save(ctx context.Context, key string, response string) error
}
