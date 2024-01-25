package cache

//go:generate mockgen -source=cache.go -package=mock -destination=mock/mockcache.go

import (
	"context"
	"time"
)

// Cache component used to store data, you can implement this by redis, memcache, lcalcache etc
type Cache[Value any] interface {
	Name(ctx context.Context) string
	// Get multi level cache need duration param, cache[i]'s duration is  math.pow(step,i) * duration
	Get(ctx context.Context, key string) (Value, error)
	Set(ctx context.Context, key string, value Value, duration time.Duration) error
	Del(ctx context.Context, key string) error
}
