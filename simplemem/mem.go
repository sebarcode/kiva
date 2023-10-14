package simplemem

import (
	"io"
	"sync"

	"github.com/ariefdarmawan/serde"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/kiva"
)

type collection struct {
	mtx   *sync.RWMutex
	items codekit.M
}

type Memory struct {
	collections map[string]*collection
}

func NewMemory() *Memory {
	mem := new(Memory)
	mem.collections = map[string]*collection{}
	return mem
}

func (mem *Memory) findCreateCollection(table string) *collection {
	c, ok := mem.collections[table]
	if !ok {
		c = &collection{
			mtx:   new(sync.RWMutex),
			items: codekit.M{},
		}
		mem.collections[table] = c
	}
	return c
}

func (mem *Memory) Get(table, id string, dest interface{}) (*kiva.ItemOptions, error) {
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
	return new(kiva.ItemOptions), nil
}

func (mem *Memory) Set(table, id string, value interface{}, opts *kiva.ItemOptions) error {
	c := mem.findCreateCollection(table)
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.items.Set(id, value)
	return nil
}

func (mem *Memory) Connect() error {
	return nil
}

func (mem *Memory) Close() {
}

func (mem *Memory) Len(table string) int {
	c := mem.findCreateCollection(table)
	return len(c.items.Keys())
}
