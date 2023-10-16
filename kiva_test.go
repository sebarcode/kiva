package kiva_test

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/kiva"
	"github.com/sebarcode/kiva/simplemem"
	"github.com/sebarcode/kiva/simplestorage"
	"github.com/smartystreets/goconvey/convey"
)

var (
	kv      *kiva.Kv
	mem     *simplemem.Memory
	storage *simplestorage.Storage
)

func TestMain(m *testing.M) {
	mem = simplemem.NewMemory()
	storage = simplestorage.NewStorage()

	kv = kiva.New("kivatest", mem, storage)
	defer kv.Close()

	os.Exit(m.Run())
}

func TestKiva(t *testing.T) {
	convey.Convey("preparing data", t, func() {
		facts := prepTestData(50)

		convey.Convey("get data not exist, should return EOF", func() {
			resFact := new(TestModel)
			e := kv.Get("facts:"+facts[10].ID, resFact)
			convey.So(e, convey.ShouldEqual, io.EOF)
		})

		convey.Convey("inject data into storage", func() {
			convey.Convey("validate data entered", func() {
				for _, fact := range facts[:20] {
					storage.Set("facts", fact.ID, fact)
				}
				convey.So(storage.Len("facts"), convey.ShouldEqual, 20)

				resFact := new(TestModel)
				e := kv.Get("facts:"+facts[10].ID, resFact)
				convey.So(e, convey.ShouldBeNil)
				convey.So(resFact.Sub.Config["random"], convey.ShouldEqual, facts[10].Sub.Config["random"])
			})

			convey.Convey("update using set", func() {
				resFact := new(TestModel)
				e := kv.Set("facts:"+facts[30].ID, facts[30])
				convey.So(e, convey.ShouldBeNil)
				convey.So(mem.Len("facts"), convey.ShouldEqual, 2)

				e = kv.Get("facts:"+facts[30].ID, resFact)
				convey.So(e, convey.ShouldBeNil)
				convey.So(resFact.Sub.Config["random"], convey.ShouldEqual, facts[30].Sub.Config["random"])

				convey.Convey("re-update using set", func() {
					resFact := new(TestModel)
					facts[30].Sequence = 999
					e := kv.Set("facts:"+facts[30].ID, facts[30])
					convey.So(e, convey.ShouldBeNil)
					convey.So(mem.Len("facts"), convey.ShouldEqual, 2)

					e = kv.Get("facts:"+facts[30].ID, resFact)
					convey.So(e, convey.ShouldBeNil)
					convey.So(resFact.Sequence, convey.ShouldEqual, facts[30].Sequence)

					convey.Convey("delete", func() {
						e := kv.Delete("facts:" + facts[30].ID)
						convey.So(e, convey.ShouldBeNil)
						convey.So(mem.Len("facts"), convey.ShouldEqual, 1)
					})
				})
			})
		})
	})
}

type Submodel struct {
	Groups []string
	Config codekit.M
}

type TestModel struct {
	ID       string
	Sequence int
	Join     time.Time
	Name     string
	Sub      *Submodel
}

func newModel(id, name string) *TestModel {
	return &TestModel{
		ID:       id,
		Name:     name,
		Sequence: codekit.RandInt(100),
		Join:     time.Now(),
		Sub: &Submodel{
			Groups: func() []string {
				n := codekit.RandInt(5) + 1
				res := make([]string, n)
				for i := 0; i < n; i++ {
					res[i] = codekit.RandomString(18)
				}
				return res
			}(),
			Config: codekit.M{}.Set("random", codekit.RandFloat(1000, 2)),
		},
	}
}

func prepTestData(n int) []*TestModel {
	res := make([]*TestModel, n)
	for i := 0; i < n; i++ {
		res[i] = newModel(uuid.NewString(), codekit.RandomString(20))
	}
	return res
}
