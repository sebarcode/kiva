package kiva

type MemoryProvider interface {
	Get(table, id string, dest interface{}) (*ItemMetadata, error)
	Set(table, id string, value interface{}) error
	Delete(table, id string) error
	Len(table string) int
	Keys(table string) []string
	Connect() error
	SetCacheOptions(table string, opt CacheOptions)
	Close()
}
