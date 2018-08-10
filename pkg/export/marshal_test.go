package export

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mbrt/gmailfilter/pkg/config"
)

func testNow() time.Time {
	// Make test deterministic, avoiding time.Now()
	now, _ := time.Parse("2006/01/02 15:04", "2018/03/08 17:00")
	return now
}

func TestEmptyEntries(t *testing.T) {
	exporter := xmlExporter{now: testNow}
	author := config.Author{Name: "Pippo Pluto", Email: "pippo@mail.com"}
	buf := new(bytes.Buffer)
	err := exporter.MarshalEntries(author, []Entry{}, buf)
	assert.Nil(t, err)
	expected := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom" xmlns:apps="http://schemas.google.com/apps/2006">
  <title>Mail Filters</title>
  <id>tag:mail.google.com,2008:filters:</id>
  <updated>2018-03-08T17:00:00Z</updated>
  <author>
    <name>Pippo Pluto</name>
    <email>pippo@mail.com</email>
  </author>
</feed>`
	assert.Equal(t, expected, buf.String())
}

func TestEntries(t *testing.T) {
	exporter := xmlExporter{now: testNow}
	author := config.Author{Name: "Pippo Pluto", Email: "pippo@mail.com"}
	entries := []Entry{
		{
			{Name: PropertyFrom, Value: "foo@baz.com"},
			{Name: PropertyMarkImportant, Value: "true"},
		},
		{
			{Name: PropertyHas, Value: "SPAM!!"},
			{Name: PropertyDelete, Value: "true"},
			{Name: PropertyApplyLabel, Value: "spam"},
		},
	}
	buf := new(bytes.Buffer)
	err := exporter.MarshalEntries(author, entries, buf)
	assert.Nil(t, err)
	expected := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom" xmlns:apps="http://schemas.google.com/apps/2006">
  <title>Mail Filters</title>
  <id>tag:mail.google.com,2008:filters:</id>
  <updated>2018-03-08T17:00:00Z</updated>
  <author>
    <name>Pippo Pluto</name>
    <email>pippo@mail.com</email>
  </author>
  <entry>
    <category term="filter"></category>
    <title>Mail Filter</title>
    <content></content>
    <apps:property name="from" value="foo@baz.com"></apps:property>
    <apps:property name="shouldAlwaysMarkAsImportant" value="true"></apps:property>
  </entry>
  <entry>
    <category term="filter"></category>
    <title>Mail Filter</title>
    <content></content>
    <apps:property name="hasTheWord" value="SPAM!!"></apps:property>
    <apps:property name="shouldTrash" value="true"></apps:property>
    <apps:property name="label" value="spam"></apps:property>
  </entry>
</feed>`
	assert.Equal(t, expected, buf.String())
}
