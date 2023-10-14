package simplestorage_test

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/sebarcode/codekit"
	"github.com/sebarcode/kiva/simplestorage"
	"github.com/smartystreets/goconvey/convey"
)

var (
	storage = simplestorage.NewStorage()
)

func TestMain(m *testing.M) {
	defer storage.Close()
	os.Exit(m.Run())
}

func TestStorage(t *testing.T) {
	convey.Convey("create 100 random data", t, func() {
		dataCount := 100
		baskets := make([]int, dataCount)
		for i := 0; i < dataCount; i++ {
			baskets[i] = codekit.RandInt(100000)
		}
		for index, basket := range baskets {
			storage.Set("basket", fmt.Sprintf("data_%05d", index), basket)
		}
		convey.So(storage.Len("basket"), convey.ShouldEqual, dataCount)
		convey.Convey("validate", func() {
			errTxt := ""
			for index, basket := range baskets {
				dt := int(0)
				err := storage.Get("basket", fmt.Sprintf("data_%05d", index), &dt)
				if err != nil {
					errTxt = err.Error()
					break
				}
				if basket != dt {
					errTxt = fmt.Sprintf("index %d: expect %d got %d", index, basket, dt)
				}
			}
			convey.So(errTxt, convey.ShouldBeBlank)
		})

		convey.Convey("get not existent data", func() {
			ne := int(00)
			err := storage.Get("basket", "not_exist_id", &ne)
			convey.So(err, convey.ShouldEqual, io.EOF)
		})
	})
}
