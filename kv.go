package kiva

import "time"

type KvWriteOptions struct {
	TTL time.Duration
}

type KvProvider interface {
	Set(key string, value interface{}, opts KvWriteOptions) error
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
