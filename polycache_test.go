package polycache

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/sync/singleflight"

	"github.com/javacode123/polycache/cache"
	cachemock "github.com/javacode123/polycache/cache/mock"
	loadermock "github.com/javacode123/polycache/loader/mock"
	"github.com/javacode123/polycache/multicache"
	multicachemock "github.com/javacode123/polycache/multicache/mock"
	"github.com/javacode123/polycache/pcerror"
	"github.com/javacode123/polycache/pclog"
	"github.com/javacode123/polycache/pclog/mock"
)

func TestNewPolyCache(t *testing.T) {
	Convey("test NewPolyCache", t, func() {

		option, cache0, cache1 := GetDefaultOption(), cachemock.NewMockCache(gomock.NewController(t)), cachemock.NewMockCache(gomock.NewController(t))
		pc := NewPolyCache[any, cache.Value](nil, nil, option, cache0, cache1)
		_, ok := pc.(*polyCache[any, cache.Value])
		So(ok, ShouldBeTrue)

	})
}

func TestPolyCache_Get(t *testing.T) {
	Convey("test Get", t, func() {
		ctx := context.TODO()
		ctl := gomock.NewController(t)
		defaultOption, mc, loader, log :=
			GetDefaultOption(), multicachemock.NewMockMultiCache(ctl), loadermock.NewMockLoader(ctl), mock.NewMockCtxLogger(ctl)
		item, key, value, mockErr := 1, "key", "value", errors.New("mock err")
		log.EXPECT().CtxDebugF(ctx, gomock.Any(), gomock.Any()).AnyTimes()
		loader.EXPECT().GenKey(ctx, item).Return(key).AnyTimes()
		pc := &polyCache[any, multicache.Value]{
			multiCache: mc,
			sf:         &singleflight.Group{},
			loader:     loader,
			log:        log,
			option:     defaultOption,
		}
		Convey("get by default option", func() {
			Convey("should return success when cache hit", func() {
				mc.EXPECT().Get(ctx, key, defaultOption.Duration).Return(value, nil)
				res, v := pc.Get(ctx, item)
				So(res, ShouldEqual, value)
				So(v, ShouldBeNil)
			})
			Convey("cache missing", func() {
				Convey("should return err when load source meet error", func() {
					mc.EXPECT().Get(ctx, key, defaultOption.Duration).Return(nil, pcerror.ErrNotFound)
					loader.EXPECT().Load(ctx, item).Return(nil, mockErr)
					_, v := pc.Get(ctx, item)
					So(v, ShouldBeError, mockErr)
				})
				Convey("should return success when load source success", func() {
					mc.EXPECT().Get(ctx, key, defaultOption.Duration).Return(nil, pcerror.ErrNotFound)
					loader.EXPECT().Load(ctx, item).Return(value, nil)
					mc.EXPECT().Set(ctx, key, value, defaultOption.Duration).Return(nil)
					v, err := pc.Get(ctx, item)
					So(err, ShouldBeNil)
					So(v, ShouldEqual, value)
				})
				Convey("should return success and skip set err when load source success", func() {
					mc.EXPECT().Get(ctx, key, defaultOption.Duration).Return(nil, pcerror.ErrNotFound)
					loader.EXPECT().Load(ctx, item).Return(value, nil)
					mc.EXPECT().Set(ctx, key, value, defaultOption.Duration).Return(mockErr)
					log.EXPECT().CtxWarnF(ctx, gomock.Any(), gomock.Any())
					v, err := pc.Get(ctx, item)
					So(err, ShouldBeNil)
					So(v, ShouldEqual, value)
				})
			})
		})
		Convey("get by custom option", func() {
			Convey("cache fist", func() {
				Convey("should return err when cache get err and set ReturnWhenCacheErr is true", func() {
					pc := &polyCache[any, multicache.Value]{
						multiCache: mc,
						sf:         &singleflight.Group{},
						loader:     loader,
						log:        log,
						option: GetDefaultOption().WithReturnErrFn(func(ctx context.Context, err error) bool {
							if err != nil {
								return true
							}
							return false
						}),
					}
					mc.EXPECT().Get(ctx, key, defaultOption.Duration).Return(nil, mockErr)
					_, e := pc.Get(ctx, item)
					So(e, ShouldBeError, mockErr)

				})
				Convey("should success with dyoption duration", func() {
					d := time.Minute
					pc := &polyCache[any, multicache.Value]{
						multiCache: mc,
						sf:         &singleflight.Group{},
						loader:     loader,
						log:        log,
						option: GetDefaultOption().WithReturnErrFn(func(ctx context.Context, err error) bool {
							if err != nil {
								return true
							}
							return false
						}),
					}
					mc.EXPECT().Get(ctx, key, d).Return(value, nil)
					v, e := pc.Get(ctx, item, WithDuration(d))
					So(e, ShouldBeNil)
					So(value, ShouldEqual, v)
					mc.EXPECT().Get(ctx, key, defaultOption.Duration).Return(value, nil)
					v, e = pc.Get(ctx, item)
					So(e, ShouldBeNil)
					So(value, ShouldEqual, v)
				})
			})
			Convey("source fist", func() {
				Convey("should return success", func() {
					nameSpace := "ns"
					pc := &polyCache[any, multicache.Value]{
						multiCache: mc,
						sf:         &singleflight.Group{},
						loader:     loader,
						log:        log,
						option:     GetDefaultOption().WithNameSpace(nameSpace).WithSourceStrategy(SsSourceFirst),
					}
					k := fmt.Sprintf("%s:%s", pc.option.NameSpace, key)
					loader.EXPECT().Load(ctx, item).Return(value, nil)
					mc.EXPECT().Set(ctx, k, value, defaultOption.Duration).Return(nil)
					v, e := pc.Get(ctx, item)
					So(e, ShouldBeNil)
					So(value, ShouldEqual, v)
				})
				Convey("should return success when load err and get cache success", func() {
					pc := &polyCache[any, multicache.Value]{
						multiCache: mc,
						sf:         &singleflight.Group{},
						loader:     loader,
						log:        log,
						option:     GetDefaultOption().WithSourceStrategy(SsSourceFirst),
					}
					loader.EXPECT().Load(ctx, item).Return(value, mockErr)
					mc.EXPECT().Get(ctx, key, defaultOption.Duration).Return(value, nil)
					v, e := pc.Get(ctx, item)
					So(e, ShouldBeNil)
					So(value, ShouldEqual, v)
				})
				Convey("should return err when load err and get cache err", func() {
					pc := &polyCache[any, multicache.Value]{
						multiCache: mc,
						sf:         &singleflight.Group{},
						loader:     loader,
						log:        log,
						option:     GetDefaultOption().WithSourceStrategy(SsSourceFirst),
					}
					loader.EXPECT().Load(ctx, item).Return(value, mockErr)
					mc.EXPECT().Get(ctx, key, defaultOption.Duration).Return(value, mockErr)
					_, e := pc.Get(ctx, item)
					So(e, ShouldBeError, mockErr)
				})
			})
			Convey("should panic", func() {
				pc := &polyCache[any, multicache.Value]{
					multiCache: mc,
					sf:         &singleflight.Group{},
					loader:     loader,
					log:        log,
					option:     GetDefaultOption().WithSourceStrategy(sourceStrategy(4)),
				}
				So(func() {
					_, _ = pc.Get(ctx, item)
				}, ShouldPanic)
			})
		})
	})
}

func TestPolyCache_Set(t *testing.T) {
	Convey("test PolyCache Set", t, func() {
		ctx := context.TODO()
		ctl := gomock.NewController(t)
		defaultOption, mc, loader, log :=
			GetDefaultOption(), multicachemock.NewMockMultiCache(ctl), loadermock.NewMockLoader(ctl), mock.NewMockCtxLogger(ctl)
		item, key, value, mockErr := 1, "key", "value", errors.New("mock err")
		loader.EXPECT().GenKey(ctx, item).Return(key).AnyTimes()
		pc := &polyCache[any, multicache.Value]{
			multiCache: mc,
			sf:         &singleflight.Group{},
			loader:     loader,
			log:        log,
			option:     defaultOption,
		}
		Convey("should return err when multiCache Set err", func() {
			dyDuration := time.Minute
			mc.EXPECT().Set(ctx, key, value, dyDuration).Return(mockErr)
			err := pc.Set(ctx, item, value, WithDuration(dyDuration))
			So(err, ShouldBeError, mockErr)
		})
		Convey("should return success", func() {
			mc.EXPECT().Set(ctx, key, value, defaultOption.Duration).Return(nil)
			err := pc.Set(ctx, item, value)
			So(err, ShouldBeNil)
		})
	})
}

func TestPolyCache_Del(t *testing.T) {
	Convey("test PolyCache Del", t, func() {
		ctx := context.TODO()
		ctl := gomock.NewController(t)
		defaultOption, mc, loader, log :=
			GetDefaultOption(), multicachemock.NewMockMultiCache(ctl), loadermock.NewMockLoader(ctl), mock.NewMockCtxLogger(ctl)
		item, key, mockErr := 1, "key", errors.New("mock err")
		loader.EXPECT().GenKey(ctx, item).Return(key).AnyTimes()
		pc := &polyCache[any, multicache.Value]{
			multiCache: mc,
			sf:         &singleflight.Group{},
			loader:     loader,
			log:        log,
			option:     defaultOption,
		}
		Convey("should return err when multiCache Set err", func() {
			mc.EXPECT().Del(ctx, key).Return(mockErr)
			err := pc.Del(ctx, item)
			So(err, ShouldBeError, mockErr)
		})
		Convey("should return success", func() {
			mc.EXPECT().Del(ctx, key).Return(nil)
			err := pc.Del(ctx, item)
			So(err, ShouldBeNil)
		})
	})
}

func TestPolyCache_Refresh(t *testing.T) {
	Convey("test PolyCache Refresh", t, func() {
		ctx := context.TODO()
		ctl := gomock.NewController(t)
		defaultOption, mc, loader, log :=
			GetDefaultOption(), multicachemock.NewMockMultiCache(ctl), loadermock.NewMockLoader(ctl), mock.NewMockCtxLogger(ctl)
		item, key, value, mockErr := 1, "key", "value", errors.New("mock err")
		loader.EXPECT().GenKey(ctx, item).Return(key).AnyTimes()
		pc := &polyCache[any, multicache.Value]{
			multiCache: mc,
			sf:         &singleflight.Group{},
			loader:     loader,
			log:        log,
			option:     defaultOption.WithDuration(time.Hour).WithLogeLevel(pclog.LevelInfo).WithDurationStep(1.5),
		}
		Convey("should return err when load  err", func() {
			loader.EXPECT().Load(ctx, item).Return(value, mockErr)
			err := pc.Refresh(ctx, item)
			So(err, ShouldBeError, mockErr)
		})
		Convey("should return success", func() {
			loader.EXPECT().Load(ctx, item).Return(value, nil)
			mc.EXPECT().Set(ctx, key, value, defaultOption.Duration).Return(nil)
			err := pc.Refresh(ctx, item, WithSourceStrategy(SsSourceFirst))
			So(err, ShouldBeNil)
		})
	})
}
