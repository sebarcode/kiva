package kiva

import (
	"errors"
	"io"
)

type Kv struct {
	mem     MemoryProvider
	storage StorageProvider
}

func New(mem MemoryProvider, storage StorageProvider) *Kv {
	kv := new(Kv)
	kv.mem = mem
	kv.storage = storage
	return kv
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
	opts, err := kv.mem.Get(id, dest)
	if err != nil {
		if err != io.EOF {
			return err
		}
		if kv.storage == nil {
			return errors.New("missing: storage")
		}
		if err := kv.storage.Get(id, dest); err != nil {
			return err
		}
		kv.mem.Set(id, dest, opts)
	}
	return nil
}

func (kv *Kv) Set(id string, value interface{}) error {
	if kv.mem == nil {
		return errors.New("missing: memory storage")
	}
	return kv.mem.Set(id, value, nil)
}
