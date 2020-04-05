package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/mbrt/gmailctl/pkg/api"
)

func openAPI() (*api.GmailAPI, error) {
	auth, err := openCredentials()
	if err != nil {
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}
	return openToken(auth)
}

func openCredentials() (*api.Authenticator, error) {
	cred, err := os.Open(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("opening credentials: %w", err)
	}
	return api.NewAuthenticator(cred)
}

func openToken(auth *api.Authenticator) (*api.GmailAPI, error) {
	token, err := os.Open(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("missing or invalid cached token: %w", err)
	}

	return auth.API(context.Background(), token)
}
