package store

import "context"

type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}