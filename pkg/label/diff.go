package label

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pmezard/go-difflib/difflib"

	"github.com/mbrt/gmailctl/pkg/filter"
)

// Diff computes the diff between two lists of labels.
//
// To compute the diff, IDs are ignored, only the properties of the labels are
// actually considered.
func Diff(upstream, local Labels) (LabelsDiff, error) {
	sort.Sort(byName(upstream))
	sort.Sort(byName(local))

	res := LabelsDiff{}
	i, j := 0, 0

	for i < len(upstream) && j < len(local) {
		ups := upstream[i]
		loc := local[j]
		cmp := strings.Compare(ups.Name, loc.Name)

		if cmp < 0 {
			// Local is ahead: it's missing a label
			res.Removed = append(res.Removed, ups)
			i++
		} else if cmp > 0 {
			// Upstream is ahead: it's missing a label
			res.Added = append(res.Added, loc)
			j++
		} else {
			// Same name, check if it's modified
			if !Equivalent(ups, loc) {
				res.Modified = append(res.Modified, ModifiedLabel{
					Old: ups,
					New: loc,
				})
			}
			i++
			j++
		}

	}

	// Consume all upstream that are not present in local
	for ; i < len(upstream); i++ {
		res.Removed = append(res.Removed, upstream[i])
	}

	// Consume all local that are not present upstream
	for ; j < len(local); j++ {
		res.Added = append(res.Added, local[j])
	}

	return res, nil
}

// LabelsDiff contains the diff of two lists of labels.
type LabelsDiff struct {
	Modified []ModifiedLabel
	Added    Labels
	Removed  Labels
}

// Empty returns true if the diff is empty.
func (d LabelsDiff) Empty() bool {
	return len(d.Added) == 0 && len(d.Removed) == 0 && len(d.Modified) == 0
}

func (d LabelsDiff) String() string {
	var old, new []string

	cleanup := func(l Label) Label {
		// Get rid of distracting information in the diff.
		return Label{
			Name:  l.Name,
			Color: l.Color,
		}
	}

	for _, ml := range d.Modified {
		old = append(old, cleanup(ml.Old).String()+"\n")
		new = append(new, ml.New.String()+"\n")
	}

	for _, l := range d.Removed {
		old = append(old, cleanup(l).String()+"\n")
	}

	for _, l := range d.Added {
		new = append(new, l.String()+"\n")
	}

	s, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        old,
		B:        new,
		FromFile: "Current",
		ToFile:   "TO BE APPLIED",
		Context:  3,
	})
	if err != nil {
		// We can't get a diff apparently, let's make something up here
		return fmt.Sprintf("Removed:\n%s\nAdded:\n%sModified:%s",
			strings.Join(old, "\n"),
			strings.Join(new, "\n"),
			fmt.Sprint(d.Modified),
		)
	}

	return s
}

// ModifiedLabel is a label in two versions, the old and the new.
type ModifiedLabel struct {
	Old Label
	New Label
}

// Validate makes sure that a diff is valid and safe to apply.
func Validate(d LabelsDiff, filters filter.Filters) error {
	for _, l := range d.Removed {
		if filters.HasLabel(l.Name) {
			return fmt.Errorf("cannot remove label %q, used in filter", l.Name)
		}
	}
	return nil
}

type byName Labels

func (b byName) Len() int {
	return len(b)
}

func (b byName) Less(i, j int) bool {
	return strings.Compare(b[i].Name, b[j].Name) == -1
}

func (b byName) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
