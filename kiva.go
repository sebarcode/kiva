package kiva

import (
	"errors"
	"fmt"
	"io"
)

type GetterFunc func(key1, key2 string, op GetKind, dest interface{}) error

type Kiva struct {
	provider Provider
	getter   GetterFunc

	defaultWriteOptions WriteOptions
}

func New(provider Provider, getter GetterFunc, opts WriteOptions) (*Kiva, error) {
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
		if e = k.getter(key, "", GetByID, dest); e != nil {
			return fmt.Errorf("kv getter error. %s", e.Error())
		}
		k.provider.Set(key, dest, k.defaultWriteOptions)
	}
	return nil
}

func (k *Kiva) GetByPrefix(prefix string, dest interface{}) error {
	if e := k.provider.GetByPrefix(prefix, dest); e != nil {
		if e != io.EOF {
			return fmt.Errorf("kv get error. %s", e.Error())
		}
		if e = k.getter(prefix, "", GetByPrefix, dest); e != nil {
			return fmt.Errorf("kv getter error. %s", e.Error())
		}
	}
	return nil
}

func (k *Kiva) GetRange(prefix string, dest interface{}) error {
	if e := k.provider.GetByPrefix(prefix, dest); e != nil {
		return fmt.Errorf("kv get error. %s", e.Error())
	}
	return nil
}
