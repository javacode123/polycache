package polycache

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/sync/singleflight"

	"github.com/javacode123/polycache/cache"
	"github.com/javacode123/polycache/loader"
	"github.com/javacode123/polycache/multicache"
	"github.com/javacode123/polycache/pcerror"
	"github.com/javacode123/polycache/pclog"
)

type PolyCache[Item, Value any] interface {
	Get(ctx context.Context, item Item, options ...DynamicOption) (Value, error)
	Set(ctx context.Context, item Item, value Value, options ...DynamicOption) error
	Del(ctx context.Context, item Item, options ...DynamicOption) error
	Refresh(ctx context.Context, item Item, options ...DynamicOption) error
}

// NewPolyCache Just need create a polyCache, It's safe for concurrent use by multiple goroutines.
func NewPolyCache[Item, Value any](
	gkf loader.GenKeyFunc[Item, Value],
	lf loader.LoadFunc[Item, Value],
	option *option,
	cache ...cache.Cache[Value],
) PolyCache[Item, Value] {

	log := pclog.NewLogger(option.LogLevel)
	mc := multicache.NewCacheWrapper[Value](log, option.DurationStep, cache...)
	pc := &polyCache[Item, Value]{
		multiCache: mc,
		sf:         &singleflight.Group{},
		loader:     loader.Wrapper(lf, gkf),
		log:        log,
		option:     option,
	}

	return pc
}

type polyCache[Item, Value any] struct {
	multiCache multicache.MultiCache[Value]
	sf         *singleflight.Group
	loader     loader.Loader[Item, Value]
	log        pclog.CtxLogger
	option     *option
}

func (pc *polyCache[Item, Value]) clone() *polyCache[Item, Value] {
	return &polyCache[Item, Value]{
		multiCache: pc.multiCache,
		sf:         pc.sf,
		loader:     pc.loader,
		log:        pc.log,
		option:     pc.option.clone(),
	}
}

func (pc *polyCache[Item, Value]) applyDynamicOption(options ...DynamicOption) *polyCache[Item, Value] {
	pcc := pc.clone()
	for _, op := range options {
		op(pcc.option)
	}
	return pcc
}

func (pc *polyCache[Item, Value]) Get(ctx context.Context, item Item, options ...DynamicOption) (Value, error) {
	pcc := pc.applyDynamicOption(options...)
	op := pcc.option
	pcc.log.CtxDebugF(ctx, "polyCache call Get with item: %v, options: %+v", item, *op)

	switch op.SourceStrategy {
	case SsCacheFirst:
		{
			value, err := pcc.getFromCache(ctx, item)
			if shouldReturn := err == nil ||
				(!errors.Is(err, pcerror.ErrNotFound) && op.ReturnWhenCacheErr != nil && op.ReturnWhenCacheErr(ctx, err)); shouldReturn {
				return value, err
			}

			value, err = pcc.getFromLoader(ctx, item, true)
			return value, err
		}
	case SsSourceFirst:
		{
			value, err := pcc.getFromLoader(ctx, item, true)
			if err != nil {
				value, err = pcc.getFromCache(ctx, item)
			}
			return value, err
		}
	case SsOnlyCache:
		{
			return pcc.getFromCache(ctx, item)
		}
	case SsOnlySource:
		{
			return pcc.getFromLoader(ctx, item, false)
		}

	default:
		panic(fmt.Errorf("not support %v", op.SourceStrategy))
	}

}

func (pc *polyCache[Item, Value]) getFromCache(ctx context.Context, item Item) (value Value, err error) {
	key := pc.getCacheKey(ctx, item)
	return pc.multiCache.Get(ctx, key, pc.option.Duration)
}

func (pc *polyCache[Item, Value]) getCacheKey(ctx context.Context, item Item) string {
	key := pc.loader.GenKey(ctx, item)
	if pc.option.NameSpace == "" {
		return key
	}

	return fmt.Sprintf("%s:%s", pc.option.NameSpace, key)
}

func (pc *polyCache[Item, Value]) getFromLoader(ctx context.Context, item Item, cacheRes bool) (value Value, err error) {
	key := pc.getCacheKey(ctx, item)
	valueInterface, err, _ := pc.sf.Do(key,
		func() (interface{}, error) {
			return pc.loader.Load(ctx, item)
		},
	)
	pc.log.CtxDebugF(ctx, "polyCache getFromLoader with key: %v, value: %v, err: %v", key, valueInterface, err)
	if err != nil {
		return
	}

	value = valueInterface.(Value)
	if cacheRes {
		setErr := pc.multiCache.Set(ctx, key, value, pc.option.Duration)
		if setErr != nil {
			pc.log.CtxWarnF(ctx, "polyCache set key: %v, value: %v, Duration: %v to cache error: %v", key, value, pc.option.Duration, err)
		}
	}
	return value, err
}

func (pc *polyCache[Item, Value]) Set(ctx context.Context, item Item, value Value, options ...DynamicOption) error {
	pcc := pc.applyDynamicOption(options...)
	key := pcc.getCacheKey(ctx, item)

	return pcc.multiCache.Set(ctx, key, value, pcc.option.Duration)
}

func (pc *polyCache[Item, Value]) Del(ctx context.Context, item Item, options ...DynamicOption) error {
	pcc := pc.applyDynamicOption(options...)
	key := pcc.getCacheKey(ctx, item)

	return pcc.multiCache.Del(ctx, key)
}

func (pc *polyCache[Item, Value]) Refresh(ctx context.Context, item Item, options ...DynamicOption) error {
	pcc := pc.applyDynamicOption(options...)
	value, err := pcc.loader.Load(ctx, item)
	if err != nil {
		return err
	}

	return pcc.multiCache.Set(ctx, pcc.getCacheKey(ctx, item), value, pcc.option.Duration)
}
