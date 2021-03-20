package impl

import (
	"testing"
)

func TestNewMatcherEmptyString(t *testing.T) {
	got, err := NewMatcher("")
	if err != nil {
		t.Errorf("NewMatcher() error = %v", err)
		return
	}
	if !got.Match(nil) && !got.Match([]byte("any")) && !got.Match([]byte("是")) {
		t.Errorf("Matcher doesn't match true always: %v", got)
	}
}
