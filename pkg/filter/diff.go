package filter

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/cnf/structhash"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

// FiltersDiff contains filters that have been added and removed locally with respect to upstream.
type FiltersDiff struct {
	Added   Filters
	Removed Filters
}

// NewMinimalFiltersDiff creates a new FiltersDiff with reordered filters, where
// similar added and removed ones are next to each other.
//
// The algorithm used is a quadratic approximation to the otherwise NP-complete
// travel salesman problem. Hopefully the number of filters is low enough to
// make this not too slow and the approximation not too bad.
func NewMinimalFiltersDiff(added, removed Filters) FiltersDiff {
	reorderWithLevenshtein(added, removed)
	return FiltersDiff{added, removed}
}

func (f FiltersDiff) String() string {
	s, err := difflib.GetContextDiffString(difflib.ContextDiff{
		A:        difflib.SplitLines(f.Removed.String()),
		B:        difflib.SplitLines(f.Added.String()),
		FromFile: "Original",
		ToFile:   "Current",
		Context:  5,
	})
	if err != nil {
		// We can't get a diff apparently, let's make something up here
		return fmt.Sprintf("Removed:\n%s\nAdded:\n%s", f.Removed, f.Added)
	}
	return s
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
	hs[i], hs[j] = hs[j], hs[i]
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

// reorderWithLevenshtein reorders the two lists to make them look as similar as
// possible based on Levenshtein distance pair-by-pair.
//
// The algorithm is a quadratic approximation of the travel salesman problem.
func reorderWithLevenshtein(f1, f2 Filters) {
	base := f1
	other := f2
	if len(f1) > len(f2) {
		base = f2
		other = f1
	}
	baseR := toRunes(base)
	otherR := toRunes(other)

	for i, b := range baseR {
		minDist := math.MaxInt64
		for j, o := range otherR[i:] {
			dist := levenshtein.DistanceForStrings(b, o, levenshtein.DefaultOptions)
			if dist < minDist {
				other[i], other[j] = other[j], other[i]
				minDist = dist
			}
		}
	}
}

func toRunes(fs Filters) [][]rune {
	res := make([][]rune, len(fs))
	for i, f := range fs {
		res[i] = []rune(fmt.Sprintf("%v", f))
	}
	return res
}
