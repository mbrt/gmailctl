// Package stringset helps with common set operations on strings.
//
// This is inspired by https://github.com/uber/kraken/blob/master/utils/stringset/stringset.go.
package stringset

// Set is a nifty little wrapper for common set operations on a map. Because it
// is equivalent to a map, make/range/len will still work with Set.
type Set map[string]struct{}

// New creates a new Set with xs.
func New(xs ...string) Set {
	s := make(Set)
	for _, x := range xs {
		s.Add(x)
	}
	return s
}

// Add adds x to s.
func (s Set) Add(x string) {
	s[x] = struct{}{}
}

// Remove removes x from s.
func (s Set) Remove(x string) {
	delete(s, x)
}

// Has returns true if x is in s.
func (s Set) Has(x string) bool {
	_, ok := s[x]
	return ok
}

// ToSlice converts s to a slice.
func (s Set) ToSlice() []string {
	var xs []string
	for x := range s {
		xs = append(xs, x)
	}
	return xs
}
