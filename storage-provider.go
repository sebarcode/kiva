package kiva

type StorageProvider interface {
	Get(table, id string, dest interface{}) error
	Connect() error
	Close()
}
