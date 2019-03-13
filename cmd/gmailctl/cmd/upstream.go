package cmd

import (
	"github.com/pkg/errors"

	"github.com/mbrt/gmailctl/pkg/api"
	"github.com/mbrt/gmailctl/pkg/filter"
)

func upstreamFilters(gmailapi api.GmailAPI) (filter.Filters, error) {
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
