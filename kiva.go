package kiva

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"
)

type GetterFunc func(key1, key2 string, op GetKind, dest interface{}) error
type CommitFunc func(key1 string, value interface{}, op CommitKind) error
type ItemReflectorFunc func(tablename string) interface{}

type Kiva struct {
	provider  Provider
	commiter  CommitFunc
	getter    GetterFunc
	reflector ItemReflectorFunc

	opts *KivaOptions

	ctx context.Context
}

func New(provider Provider, reflector ItemReflectorFunc, getter GetterFunc, committer CommitFunc, opts *KivaOptions) (*Kiva, error) {
	if e := provider.Connect(); e != nil {
		return nil, errors.New("unable to connect to provider. " + e.Error())
	}

	k := new(Kiva)
	k.ctx = context.Background()
	k.provider = provider
	k.getter = getter
	k.commiter = committer
	k.reflector = reflector
	k.opts = opts

	k.provider.SetContext(k.ctx)

	if k.opts.SyncBatch.EveryInSecond > 0 {
		go k.Sync()
	}

	return k, nil
}

func (k *Kiva) Get(key string, dest interface{}) error {
	opts, e := k.provider.Get(key, dest)
	if e != nil {
		if k.getter == nil {
			return fmt.Errorf("kv getter: invalid getter")
		}
		if e = k.getter(key, "", GetByID, dest); e != nil {
			return fmt.Errorf("kv getter: %s", e.Error())
		}
		if e != nil {
			return fmt.Errorf("kv getter: %s", e.Error())
		}

		destValue := reflect.Indirect(reflect.ValueOf(dest)).Interface()
		if e = k.provider.Set(key, destValue, &k.opts.DefaultWrite); e != nil {
			return fmt.Errorf("kv setter: %s", e.Error())
		}
		opts = &ItemOptions{
			Expiry:        time.Now().Add(k.opts.DefaultWrite.TTL),
			SyncDirection: SyncToHots,
			ExpiryKind:    k.opts.DefaultWrite.ExpiryKind,
			SyncKind:      k.opts.DefaultWrite.SyncKind,
		}
	}
	if opts.ExpiryKind == ExpiryExtended {
		opts.Expiry = opts.Expiry.Add(opts.ExpiryExtendDuration)
	}
	if opts.Expiry.Before(time.Now()) {
		k.provider.Delete(key)
		return errors.New("item is expired")
	}
	return nil
}

func (k *Kiva) GetByPattern(pattern string, dest interface{}, runGetterIfEmpty bool) error {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("output should be ptr of slice")
	}

	if rv.Type().Elem().Kind() != reflect.Slice {
		return fmt.Errorf("output should be ptr of slice")
	}

	keys := k.Keys(pattern)
	if e := k.getByKeys(dest, keys...); e != nil {
		return fmt.Errorf("getter error: %s", e.Error())
	}

	if runGetterIfEmpty {
		destLen := rv.Elem().Len()
		if destLen == 0 && k.getter != nil {
			if e := k.getter(pattern, "", GetByPattern, dest); e != nil {
				return fmt.Errorf("getter error: %s", e.Error())
			}
		}
	}
	return nil
}

func (k *Kiva) GetRange(from, to string, dest interface{}, runGetterIfEmpty bool) error {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("output should be ptr of slice")
	}

	if rv.Type().Elem().Kind() != reflect.Slice {
		return fmt.Errorf("output should be ptr of slice")
	}

	keys := k.KeyRanges(from, to)
	if e := k.getByKeys(dest, keys...); e != nil {
		return fmt.Errorf("getter error: %s", e.Error())
	}

	if runGetterIfEmpty {
		destLen := rv.Elem().Len()
		if destLen == 0 && k.getter != nil {
			if e := k.getter(from, to, GetRange, dest); e != nil {
				return fmt.Errorf("getter error: %s", e.Error())
			}
		}
	}
	return nil
}

func (k *Kiva) getByKeys(dest interface{}, keys ...string) error {
	rtSlice := reflect.TypeOf(dest).Elem()
	rtElem := rtSlice.Elem()

	buffers := reflect.MakeSlice(rtSlice, len(keys), len(keys))
	for i, key := range keys {
		newElem := reflect.New(rtElem).Interface()
		var (
			err error
		)
		if _, err = k.provider.Get(key, newElem); err != nil {
			return fmt.Errorf("read data erorr. key %s. %s", key, err.Error())
		}
		buffers.Index(i).Set(reflect.ValueOf(newElem).Elem())
	}
	reflect.ValueOf(dest).Elem().Set(buffers)
	return nil
}

func (k *Kiva) Set(key string, value interface{}, opts *WriteOptions, syncToDB bool) error {
	if opts == nil {
		opts = &k.opts.DefaultWrite
	}
	if e := k.provider.Set(key, value, opts); e != nil {
		return e
	}
	if (syncToDB && opts.SyncKind == SyncNow) && k.commiter != nil {
		if e := k.commiter(key, value, CommitSave); e != nil {
			return fmt.Errorf("commit error. %s", e.Error())
		}
		k.provider.UpdateLastSyncTime(key)
	}
	return nil
}

func (k *Kiva) Delete(syncToDB bool, keys ...string) {
	for _, key := range keys {
		k.provider.Delete(key)
		if syncToDB && k.commiter != nil {
			k.commiter(key, nil, CommitDelete)
		}
	}
}

func (k *Kiva) DeleteRange(from, to string, syncToDB bool) {
	keys := k.KeyRanges(from, to)
	k.Delete(syncToDB, keys...)
}

func (k *Kiva) DeleteByPattern(pattern string, syncToDB bool) {
	keys := k.Keys(pattern)
	k.Delete(syncToDB, keys...)
}

func (k *Kiva) Keys(pattern string) []string {
	return k.provider.Keys(pattern)
}

func (k *Kiva) KeyRanges(from, to string) []string {
	return k.provider.KeyRanges(from, to)
}
