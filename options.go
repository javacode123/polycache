package polycache

import (
	"context"
	"time"

	"github.com/javacode123/polycache/pclog"
)

type sourceStrategy int

const (
	SsCacheFirst  = iota // query from cache, if query miss, query from source and set to cache
	SsSourceFirst        // query from source and set to cache, if query miss, query from cache
	SsOnlyCache          // just query from cache
	SsOnlySource         // just query from source and do not cache query result
)

const (
	defaultOneHourDuration = time.Hour
	defaultDurationStep    = 1.5
)

type ReturnWhenCacheErr func(ctx context.Context, err error) bool

type option struct {
	NameSpace          string // Different PolyCache will store value by key which start with NameSpace
	SourceStrategy     sourceStrategy
	Duration           time.Duration // Entry can be evicted after time.Duration
	ReturnWhenCacheErr ReturnWhenCacheErr
	LogLevel           pclog.Level
	DurationStep       float64 // When use multiLevel cache, cache[level]'duration = time.Duration(float64(duration) * math.Pow(w.durationStep, float64(level)))
}

func (op *option) clone() *option {
	return &option{
		NameSpace:          op.NameSpace,
		SourceStrategy:     op.SourceStrategy,
		Duration:           op.Duration,
		ReturnWhenCacheErr: op.ReturnWhenCacheErr,
		LogLevel:           op.LogLevel,
		DurationStep:       op.DurationStep,
	}
}

func GetDefaultOption() *option {
	return &option{
		NameSpace:      "",
		SourceStrategy: SsCacheFirst,
		Duration:       defaultOneHourDuration,
		ReturnWhenCacheErr: func(ctx context.Context, err error) bool {
			return false
		},
		LogLevel:     pclog.LevelInfo,
		DurationStep: defaultDurationStep,
	}
}

func (op *option) WithNameSpace(ns string) *option {
	op.NameSpace = ns
	return op
}

func (op *option) WithSourceStrategy(ss sourceStrategy) *option {
	op.SourceStrategy = ss
	return op
}

func (op *option) WithDuration(duration time.Duration) *option {
	op.Duration = duration
	return op
}

func (op *option) WithReturnErrFn(fn ReturnWhenCacheErr) *option {
	op.ReturnWhenCacheErr = fn
	return op
}

func (op *option) WithLogeLevel(level pclog.Level) *option {
	op.LogLevel = level
	return op
}

func (op *option) WithDurationStep(ds float64) *option {
	op.DurationStep = ds
	return op
}

type DynamicOption func(op *option)

func WithSourceStrategy(ss sourceStrategy) DynamicOption {
	return func(op *option) {
		op.SourceStrategy = ss
	}
}

func WithDuration(duration time.Duration) DynamicOption {
	return func(op *option) {
		op.Duration = duration
	}
}
