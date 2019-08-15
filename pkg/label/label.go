package label

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	cfgv3 "github.com/mbrt/gmailctl/pkg/config/v1alpha3"
)

// Labels is a list of labels.
type Labels []Label

func (ls Labels) String() string {
	var ss []string
	for _, l := range ls {
		ss = append(ss, l.String())
	}
	return strings.Join(ss, "\n")
}

// Validate checks the given labels for possible issues.
func (ls Labels) Validate() error {
	lmap := stringset{}

	for _, l := range ls {
		n := l.Name
		if n == "" {
			return errors.New("invalid label without a name")
		}
		if strings.HasPrefix(n, "/") {
			return errors.Errorf("label '%s' shouldn't start with /", n)
		}
		if strings.HasSuffix(n, "/") {
			return errors.Errorf("label '%s' shouldn't end with /", n)
		}
		if _, ok := lmap[n]; ok {
			return errors.Errorf("label '%s' provided multiple times", n)
		}
		lmap[n] = struct{}{}
	}

	// Check that the labels have all the parents.
	// e.g. A/B requires A to be in the list.
	for _, l := range ls {
		if err := checkPrefix(lmap, l.Name); err != nil {
			return err
		}
	}

	return nil
}

type stringset map[string]struct{}

func checkPrefix(m stringset, n string) error {
	i := strings.LastIndex(n, "/")
	if i < 0 {
		return nil
	}
	prefix := n[:i]
	if _, ok := m[prefix]; !ok {
		return errors.Errorf("label '%s' requires label '%s' to be present",
			n, prefix)
	}
	return checkPrefix(m, prefix)
}

// Label contains information about a Gmail label.
type Label struct {
	ID          string
	Name        string
	Color       *Color
	NumMessages int
}

func (l Label) String() string {
	var ss []string

	if l.ID != "" {
		ss = append(ss, fmt.Sprintf("%s [%s]", l.Name, l.ID))
	} else {
		ss = append(ss, l.Name)
	}
	if l.Color != nil {
		ss = append(ss, fmt.Sprintf("color: %s, %s",
			l.Color.Background, l.Color.Text))
	}
	if l.NumMessages > 0 {
		ss = append(ss, fmt.Sprintf("num messages: %d", l.NumMessages))
	}

	return strings.Join(ss, "; ")
}

// Color is the color of a label.
//
// See https://developers.google.com/gmail/api/v1/reference/users/labels
// for the list of possible colors.
type Color struct {
	Background string
	Text       string
}

// Equivalent returns true if two labels can be considered equal, despite a
// different ID or number of messages.
//
// Unspecified color is also ignored.
func Equivalent(upstream, local Label) bool {
	// Ignore ID
	if upstream.Name != local.Name {
		return false
	}

	upsHasColor := upstream.Color != nil
	locHasColor := local.Color != nil
	if !locHasColor {
		// We don't care about the color locally
		return true
	}
	if !upsHasColor {
		// Going to add the color.
		return false
	}
	// Need to check if the color is the same
	return *upstream.Color == *local.Color
}

// FromConfig creates labels from the config format.
func FromConfig(ls []cfgv3.Label) Labels {
	var res Labels

	for _, l := range ls {
		var color *Color
		if l.Color != nil {
			color = &Color{
				Background: l.Color.Background,
				Text:       l.Color.Text,
			}
		}
		res = append(res, Label{
			Name:  l.Name,
			Color: color,
		})
	}

	return res
}
