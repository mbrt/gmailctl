package cmd

import (
	"fmt"

	"github.com/mbrt/gmailctl/pkg/api"
	papply "github.com/mbrt/gmailctl/pkg/apply"
)

func upstreamConfig(gmailapi *api.GmailAPI) (papply.GmailConfig, error) {
	f, err := gmailapi.ListFilters()
	if err != nil {
		if len(f) == 0 {
			return papply.GmailConfig{}, fmt.Errorf("getting filters from Gmail: %w", err)
		}
		// We have some filters, let's work with what we have and issue a warning.
		stderrPrintf("Warning: Error getting one or more filters from Gmail: %sThey will be ignored in the diff.\n", err)
	}
	l, err := gmailapi.ListLabels()
	return papply.GmailConfig{
		Labels:  l,
		Filters: f,
	}, err
}
