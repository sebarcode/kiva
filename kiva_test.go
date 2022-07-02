package kiva_test

import (
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"

	"git.kanosolution.net/kano/appkit"
	"git.kanosolution.net/kano/dbflex"
	"github.com/ariefdarmawan/datahub"
	_ "github.com/ariefdarmawan/flexmgo"
	"github.com/eaciit/toolkit"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/kiva"
	"github.com/sebarcode/kiva/kvsimple"
	"github.com/smartystreets/goconvey/convey"
)

var (
	logger  = appkit.LogWithPrefix("kvtest")
	connTxt = "mongodb://localhost:27017/testdb"
	h       = datahub.NewHubWithOpts(datahub.GeneralDbConnBuilder(connTxt), &datahub.HubOptions{UsePool: true, PoolSize: 20})
)

func TestSingle(t *testing.T) {
	convey.Convey("Preparing", t, func() {
		kv, err := prepareKiva()
		convey.So(err, convey.ShouldBeNil)

		convey.Convey("classic db data", func() {
			err := h.SaveAny("TestTable", codekit.M{}.Set("_id", "Key1").Set("Value", 100))
			convey.So(err, convey.ShouldBeNil)
			convey.Convey("get data", func() {
				destInt := int(0)
				err = kv.Get("Key1", &destInt)
				convey.So(err, convey.ShouldBeNil)
				convey.So(destInt, convey.ShouldEqual, 100)
			})
		})

		convey.Convey("injecting object data", func() {
			data := allTypes{
				ID:      "Data1",
				Name:    "Data1 Name",
				Age:     40,
				Salary:  8000,
				Roles:   []string{"CEO", "Founder"},
				Created: time.Now(),
			}

			err := kv.Set("Data1", &data, nil, false)
			convey.So(err, convey.ShouldBeNil)
			convey.Convey("validate object data", func() {
				getData := allTypes{}
				err := kv.Get("Data1", &getData)
				convey.So(err, convey.ShouldBeNil)
				convey.So(getData.Name, convey.ShouldEqual, data.Name)
				convey.So(getData.Age, convey.ShouldEqual, data.Age)
				convey.So(getData.Salary, convey.ShouldEqual, data.Salary)
				convey.So(getData.Created.UnixMilli(), convey.ShouldAlmostEqual, data.Created.UnixMilli())
			})
		})
	})
}

func TestSlice(t *testing.T) {
	convey.Convey("inject", t, func() {
		k, e := prepareKiva()
		convey.So(e, convey.ShouldBeNil)
		sources := make([]allTypes, 1000)

		for i := 0; i < 1000; i++ {
			sources[i] = allTypes{
				ID:      fmt.Sprintf("Data_%04d", i),
				Name:    fmt.Sprintf("Data_%d's Name", i),
				Age:     codekit.RandInt(30) + 15,
				Created: time.Now(),
			}
			sources[i].Salary = float64(sources[i].Age) * 100 * float64(toolkit.RandInt(10)/10)
			if e = k.Set(sources[i].ID, sources[i], nil, false); e != nil {
				break
			}
		}
		convey.So(e, convey.ShouldBeNil)

		convey.Convey("get data by pattern", func() {
			pattern := "Data_*"
			resDatas := []allTypes{}
			e := k.GetByPattern(pattern, &resDatas, false)
			convey.So(e, convey.ShouldBeNil)
			convey.So(len(resDatas), convey.ShouldEqual, len(sources))

			convey.Convey("validate", func() {
				//random check of 3 elements
				for i := 0; i < 3; i++ {
					index := codekit.RandInt(999)
					output := resDatas[index]
					var source allTypes
				getSource:
					for y := 0; y < 1000; y++ {
						if sources[y].ID == output.ID {
							source = sources[y]
							break getSource
						}
					}

					convey.So(output.Name, convey.ShouldEqual, source.Name)
					convey.So(output.Age, convey.ShouldEqual, source.Age)
					convey.So(output.Salary, convey.ShouldEqual, source.Salary)
					convey.So(output.Created.UnixMilli(), convey.ShouldAlmostEqual, source.Created.UnixMilli())
				}
			})
		})

		convey.Convey("get data by range", func() {
			from := "Data_0200"
			to := "Data_0299"
			resDatas := []allTypes{}
			e := k.GetRange(from, to, &resDatas, false)
			convey.So(e, convey.ShouldBeNil)
			convey.So(len(resDatas), convey.ShouldEqual, 100)

			convey.Convey("validate", func() {
				//random check of 3 elements
				for i := 0; i < 3; i++ {
					index := codekit.RandInt(len(resDatas))
					output := resDatas[index]
					var source allTypes
				getSource:
					for y := 0; y < 1000; y++ {
						if sources[y].ID == output.ID {
							source = sources[y]
							break getSource
						}
					}

					convey.So(output.Name, convey.ShouldEqual, source.Name)
					convey.So(output.Age, convey.ShouldEqual, source.Age)
					convey.So(output.Salary, convey.ShouldEqual, source.Salary)
					convey.So(output.Created.UnixMilli(), convey.ShouldAlmostEqual, source.Created.UnixMilli())
				}
			})
		})

		convey.Convey("delete", func() {
			k.Delete(false, "Data_0301")
			k.DeleteRange("Data_0320", "Data_0349", false)
			keys := k.Keys("Data_*")
			convey.So(len(keys), convey.ShouldEqual, 1000-31)

			keys = k.KeyRanges("Data_0300", "Data_0399")
			convey.So(len(keys), convey.ShouldEqual, 100-31)
		})
	})
}

func TestSyncDB(t *testing.T) {
	convey.Convey("set 10 key with syncToDb", t, func() {
		k, e := prepareKiva()
		convey.So(e, convey.ShouldBeNil)
		sources := make([]allTypes, 100)
		for i := 0; i < 100; i++ {
			sources[i] = allTypes{
				ID:      fmt.Sprintf("DB_%04d", i),
				Name:    fmt.Sprintf("DB_%d's Name", i),
				Age:     codekit.RandInt(30) + 15,
				Created: time.Now(),
			}
			sources[i].Salary = float64(sources[i].Age) * 100 * float64(toolkit.RandInt(10)/10)
			if e = k.Set(sources[i].ID, sources[i], nil, true); e != nil {
				break
			}
		}
		convey.So(e, convey.ShouldBeNil)

		convey.Convey("validate provider", func() {
			resDatas := []allTypes{}
			e := k.GetByPattern("DB_*", &resDatas, false)
			convey.So(e, convey.ShouldBeNil)
			convey.So(len(resDatas), convey.ShouldEqual, len(sources))

			//random check of 3 elements
			for i := 0; i < 3; i++ {
				index := codekit.RandInt(len(resDatas))
				output := resDatas[index]
				var source allTypes
			getSource:
				for y := 0; y < 1000; y++ {
					if sources[y].ID == output.ID {
						source = sources[y]
						break getSource
					}
				}
				convey.So(output.Name, convey.ShouldEqual, source.Name)
				convey.So(output.Age, convey.ShouldEqual, source.Age)
				convey.So(output.Salary, convey.ShouldEqual, source.Salary)
				convey.So(output.Created.UnixMilli(), convey.ShouldAlmostEqual, source.Created.UnixMilli())
			}
		})

		convey.Convey("validate db", func() {
			resDatas := []allTypes{}
			e := h.PopulateByFilter("TestTable", dbflex.StartWith("_id", "DB_"), 0, &resDatas)
			convey.So(e, convey.ShouldBeNil)
			convey.So(len(resDatas), convey.ShouldEqual, len(sources))

			//random check of 3 elements
			for i := 0; i < 3; i++ {
				index := codekit.RandInt(len(resDatas))
				output := resDatas[index]
				var source allTypes
			getSource:
				for y := 0; y < 100; y++ {
					if sources[y].ID == output.ID {
						source = sources[y]
						break getSource
					}
				}
				convey.So(output.Name, convey.ShouldEqual, source.Name)
				convey.So(output.Age, convey.ShouldEqual, source.Age)
				convey.So(output.Salary, convey.ShouldEqual, source.Salary)
				convey.So(output.Created.UnixMilli(), convey.ShouldAlmostEqual, source.Created.UnixMilli())
			}
		})
	})
}

func prepareKiva() (*kiva.Kiva, error) {
	/*
		provider, err := kvredis.New("", logger, nil)
		if err != nil {
			return nil, err
		}
	*/
	provider := kvsimple.New(&kiva.WriteOptions{TTL: 30 * time.Minute})
	kv, err := kiva.New(
		provider,
		func(key1, key2 string, kind kiva.GetKind, res interface{}) error {
			var f *dbflex.Filter
			switch kind {
			case kiva.GetByID:
				f = dbflex.Eq("_id", key1)

			case kiva.GetByPattern:
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
		},
		func(key1 string, value interface{}, op kiva.CommitKind) error {
			isStruct := reflect.Indirect(reflect.ValueOf(value)).Kind() == reflect.Struct
			w := dbflex.Eq("_id", key1)
			switch op {
			case kiva.CommitSave:
				if isStruct {
					return h.SaveAny("TestTable", value)
				} else {
					return h.SaveAny("TestTable", codekit.M{}.Set("_id", key1).Set("Value", value))
				}

			case kiva.CommitDelete:
				cmd := dbflex.From("TestTable").Where(w).Delete()
				_, e := h.Execute(cmd, nil)
				return e
			}
			return nil
		}, //committerfunc
		&kiva.WriteOptions{
			TTL: 15 * time.Minute,
		})
	return kv, err
}

type allTypes struct {
	ID      string `json:"_id" bson:"_id"`
	Name    string
	Age     int
	Salary  float64
	Roles   []string
	Created time.Time
}
