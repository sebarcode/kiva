package simplemem

import "github.com/sebarcode/kiva"

type memory struct {
}

func NewMemory() *memory {
	mem := new(memory)
	return mem
}

func (mem *memory) Get(table, id string, dest interface{}) (*kiva.ItemOptions, error) {
	panic("not implemented") // TODO: Implement
}

func (mem *memory) Set(table, id string, value interface{}, opts *kiva.ItemOptions) error {
	panic("not implemented") // TODO: Implement
}

func (mem *memory) Connect() error {
	return nil
}

func (mem *memory) Close() {
}
