package kiva

type MemoryProvider interface {
	Get(table, id string, dest interface{}) (*ItemOptions, error)
	Set(table, id string, value interface{}, opts *ItemOptions) error
	Connect() error
	Close()
}
