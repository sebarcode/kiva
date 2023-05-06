package kvsimple

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/ariefdarmawan/serde"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/kiva"
)

type providerItem struct {
	data interface{}
	opts *kiva.ItemOptions
}

type SimpleProvider struct {
	defaultWriteOptions *kiva.WriteOptions
	keys                []string
	data                map[string]*providerItem

	mtx *sync.RWMutex
	ctx context.Context
}

func New() kiva.Provider {
	s := new(SimpleProvider)
	s.data = make(map[string]*providerItem)
	s.keys = []string{}
	s.mtx = new(sync.RWMutex)
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
	if opts == nil {
		opts = p.defaultWriteOptions
	}

	item := providerItem{
		data: value,
		opts: &kiva.ItemOptions{
			Expiry:               time.Now().Add(opts.TTL),
			SyncDirection:        kiva.SyncToPersistent,
			ExpiryKind:           opts.ExpiryKind,
			ExpiryExtendDuration: opts.TTL,
			SyncKind:             opts.SyncKind,
			SyncEveryInSecond:    opts.SyncEveryInSecond,
			LastSync:             time.Now(),
		},
	}

	p.data[key] = &item

	found := false
	alreadyExist := false
	strCompare := -2
	cutOffIndex := -1
	for index, simpleKey := range p.keys {
		strCompare = strings.Compare(key, simpleKey)
		if strCompare < 0 {
			cutOffIndex = index
			found = true
			break
		} else if strCompare == 0 {
			alreadyExist = true
			break
		}
	}
	if alreadyExist {
		return nil
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

func (p *SimpleProvider) Get(key string, dest interface{}) (*kiva.ItemOptions, error) {
	v, ok := p.data[key]
	if !ok {
		return nil, io.EOF
	}
	e := serde.Serde(v.data, dest)
	if e != nil {
		return nil, fmt.Errorf("cast: %s", e.Error())
	}
	return v.opts, nil
}

func (p *SimpleProvider) HasKey(key string) bool {
	_, ok := p.data[key]
	return ok
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
	keys := []string{}
	if pattern == "*" {
		return p.keys
	}
	pattern = strings.TrimSuffix(pattern, "*")
	for _, k := range p.keys {
		if strings.HasPrefix(k, pattern) {
			keys = append(keys, k)
		}
	}
	return keys
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

func (p *SimpleProvider) ChangeSyncOpts(key string, opts *kiva.ItemOptions) error {
	item, hasItem := p.data[key]
	if !hasItem {
		return errors.New("ket not found")
	}

	item.opts.SyncDirection = opts.SyncDirection
	item.opts.SyncKind = opts.SyncKind
	item.opts.SyncEveryInSecond = opts.SyncEveryInSecond
	p.data[key] = item

	return nil
}

func (p *SimpleProvider) RenewExpiry(key string) error {
	item, hasItem := p.data[key]
	if !hasItem {
		return errors.New("ket not found")
	}

	item.opts.Expiry = time.Now().Add(item.opts.ExpiryExtendDuration)
	p.data[key] = item

	return nil
}

func (p *SimpleProvider) UpdateLastSyncTime(key string) error {
	item, hasItem := p.data[key]
	if !hasItem {
		return errors.New("ket not found")
	}

	item.opts.LastSync = time.Now()
	item.opts.SyncDirection = kiva.SyncToHots
	p.data[key] = item

	return nil
}

func (p *SimpleProvider) ItemOpts(key string) *kiva.ItemOptions {
	item, hasItem := p.data[key]
	if !hasItem {
		return nil
	}

	return item.opts
}
