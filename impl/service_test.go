package impl

import (
	"github.com/dgraph-io/badger/v3"
	"io/ioutil"
	"os"
	"testing"
)

func TestInMemoryWorks(t *testing.T) {
	NewServer(func(bo *badger.Options) {
		bo.WithDir("any").WithValueDir("any").WithInMemory(true)
	})
	_, err := ioutil.ReadDir("any")
	if _, ok := err.(*os.PathError); !ok {
		t.Error("got an unexpected error, or nil error when it should have been os.PathError", err)
	}
}

func TestCreatesAFile(t *testing.T) {
	NewServer(func(bo *badger.Options) {
		bo.WithDir("any").WithValueDir("any")
	})
	defer os.RemoveAll("any")
	_, err := ioutil.ReadDir("any")
	if err != nil {
		t.Error("got an error unexpectedly", err)
	}
}
