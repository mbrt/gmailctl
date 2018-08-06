package main

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func testNow() time.Time {
	// Make test deterministic, avoiding time.Now()
	now, _ := time.Parse("2006/01/02 15:04", "2018/03/08 17:00")
	return now
}

func TestMarshalEmptyEntries(t *testing.T) {
	exporter := xmlExporter{now: testNow}
	author := Author{Name: "Pippo Pluto", Email: "pippo@mail.com"}
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
