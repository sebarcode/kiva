package kiva

import (
	"time"
)

type SyncKindEnum string

const (
	SyncNow   SyncKindEnum = "NOW"
	SyncBatch SyncKindEnum = "BATCH"
)

type WriteOptions struct {
	TTL        time.Duration
	SyncKind   SyncKindEnum
	ExpiryKind ExpiryKindEnum
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
