package label

// Label contains information about a Gmail label.
type Label struct {
	ID          string
	Name        string
	Color       *Color
	NumMessages int
}

// Color is the color of a label.
//
// See https://developers.google.com/gmail/api/v1/reference/users/labels
// for the list of possible colors.
type Color struct {
	Background string
	Text       string
}
