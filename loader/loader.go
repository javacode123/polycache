package loader

//go:generate mockgen -source=loader.go -package=mock -destination=mock/mockloader.go

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/javacode123/polycache/pcerror"
)

// Loader component, used to load data from source and gen key by item which need store in KV cache
type Loader[Item, Value any] interface {
	// Load Value from source
	Load(ctx context.Context, notHitItem Item) (Value, error)
	// GenKey generate key from item, use key to identity every Value
	GenKey(ctx context.Context, notHitItem Item) string
}

type LoadFunc[Item, Value any] func(ctx context.Context, notHitItem Item) (Value, error)
type GenKeyFunc[Item, Value any] func(ctx context.Context, notHitItem Item) string

type load[Item, Value any] struct {
	loadFunc   LoadFunc[Item, Value]
	genKeyFunc GenKeyFunc[Item, Value]
}

func Wrapper[Item, Value any](lf LoadFunc[Item, Value], gkf GenKeyFunc[Item, Value]) Loader[Item, Value] {
	return &load[Item, Value]{
		loadFunc: func(ctx context.Context, notHitItem Item) (v Value, err error) {
			defer func() {
				if panicErr := recover(); panicErr != nil {
					stack := string(debug.Stack())
					err = pcerror.ErrPanic.WithCauseAndStack(fmt.Errorf("load data panic=%s", panicErr), stack)
				}
			}()
			return lf(ctx, notHitItem)
		},
		genKeyFunc: gkf,
	}

}

func (l *load[Item, Value]) Load(ctx context.Context, notHitItem Item) (Value, error) {
	return l.loadFunc(ctx, notHitItem)
}

func (l *load[Item, Value]) GenKey(ctx context.Context, notHitItem Item) string {
	return l.genKeyFunc(ctx, notHitItem)
}
