package kiva

import "time"

type WriteOptions struct {
	TTL         time.Duration
	MaxMemory   int
	MaxItemSize int
}

type Provider interface {
	Connect() error
	Close()
	Set(key string, value interface{}, opts *WriteOptions) error
	Get(key string, dest interface{}) error
	GetByPattern(pattern string, dest interface{}) error
	GetRange(keyFrom, keyTo string, dest interface{}) error
	Delete(key string)
	DeleteByPattern(pattern string)
	DeleteRange(keyFrom, keyTo string)
	//Keys(pattern string) []string
}

type GetKind string

const (
	GetByID      GetKind = "eq"
	GetByPattern GetKind = "startsWith"
	GetRange     GetKind = "between"
)
