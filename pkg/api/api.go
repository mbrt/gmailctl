package api

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gmailv1 "google.golang.org/api/gmail/v1"

	"github.com/mbrt/gmailfilter/pkg/filter"
)

// GmailAPI is a wrapper around the Gmail APIs.
type GmailAPI interface {
	ListFilters() (filter.Filters, error)
	DeleteFilters(ids []string) error
	AddFilters(fs filter.Filters) error
}

// NewGmailAPI creates a new GmailAPI object by giving credentials and token JSON file contents.
func NewGmailAPI(ctx context.Context, credentials, token io.Reader) (GmailAPI, error) {
	// TODO: refactor
	credBytes, err := ioutil.ReadAll(credentials)
	if err != nil {
		return nil, errors.Wrap(err, "error reading credentials")
	}
	cfg, err := google.ConfigFromJSON(credBytes, gmailv1.GmailSettingsBasicScope)
	if err != nil {
		return nil, errors.Wrap(err, "error creating config from credentials")
	}
	tok := &oauth2.Token{}
	err = json.NewDecoder(token).Decode(tok)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding token")
	}

	client := cfg.Client(ctx, tok)
	srv, err := gmailv1.New(client)
	if err != nil {
		return nil, errors.Wrap(err, "error creating gmail client")
	}

	return gmailAPI{srv}, nil
}

type gmailAPI struct {
	service *gmailv1.Service
}

func (g gmailAPI) ListFilters() (filter.Filters, error) {
	panic("not implemented")
}

func (g gmailAPI) DeleteFilters(ids []string) error {
	panic("not implemented")
}

func (g gmailAPI) AddFilters(fs filter.Filters) error {
	panic("not implemented")
}
