package cmd

import (
	"context"
	"os"

	"github.com/pkg/errors"

	"github.com/mbrt/gmailctl/pkg/api"
)

func openAPI() (*api.GmailAPI, error) {
	auth, err := openCredentials()
	if err != nil {
		return nil, errors.Wrap(err, "invalid credentials")
	}
	return openToken(auth)
}

func openCredentials() (*api.Authenticator, error) {
	cred, err := os.Open(credentialsPath)
	if err != nil {
		return nil, errors.Wrap(err, "cannot open credentials")
	}
	return api.NewAuthenticator(cred)
}

func openToken(auth *api.Authenticator) (*api.GmailAPI, error) {
	token, err := os.Open(tokenPath)
	if err != nil {
		return nil, errors.Wrap(err, "missing or invalid cached token")
	}

	return auth.API(context.Background(), token)
}
