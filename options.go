package kiva

import "time"

type ItemMetadata struct {
	Created  time.Time
	LastUsed time.Time
}

func (i *ItemMetadata) IsExpired(opt CacheOptions) bool {
	switch opt.ExpiryBy {
	case ExpiryByLastUsed:
		return time.Since(i.LastUsed) >= opt.ExpiryPeriod

	case ExpiryByCreated:
		return time.Since(i.Created) >= opt.ExpiryPeriod

	default:
		return false
	}
}

type ExpiryByType string

const (
	ExpiryByCreated  ExpiryByType = "ByCreatedTime"
	ExpiryByLastUsed ExpiryByType = "ByLastUsed"
)

type CacheOptions struct {
	Size         int
	ExpiryBy     ExpiryByType
	ExpiryPeriod time.Duration
	SyncEvery    time.Duration
}
