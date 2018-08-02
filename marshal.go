package main

import (
	"encoding/xml"
	"io"
	"time"
)

// MarshalEntries exports the given entries to the Gmail xml format.
func MarshalEntries(author Author, entries []Entry, w io.Writer) error {
	doc := toXml(author, entries)
	out, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	w.Write([]byte(xml.Header))
	w.Write(out)
	return nil
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
	XMLName      xml.Name  `xml:"entry"`
	CategoryTerm string    `xml:"category>term,attr"`
	Title        string    `xml:"title"`
	ID           string    `xml:"id"`
	Updated      time.Time `xml:"updated"`
	Content      string    `xml:"content"`
	Properties   []xmlProperty
}

type xmlProperty struct {
	XMLName xml.Name `xml:"apps:property"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value,attr"`
}

func toXml(author Author, entries []Entry) xmlDoc {
	res := xmlDoc{
		XMLNS:       "http://www.w3.org/2005/Atom",
		XMLNSApps:   "http://schemas.google.com/apps/2006",
		Title:       "Mail Filters",
		ID:          "tag:mail.google.com,2008:filters:",
		Updated:     time.Now(),
		AuthorName:  author.Name,
		AuthorEmail: author.Email,
		Entries:     entriesToXml(entries),
	}
	return res
}

func entriesToXml(entries []Entry) []xmlEntry {
	res := make([]xmlEntry, len(entries))
	for i, entry := range entries {
		xentry := xmlEntry{
			CategoryTerm: "Filter",
			Title:        "Mail Filter",
			Content:      "",
			Properties:   propertiesToXml(entry.Properties),
		}
		res[i] = xentry
	}
	return res
}

func propertiesToXml(props []Property) []xmlProperty {
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
