package kiva

type MemoryProvider interface {
	Get(id string, dest interface{}) (*ItemOptions, error)
	Set(id string, value interface{}, opts *ItemOptions) error
	Connect() error
	Close()
}
