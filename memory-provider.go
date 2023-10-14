package kiva

type MemoryProvider interface {
	Get(table, id string, dest interface{}) (*ItemOptions, error)
	Set(table, id string, value interface{}, opts *ItemOptions) error
	Len(table string) int
	Connect() error
	Close()
}
