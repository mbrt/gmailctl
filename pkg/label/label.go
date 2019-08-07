package label

import (
	"fmt"
	"strings"

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
func Equivalent(l1, l2 Label) bool {
	// Ignore ID
	if l1.Name != l2.Name {
		return false
	}

	l1HasColor := l1.Color != nil
	l2HasColor := l2.Color != nil
	if l1HasColor != l2HasColor {
		return false
	}
	if !l1HasColor {
		// Both are not colored
		return true
	}
	// Need to check if the color is the same
	return *l1.Color == *l2.Color
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
