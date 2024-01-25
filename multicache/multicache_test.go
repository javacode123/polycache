package multicache

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/javacode123/polycache/cache"
	cachemock "github.com/javacode123/polycache/cache/mock"
	"github.com/javacode123/polycache/pcerror"
	logmock "github.com/javacode123/polycache/pclog/mock"
)

func TestNewCacheWrapper(t *testing.T) {
	Convey("test NewCacheWrapper", t, func() {
		Convey("should panic when no cache provided", func() {
			So(
				func() {
					_ = NewCacheWrapper[int](nil, 0)
				}, ShouldPanic,
			)
		})
		Convey("should success", func() {
			ctl := gomock.NewController(t)
			log, durationStep, cacheComponent := logmock.NewMockCtxLogger(ctl), 1.5, cachemock.NewMockCache(ctl)
			c := NewCacheWrapper[cache.Value](log, durationStep, cacheComponent)
			a, ok := c.(*multiCache[cache.Value])
			So(ok, ShouldBeTrue)
			So(a.log, ShouldEqual, log)
			So(a.durationStep, ShouldEqual, durationStep)
			So(a.caches, ShouldResemble, []cache.Cache[cache.Value]{cacheComponent})
		})
	})
}

func TestMultiCache_Get(t *testing.T) {
	Convey("test Get", t, func() {
		ctx := context.TODO()
		ctl := gomock.NewController(t)
		log, durationStep, cache1, cache2 := logmock.NewMockCtxLogger(ctl), 1.5, cachemock.NewMockCache(ctl), cachemock.NewMockCache(ctl)
		cache1.EXPECT().Name(ctx).Return("cache1").AnyTimes()
		cache2.EXPECT().Name(ctx).Return("cache2").AnyTimes()
		log.EXPECT().CtxDebugF(gomock.Any(), gomock.Any(), gomock.Any()).Return().AnyTimes()
		w := multiCache[cache.Value]{log: log, durationStep: durationStep, caches: []cache.Cache[cache.Value]{cache1, cache2}}
		key, value, duration, mockErr := "key", "value", time.Hour, errors.New("err")
		Convey("should return err when cache1 get error", func() {
			cache1.EXPECT().Get(ctx, key).Return(nil, mockErr)
			_, err := w.Get(ctx, key, duration)
			So(err, ShouldBeError, mockErr)
		})

		Convey("should return err when cache1 not found and cache2 get error", func() {
			cache1.EXPECT().Get(ctx, key).Return(nil, pcerror.ErrNotFound)
			cache2.EXPECT().Get(ctx, key).Return(nil, mockErr)
			_, err := w.Get(ctx, key, duration)
			So(err, ShouldBeError, mockErr)
		})

		Convey("should return err when cache1 not found and cache2 not found", func() {
			cache1.EXPECT().Get(ctx, key).Return(nil, pcerror.ErrNotFound)
			cache2.EXPECT().Get(ctx, key).Return(nil, pcerror.ErrNotFound)
			_, err := w.Get(ctx, key, duration)
			So(err, ShouldBeError, pcerror.ErrNotFound)
		})

		Convey("should return value when cache1 found", func() {
			cache1.EXPECT().Get(ctx, key).Return(value, nil)
			v, err := w.Get(ctx, key, duration)
			So(err, ShouldBeNil)
			So(v, ShouldEqual, value)
		})

		Convey("should return value and set cache when cache1 not found and cache2 found", func() {
			cache1.EXPECT().Get(ctx, key).Return(nil, pcerror.ErrNotFound)
			cache2.EXPECT().Get(ctx, key).Return(value, nil)
			levelDuration := time.Duration(float64(duration) * math.Pow(w.durationStep, float64(0)))
			cache1.EXPECT().Set(ctx, key, value, levelDuration).Return(nil)
			v, err := w.Get(ctx, key, duration)
			So(err, ShouldBeNil)
			So(v, ShouldEqual, value)
		})

		Convey("should return value set cache and skip err when cache1 not found and cache2 found and cache1 set meet err", func() {
			cache1.EXPECT().Get(ctx, key).Return(nil, pcerror.ErrNotFound)
			cache2.EXPECT().Get(ctx, key).Return(value, nil)
			levelDuration := time.Duration(float64(duration) * math.Pow(w.durationStep, float64(0)))
			cache1.EXPECT().Set(ctx, key, value, levelDuration).Return(mockErr)
			log.EXPECT().CtxWarnF(ctx, gomock.Any(), gomock.Any()).Return()
			v, err := w.Get(ctx, key, duration)
			So(err, ShouldBeNil)
			So(v, ShouldEqual, value)
		})
	})
}

func TestMultiCache_Set(t *testing.T) {
	Convey("test Set", t, func() {
		ctx := context.TODO()
		ctl := gomock.NewController(t)
		log, durationStep, cache1, cache2 := logmock.NewMockCtxLogger(ctl), 1.5, cachemock.NewMockCache(ctl), cachemock.NewMockCache(ctl)
		cache1.EXPECT().Name(ctx).Return("cache1").AnyTimes()
		cache2.EXPECT().Name(ctx).Return("cache2").AnyTimes()
		log.EXPECT().CtxDebugF(gomock.Any(), gomock.Any(), gomock.Any()).Return().AnyTimes()
		w := multiCache[cache.Value]{log: log, durationStep: durationStep, caches: []cache.Cache[cache.Value]{cache1, cache2}}
		key, value, duration, mockErr := "key", "value", time.Hour, errors.New("err")
		Convey("should return err when cache2 set error", func() {
			cache2.EXPECT().Set(ctx, key, value, time.Duration(float64(duration)*math.Pow(w.durationStep, float64(1)))).Return(mockErr)
			err := w.Set(ctx, key, value, duration)
			So(err, ShouldBeError, mockErr)
		})
		Convey("should return err when cache1 set error", func() {
			cache2.EXPECT().Set(ctx, key, value, time.Duration(float64(duration)*durationStep)).Return(nil)
			cache1.EXPECT().Set(ctx, key, value, duration).Return(mockErr)
			err := w.Set(ctx, key, value, duration)
			So(err, ShouldBeError, mockErr)
		})
		Convey("should return success", func() {
			cache2.EXPECT().Set(ctx, key, value, time.Duration(float64(duration)*durationStep)).Return(nil)
			cache1.EXPECT().Set(ctx, key, value, duration).Return(nil)
			err := w.Set(ctx, key, value, duration)
			So(err, ShouldBeNil)
		})
	})
}

func TestMultiCache_Del(t *testing.T) {
	Convey("test Del", t, func() {
		ctx := context.TODO()
		ctl := gomock.NewController(t)
		log, durationStep, cache1, cache2 := logmock.NewMockCtxLogger(ctl), 1.5, cachemock.NewMockCache(ctl), cachemock.NewMockCache(ctl)
		cache1.EXPECT().Name(ctx).Return("cache1").AnyTimes()
		cache2.EXPECT().Name(ctx).Return("cache2").AnyTimes()
		log.EXPECT().CtxDebugF(gomock.Any(), gomock.Any(), gomock.Any()).Return().AnyTimes()
		w := multiCache[cache.Value]{log: log, durationStep: durationStep, caches: []cache.Cache[cache.Value]{cache1, cache2}}
		key, mockErr := "key", errors.New("err")
		Convey("should return err when cache2 del error", func() {
			cache2.EXPECT().Del(ctx, key).Return(mockErr)
			err := w.Del(ctx, key)
			So(err, ShouldBeError, mockErr)
		})
		Convey("should return err when cache1 del error", func() {
			cache2.EXPECT().Del(ctx, key).Return(nil)
			cache1.EXPECT().Del(ctx, key).Return(mockErr)
			err := w.Del(ctx, key)
			So(err, ShouldBeError, mockErr)
		})
		Convey("should return success", func() {
			cache2.EXPECT().Del(ctx, key).Return(nil)
			cache1.EXPECT().Del(ctx, key).Return(nil)
			err := w.Del(ctx, key)
			So(err, ShouldBeNil)
		})
	})
}
