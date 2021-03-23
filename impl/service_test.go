package impl

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestInMemoryWorks(t *testing.T) {
	NewServer("any", true)
	_, err := ioutil.ReadDir("any")
	if _, ok := err.(*os.PathError); !ok {
		t.Error("got an unexpected error, or nil error when it should have been os.PathError", err)
	}
}

func TestCreatesAFile(t *testing.T) {
	NewServer("any", false)
	defer os.RemoveAll("any")
	_, err := ioutil.ReadDir("any")
	if err != nil {
		t.Error("got an error unexpectedly", err)
	}
}
