package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/mbrt/gmailctl/pkg/api"
)

// Authenticator is the APIProvider used by all gmailctl commands.
var Authenticator APIProvider = localCredentialsProvider{}

// APIProvider is the integration point between gmailctl commands and GMail
// APIs providers.
type APIProvider interface {
	Open(ctx context.Context) (*api.GmailAPI, error)
}

type localCredentialsProvider struct{}

func (l localCredentialsProvider) Open(ctx context.Context) (*api.GmailAPI, error) {
	auth, err := openCredentials()
	if err != nil {
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}
	return openToken(ctx, auth)
}

func openAPI() (*api.GmailAPI, error) {
	return Authenticator.Open(context.Background())
}

func openCredentials() (*api.Authenticator, error) {
	cred, err := os.Open(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("opening credentials: %w", err)
	}
	return api.NewAuthenticator(cred)
}

func openToken(ctx context.Context, auth *api.Authenticator) (*api.GmailAPI, error) {
	token, err := os.Open(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("missing or invalid cached token: %w", err)
	}

	return auth.API(ctx, token)
}
