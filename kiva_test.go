package kiva_test

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/ariefdarmawan/serde"
	"github.com/eaciit/toolkit"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/kiva"
	"github.com/sebarcode/kiva/kvsimple"
	"github.com/smartystreets/goconvey/convey"
)

type storage map[string]interface{}

var (
	tableName     = "dataku"
	sourceStorage = map[string]storage{}
)

func init() {
	sourceStorage[tableName] = storage{}
}

func TestSingle(t *testing.T) {
	convey.Convey("Preparing", t, func() {
		kv, err := prepareKiva()
		convey.So(err, convey.ShouldBeNil)
		sourceStorage[tableName]["Key1"] = map[string]interface{}{"_id": 1, "Value": 100}
		convey.Convey("classic db data", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.Convey("get data", func() {
				destInt := int(0)
				err = kv.Get(tableName+":Key1", &destInt)
				convey.So(err, convey.ShouldBeNil)
				convey.So(destInt, convey.ShouldEqual, 100)

				sourceStorage[tableName]["Key1"] = map[string]interface{}{"_id": 1, "Value": 150}
				convey.Convey("should read from memory", func() {
					err = kv.Get(tableName+":Key1", &destInt)
					convey.So(err, convey.ShouldBeNil)
					convey.So(destInt, convey.ShouldEqual, 100)

					kv.Delete(false, tableName+":Key1")
					convey.Convey("should read from db", func() {
						err = kv.Get(tableName+":Key1", &destInt)
						convey.So(err, convey.ShouldBeNil)
						convey.So(destInt, convey.ShouldEqual, 150)
					})
				})
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

			err := kv.Set(tableName+":"+"Data1", &data, nil, false)
			convey.So(err, convey.ShouldBeNil)
			convey.Convey("validate object data", func() {
				getData := allTypes{}
				err := kv.Get(tableName+":"+"Data1", &getData)
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
			if e = k.Set(tableName+":"+sources[i].ID, sources[i], nil, false); e != nil {
				break
			}
		}
		convey.So(e, convey.ShouldBeNil)

		convey.Convey("get data by pattern", func() {
			pattern := tableName + ":" + "Data_*"
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
			from := tableName + ":" + "Data_0200"
			to := tableName + ":" + "Data_0299"
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
			k.Delete(false, tableName+":"+"Data_0301")
			k.DeleteRange(tableName+":"+"Data_0320", tableName+":"+"Data_0349", false)
			keys := k.Keys(tableName + ":" + "Data_*")
			convey.So(len(keys), convey.ShouldEqual, 1000-31)

			keys = k.KeyRanges(tableName+":"+"Data_0300", tableName+":"+"Data_0399")
			convey.So(len(keys), convey.ShouldEqual, 100-31)

			convey.Convey("delete by pattern", func() {
				resDatas := []allTypes{}
				pattern := tableName + ":" + "Data_*"
				k.DeleteByPattern(pattern, true)
				e = k.GetByPattern(pattern, &resDatas, false)
				convey.So(e, convey.ShouldBeNil)
				convey.So(len(resDatas), convey.ShouldEqual, 0)
			})
		})
	})
}

func TestSync(t *testing.T) {
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
			if e = k.Set(tableName+":"+sources[i].ID, sources[i], nil, true); e != nil {
				break
			}
		}
		convey.So(e, convey.ShouldBeNil)

		convey.Convey("validate provider", func() {
			resDatas := []allTypes{}
			e := k.GetByPattern(tableName+":"+"DB_*", &resDatas, false)
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
			for _, v := range sourceStorage[tableName] {
				va, castOK := v.(map[string]interface{})["Value"].(allTypes)
				if !castOK {
					continue
				}
				id := va.ID
				if strings.HasPrefix(id, "DB_") {
					resDatas = append(resDatas, va)
				}
			}
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

			convey.Convey("sync get ranges", func() {
				convey.Convey("delete data on persistent storage", func() {
					tempStorage := sourceStorage[tableName]
					for k := range sourceStorage[tableName] {
						if strings.Compare(k, "DB_0015") >= 0 && strings.Compare(k, "DB_0020") <= 0 {
							delete(tempStorage, k)
						}
					}
					sourceStorage[tableName] = tempStorage

					convey.Convey("storage should still have data", func() {
						e = k.GetRange(tableName+":"+"DB_0015", tableName+":"+"DB_0020", &resDatas, true)
						convey.So(e, convey.ShouldBeNil)
						convey.So(len(resDatas), convey.ShouldEqual, 6)
					})
				})
			})
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
		myGetter,
		mySetter,
		&kiva.KivaOptions{
			DefaultWrite: kiva.WriteOptions{
				TTL: 15 * time.Minute,
			},
			SyncBatch: kiva.SyncBatchOptions{},
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

func myGetter(key1, key2 string, kind kiva.GetKind, dest interface{}) error {
	tableName, keyFind, err := kiva.ParseKey(key1)
	if err != nil {
		return err
	}

	_, keyTo, err := kiva.ParseKey(key2)
	if kind == kiva.GetRange && err != nil {
		return err
	}

	storage, ok := sourceStorage[tableName]
	if !ok {
		return errors.New("storage not exists for table " + tableName)
	}

	var (
		item interface{}
	)
	single := false
	items := []interface{}{}
	switch kind {
	case kiva.GetByID:
		data, ok := storage[keyFind]
		if !ok {
			return io.EOF
		}
		single = true
		item = data

	case kiva.GetByPattern:
		for k, v := range storage {
			if strings.HasPrefix(k, keyFind) {
				items = append(items, v)
			}
		}

	case kiva.GetRange:
		for k, v := range storage {
			if strings.Compare(k, key1) >= 0 && strings.Compare(k, keyTo) <= 0 {
				items = append(items, v)
			}
		}
	}

	var e error
	if single {
		e = serde.Serde(item.(map[string]interface{})["Value"], dest)
	} else {
		e = serde.Serde(items, items)
	}
	return e
}

func mySetter(key string, value interface{}, op kiva.CommitKind) error {
	tableName, keyFind, err := kiva.ParseKey(key)
	if err != nil {
		return err
	}
	storage, ok := sourceStorage[tableName]
	if !ok {
		return errors.New("storage not exist for " + tableName)
	}

	switch op {
	case kiva.CommitSave:
		storage[keyFind] = map[string]interface{}{"_id": keyFind, "Value": value}
		sourceStorage[tableName] = storage

	case kiva.CommitDelete:
		delete(storage, keyFind)
	}
	return nil
}
