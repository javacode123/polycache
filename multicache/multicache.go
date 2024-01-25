package multicache

//go:generate mockgen -source=multicache.go -package=mock -destination=mock/mockmulticache.go

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/javacode123/polycache/cache"
	"github.com/javacode123/polycache/pcerror"
	"github.com/javacode123/polycache/pclog"
)

// MultiCache wrapper many cache components
type MultiCache[Value any] interface {
	Get(ctx context.Context, key string, duration time.Duration) (Value, error)
	Set(ctx context.Context, key string, value Value, duration time.Duration) error
	Del(ctx context.Context, key string) error
}

type multiCache[Value any] struct {
	log          pclog.CtxLogger
	caches       []cache.Cache[Value]
	durationStep float64
}

func NewCacheWrapper[Value any](log pclog.CtxLogger, durationStep float64, caches ...cache.Cache[Value]) MultiCache[Value] {
	if len(caches) == 0 {
		panic("no cache provided")
	}
	w := &multiCache[Value]{
		caches:       make([]cache.Cache[Value], 0, len(caches)),
		durationStep: durationStep,
		log:          log,
	}

	for _, c := range caches {
		w.caches = append(w.caches, c)
	}

	return w
}

func (w *multiCache[Value]) Get(ctx context.Context, key string, duration time.Duration) (value Value, err error) {
	level := 0
	for i, c := range w.caches {
		value, err = c.Get(ctx, key)
		w.log.CtxDebugF(ctx, "get from cache: %s level: %d/%d with key: %v, value: %v,  err: %v", c.Name(ctx), i, len(w.caches)-1, key, value, err)
		if err == nil {
			level = i
			break
		}
		if errors.Is(err, pcerror.ErrNotFound) {
			continue
		}
		break
	}

	if err != nil {
		return value, err
	}

	// Flush cache up
	for i := 0; i < level; i++ {
		setErr := w.setCacheByLevel(ctx, key, value, duration, i)
		if setErr != nil {
			w.log.CtxWarnF(ctx, "setCacheByLevel failed and skipped, cache: %s key: %v, value: %v, level: %v/%d err: %v", w.caches[i].Name(ctx), key, value, i, len(w.caches)-1, setErr)
		}
	}

	return value, nil
}

func (w *multiCache[Value]) setCacheByLevel(ctx context.Context, key string, value Value, duration time.Duration, level int) error {
	duration = time.Duration(float64(duration) * math.Pow(w.durationStep, float64(level)))
	err := w.caches[level].Set(ctx, key, value, duration)
	w.log.CtxDebugF(ctx, "setCacheByLevel cache: %s, level: %d/%d, key: %v, value: %v, duration:%v, err: %v", w.caches[level].Name(ctx), level, len(w.caches)-1, key, value, duration, err)
	return err
}

func (w *multiCache[Value]) Set(ctx context.Context, key string, value Value, duration time.Duration) (err error) {
	for i := len(w.caches) - 1; i >= 0; i-- {
		err = w.setCacheByLevel(ctx, key, value, duration, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *multiCache[Value]) Del(ctx context.Context, key string) error {
	for i := len(w.caches) - 1; i >= 0; i-- {
		err := w.caches[i].Del(ctx, key)
		w.log.CtxDebugF(ctx, "del from cache: %s with key: %v, level: %d/%d err: %v", w.caches[i].Name(ctx), key, i, len(w.caches)-1, err)
		if err != nil {
			return err
		}
	}
	return nil
}
