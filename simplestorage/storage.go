package simplestorage

type storage struct {
}

func NewStorage() *storage {
	s := new(storage)
	return s
}

func (s *storage) Connect() error {
	return nil
}

func (s *storage) Close() {
}

func (s *storage) Get(id string, dest interface{}) error {
	panic("not implemented") // TODO: Implement
}
