package kiva

type MemoryProvider interface {
	Get(table, id string, dest interface{}) (*ItemOptions, error)
	Set(table, id string, value interface{}, opts *ItemOptions) error
	Delete(table, id string) error
	Len(table string) int
	Connect() error
	Close()
}
