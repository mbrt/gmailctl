package xml

import (
	"encoding/xml"
	"io"
	"time"

	"github.com/mbrt/gmailctl/internal/engine/config/v1alpha3"
	"github.com/mbrt/gmailctl/internal/engine/filter"
)

// DefaultExporter returns a default XML exporter.
func DefaultExporter() Exporter {
	return Exporter{now: defaultNow}
}

// NewWithTime returns a new exporter with the given time function.
func NewWithTime(now func() time.Time) Exporter {
	return Exporter{now}
}

var defaultNow = time.Now

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

// Exporter exports the given entries to the Gmail xml format.
type Exporter struct {
	// Allows to be mocked away
	now func() time.Time
}

// Export exports Gmail filters into the Gmail xml format.
func (x Exporter) Export(author v1alpha3.Author, filters filter.Filters, w io.Writer) error {
	doc, err := x.toXML(author, filters)
	if err != nil {
		return err
	}
	out, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(xml.Header))
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, "\n")
	return err
}

func (x Exporter) toXML(author v1alpha3.Author, filters filter.Filters) (xmlDoc, error) {
	entries, err := x.entriesToXML(filters)
	res := xmlDoc{
		XMLNS:       "http://www.w3.org/2005/Atom",
		XMLNSApps:   "http://schemas.google.com/apps/2006",
		Title:       "Mail Filters",
		ID:          "tag:mail.google.com,2008:filters:",
		Updated:     x.now(),
		AuthorName:  author.Name,
		AuthorEmail: author.Email,
		Entries:     entries,
	}
	return res, err
}

func (x Exporter) entriesToXML(filters filter.Filters) ([]xmlEntry, error) {
	res := make([]xmlEntry, len(filters))
	for i, f := range filters {
		props, err := x.propertiesToXML(f)
		if err != nil {
			return nil, err
		}
		xentry := xmlEntry{
			Category:   xmlCategory{"filter"},
			Title:      "Mail Filter",
			Content:    "",
			Properties: props,
		}
		res[i] = xentry
	}
	return res, nil
}

func (x Exporter) propertiesToXML(f filter.Filter) ([]xmlProperty, error) {
	res := x.criteriaProperties(f.Criteria)
	ap, err := x.actionProperties(f.Action)
	if err != nil {
		return nil, err
	}
	res = append(res, ap...)
	return res, nil
}

func (x Exporter) criteriaProperties(c filter.Criteria) []xmlProperty {
	res := []xmlProperty{}
	res = x.appendStringProperty(res, PropertyFrom, c.From)
	res = x.appendStringProperty(res, PropertyTo, c.To)
	res = x.appendStringProperty(res, PropertySubject, c.Subject)
	res = x.appendStringProperty(res, PropertyHas, c.Query)
	return res
}

func (x Exporter) actionProperties(a filter.Actions) ([]xmlProperty, error) {
	res := []xmlProperty{}
	res = x.appendBoolProperty(res, PropertyArchive, a.Archive)
	res = x.appendBoolProperty(res, PropertyDelete, a.Delete)
	res = x.appendBoolProperty(res, PropertyMarkImportant, a.MarkImportant)
	res = x.appendBoolProperty(res, PropertyMarkNotImportant, a.MarkNotImportant)
	res = x.appendBoolProperty(res, PropertyMarkRead, a.MarkRead)
	res = x.appendBoolProperty(res, PropertyMarkNotSpam, a.MarkNotSpam)
	res = x.appendBoolProperty(res, PropertyStar, a.Star)
	res = x.appendStringProperty(res, PropertyApplyLabel, a.AddLabel)
	res = x.appendStringProperty(res, PropertyForward, a.Forward)

	if a.Category != "" {
		cat, err := categoryToSmartLabel(a.Category)
		if err != nil {
			return nil, err
		}
		res = x.appendStringProperty(res, PropertyApplyCategory, cat)
	}

	return res, nil
}

func (x Exporter) appendStringProperty(res []xmlProperty, name, value string) []xmlProperty {
	if value == "" {
		return res
	}
	p := xmlProperty{
		Name:  name,
		Value: value,
	}
	return append(res, p)
}

func (x Exporter) appendBoolProperty(res []xmlProperty, name string, value bool) []xmlProperty {
	if !value {
		return res
	}
	p := xmlProperty{
		Name:  name,
		Value: "true",
	}
	return append(res, p)
}
