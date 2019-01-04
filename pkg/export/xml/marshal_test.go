package xml

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	cfgv1 "github.com/mbrt/gmailctl/pkg/config/v1alpha1"
	"github.com/mbrt/gmailctl/pkg/filter"
	"github.com/mbrt/gmailctl/pkg/gmail"
)

func testNow() time.Time {
	// Make test deterministic, avoiding time.Now()
	now, _ := time.Parse("2006/01/02 15:04", "2018/03/08 17:00")
	return now
}

func TestEmptyEntries(t *testing.T) {
	exporter := xmlExporter{now: testNow}
	author := cfgv1.Author{Name: "Pippo Pluto", Email: "pippo@mail.com"}
	buf := new(bytes.Buffer)
	err := exporter.Export(author, filter.Filters{}, buf)
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
</feed>
`
	assert.Equal(t, expected, buf.String())
}

func TestSomeEntries(t *testing.T) {
	exporter := xmlExporter{now: testNow}
	author := cfgv1.Author{Name: "Pippo Pluto", Email: "pippo@mail.com"}
	filters := filter.Filters{
		{
			Action: filter.Action{
				MarkImportant: true,
			},
			Criteria: filter.Criteria{
				From: "foo@baz.com",
			},
		},
		{
			Action: filter.Action{
				Delete:   true,
				AddLabel: "spam",
			},
			Criteria: filter.Criteria{
				Query: "SPAM!!",
			},
		},
	}
	buf := new(bytes.Buffer)
	err := exporter.Export(author, filters, buf)
	assert.Nil(t, err)
	expected := `
<?xml version="1.0" encoding="UTF-8"?>
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
</feed>
`
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(buf.String()))
}

func TestAllEntries(t *testing.T) {
	exporter := xmlExporter{now: testNow}
	author := cfgv1.Author{Name: "Pippo Pluto", Email: "pippo@mail.com"}
	filters := filter.Filters{
		{
			Action: filter.Action{
				Archive:       true,
				Delete:        true,
				MarkImportant: true,
				MarkRead:      true,
				Category:      gmail.CategoryPromotions,
				AddLabel:      "MyLabel",
			},
			Criteria: filter.Criteria{
				From:    "foo@baz.com",
				To:      "me@gmail.com",
				Subject: "subject",
				Query:   "has words",
			},
		},
	}
	buf := new(bytes.Buffer)
	err := exporter.Export(author, filters, buf)
	assert.Nil(t, err)
	expected := `
<?xml version="1.0" encoding="UTF-8"?>
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
    <apps:property name="to" value="me@gmail.com"></apps:property>
    <apps:property name="subject" value="subject"></apps:property>
    <apps:property name="hasTheWord" value="has words"></apps:property>
    <apps:property name="shouldArchive" value="true"></apps:property>
    <apps:property name="shouldTrash" value="true"></apps:property>
    <apps:property name="shouldAlwaysMarkAsImportant" value="true"></apps:property>
    <apps:property name="shouldMarkAsRead" value="true"></apps:property>
    <apps:property name="label" value="MyLabel"></apps:property>
    <apps:property name="smartLabelToApply" value="^smartlabel_promo"></apps:property>
  </entry>
</feed>`
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(buf.String()))
}
