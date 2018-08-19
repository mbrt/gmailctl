package api

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/pkg/errors"
	gmailv1 "google.golang.org/api/gmail/v1"
)

// NewAuthenticator creates an Authenticator instance from credentials JSON file contents.
//
// Credentials can be obtained by creating a new OAuth client ID at the Google API console
// https://console.developers.google.com/apis/credentials.
func NewAuthenticator(credentials io.Reader) (Authenticator, error) {
	cfg, err := clientFromCredentials(credentials)
	if err != nil {
		return nil, errors.Wrap(err, "error creating config from credentials")
	}
	return authenticator{cfg}, nil
}

// Authenticator encapsulates authentication operations for Gmail APIs.
type Authenticator interface {
	// API creates a GmailAPI instance from a token JSON file contents.
	//
	// If no token is available, AuthURL and CacheToken can be used to
	// obtain one.
	API(ctx context.Context, token io.Reader) (GmailAPI, error)

	// AuthURL returns the URL the user has to visit to authorize the
	// application and obtain an auth code.
	AuthURL() string
	// CacheToken creates and caches a token JSON file from an auth code.
	//
	// The token can be subsequently used to authorize a GmailAPI instance.
	CacheToken(ctx context.Context, authCode string, token io.Writer) error
}

type authenticator struct {
	cfg *oauth2.Config
}

func (a authenticator) API(ctx context.Context, token io.Reader) (GmailAPI, error) {
	tok, err := parseToken(token)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding token")
	}

	client := a.cfg.Client(ctx, tok)
	srv, err := gmailv1.New(client)
	if err != nil {
		return nil, errors.Wrap(err, "error creating gmail client")
	}

	// Lazy load the LabelMap
	return &gmailAPI{srv, nil, &sync.Mutex{}}, nil
}

func (a authenticator) AuthURL() string {
	return a.cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
}

func (a authenticator) CacheToken(ctx context.Context, authCode string, token io.Writer) error {
	tok, err := a.cfg.Exchange(ctx, authCode)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve token from web")
	}
	return json.NewEncoder(token).Encode(tok)
}

func clientFromCredentials(credentials io.Reader) (*oauth2.Config, error) {
	credBytes, err := ioutil.ReadAll(credentials)
	if err != nil {
		return nil, errors.Wrap(err, "error reading credentials")
	}
	return google.ConfigFromJSON(credBytes,
		gmailv1.GmailSettingsBasicScope,
		gmailv1.GmailMetadataScope,
	)
}

func parseToken(token io.Reader) (*oauth2.Token, error) {
	tok := &oauth2.Token{}
	err := json.NewDecoder(token).Decode(tok)
	return tok, err
}
