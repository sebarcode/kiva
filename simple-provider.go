package kiva

import (
	"strings"
	"time"
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
	return
}

func (p *SimpleProvider) Set(key string, value interface{}, opts *WriteOptions) error {
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
	}
	return nil
}

func (p *SimpleProvider) Get(key string, dest interface{}) error {
	panic("not implemented") // TODO: Implement
}

func (p *SimpleProvider) Delete(key string) {
	panic("not implemented") // TODO: Implement
}

func (p *SimpleProvider) Keys(pattern string) []string {
	panic("not implemented") // TODO: Implement
}

func (p *SimpleProvider) KeyRanges(from string, to string) []string {
	panic("not implemented") // TODO: Implement
}
