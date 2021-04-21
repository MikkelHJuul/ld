package impl

import (
	"github.com/dgraph-io/badger/v3"
	"io/ioutil"
	"os"
	"testing"
)

func TestCreatesAFile(t *testing.T) {
	_, _ = NewServer(func(bo *badger.Options) {
		*bo = bo.WithDir("any").WithValueDir("any")
	})
	defer os.RemoveAll("any")
	_, err := ioutil.ReadDir("any")
	if err != nil {
		t.Error("got an error unexpectedly", err)
	}
}
