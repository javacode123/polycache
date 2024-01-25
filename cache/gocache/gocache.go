package gocache

import (
	"context"
	"time"

	"github.com/javacode123/polycache/cache"
	"github.com/javacode123/polycache/pcerror"
	gocache "github.com/patrickmn/go-cache"
)

type cacheWrapper[Value any] struct {
	goCache *gocache.Cache
}

func NewCache[Value any](cleanupInterval time.Duration) cache.Cache[Value] {
	// unUseDuration is invalid, just init gocache need default duration, polycache will use duration from option which user set
	unUseDuration := time.Hour
	return &cacheWrapper[Value]{
		goCache: gocache.New(unUseDuration, cleanupInterval),
	}
}

func (c *cacheWrapper[Value]) Name(ctx context.Context) string {
	return "goCache"
}

func (c *cacheWrapper[Value]) Get(ctx context.Context, key string) (value Value, err error) {
	valueInterface, found := c.goCache.Get(key)
	if !found {
		return value, pcerror.ErrNotFound
	}

	value = valueInterface.(Value)
	return value, nil
}

func (c *cacheWrapper[Value]) Set(ctx context.Context, key string, value Value, duration time.Duration) error {
	c.goCache.Set(key, value, duration)
	return nil
}

func (c *cacheWrapper[Value]) Del(ctx context.Context, key string) error {
	c.goCache.Delete(key)
	return nil
}
