package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCriteria(t *testing.T) {
	tests := []struct {
		name           string
		criteria       Criteria
		gmailSearch    string
		gmailSearchURL string
	}{
		{
			name:           "from criteria",
			criteria:       Criteria{From: "someone@gmail.com"},
			gmailSearch:    "from:someone@gmail.com",
			gmailSearchURL: "https://mail.google.com/mail/u/0/#search/from%3Asomeone%40gmail.com",
		},
		{
			name: "complicated query",
			criteria: Criteria{
				Query: "{from:noreply@acme.com to:{me@google.com me@acme.com}}",
			},
			gmailSearch:    "{from:noreply@acme.com to:{me@google.com me@acme.com}}",
			gmailSearchURL: "https://mail.google.com/mail/u/0/#search/%7Bfrom%3Anoreply%40acme.com+to%3A%7Bme%40google.com+me%40acme.com%7D%7D",
		},
		{
			name: "all fileds",
			criteria: Criteria{
				From:    "someone@gmail.com",
				To:      "me@gmail.com",
				Subject: "Hello world",
				Query:   "unsubscribe",
			},
			gmailSearch:    "from:someone@gmail.com to:me@gmail.com subject:Hello world unsubscribe",
			gmailSearchURL: "https://mail.google.com/mail/u/0/#search/from%3Asomeone%40gmail.com+to%3Ame%40gmail.com+subject%3AHello+world+unsubscribe",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.criteria.ToGmailSearch(), tc.gmailSearch)
			assert.Equal(t, tc.criteria.ToGmailSearchURL(), tc.gmailSearchURL)
		})
	}
}
