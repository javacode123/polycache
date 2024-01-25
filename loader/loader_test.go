package loader

import (
	"context"
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/javacode123/polycache/pcerror"
)

func TestWrapper(t *testing.T) {
	Convey("test Wrapper", t, func() {
		ctx := context.TODO()
		Convey("test load", func() {
			Convey("should recover panic", func() {
				loadFn := func(ctx context.Context, notHitItem int64) (v string, err error) {
					panic("panic")
				}
				w := Wrapper[int64, string](loadFn, nil)
				_, err := w.Load(ctx, 1)
				So(errors.Is(err, pcerror.ErrPanic), ShouldBeTrue)
			})
			Convey("should return success", func() {
				value := "value"
				loadFn := func(ctx context.Context, notHitItem int64) (v string, err error) {
					return value, nil
				}
				w := Wrapper[int64, string](loadFn, nil)
				v, err := w.Load(ctx, 1)
				So(err, ShouldBeNil)
				So(v, ShouldEqual, value)
			})
		})

		Convey("test genKey", func() {
			genKeyFn := func(ctx context.Context, notHitItem int64) string {
				return "key"
			}
			w := Wrapper[int64, string](nil, genKeyFn)
			key := w.GenKey(ctx, 1)
			So(key, ShouldEqual, "key")
		})
	})
}
