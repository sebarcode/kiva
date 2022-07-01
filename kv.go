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
	Delete(key string)
	Keys(pattern string) []string
	KeyRanges(from, to string) []string
}

type GetKind string
type CommitKind string

const (
	GetByID      GetKind = "eq"
	GetByPattern GetKind = "pattern"
	GetRange     GetKind = "between"
)

const (
	CommitSave   CommitKind = "save"
	CommitDelete CommitKind = "delete"
)
