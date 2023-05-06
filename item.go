package kiva

import (
	"time"

	"github.com/ariefdarmawan/serde"
)

type Item struct {
	Data   interface{}
	Expiry time.Time
	Synced bool
	Error  error
}

func (i *Item) Set(v interface{}) {
	i.Data = v
}

func (i *Item) StoreTo(dest interface{}) error {
	return serde.Serde(i.Data, dest)
}
