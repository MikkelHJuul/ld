package impl

import "regexp"

type Matcher interface {
	Match([]byte) bool
}
type MatcherFunc func([]byte) bool

func (matcher MatcherFunc) Match(b []byte) bool {
	return matcher(b)
}

func NewMatcher(pattern string) (Matcher, error) {
	if pattern != "" {
		return regexp.Compile(pattern)
	}
	return MatcherFunc(func(_ []byte) bool { return true }), nil
}
