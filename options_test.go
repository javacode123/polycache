package polycache

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v4"
	"github.com/bxcodec/faker/v4/pkg/options"
	"github.com/javacode123/polycache/pclog"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetDefaultOption(t *testing.T) {
	Convey("test DefaultOption", t, func() {
		d := &option{
			NameSpace:      "",
			SourceStrategy: SsCacheFirst,
			Duration:       defaultOneHourDuration,
			ReturnWhenCacheErr: func(ctx context.Context, err error) bool {
				return false
			},
			LogLevel:     pclog.LevelInfo,
			DurationStep: defaultDurationStep,
		}
		g := GetDefaultOption()
		So(d.ReturnWhenCacheErr(nil, nil), ShouldEqual, g.ReturnWhenCacheErr(nil, nil))
		d.ReturnWhenCacheErr = nil
		g.ReturnWhenCacheErr = nil
		So(d, ShouldResemble, g)
	})
}

func TestCloneOption(t *testing.T) {
	Convey("test CloneOption", t, func() {
		o := &option{}
		_ = faker.FakeData(o, options.WithFieldsToIgnore("ReturnWhenCacheErr"))
		c := o.clone()
		So(c, ShouldResemble, o)
	})
}
