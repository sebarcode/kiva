package simplestorage

import (
	"fmt"
	"io"
	"sync"

	"github.com/ariefdarmawan/serde"
	"github.com/sebarcode/codekit"
)

type collection struct {
	mtx   *sync.RWMutex
	items codekit.M
}

func (c *collection) get(id string, dest interface{}) error {
	c.mtx.RLock()
	defer c.mtx.RUnlock()

	d, ok := c.items[id]
	if !ok {
		return io.EOF
	}

	if err := serde.Serde(d, dest); err != nil {
		return fmt.Errorf("serde: %s", err.Error())
	}

	return nil
}

func (c *collection) Len() int {
	return len(c.items.Keys())
}

type storage struct {
	collections map[string]*collection
}

func NewStorage() *storage {
	s := new(storage)
	s.collections = map[string]*collection{}
	return s
}

func (s *storage) Connect() error {
	return nil
}

func (s *storage) Close() {
}

func (s *storage) Get(table, id string, dest interface{}) error {
	c, ok := s.collections[table]
	if !ok {
		return io.EOF
	}
	return c.get(id, dest)
}

func (s *storage) Set(table, id string, value interface{}) error {
	c, ok := s.collections[table]
	if !ok {
		c = &collection{
			items: codekit.M{},
			mtx:   new(sync.RWMutex),
		}
		s.collections[table] = c
	}
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.items.Set(id, value)
	return nil
}

func (s *storage) Len(table string) int {
	c, ok := s.collections[table]
	if !ok {
		return 0
	}
	return c.Len()
}
