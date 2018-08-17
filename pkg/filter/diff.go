package filter

import (
	"sort"
	"strings"

	"github.com/cnf/structhash"
)

// FiltersDiff contains filters that have been added and removed locally with respect to upstream.
type FiltersDiff struct {
	Added   Filters
	Removed Filters
}

// NewMinimalFiltersDiff creates a new FiltersDiff with reordered filters, where
// similar added and removed ones are next to each other.
//
// The algorithm used is maximum bipartite matching, so the complexity is around O(N^3).
// Hopefully the number of filters is low enough to make this not too bad.
func NewMinimalFiltersDiff(added, removed Filters) FiltersDiff {
	reorderWithHammingDistance(added, removed)
	return FiltersDiff{added, removed}
}

func (f FiltersDiff) String() string {
	panic("TODO")
}

// Diff computes the diff between two lists of filters.
//
// To compute the diff, IDs are ignored, only the contents of the filters are actually considered.
func Diff(upstream, local Filters) (FiltersDiff, error) {
	hashedUp := newHashedFilters(upstream)
	hashedLoc := newHashedFilters(local)

	added := Filters{}
	removed := Filters{}

	i, j := 0, 0
	for i < len(upstream) && j < len(local) {
		ups := hashedUp[i]
		loc := hashedLoc[j]
		cmp := strings.Compare(ups.hash, loc.hash)

		if cmp < 0 {
			// Local is ahead: it is missing one filter
			removed = append(removed, ups.filter)
			i++

		} else if cmp > 0 {
			// Upstream is ahead: it is missing one filter
			added = append(added, loc.filter)
			j++
		} else {
			// All good
			i++
			j++
		}
	}

	// Consume all upstream that are not present in local
	for ; i < len(upstream); i++ {
		removed = append(removed, hashedUp[i].filter)
	}

	// Consume all local that are not present upstream
	for ; j < len(local); j++ {
		added = append(added, hashedLoc[j].filter)
	}

	return NewMinimalFiltersDiff(added, removed), nil
}

type hashedFilter struct {
	hash   string
	filter Filter
}

type hashedFilters []hashedFilter

func (hs hashedFilters) Len() int {
	return len(hs)
}

func (hs hashedFilters) Less(i, j int) bool {
	return strings.Compare(hs[i].hash, hs[j].hash) == -1
}

func (hs hashedFilters) Swap(i, j int) {
	tmp := hs[i]
	hs[i] = hs[j]
	hs[j] = tmp
}

func newHashedFilters(fs Filters) hashedFilters {
	res := make(hashedFilters, len(fs))

	for i, f := range fs {
		res[i] = hashFilter(f)
	}
	// By sorting we can compare two instances by going element-by-element in order
	sort.Sort(res)

	return res
}

func hashFilter(f Filter) hashedFilter {
	h, err := structhash.Hash(f, 1)
	if err != nil {
		panic("hash cannot fail, unreachable")
	}
	return hashedFilter{h, f}
}

func reorderWithHammingDistance(f1, f2 Filters) {
	// We use bipartite matching to match the two filters and order them accordingly
	panic("TODO")
}
