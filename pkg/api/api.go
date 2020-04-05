package api

import (
	"fmt"

	gmailv1 "google.golang.org/api/gmail/v1"

	exportapi "github.com/mbrt/gmailctl/pkg/export/api"
	"github.com/mbrt/gmailctl/pkg/filter"
	"github.com/mbrt/gmailctl/pkg/label"
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
	return exportapi.Import(apires.Filter, lmap)
}

// DeleteFilters deletes all the given filter IDs.
func (g *GmailAPI) DeleteFilters(ids []string) error {
	for _, id := range ids {
		err := g.service.Users.Settings.Filters.Delete(gmailUser, id).Do()
		if err != nil {
			return fmt.Errorf("deleting filter %q: %w", id, err)
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

	gfilters, err := exportapi.Export(fs, lmap)
	if err != nil {
		return err
	}

	for i, gfilter := range gfilters {
		_, err = g.service.Users.Settings.Filters.Create(gmailUser, gfilter).Do()
		if err != nil {
			return fmt.Errorf("creating filter %d: %w", i, err)
		}
	}

	return nil
}

// ListLabels lists the user labels.
func (g *GmailAPI) ListLabels() (label.Labels, error) {
	apires, err := g.service.Users.Labels.List(gmailUser).Do()
	if err != nil {
		return nil, err
	}

	var res label.Labels

	for _, lb := range apires.Labels {
		// We are only interested in user labels.
		if lb.Type == labelTypeSystem {
			continue
		}

		var color *label.Color
		if lb.Color != nil {
			color = &label.Color{
				Background: lb.Color.BackgroundColor,
				Text:       lb.Color.TextColor,
			}
		}

		res = append(res, label.Label{
			ID:    lb.Id,
			Name:  lb.Name,
			Color: color,
		})
	}

	return res, nil
}

// DeleteLabels deletes all the given label IDs.
func (g *GmailAPI) DeleteLabels(ids []string) error {
	for _, id := range ids {
		err := g.service.Users.Labels.Delete(gmailUser, id).Do()
		if err != nil {
			return fmt.Errorf("deleting label %q: %w", id, err)
		}
	}
	return nil

}

// AddLabels creates the given labels.
func (g *GmailAPI) AddLabels(lbs label.Labels) error {
	for _, lb := range lbs {
		_, err := g.service.Users.Labels.Create(gmailUser, labelToGmailAPI(lb)).Do()
		if err != nil {
			return fmt.Errorf("creating label %q: %w", lb.Name, err)
		}
	}
	return nil
}

// UpdateLabels modifies the given labels.
//
// The label ID is required for the edit to be successful.
func (g *GmailAPI) UpdateLabels(lbs label.Labels) error {
	for _, lb := range lbs {
		if lb.ID == "" {
			return fmt.Errorf("label %q has empty ID", lb.Name)
		}
		_, err := g.service.Users.Labels.Patch(gmailUser, lb.ID, labelToGmailAPI(lb)).Do()
		if err != nil {
			return fmt.Errorf("patching label %q: %w", lb.Name, err)
		}
	}
	return nil
}

func (g *GmailAPI) getLabelMap() (exportapi.LabelMap, error) {
	labels, err := g.ListLabels()
	if err != nil {
		return exportapi.LabelMap{}, err
	}
	return exportapi.NewLabelMap(labels), nil
}

func labelToGmailAPI(lb label.Label) *gmailv1.Label {
	var color *gmailv1.LabelColor
	if lb.Color != nil {
		color = &gmailv1.LabelColor{
			BackgroundColor: lb.Color.Background,
			TextColor:       lb.Color.Text,
		}
	}
	return &gmailv1.Label{
		Name:  lb.Name,
		Color: color,
	}
}
