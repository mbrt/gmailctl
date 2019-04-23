package xml

import (
	"encoding/xml"
	"io"
	"time"

	cfgv2 "github.com/mbrt/gmailctl/pkg/config/v1alpha2"
	"github.com/mbrt/gmailctl/pkg/filter"
)

// Exporter exports the given entries to the Gmail xml format.
type Exporter interface {
	// Export exports Gmail filters into the Gmail xml format.
	Export(author cfgv2.Author, filters filter.Filters, w io.Writer) error
}

// DefaultExporter returns a default implementation of the XMLExporter interface.
func DefaultExporter() Exporter {
	return xmlExporter{now: defaultNow}
}

// nowFunc returns the current time
type nowFunc func() time.Time

var defaultNow nowFunc = func() time.Time { return time.Now() }

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

type xmlExporter struct {
	// Allows to be mocked away
	now nowFunc
}

func (x xmlExporter) Export(author cfgv2.Author, filters filter.Filters, w io.Writer) error {
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

func (x xmlExporter) toXML(author cfgv2.Author, filters filter.Filters) (xmlDoc, error) {
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

func (x xmlExporter) entriesToXML(filters filter.Filters) ([]xmlEntry, error) {
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

func (x xmlExporter) propertiesToXML(f filter.Filter) ([]xmlProperty, error) {
	res := x.criteriaProperties(f.Criteria)
	ap, err := x.actionProperties(f.Action)
	if err != nil {
		return nil, err
	}
	res = append(res, ap...)
	return res, nil
}

func (x xmlExporter) criteriaProperties(c filter.Criteria) []xmlProperty {
	res := []xmlProperty{}
	res = x.appendStringProperty(res, PropertyFrom, c.From)
	res = x.appendStringProperty(res, PropertyTo, c.To)
	res = x.appendStringProperty(res, PropertySubject, c.Subject)
	res = x.appendStringProperty(res, PropertyHas, c.Query)
	return res
}

func (x xmlExporter) actionProperties(a filter.Actions) ([]xmlProperty, error) {
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

func (x xmlExporter) appendStringProperty(res []xmlProperty, name, value string) []xmlProperty {
	if value == "" {
		return res
	}
	p := xmlProperty{
		Name:  name,
		Value: value,
	}
	return append(res, p)
}

func (x xmlExporter) appendBoolProperty(res []xmlProperty, name string, value bool) []xmlProperty {
	if !value {
		return res
	}
	p := xmlProperty{
		Name:  name,
		Value: "true",
	}
	return append(res, p)
}
