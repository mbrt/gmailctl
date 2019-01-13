package filter

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/cnf/structhash"
	"github.com/eapache/queue"
	"github.com/pmezard/go-difflib/difflib"
)

// Diff computes the diff between two lists of filters.
//
// To compute the diff, IDs are ignored, only the contents of the filters are actually considered.
func Diff(upstream, local Filters) (FiltersDiff, error) {
	added, removed := changedFilters(upstream, local)
	return NewMinimalFiltersDiff(added, removed), nil
}

// NewMinimalFiltersDiff creates a new FiltersDiff with reordered filters, where
// similar added and removed ones are next to each other.
//
// The algorithm used is a quadratic approximation to the otherwise NP-complete
// travel salesman problem. Hopefully the number of filters is low enough to
// make this not too slow and the approximation not too bad.
func NewMinimalFiltersDiff(added, removed Filters) FiltersDiff {
	reorderWithMaxFlow(added, removed)
	return FiltersDiff{added, removed}
}

// FiltersDiff contains filters that have been added and removed locally with respect to upstream.
type FiltersDiff struct {
	Added   Filters
	Removed Filters
}

// Empty returns true if the diff is empty.
func (f FiltersDiff) Empty() bool {
	return len(f.Added) == 0 && len(f.Removed) == 0
}

func (f FiltersDiff) String() string {
	s, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
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

func changedFilters(upstream, local Filters) (added, removed Filters) {
	hupstream := newHashedFilters(upstream)
	hlocal := newHashedFilters(local)

	i, j := 0, 0
	for i < len(upstream) && j < len(local) {
		ups := hupstream[i]
		loc := hlocal[j]
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
		removed = append(removed, hupstream[i].filter)
	}

	// Consume all local that are not present upstream
	for ; j < len(local); j++ {
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
	res := make(hashedFilters, len(fs))

	for i, f := range fs {
		res[i] = hashFilter(f)
	}
	// By sorting we can compare two instances by going element-by-element
	// in order
	sort.Sort(res)

	return res
}

func hashFilter(f Filter) hashedFilter {
	// We have to hash only the contents, not the ID
	noIDFilter := Filter{
		Action:   f.Action,
		Criteria: f.Criteria,
	}
	h, err := structhash.Hash(noIDFilter, 1)
	if err != nil {
		panic("hash cannot fail, unreachable")
	}
	return hashedFilter{h, f}
}

// reorderWithMaxFlow reorders the two lists to make them look as similar as
// possible based on bipartite matching.
func reorderWithMaxFlow(f1, f2 Filters) {
}

type graph []vertex

type edge struct {
	Dest int
	Cap  int
	Flow int
}

type vertex struct {
	ID       int
	OutEdges []edge
}

type auxGraph []auxVertex

type auxVertex struct {
	ID int
	// OutEdges maps dest to residual flow
	OutEdges   map[int]int
	PathParent int
	Visited    bool
}

func maxFlow(g graph, src, dest int) {
	aux := toAux(g)
	auxMaxFlow(aux, src, dest)
	copyFlow(aux, g)
}

func toAux(g graph) auxGraph {
	var res auxGraph
	for _, v := range g {
		av := auxVertex{
			ID:         v.ID,
			PathParent: -1,
			OutEdges:   map[int]int{},
		}
		for _, e := range v.OutEdges {
			av.OutEdges[e.Dest] = e.Cap
		}
		res = append(res, av)
	}
	return res
}

func auxMaxFlow(g auxGraph, src, dest int) {
	for {
		hasPath := findAugmentingPath(g, src, dest)
		if !hasPath {
			break
		}
		mfp := maxFlowPath(g, dest)
		increaseFlow(g, dest, mfp)
	}
}

func findAugmentingPath(g auxGraph, src, dest int) bool {
	// Initialize visit
	for i := range g {
		g[i].Visited = false
	}

	q := queue.New()
	q.Add(src)
	g[src].PathParent = -1
	g[src].Visited = true

	for q.Length() > 0 {
		currID := q.Remove().(int)
		for dst := range g[currID].OutEdges {
			dstv := &g[dst]
			if dstv.Visited {
				continue
			}

			dstv.PathParent = currID
			dstv.Visited = true
			if dst == dest {
				return true
			}

			q.Add(dst)
		}
	}

	return false
}

func maxFlowPath(g auxGraph, dest int) int {
	max := math.MaxInt64
	curr := &g[dest]

	for curr.PathParent != -1 {
		parent := &g[curr.PathParent]
		resflow := parent.OutEdges[curr.ID]
		if resflow < max {
			max = resflow
		}
		curr = parent
	}

	return max
}

func increaseFlow(g auxGraph, dest, q int) {
	curr := &g[dest]

	for curr.PathParent != -1 {
		decreaseFlowEdge(g, curr.PathParent, curr.ID, q)
		increaseFlowEdge(g, curr.ID, curr.PathParent, q)
		parent := &g[curr.PathParent]
		curr = parent
	}
}

func decreaseFlowEdge(g auxGraph, src, dest, q int) {
	srcv := &g[src]
	flow := srcv.OutEdges[dest]
	flow -= q
	if flow > 0 {
		srcv.OutEdges[dest] = flow
	} else {
		delete(srcv.OutEdges, dest)
	}
}

func increaseFlowEdge(g auxGraph, src, dest, q int) {
	srcv := &g[src]
	flow, ok := srcv.OutEdges[dest]
	if ok {
		srcv.OutEdges[dest] = flow + q
	} else {
		srcv.OutEdges[dest] = q
	}
}

func copyFlow(aux auxGraph, g graph) {
	for i := range g {
		dstv := &g[i]
		srcv := &aux[i]
		for j, e := range dstv.OutEdges {
			resFlow := srcv.OutEdges[e.Dest]
			dstv.OutEdges[j].Flow = e.Cap - resFlow
		}
	}
}
