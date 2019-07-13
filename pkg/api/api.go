package api

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
	gmailv1 "google.golang.org/api/gmail/v1"

	exportapi "github.com/mbrt/gmailctl/pkg/export/api"
	"github.com/mbrt/gmailctl/pkg/filter"
)

const (
	gmailUser = "me"

	labelTypeSystem = "system"
)

// GmailAPI is a wrapper around the Gmail APIs.
type GmailAPI struct {
	service  *gmailv1.Service
	labelmap *exportapi.LabelMap // don't use without locking
	mutex    *sync.Mutex
}

// ListFilters returns the list of Gmail filters in the settings.
func (g *GmailAPI) ListFilters() (filter.Filters, error) {
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

// DeleteFilters deletes all the given filter IDs.
func (g *GmailAPI) DeleteFilters(ids []string) error {
	for _, id := range ids {
		err := g.service.Users.Settings.Filters.Delete(gmailUser, id).Do()
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error deleting filter '%s'", id))
		}
	}
	return nil
}

// AddFilters creates the given filters.
func (g *GmailAPI) AddFilters(fs filter.Filters) error {
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

// ListLabels lists the user labels.
func (g *GmailAPI) ListLabels() ([]filter.Label, error) {
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

// LabelMap returns a map of label ids and names.
//
// Deprecated: build the LabelMap directly with the list of labels.
func (g *GmailAPI) LabelMap() (exportapi.LabelMap, error) {
	_, err := g.refreshLabelMap()
	if err != nil {
		return exportapi.LabelMap{}, err
	}
	return *g.labelmap, nil
}

func (g *GmailAPI) getLabelMap() (exportapi.LabelMap, error) {
	if err := g.initLabelMap(); err != nil {
		return exportapi.LabelMap{}, errors.Wrap(err, "cannot get list of labels")
	}
	g.mutex.Lock()
	res := *g.labelmap
	g.mutex.Unlock()

	return res, nil
}

func (g *GmailAPI) initLabelMap() error {
	g.mutex.Lock()
	needInit := (g.labelmap == nil)
	g.mutex.Unlock()

	if !needInit {
		return nil
	}

	_, err := g.refreshLabelMap()
	return err
}

func (g *GmailAPI) refreshLabelMap() (map[string]string, error) {
	idNameMap, err := g.fetchLabelsIDNameMap()
	if err != nil {
		return nil, err
	}
	newLabelMap := exportapi.NewLabelMap(idNameMap)

	g.mutex.Lock()
	g.labelmap = &newLabelMap
	g.mutex.Unlock()

	return idNameMap, nil
}

func (g *GmailAPI) fetchLabelsIDNameMap() (map[string]string, error) {
	apires, err := g.service.Users.Labels.List(gmailUser).Do()
	if err != nil {
		return nil, err
	}
	idNameMap := map[string]string{}
	for _, label := range apires.Labels {
		// We are only interested in user labels.
		if label.Type != labelTypeSystem {
			idNameMap[label.Id] = label.Name
		}
	}
	return idNameMap, nil
}
