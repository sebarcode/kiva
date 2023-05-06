package kiva

import "context"

type Provider interface {
	Connect() error
	Close()
	Set(key string, value interface{}, opts *WriteOptions) error
	Get(key string, dest interface{}) (*ItemOptions, error)
	Delete(key string)
	HasKey(string) bool
	Keys(pattern string) []string
	KeyRanges(from, to string) []string
	SetContext(ctx context.Context)
	Context() context.Context
	ChangeSyncOpts(key string, opts *ItemOptions) error
}
