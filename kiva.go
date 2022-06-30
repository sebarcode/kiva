package kiva

import (
	"errors"
	"fmt"
	"io"
	"reflect"
)

type GetterFunc func(key1, key2 string, op GetKind, dest interface{}) error

type Kiva struct {
	provider Provider
	getter   GetterFunc

	defaultWriteOptions *WriteOptions
}

func New(provider Provider, getter GetterFunc, opts *WriteOptions) (*Kiva, error) {
	if e := provider.Connect(); e != nil {
		return nil, errors.New("unable to connect to provider. " + e.Error())
	}

	k := new(Kiva)
	k.provider = provider
	k.getter = getter
	k.defaultWriteOptions = opts

	return k, nil
}

func (k *Kiva) Get(key string, dest interface{}) error {
	if e := k.provider.Get(key, dest); e != nil {
		if e != io.EOF {
			return fmt.Errorf("kv get error. %s", e.Error())
		}
		if k.getter == nil {
			return fmt.Errorf("kv get error. %s", io.EOF.Error())
		}
		if e = k.getter(key, "", GetByID, dest); e != nil {
			return fmt.Errorf("kv getter error. %s", e.Error())
		}

		destValue := reflect.Indirect(reflect.ValueOf(dest)).Interface()
		if e = k.provider.Set(key, destValue, k.defaultWriteOptions); e != nil {
			return fmt.Errorf("kv setter error. %s", e.Error())
		}
	}
	return nil
}

func (k *Kiva) GetByPattern(pattern string, dest interface{}) error {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("output should be ptr of slice")
	}

	if rv.Type().Elem().Kind() != reflect.Slice {
		return fmt.Errorf("output should be ptr of slice")
	}

	if e := k.provider.GetByPattern(pattern, dest); e != nil {
		if e != io.EOF {
			return fmt.Errorf("kv get error. %s", e.Error())
		}
		if e = k.getter(pattern, "", GetByPattern, dest); e != nil {
			return fmt.Errorf("kv getter error. %s", e.Error())
		}
	}
	return nil
}

func (k *Kiva) GetRange(from, to string, dest interface{}) error {
	if e := k.provider.GetRange(from, to, dest); e != nil {
		return fmt.Errorf("kv get error. %s", e.Error())
	}
	return nil
}

func (k *Kiva) Set(key string, value interface{}, opts *WriteOptions) error {
	if opts == nil {
		opts = k.defaultWriteOptions
	}
	return k.provider.Set(key, value, opts)
}
