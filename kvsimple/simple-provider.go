package kvsimple

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/sebarcode/codekit"
	"github.com/sebarcode/kiva"
)

type SimpleProvider struct {
	defaultWriteOptions *kiva.WriteOptions
	keys                []string
	data                map[string]*kiva.Item

	mtx *sync.RWMutex
	ctx context.Context
}

func New(opts *kiva.WriteOptions) kiva.Provider {
	s := new(SimpleProvider)
	s.defaultWriteOptions = opts
	s.data = make(map[string]*kiva.Item)

	s.mtx = new(sync.RWMutex)

	if opts == nil {
		opts = new(kiva.WriteOptions)
	}
	if opts.TTL == 0 {
		opts.TTL = 24 * time.Hour
	}

	return s
}

func (p *SimpleProvider) Connect() error {
	return nil
}

func (p *SimpleProvider) Close() {
}

func (p *SimpleProvider) Context() context.Context {
	if p.ctx == nil {
		p.ctx = context.Background()
		return p.ctx
	}
	return p.ctx
}

func (p *SimpleProvider) SetContext(ctx context.Context) {
	p.ctx = ctx
}

func (p *SimpleProvider) Set(key string, value interface{}, opts *kiva.WriteOptions) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if codekit.IsPointer(value) {
		value = reflect.Indirect(reflect.ValueOf(value)).Interface()
	}
	item := &kiva.Item{
		Data: value,
	}
	if opts == nil {
		opts = p.defaultWriteOptions
	}
	item.Expiry = time.Now().Add(opts.TTL)
	p.data[key] = item

	found := false
	strCompare := -2
	cutOffIndex := -1
	for index, simpleKey := range p.keys {
		strCompare = strings.Compare(key, simpleKey)
		if strCompare < 0 {
			cutOffIndex = index
			found = true
			break
		} else if strCompare == 0 {
			break
		}
	}
	if found {
		var newKeys []string
		if cutOffIndex > 0 {
			newKeys = p.keys[:cutOffIndex]
		}
		newKeys = append(newKeys, key)
		if cutOffIndex < len(p.keys)-1 {
			newKeys = append(newKeys, p.keys[cutOffIndex:]...)
		}
		p.keys = newKeys
	} else {
		p.keys = append(p.keys, key)
	}
	return nil
}

func (p *SimpleProvider) Get(key string) *kiva.Item {
	v, ok := p.data[key]
	if ok {
		return v
	}
	return &kiva.Item{
		Error: errors.New("key is not exists"),
	}
}

func (p *SimpleProvider) Delete(key string) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	delete(p.data, key)
	keys := []string{}
	for _, k := range p.keys {
		if k != key {
			keys = append(keys, k)
		}
	}
	p.keys = keys
}

func (p *SimpleProvider) Keys(pattern string) []string {
	return p.keys
}

func (p *SimpleProvider) KeyRanges(from string, to string) []string {
	inRangeKeys := []string{}
	for _, key := range p.keys {
		if strings.Compare(key, from) >= 0 && strings.Compare(key, to) <= 0 {
			inRangeKeys = append(inRangeKeys, key)
		}
	}
	return inRangeKeys
}
