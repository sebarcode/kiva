package kiva

import (
	"fmt"
	"io"
)

type GetterFunc func(key1, key2 string, op GetKind, dest interface{}) error

type Kiva struct {
	provider KvProvider
	getter   GetterFunc
}

func New(provider KvProvider, getter GetterFunc) (*Kiva, error)

func (k *Kiva) Get(key string, dest interface{}) error {
	if e := k.provider.Get(key, dest); e != nil {
		if e != io.EOF {
			return fmt.Errorf("kv get error. %s", e.Error())
		}
		if e = k.getter(key, "", GetByID, dest); e != nil {
			return fmt.Errorf("kv getter error. %s", e.Error())
		}
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
