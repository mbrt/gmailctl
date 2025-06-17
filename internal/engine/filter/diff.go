package filter

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	"github.com/pmezard/go-difflib/difflib"

	"github.com/mbrt/gmailctl/internal/graph"
	"github.com/mbrt/gmailctl/internal/reporting"
)

// Diff computes the diff between two lists of filters.
//
// To compute the diff, IDs are ignored, only the contents of the filters are actually considered.
func Diff(upstream, local Filters, debugInfo bool, contextLines int, colorize bool) (FiltersDiff, error) {
	// Computing the diff is very expensive, so we have to minimize the number of filters
	// we have to analyze. To do so, we get rid of the filters that are exactly the same,
	// by hashing them.
	added, removed := changedFilters(upstream, local)
	return NewMinimalFiltersDiff(added, removed, debugInfo, contextLines, colorize), nil
}

// NewMinimalFiltersDiff creates a new FiltersDiff with reordered filters, where
// similar added and removed ones are next to each other.
//
// The algorithm used is a quadratic approximation to the otherwise NP-complete
// travel salesman problem. Hopefully the number of filters is low enough to
// make this not too slow and the approximation not too bad.
func NewMinimalFiltersDiff(added, removed Filters, printDebugInfo bool, contextLines int, colorize bool) FiltersDiff {
	if len(added) > 0 && len(removed) > 0 {
		added, removed = reorderWithHungarian(added, removed)
	}
	return FiltersDiff{added, removed, printDebugInfo, contextLines, colorize}
}

// FiltersDiff contains filters that have been added and removed locally with respect to upstream.
type FiltersDiff struct {
	Added          Filters
	Removed        Filters
	PrintDebugInfo bool
	ContextLines   int
	Colorize       bool
}

// Empty returns true if the diff is empty.
func (f FiltersDiff) Empty() bool {
	return len(f.Added) == 0 && len(f.Removed) == 0
}

func (f FiltersDiff) String() string {
	var removed, added string
	if f.PrintDebugInfo {
		removed = f.Removed.DebugString()
		added = f.Added.DebugString()
	} else {
		removed = f.Removed.String()
		added = f.Added.String()
	}

	s, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(removed),
		B:        difflib.SplitLines(added),
		FromFile: "Current",
		ToFile:   "TO BE APPLIED",
		Context:  f.ContextLines,
	})
	if err != nil {
		// We can't get a diff apparently, let's make something up here
		return fmt.Sprintf("Removed:\n%s\nAdded:\n%s", removed, added)
	}
	if f.Colorize {
		s = reporting.ColorizeDiff(s)
	}

	return s
}

func changedFilters(upstream, local Filters) (added, removed Filters) {
	hupstream := newHashedFilters(upstream)
	hlocal := newHashedFilters(local)

	i, j := 0, 0
	for i < len(hupstream) && j < len(hlocal) {
		ups := hupstream[i]
		loc := hlocal[j]
		cmp := strings.Compare(ups.hash, loc.hash)

		switch {
		case cmp < 0:
			// Local is ahead: it is missing one filter
			removed = append(removed, ups.filter)
			i++
		case cmp > 0:
			// Upstream is ahead: it is missing one filter
			added = append(added, loc.filter)
			j++
		default:
			// All good
			i++
			j++
		}
	}

	// Consume all upstream that are not present in local
	for ; i < len(hupstream); i++ {
		removed = append(removed, hupstream[i].filter)
	}

	// Consume all local that are not present upstream
	for ; j < len(hlocal); j++ {
		added = append(added, hlocal[j].filter)
	}

	return added, removed
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
	// Remove duplicates while creating the filters
	// Gmail doesn't support them, so we might as well do it here.
	uniqueFs := map[string]Filter{}
	for _, f := range fs {
		hf := hashFilter(f)
		uniqueFs[hf.hash] = f
	}

	// By sorting we can compare two instances by going element-by-element
	// in order
	res := hashedFilters{}
	for h, f := range uniqueFs {
		res = append(res, hashedFilter{h, f})
	}
	sort.Sort(res)

	return res
}

func hashFilter(f Filter) hashedFilter {
	// We have to hash only the contents, not the ID
	noIDFilter := Filter{
		Action:   f.Action,
		Criteria: f.Criteria,
	}
	h := hashStruct(noIDFilter)
	return hashedFilter{h, f}
}

func hashStruct(a interface{}) string {
	h := sha256.New()
	if _, err := h.Write([]byte(fmt.Sprintf("%#v", a))); err != nil {
		// This should be unreachable.
		panic(err)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// reorderWithHungarian reorders the two lists to make them look as similar as
// possible based on the hungarian algorithm.
//
// See https://en.wikipedia.org/wiki/Hungarian_algorithm
func reorderWithHungarian(f1, f2 Filters) (Filters, Filters) {
	c := costMatrix(f1, f2)
	mapping := hungarian(c)
	return reorderWithMapping(f1, f2, mapping)
}

func costMatrix(fs1, fs2 Filters) [][]float64 {
	// Compute the strings only once at the beginning
	ss1 := filterStrings(fs1)
	ss2 := filterStrings(fs2)

	var c [][]float64
	for i, s1 := range ss1 {
		c = append(c, nil)
		for _, s2 := range ss2 {
			c[i] = append(c[i], diffCost(s1, s2))
		}
	}

	return c
}

type filterLines []string

func filterStrings(fs Filters) []filterLines {
	var res []filterLines
	for _, f := range fs {
		res = append(res, difflib.SplitLines(f.String()))
	}
	return res
}

func diffCost(s1, s2 filterLines) float64 {
	m := difflib.NewMatcher(s1, s2)
	// Ratio returns a measure of similarity between 0 and 1.
	// We have to return a cost instead.
	return 1 - m.Ratio()
}

func hungarian(c [][]float64) []int {
	if len(c) == 0 {
		return nil
	}

	var mnk graph.Munkres
	mnk.Init(len(c), len(c[0]))
	mnk.SetCostMatrix(c)
	mnk.Run()
	return mnk.Links
}

func reorderWithMapping(f1, f2 Filters, mapping []int) (Filters, Filters) {
	var r1, r2 Filters

	mappedF1 := map[int]struct{}{}
	mappedF2 := map[int]struct{}{}

	// mapping[i] = j means that filter1[i] is matched with filter2[j]
	for i, j := range mapping {
		if j < 0 {
			continue
		}
		r1 = append(r1, f1[i])
		r2 = append(r2, f2[j])
		mappedF1[i] = struct{}{}
		mappedF2[j] = struct{}{}
	}

	// Add unmapped filters
	for i, f := range f1 {
		if _, ok := mappedF1[i]; !ok {
			r1 = append(r1, f)
		}
	}
	for i, f := range f2 {
		if _, ok := mappedF2[i]; !ok {
			r2 = append(r2, f)
		}
	}

	return r1, r2
}
