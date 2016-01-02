package typical

import (
	"fmt"
	"testing"

	. "github.com/luci/luci-go/common/testing/assertions"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTypical(t *testing.T) {
	t.Parallel()

	Convey("typical package methods", t, func() {
		Convey("can create static data objects", func() {
			Convey("singleton data", func() {
				So(Data("foo").First(), ShouldEqual, "foo")
				So(Data(nil).First(), ShouldEqual, nil)
				So(Data().First() == NoData, ShouldBeTrue)
			})

			Convey("multiple data", func() {
				So(Data("foo", 10).First(), ShouldEqual, "foo")
				So(Data("foo", 20).All(), ShouldResembleV, []interface{}{"foo", 20})
				So(Data().All(), ShouldResembleV, []interface{}{})
			})

			Convey("error", func() {
				So(Error(fmt.Errorf("hey")).Error(), ShouldErrLike, "hey")
				So(Error(nil), ShouldResembleV, Data())
				So(Error(nil).Error(), ShouldBeNil)

				So(func() { Error(fmt.Errorf("hey")).First() }, ShouldPanicLike, "hey")
				So(func() { Error(fmt.Errorf("hey")).All() }, ShouldPanicLike, "hey")
			})
		})

		Convey("registering non-functions fails", func() {
			So(func() { RegisterCommonFunction("hi", nil) }, ShouldPanicLike, "string must be a function")
		})
	})
}
