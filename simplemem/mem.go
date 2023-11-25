package simplemem

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ariefdarmawan/serde"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/kiva"
)

type Memory struct {
	collections map[string]*collection
	cancelFn    context.CancelFunc
}

func NewMemory() *Memory {
	mem := new(Memory)
	mem.collections = map[string]*collection{}
	return mem
}

func (mem *Memory) SetCacheOptions(table string, opts kiva.CacheOptions) {
	if table == "" {
		return
	}

	c := mem.findCreateCollection(table)
	c.setOpts(opts)
}

func (mem *Memory) findCreateCollection(table string) *collection {
	c, ok := mem.collections[table]
	if !ok {
		c = &collection{
			mtx:   new(sync.RWMutex),
			items: codekit.M{},
			metas: map[string]kiva.ItemMetadata{},
		}
		mem.collections[table] = c
	}
	return c
}

func (mem *Memory) Get(table, id string, dest interface{}) (*kiva.ItemMetadata, error) {
	c := mem.findCreateCollection(table)
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	d, ok := c.items[id]
	if !ok {
		return nil, io.EOF
	}
	if err := serde.Serde(d, dest); err != nil {
		return nil, err
	}
	meta, ok := c.metas[id]
	if ok {
		meta.LastUsed = time.Now()
	} else {
		meta = kiva.ItemMetadata{
			Created:  time.Now(),
			LastUsed: time.Now(),
		}
	}
	c.metas[id] = meta
	return &meta, nil
}

func (mem *Memory) Set(table, id string, value interface{}) error {
	c := mem.findCreateCollection(table)
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.items.Set(id, value)
	meta, hasMeta := c.metas[id]
	if hasMeta {
		meta.LastUsed = time.Now()
	} else {
		c.keys = append(c.keys, id)
		meta = kiva.ItemMetadata{
			Created:  time.Now(),
			LastUsed: time.Now(),
		}
	}
	c.metas[id] = meta
	return nil
}

func (mem *Memory) Delete(table, id string) error {
	c, ok := mem.collections[table]
	if !ok {
		return fmt.Errorf("invalid collection: %s", table)
	}

	c.mtx.Lock()
	defer c.mtx.Unlock()
	delete(c.items, id)
	delete(c.metas, id)
	newKeys := []string{}
	for _, key := range c.keys {
		if key != id {
			newKeys = append(newKeys, key)
		}
	}
	c.keys = newKeys
	return nil
}

func (mem *Memory) Connect() error {
	return nil
}

func (mem *Memory) Close() {
	for _, c := range mem.collections {
		c.Close()
	}

	if mem.cancelFn != nil {
		mem.cancelFn()
	}
}

func (mem *Memory) Len(table string) int {
	c, ok := mem.collections[table]
	if !ok {
		return 0
	}
	return len(c.keys)
}

func (mem *Memory) Keys(table string) []string {
	c, ok := mem.collections[table]
	if !ok {
		return []string{}
	}
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.keys
}
