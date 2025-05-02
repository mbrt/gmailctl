package cmd

import (
	"context"
	"fmt"

	"google.golang.org/api/gmail/v1"

	"github.com/mbrt/gmailctl/internal/engine/api"
)

// APIProvider is the APIProvider used by all gmailctl commands.
var APIProvider GmailAPIProvider

// GmailAPIProvider is the integration point between gmailctl commands and GMail
// APIs providers.
type GmailAPIProvider interface {
	// Service returns the GMail API service.
	Service(ctx context.Context, cfgDir string) (*gmail.Service, error)
	// ResetConfig cleans up the configuration.
	ResetConfig(cfgDir string) error
	// InitConfig initializes the configuration.
	InitConfig(cfgDir string, port int) error
}

// APIKeyProvider is the interface implemented by API providers with
// an API key.
type APIKeyProvider interface {
	// APIKey returns the API key used to authenticate with GMail APIs.
	// Returns an empty string if not available.
	APIKey() string
}

// TokenRefresher allows to refresh API tokens if they are expired.
type TokenRefresher interface {
	// RefreshToken refreshes the token if it is expired.
	RefreshToken(ctx context.Context, cfgDir string, port int) error
}

func openAPI() (*api.GmailAPI, error) {
	srv, err := APIProvider.Service(context.Background(), cfgDir)
	if err != nil {
		return nil, fmt.Errorf("in Authenticator.Service: %w", err)
	}
	if kprov, ok := APIProvider.(APIKeyProvider); ok {
		return api.NewWithAPIKey(srv, kprov.APIKey()), nil
	}
	return api.NewFromService(srv), nil
}
