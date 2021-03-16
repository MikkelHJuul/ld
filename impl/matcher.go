package impl

import "regexp"

//Matcher is simply an interface that regexp.Regexp fulfill (with the method we care for)
//This way we can implement another Matcher fulfilling this interface
type Matcher interface {
	Match([]byte) bool
}
type MatcherFunc func([]byte) bool

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
