package kiva

import (
	"io"
	"reflect"
	"strings"
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
}

func NewSimpleProvider(opts *WriteOptions) Provider {
	s := new(SimpleProvider)
	s.defaultWriteOptions = opts
	s.data = make(map[string]simpleProvideItem)
	return s
}

func (p *SimpleProvider) Connect() error {
	return nil
}

func (p *SimpleProvider) Close() {
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
	if opts.TTL != time.Duration(0) {
		item.Expiry = time.Now().Add(opts.TTL)
	}
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
