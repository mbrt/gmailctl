package cmd

import (
	"github.com/mbrt/gmailctl/internal/engine/api"
	papply "github.com/mbrt/gmailctl/internal/engine/apply"
)

func upstreamConfig(gmailapi *api.GmailAPI) (papply.GmailConfig, error) {
	cfg, err := papply.FromAPI(gmailapi)
	if err != nil {
		if len(cfg.Filters) == 0 {
			return papply.GmailConfig{}, err
		}
		// We have some filters, let's work with what we have and issue a warning.
		stderrPrintf("Warning: Error getting one or more filters from Gmail: %sThey will be ignored in the diff.\n", err)
	}
	return cfg, nil
}
