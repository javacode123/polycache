package example

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/javacode123/polycache"
	"github.com/javacode123/polycache/cache/gocache"
	rediscache "github.com/javacode123/polycache/cache/redis"
	"github.com/javacode123/polycache/pclog"
	"github.com/redis/go-redis/v9"
)

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (u User) MarshalBinary() (data []byte, err error) {
	return json.Marshal(u)
}

func TestGetWithLocalCache(t *testing.T) {
	// config your cache component
	cleanupInterval := time.Minute
	localCache := gocache.NewCache[User](cleanupInterval)
	// Singleton, It's safe for concurrent use by multiple goroutines.
	pc := polycache.NewPolyCache[int64, User](
		// key func, generate key from your req
		func(ctx context.Context, notHitItem int64) string {
			return fmt.Sprintf("id_%d", notHitItem)
		},
		// load source func: where the value from and you need cache the result
		func(ctx context.Context, notHitItem int64) (User, error) {
			return User{
				ID:   notHitItem,
				Name: "user_name",
			}, nil
		},
		// custom settings
		polycache.GetDefaultOption().WithDuration(time.Hour).WithLogeLevel(pclog.LevelDebug),
		localCache,
	)
	ctx := context.Background()
	// fist time to call Get method, it will use singleflight call load source func, and set KV to cache
	res, err := pc.Get(ctx, 1)
	if err != nil {
		panic(err)
	}
	fmt.Printf("get user: %+v \n", res)

	// second time to call Get method, will get from localcache
	res, err = pc.Get(ctx, 1)
	if err != nil {
		panic(err)
	}
	fmt.Printf("get user: %+v \n", res)

	// use dynamicOptions
	res, err = pc.Get(ctx, 2, polycache.WithSourceStrategy(polycache.SsSourceFirst))
	fmt.Printf("get user: %+v \n", res)

}

func TestGetWithMultiLevelCache(t *testing.T) {
	// redis cli
	redisCli := redis.NewClient(&redis.Options{Addr: miniredis.RunT(t).Addr()})
	cleanupInterval := time.Hour
	// first level cache is localcache, second level cache is redis
	fistLevelCache, secondLevelCache := gocache.NewCache[User](cleanupInterval), rediscache.NewCache[User](redisCli)
	pc := polycache.NewPolyCache[int64, User](
		// key func
		func(ctx context.Context, notHitItem int64) string {
			return fmt.Sprintf("id_%d", notHitItem)
		},
		// load source func: where the value from and you need cache the result
		func(ctx context.Context, notHitItem int64) (User, error) {
			return User{
				ID:   notHitItem,
				Name: "user_name",
			}, nil
		},
		// settings, durationStep used to calculate the duration. multiLevelCache[i]'duration = math.pow(expireTimeStep,i) * duration. Prevent multiLevelCache expiring simultaneously.
		polycache.GetDefaultOption().WithDuration(time.Hour).WithLogeLevel(pclog.LevelDebug).WithDurationStep(1.5),
		// multiLevelCache
		fistLevelCache, secondLevelCache,
	)
	ctx := context.Background()
	// fist time to call Get method, it will use singleflight call load source func, and set KV to localcahe and redis
	res, err := pc.Get(ctx, 1)
	if err != nil {
		panic(err)
	}
	fmt.Printf("get user: %+v \n", res)

	// second time to call Get method, will get from localcache
	res, err = pc.Get(ctx, 1)
	if err != nil {
		panic(err)
	}
	fmt.Printf("get user: %+v \n", res)

	res, err = pc.Get(ctx, 2, polycache.WithSourceStrategy(polycache.SsSourceFirst))
	fmt.Printf("get user: %+v \n", res)

}
