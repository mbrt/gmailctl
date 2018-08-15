package api

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"sync"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gmailv1 "google.golang.org/api/gmail/v1"

	exportapi "github.com/mbrt/gmailfilter/pkg/export/api"
	"github.com/mbrt/gmailfilter/pkg/filter"
)

const gmailUser = "me"

// GmailAPI is a wrapper around the Gmail APIs.
type GmailAPI interface {
	ListFilters() (filter.Filters, error)
	DeleteFilters(ids []string) error
	AddFilters(fs filter.Filters) error
	ListLabels() ([]filter.Label, error)
}

// NewGmailAPI creates a new GmailAPI object by giving credentials and token JSON file contents.
func NewGmailAPI(ctx context.Context, credentials, token io.Reader) (GmailAPI, error) {
	cfg, err := clientFromCredentials(credentials)
	if err != nil {
		return nil, errors.Wrap(err, "error creating config from credentials")
	}
	tok, err := parseToken(token)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding token")
	}

	client := cfg.Client(ctx, tok)
	srv, err := gmailv1.New(client)
	if err != nil {
		return nil, errors.Wrap(err, "error creating gmail client")
	}

	// Lazy load the LabelMap
	return &gmailAPI{srv, nil, &sync.Mutex{}}, nil
}

func clientFromCredentials(credentials io.Reader) (*oauth2.Config, error) {
	credBytes, err := ioutil.ReadAll(credentials)
	if err != nil {
		return nil, errors.Wrap(err, "error reading credentials")
	}
	return google.ConfigFromJSON(credBytes, gmailv1.GmailSettingsBasicScope)
}

func parseToken(token io.Reader) (*oauth2.Token, error) {
	tok := &oauth2.Token{}
	err := json.NewDecoder(token).Decode(tok)
	return tok, err
}

type gmailAPI struct {
	service  *gmailv1.Service
	labelmap exportapi.LabelMap // don't use without locking
	mutex    *sync.Mutex
}

func (g *gmailAPI) ListFilters() (filter.Filters, error) {
	lmap, err := g.getLabelMap()
	if err != nil {
		return nil, err
	}

	apires, err := g.service.Users.Settings.Filters.List(gmailUser).Do()
	if err != nil {
		return nil, err
	}
	return exportapi.DefaulImporter().Import(apires.Filter, lmap)
}

func (g *gmailAPI) DeleteFilters(ids []string) error {
	panic("not implemented")
}

func (g *gmailAPI) AddFilters(fs filter.Filters) error {
	panic("not implemented")
}

func (g *gmailAPI) ListLabels() ([]filter.Label, error) {
	idNameMap, err := g.refreshLabelMap()
	if err != nil {
		return nil, err
	}

	i := 0
	res := make([]filter.Label, len(idNameMap))
	for id, name := range idNameMap {
		res[i] = filter.Label{ID: id, Name: name}
		i++
	}

	return res, nil
}

func (g *gmailAPI) getLabelMap() (exportapi.LabelMap, error) {
	if err := g.initLabelMap(); err != nil {
		return nil, errors.Wrap(err, "cannot get list of labels")
	}
	g.mutex.Lock()
	res := g.labelmap
	g.mutex.Unlock()

	return res, nil
}

func (g *gmailAPI) initLabelMap() error {
	g.mutex.Lock()
	needInit := (g.labelmap == nil)
	g.mutex.Unlock()

	if !needInit {
		return nil
	}

	_, err := g.refreshLabelMap()
	return err
}

func (g *gmailAPI) refreshLabelMap() (map[string]string, error) {
	idNameMap, err := g.fetchLabelsIDNameMap()
	if err != nil {
		return nil, err
	}
	newLabelMap := exportapi.NewDefaultLabelMap(idNameMap)

	g.mutex.Lock()
	g.labelmap = newLabelMap
	g.mutex.Unlock()

	return idNameMap, nil
}

func (g *gmailAPI) fetchLabelsIDNameMap() (map[string]string, error) {
	apires, err := g.service.Users.Labels.List(gmailUser).Do()
	if err != nil {
		return nil, err
	}
	idNameMap := map[string]string{}
	for _, label := range apires.Labels {
		idNameMap[label.Id] = label.Name
	}
	return idNameMap, nil
}
