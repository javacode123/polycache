package redis

import (
	"context"
	"errors"
	"time"

	"github.com/javacode123/polycache/cache"
	"github.com/javacode123/polycache/pcerror"
	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// wrapper redis client to implement Cache
type redisWrapper[Value any] struct {
	redis *redis.Client
}

func NewCache[Value any](redisCli *redis.Client) cache.Cache[Value] {
	return &redisWrapper[Value]{
		redis: redisCli,
	}
}

func (c *redisWrapper[Value]) Name(ctx context.Context) string {
	return "redisCache"
}

func (c *redisWrapper[Value]) Get(ctx context.Context, key string) (value Value, err error) {
	buf, err := c.redis.Get(ctx, key).Bytes()
	if err != nil {
		if notFound := errors.Is(err, redis.Nil); notFound {
			return value, pcerror.ErrNotFound
		}
		return value, err
	}

	err = json.Unmarshal(buf, &value)
	return value, err
}

func (c *redisWrapper[Value]) Set(ctx context.Context, key string, value Value, duration time.Duration) error {
	buf, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, key, buf, duration).Err()
}

func (c *redisWrapper[Value]) Del(ctx context.Context, key string) error {
	return c.redis.Del(ctx, key).Err()
}
