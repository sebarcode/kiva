package kiva

import (
	"context"
	"io"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/sebarcode/codekit"
)

type simpleProvideItem struct {
	Data   interface{}
	Expiry time.Time
}

type SimpleProvider struct {
	defaultWriteOptions *WriteOptions
	keys                []string
	data                map[string]simpleProvideItem

	ctx    context.Context
	cancel context.CancelFunc
	mtx    *sync.RWMutex
}

func NewSimpleProvider(opts *WriteOptions) Provider {
	s := new(SimpleProvider)
	s.defaultWriteOptions = opts
	s.data = make(map[string]simpleProvideItem)

	ctx, cf := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancel = cf
	s.mtx = new(sync.RWMutex)

	if opts == nil {
		opts = new(WriteOptions)
	}
	if opts.TTL == 0 {
		opts.TTL = 24 * time.Hour
	}

	go s.DataCleansing()
	return s
}

func (p *SimpleProvider) Connect() error {
	return nil
}

func (p *SimpleProvider) Close() {
	if p.ctx != nil {
		p.cancel()
		p.ctx = nil
	}
}

func (p *SimpleProvider) DataCleansing() {
	for {
		select {
		case <-time.After(1 * time.Second):
			p.mtx.Lock()
			func() {
				defer p.mtx.Unlock()

				removedKeys := []string{}
				for k, i := range p.data {
					if i.Expiry.After(time.Now()) {
						removedKeys = append(removedKeys, k)
					}
				}

				for _, key := range removedKeys {
					delete(p.data, key)
				}

				newKeys := []string{}
				for _, key := range p.keys {
					if !codekit.HasMember(removedKeys, key) {
						newKeys = append(newKeys, key)
					}
				}
				p.keys = newKeys
			}()

		case <-p.ctx.Done():
			return
		}
	}
}

func (p *SimpleProvider) Set(key string, value interface{}, opts *WriteOptions) error {
	if codekit.IsPointer(value) {
		value = reflect.Indirect(reflect.ValueOf(value)).Interface()
	}
	item := simpleProvideItem{
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

func (p *SimpleProvider) Get(key string, dest interface{}) error {
	v, ok := p.data[key]
	if ok {
		reflect.ValueOf(dest).Elem().Set(reflect.ValueOf(v.Data))
		if p.defaultWriteOptions.TTL != 0 && time.Until(v.Expiry) < 5*time.Second {
			v.Expiry = time.Now().Add(p.defaultWriteOptions.TTL)
			p.data[key] = v
		}
		return nil
	}
	return io.EOF
}

func (p *SimpleProvider) Delete(key string) {
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
