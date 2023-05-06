package kiva

import (
	"time"
)

type SyncKind string

const (
	SyncNow   SyncKind = "NOW"
	SyncBatch SyncKind = "BATCH"
)

type WriteOptions struct {
	TTL         time.Duration
	MaxMemory   int
	MaxItemSize int
	SyncKind    SyncKind
}

type SyncBatchOptions struct {
	EveryInSecond       int
	SyncTimeoutInSecond int
}

type KivaOptions struct {
	DefaultWrite WriteOptions
	SyncBatch    SyncBatchOptions
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
