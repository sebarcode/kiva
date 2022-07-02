package kiva

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
)

func BaseGetter(h *datahub.Hub, field string) GetterFunc {
	return func(key1, key2 string, kind GetKind, res interface{}) error {
		tableName, keyFind, err := ParseKey(key1)
		if err != nil {
			return err
		}

		_, keyTo, err := ParseKey(key2)
		if kind == GetRange && err != nil {
			return err
		}

		var f *dbflex.Filter
		single := false
		switch kind {
		case GetByID:
			f = dbflex.Eq(field, keyFind)
			single = true

		case GetByPattern:
			f = dbflex.StartWith(field, keyFind)

		case GetRange:
			f = dbflex.Range(field, keyFind, keyTo)
		}

		isPrimitive := codekit.HasMember([]reflect.Kind{reflect.String, reflect.Int, reflect.Float32, reflect.Float64},
			reflect.Indirect(reflect.ValueOf(res)).Kind())

		if isPrimitive {
			ms := []codekit.M{}
			if e := h.PopulateByFilter(tableName, f, 0, &ms); e != nil {
				return e
			}
			if len(ms) == 0 {
				return io.EOF
			}
			*(res.(*int)) = ms[0].GetInt("Value")
		} else {
			var e error
			if single {
				e = h.GetAnyByFilter(tableName, f, res)
			} else {
				e = h.PopulateByFilter(tableName, f, 0, res)
			}
			return e
		}
		return nil
	}
}

func BaseCommitter(h *datahub.Hub, field string) CommitFunc {
	return func(key string, value interface{}, op CommitKind) error {
		tableName, keyFind, err := ParseKey(key)
		if err != nil {
			return err
		}

		isStruct := reflect.Indirect(reflect.ValueOf(value)).Kind() == reflect.Struct
		w := dbflex.Eq(field, keyFind)
		switch op {
		case CommitSave:
			if isStruct {
				return h.SaveAny(tableName, value)
			} else {
				return h.SaveAny(tableName, codekit.M{}.Set(field, keyFind).Set("Value", value))
			}

		case CommitDelete:
			cmd := dbflex.From(tableName).Where(w).Delete()
			_, e := h.Execute(cmd, nil)
			return e
		}
		return nil
	}
}

func ParseKey(txt string) (table, key string, err error) {
	keys := strings.Split(txt, ":")
	if len(keys) != 2 {
		err = fmt.Errorf("txt pattern should be table:key")
		return
	}
	if keys[0] == "" {
		err = errors.New("table can't be blank")
		return
	}
	if keys[1] == "" {
		err = errors.New("key  can't be blank")
		return
	}
	table = keys[0]
	key = keys[1]
	return
}
