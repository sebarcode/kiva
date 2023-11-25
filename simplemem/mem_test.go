package simplemem_test

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/sebarcode/codekit"
	"github.com/sebarcode/kiva/simplemem"
	"github.com/sebarcode/logger"
	"github.com/smartystreets/goconvey/convey"
)

var (
	mem  = simplemem.NewMemory()
	logw = logger.NewLogEngine(true, false, "", "", "")
)

func TestMain(m *testing.M) {
	err := mem.Connect()
	if err != nil {
		logw.Error(err.Error())
		os.Exit(-1)
	}
	defer mem.Close()
	os.Exit(m.Run())
}

func TestMemory(t *testing.T) {
	convey.Convey("create 100 random data", t, func() {
		dataCount := 100
		baskets := make([]int, dataCount)
		for i := 0; i < dataCount; i++ {
			baskets[i] = codekit.RandInt(100000)
		}
		for index, basket := range baskets {
			mem.Set("basket", fmt.Sprintf("data_%05d", index), basket)
		}
		convey.So(mem.Len("basket"), convey.ShouldEqual, dataCount)
		convey.Convey("validate", func() {
			errTxt := ""
			for index, basket := range baskets {
				dt := int(0)
				_, err := mem.Get("basket", fmt.Sprintf("data_%05d", index), &dt)
				if err != nil {
					errTxt = err.Error()
					break
				}
				if basket != dt {
					errTxt = fmt.Sprintf("index %d: expect %d got %d", index, basket, dt)
				}
			}
			convey.So(errTxt, convey.ShouldBeBlank)

			convey.Convey("delete", func() {
				err := mem.Delete("basket", fmt.Sprintf("data_%05d", 10))
				convey.So(err, convey.ShouldBeNil)
				convey.So(mem.Len("basket"), convey.ShouldEqual, 99)
			})
		})

		convey.Convey("get non-existent data", func() {
			ne := int(00)
			_, err := mem.Get("basket", "not_exist_id", &ne)
			convey.So(err, convey.ShouldEqual, io.EOF)
		})
	})
}
