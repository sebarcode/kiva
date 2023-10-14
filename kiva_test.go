package kiva_test

import (
	"os"
	"testing"

	"github.com/sebarcode/kiva"
	"github.com/sebarcode/kiva/simplemem"
	"github.com/sebarcode/kiva/simplestorage"
)

var (
	kv      *kiva.Kv
	mem     kiva.MemoryProvider
	storage kiva.StorageProvider
)

func TestMain(m *testing.M) {
	mem = simplemem.NewMemory()
	storage = simplestorage.NewStorage()

	kv := kiva.New("kivatest", mem, storage)
	defer kv.Close()

	os.Exit(m.Run())
}
