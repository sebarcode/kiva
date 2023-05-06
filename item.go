package kiva

import (
	"time"
)

type ExpiryKindEnum string
type SyncDirectionEnum string

const (
	ExpiryAbsolute ExpiryKindEnum = "ABSOLUTE"
	ExpiryExtended ExpiryKindEnum = "EXTENDED"

	SyncToPersistent SyncDirectionEnum = "UPDATE_PERSISTENT"
	SyncToHots       SyncDirectionEnum = "UPDATE_HOT_STORAGE"
)

type ItemOptions struct {
	Expiry               time.Time
	SyncDirection        SyncDirectionEnum
	ExpiryKind           ExpiryKindEnum
	ExpiryExtendDuration time.Duration
	SyncKind             SyncKindEnum
	SyncEveryInSecond    int
	LastSync             time.Time
}
