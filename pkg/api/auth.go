package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gmailv1 "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// NewAuthenticator creates an Authenticator instance from credentials JSON file contents.
//
// Credentials can be obtained by creating a new OAuth client ID at the Google API console
// https://console.developers.google.com/apis/credentials.
func NewAuthenticator(credentials io.Reader) (*Authenticator, error) {
	cfg, err := clientFromCredentials(credentials)
	if err != nil {
		return nil, fmt.Errorf("creating config from credentials: %w", err)
	}
	return &Authenticator{cfg}, nil
}

// Authenticator encapsulates authentication operations for Gmail APIs.
type Authenticator struct {
	cfg *oauth2.Config
}

// API creates a GmailAPI instance from a token JSON file contents.
//
// If no token is available, AuthURL and CacheToken can be used to
// obtain one.
func (a Authenticator) API(ctx context.Context, token io.Reader) (*GmailAPI, error) {
	tok, err := parseToken(token)
	if err != nil {
		return nil, fmt.Errorf("decoding token: %w", err)
	}

	srv, err := gmailv1.NewService(ctx, option.WithTokenSource(a.cfg.TokenSource(ctx, tok)))
	if err != nil {
		return nil, fmt.Errorf("creating gmail client: %w", err)
	}

	return &GmailAPI{srv}, nil
}

// AuthURL returns the URL the user has to visit to authorize the
// application and obtain an auth code.
func (a Authenticator) AuthURL() string {
	return a.cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
}

// CacheToken creates and caches a token JSON file from an auth code.
//
// The token can be subsequently used to authorize a GmailAPI instance.
func (a Authenticator) CacheToken(ctx context.Context, authCode string, token io.Writer) error {
	tok, err := a.cfg.Exchange(ctx, authCode)
	if err != nil {
		return fmt.Errorf("unable to retrieve token from web: %w", err)
	}
	return json.NewEncoder(token).Encode(tok)
}

func clientFromCredentials(credentials io.Reader) (*oauth2.Config, error) {
	credBytes, err := ioutil.ReadAll(credentials)
	if err != nil {
		return nil, fmt.Errorf("reading credentials: %w", err)
	}
	return google.ConfigFromJSON(credBytes,
		gmailv1.GmailSettingsBasicScope,
		gmailv1.GmailMetadataScope,
		gmailv1.GmailLabelsScope,
	)
}

func parseToken(token io.Reader) (*oauth2.Token, error) {
	tok := &oauth2.Token{}
	err := json.NewDecoder(token).Decode(tok)
	return tok, err
}
