package xml

import (
	"encoding/xml"
	"io"
	"time"

	"github.com/mbrt/gmailfilter/pkg/config"
)

// Exporter exports the given entries to the Gmail xml format.
type Exporter interface {
	// MarshalEntries exports the given entries to the Gmail xml format.
	MarshalEntries(author config.Author, entries []Entry, w io.Writer) error
}

// DefaultExporter returns a default implementation of the XMLExporter interface.
func DefaultExporter() Exporter {
	return xmlExporter{now: defaultNow}
}

// nowFunc returns the current time
type nowFunc func() time.Time

var defaultNow nowFunc = func() time.Time { return time.Now() }

type xmlExporter struct {
	// Allows to be mocked away
	now nowFunc
}

func (x xmlExporter) MarshalEntries(author config.Author, entries []Entry, w io.Writer) error {
	doc := x.toXML(author, entries)
	out, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(xml.Header))
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	return err
}

type xmlDoc struct {
	XMLName     xml.Name  `xml:"feed"`
	XMLNS       string    `xml:"xmlns,attr"`
	XMLNSApps   string    `xml:"xmlns:apps,attr"`
	Title       string    `xml:"title"`
	ID          string    `xml:"id"`
	Updated     time.Time `xml:"updated"`
	AuthorName  string    `xml:"author>name"`
	AuthorEmail string    `xml:"author>email"`
	Entries     []xmlEntry
}

type xmlEntry struct {
	XMLName    xml.Name    `xml:"entry"`
	Category   xmlCategory `xml:"category"`
	Title      string      `xml:"title"`
	Content    string      `xml:"content"`
	Properties []xmlProperty
}

type xmlCategory struct {
	Term string `xml:"term,attr"`
}

type xmlProperty struct {
	XMLName xml.Name `xml:"apps:property"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value,attr"`
}

func (x xmlExporter) toXML(author config.Author, entries []Entry) xmlDoc {
	res := xmlDoc{
		XMLNS:       "http://www.w3.org/2005/Atom",
		XMLNSApps:   "http://schemas.google.com/apps/2006",
		Title:       "Mail Filters",
		ID:          "tag:mail.google.com,2008:filters:",
		Updated:     x.now(),
		AuthorName:  author.Name,
		AuthorEmail: author.Email,
		Entries:     x.entriesToXML(entries),
	}
	return res
}

func (x xmlExporter) entriesToXML(entries []Entry) []xmlEntry {
	res := make([]xmlEntry, len(entries))
	for i, entry := range entries {
		xentry := xmlEntry{
			Category:   xmlCategory{"filter"},
			Title:      "Mail Filter",
			Content:    "",
			Properties: x.propertiesToXML(entry),
		}
		res[i] = xentry
	}
	return res
}

func (x xmlExporter) propertiesToXML(props []Property) []xmlProperty {
	res := make([]xmlProperty, len(props))
	for i, prop := range props {
		xprop := xmlProperty{
			Name:  prop.Name,
			Value: prop.Value,
		}
		res[i] = xprop
	}
	return res
}
