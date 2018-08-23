package api

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
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
	LabelMap() (exportapi.LabelMap, error)
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
	for _, id := range ids {
		err := g.service.Users.Settings.Filters.Delete(gmailUser, id).Do()
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error deleting filter '%s'", id))
		}
	}
	return nil
}

func (g *gmailAPI) AddFilters(fs filter.Filters) error {
	lmap, err := g.getLabelMap()
	if err != nil {
		return err
	}

	gfilters, err := exportapi.DefaulExporter().Export(fs, lmap)
	if err != nil {
		return err
	}

	for i, gfilter := range gfilters {
		_, err = g.service.Users.Settings.Filters.Create(gmailUser, gfilter).Do()
		if err != nil {
			return errors.Wrapf(err, "error creating filter '%d'", i)
		}
	}

	return nil
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

func (g *gmailAPI) LabelMap() (exportapi.LabelMap, error) {
	_, err := g.refreshLabelMap()
	if err != nil {
		return nil, err
	}
	return g.labelmap, nil
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
