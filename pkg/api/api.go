package api

import (
	"fmt"

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
	service *gmailv1.Service
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
	apires, err := g.service.Users.Labels.List(gmailUser).Do()
	if err != nil {
		return nil, err
	}

	var res []filter.Label

	for _, label := range apires.Labels {
		// We are only interested in user labels.
		if label.Type == labelTypeSystem {
			continue
		}

		var color *filter.LabelColor
		if label.Color != nil {
			color = &filter.LabelColor{
				Background: label.Color.BackgroundColor,
				Text:       label.Color.TextColor,
			}
		}

		res = append(res, filter.Label{
			ID:          label.Id,
			Name:        label.Name,
			Color:       color,
			NumMessages: int(label.MessagesTotal),
		})
	}

	return res, nil
}

func (g *GmailAPI) getLabelMap() (exportapi.LabelMap, error) {
	labels, err := g.ListLabels()
	if err != nil {
		return exportapi.LabelMap{}, err
	}
	return exportapi.NewLabelMap(labels), nil
}
