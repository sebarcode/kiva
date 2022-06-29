package kiva

import "time"

type WriteOptions struct {
	TTL time.Duration
}

type Provider interface {
	Connect() error
	Close()
	Set(key string, value interface{}, opts WriteOptions) error
	Get(key string, dest interface{}) error
	GetByPrefix(prefix string, dest interface{}) error
	GetRange(keyFrom, keyTo string, dest interface{}) error
	Delete(key string)
	DeleteByPrefix(prefix string)
	DeleteRange(keyFrom, keyTo string)
}

type GetKind string

const (
	GetByID     GetKind = "eq"
	GetByPrefix GetKind = "startsWith"
	GetRange    GetKind = "between"
)
