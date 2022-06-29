package kiva_test

import (
	"io"
	"testing"

	"git.kanosolution.net/kano/dbflex"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/kiva"
	"github.com/sebarcode/kiva/kvredis"
	"github.com/smartystreets/goconvey/convey"
)

var (
	h       *datahub.Hub
	connTxt string
)

func TestKiva(t *testing.T) {
	convey.Convey("Preparing", t, func() {
		h := datahub.NewHubWithOpts(datahub.GeneralDbConnBuilder(connTxt), &datahub.HubOptions{UsePool: true, PoolSize: 20})

		h.SaveAny("TestTable", codekit.M{}.Set("_id", "Key1").Set("Value", 100))

		provider, err := kvredis.New()
		convey.So(err, convey.ShouldBeNil)

		kv, err = kiva.New(provider, func(key1, key2 string, kind kiva.GetKind, res interface{}) error {
			var f *dbflex.Filter
			switch kind {
			case kiva.GetByID:
				f = dbflex.Eq("_id", key1)

			case kiva.GetByPrefix:
				f = dbflex.StartWith("_id", key1)

			case kiva.GetRange:
				f = dbflex.Range("_id", key1, key2)
			}

			ms := []codekit.M{}
			if e := h.PopulateByFilter("TestTable", f, 0, &ms); e != nil {
				return e
			}
			if len(ms) == 0 {
				return io.EOF
			}
			*(res.(*int)) = ms[0].GetInt("Value")
			return nil
		})
		convey.So(err, convey.ShouldBeNil)

		convey.Convey("get data", func() {
			kv.Get("key1")
		})
	})
}
