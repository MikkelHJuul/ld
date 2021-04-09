package impl

import "regexp"

//Matcher is simply an interface that regexp.Regexp fulfill (with the method we care for)
//This way we can implement another Matcher fulfilling this interface
type Matcher interface {
	Match([]byte) bool
}

// MatcherFunc is any `func([]byte) bool` acting as a named target for implementation
type MatcherFunc func([]byte) bool

// Match implements interface Matcher for MatcherFunc, delegating to the underlying function
func (matcher MatcherFunc) Match(b []byte) bool {
	return matcher(b)
}

//NewMatcher returns a matcher that is the regexp.Regexp or a function that is always true
func NewMatcher(pattern string) (Matcher, error) {
	if pattern != "" {
		return regexp.Compile(pattern)
	}
	return MatcherFunc(func(_ []byte) bool { return true }), nil
}
