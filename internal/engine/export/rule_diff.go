package export

import (
	"io"

	"github.com/mbrt/gmailctl/internal/engine/config/v1alpha3"
	"github.com/mbrt/gmailctl/internal/engine/export/xml"
	"github.com/mbrt/gmailctl/internal/engine/filter"
)

// RuleDiffExporter exports newly added or modified rules to Gmail XML format.
type RuleDiffExporter struct {
	exporter xml.Exporter
}

// NewRuleDiffExporter creates a new RuleDiffExporter using the default XML exporter.
func NewRuleDiffExporter() RuleDiffExporter {
	return RuleDiffExporter{
		exporter: xml.DefaultExporter(),
	}
}

// NewRuleDiffExporterWithXMLExporter creates a RuleDiffExporter with a custom XML exporter.
func NewRuleDiffExporterWithXMLExporter(exp xml.Exporter) RuleDiffExporter {
	return RuleDiffExporter{
		exporter: exp,
	}
}

// ExportNew filters local rules against upstream rules and exports newly added
// or modified rules to the writer in Gmail XML format.
func (e RuleDiffExporter) ExportNew(author v1alpha3.Author, local, upstream filter.Filters, w io.Writer) error {
	newFilters := filter.FilterNew(local, upstream)
	return e.exporter.Export(author, newFilters, w)
}
