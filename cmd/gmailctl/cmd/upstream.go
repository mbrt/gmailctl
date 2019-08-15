package cmd

import (
	"github.com/pkg/errors"

	"github.com/mbrt/gmailctl/pkg/api"
	papply "github.com/mbrt/gmailctl/pkg/apply"
	"github.com/mbrt/gmailctl/pkg/filter"
)

var errLabelsDisabled = errors.New("label management disabled")

func upstreamConfig(gmailapi *api.GmailAPI) (papply.GmailConfig, error) {
	f, err := gmailapi.ListFilters()
	if err != nil {
		if len(f) == 0 {
			return papply.GmailConfig{}, errors.Wrap(err, "cannot get filters from Gmail")
		}
		// We have some filters, let's work with what we have and issue a warning.
		stderrPrintf("Warning: Error getting one or more filters from Gmail: %sThey will be ignored in the diff.\n", err)
	}
	l, err := gmailapi.ListLabels()
	if err != nil {
		stderrPrintf("Warning: Error getting labels from Gmail: %s. Labels management will be disabled\n", err)
		err = errLabelsDisabled
	}
	return papply.GmailConfig{
		Labels:  l,
		Filters: f,
	}, err
}

// TODO REMOVE
func upstreamFilters(gmailapi *api.GmailAPI) (filter.Filters, error) {
	f, err := gmailapi.ListFilters()
	if err != nil {
		if len(f) == 0 {
			return f, errors.Wrap(err, "cannot get filters from Gmail")
		}
		// We have some filters, let's work with what we have and issue a warning.
		stderrPrintf("Warning: Error getting one or more filters from Gmail: %sThey will be ignored in the diff.\n", err)
		return f, nil
	}
	return f, nil
}
