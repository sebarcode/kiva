package kiva

type StorageProvider interface {
	Get(id string, dest interface{}) error
	Connect() error
	Close()
}
