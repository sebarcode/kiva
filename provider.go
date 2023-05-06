package kiva

import "context"

type Provider interface {
	Connect() error
	Close()
	Set(key string, value interface{}, opts *WriteOptions) error
	Get(key string) *Item
	Delete(key string)
	Keys(pattern string) []string
	KeyRanges(from, to string) []string
	SetContext(ctx context.Context)
	Context() context.Context
}
