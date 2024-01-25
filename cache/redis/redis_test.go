package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/javacode123/polycache/pcerror"
	"github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetName(t *testing.T) {
	Convey("TestGetName", t, func() {
		c := NewCache[string](newMockRedis(t))
		So(c.Name(context.TODO()), ShouldEqual, "redisCache")
	})
}

func TestGet(t *testing.T) {
	Convey("TestGet", t, func() {
		c := &redisWrapper[string]{
			redis: newMockRedis(t),
		}
		ctx := context.TODO()
		Convey("should return error when key not found", func() {
			_, err := c.Get(ctx, "key")
			So(err, ShouldEqual, pcerror.ErrNotFound)
		})

		Convey("should return value when key found", func() {
			value := "value"
			buf, _ := json.Marshal(value)
			err := c.redis.Set(ctx, "key", buf, time.Hour).Err()
			So(err, ShouldBeNil)
			v, err := c.Get(ctx, "key")
			So(err, ShouldBeNil)
			So(v, ShouldEqual, value)
		})
	})
}

func TestSet(t *testing.T) {
	Convey("TestSet", t, func() {
		redisCli := newMockRedis(t)
		c := &redisWrapper[string]{
			redis: redisCli,
		}
		ctx := context.TODO()
		key, value := "key", "value"
		err := c.Set(ctx, key, value, time.Hour)
		So(err, ShouldBeNil)
		buf, _ := redisCli.Get(ctx, key).Bytes()
		v := ""
		json.Unmarshal(buf, &v)
		So(v, ShouldEqual, value)
	})
}

func TestDel(t *testing.T) {
	Convey("TestDel", t, func() {
		redisCli := newMockRedis(t)
		c := &redisWrapper[string]{
			redis: redisCli,
		}
		ctx := context.TODO()
		key, value := "key", "value"
		redisCli.Set(ctx, key, value, time.Hour)
		err := c.Del(ctx, key)
		So(err, ShouldBeNil)
		e := redisCli.Get(ctx, key).Err()
		So(e, ShouldEqual, redis.Nil)
	})
}

func newMockRedis(t *testing.T) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: miniredis.RunT(t).Addr()})
}
