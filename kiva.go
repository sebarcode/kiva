package kiva

import (
	"errors"
	"io"
	"strings"
)

type Kv struct {
	mem     MemoryProvider
	storage StorageProvider
	group   string
}

func New(group string, mem MemoryProvider, storage StorageProvider) *Kv {
	kv := new(Kv)
	kv.mem = mem
	kv.storage = storage
	kv.group = group
	return kv
}

func (kv *Kv) parseKey(key string) (string, string, error) {
	keys := strings.Split(key, ":")
	if len(keys) < 3 {
		return "", "", errors.New("invalid: key: should be 3 segments")
	}
	if keys[0] != kv.group {
		return "", "", errors.New("invalid: group")
	}
	return keys[1], keys[2], nil
}

func (kv *Kv) makeKeys(table, id string) string {
	return strings.Join([]string{kv.group, table, id}, ":")
}

func (kv *Kv) Close() {
	if kv.mem != nil {
		kv.mem.Close()
	}

	if kv.storage != nil {
		kv.storage.Close()
	}
}

func (kv *Kv) Get(id string, dest interface{}) error {
	if kv.mem == nil {
		return errors.New("missing: memory storage")
	}
	table, id, err := kv.parseKey(id)
	if err != nil {
		return err
	}
	opts, err := kv.mem.Get(table, id, dest)
	if err != nil {
		if err != io.EOF {
			return err
		}
		if kv.storage == nil {
			return errors.New("missing: storage")
		}
		if err := kv.storage.Get(table, id, dest); err != nil {
			return err
		}
		kv.mem.Set(table, id, dest, opts)
	}
	return nil
}

func (kv *Kv) Set(id string, value interface{}) error {
	table, id, err := kv.parseKey(id)
	if err != nil {
		return err
	}
	if kv.mem == nil {
		return errors.New("missing: memory storage")
	}
	return kv.mem.Set(table, id, value, nil)
}
