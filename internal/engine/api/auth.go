package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
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
	return &Authenticator{
		State: generateOauthState(),
		cfg:   cfg,
	}, nil
}

// Authenticator encapsulates authentication operations for Gmail APIs.
type Authenticator struct {
	State string
	cfg   *oauth2.Config
}

// Service creates a Gmail API service from a token JSON file contents.
//
// If no token is available, AuthURL and CacheToken can be used to
// obtain one.
func (a Authenticator) Service(ctx context.Context, token io.Reader) (*gmail.Service, error) {
	tok, err := parseToken(token)
	if err != nil {
		return nil, fmt.Errorf("decoding token: %w", err)
	}
	return gmail.NewService(ctx, option.WithTokenSource(a.cfg.TokenSource(ctx, tok)))
}

// API creates a GmailAPI instance from a token JSON file contents.
//
// If no token is available, AuthURL and CacheToken can be used to
// obtain one.
func (a Authenticator) API(ctx context.Context, token io.Reader) (*GmailAPI, error) {
	srv, err := a.Service(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("creating gmail client: %w", err)
	}
	return NewFromService(srv), nil
}

// AuthURL returns the URL the user has to visit to authorize the
// application and obtain an auth code.
func (a Authenticator) AuthURL(redirectURL string) string {
	a.cfg.RedirectURL = redirectURL
	return a.cfg.AuthCodeURL(a.State, oauth2.AccessTypeOffline)
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
	credBytes, err := io.ReadAll(credentials)
	if err != nil {
		return nil, fmt.Errorf("reading credentials: %w", err)
	}
	return google.ConfigFromJSON(credBytes,
		gmail.GmailSettingsBasicScope,
		gmail.GmailLabelsScope,
	)
}

func parseToken(token io.Reader) (*oauth2.Token, error) {
	tok := &oauth2.Token{}
	err := json.NewDecoder(token).Decode(tok)
	return tok, err
}

func generateOauthState() string {
	b := make([]byte, 128)
	if _, err := rand.Read(b); err != nil {
		// We can't really afford errors in secure random number generation.
		panic(err)
	}
	state := base64.URLEncoding.EncodeToString(b)
	return state
}
