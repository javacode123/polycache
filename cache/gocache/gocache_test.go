package gocache

import (
	"context"
	"testing"
	"time"

	"github.com/javacode123/polycache/pcerror"
	gocache "github.com/patrickmn/go-cache"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetName(t *testing.T) {
	Convey("TestGetName", t, func() {
		c := NewCache[string](time.Hour)
		So(c.Name(context.TODO()), ShouldEqual, "goCache")
	})
}

func TestGet(t *testing.T) {
	Convey("TestGet", t, func() {
		goCache := gocache.New(time.Hour, time.Hour)
		c := &cacheWrapper[string]{
			goCache: goCache,
		}
		ctx := context.TODO()
		Convey("should return error when key not found", func() {
			_, err := c.Get(ctx, "key")
			So(err, ShouldEqual, pcerror.ErrNotFound)
		})

		Convey("should return value when key found", func() {
			value := "value"
			goCache.Set("key", value, time.Hour)
			value, err := c.Get(ctx, "key")
			So(err, ShouldBeNil)
			So(value, ShouldEqual, value)
		})
	})
}

func TestSet(t *testing.T) {
	Convey("TestSet", t, func() {
		goCache := gocache.New(time.Hour, time.Hour)
		c := &cacheWrapper[string]{
			goCache: goCache,
		}
		ctx := context.TODO()
		key, value := "key", "value"
		err := c.Set(ctx, key, value, time.Hour)
		So(err, ShouldBeNil)
		v, _ := goCache.Get(key)
		So(v, ShouldEqual, value)
	})
}

func TestDel(t *testing.T) {
	Convey("TestDel", t, func() {
		goCache := gocache.New(time.Hour, time.Hour)
		c := &cacheWrapper[string]{
			goCache: goCache,
		}
		ctx := context.TODO()
		key, value := "key", "value"
		goCache.Set(key, value, time.Hour)
		err := c.Del(ctx, key)
		So(err, ShouldBeNil)
		_, found := goCache.Get(key)
		So(found, ShouldBeFalse)
	})
}
