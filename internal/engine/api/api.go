package api

import (
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/googleapi"

	"github.com/mbrt/gmailctl/internal/engine/export/api"
	"github.com/mbrt/gmailctl/internal/engine/filter"
	"github.com/mbrt/gmailctl/internal/engine/label"
	"github.com/mbrt/gmailctl/internal/errors"
)

const (
	gmailUser       = "me"
	labelTypeSystem = "system"
	labelDocsURL    = "https://developers.google.com/gmail/api/v1/reference/users/labels#resource"
	authExpiredURL  = "https://github.com/mbrt/gmailctl#oauth2-authentication-errors"
)

// NewFromService creates a new GmailAPI instance from the given Gmail service.
func NewFromService(s *gmail.Service) *GmailAPI {
	return &GmailAPI{s, nil}
}

// NewWithAPIKey creates a new GmailAPI instance from the given Gmail service and API key.
func NewWithAPIKey(s *gmail.Service, key string) *GmailAPI {
	return &GmailAPI{s, []googleapi.CallOption{keyOption(key)}}
}

// GmailAPI is a wrapper around the Gmail APIs.
type GmailAPI struct {
	service *gmail.Service
	opts    []googleapi.CallOption
}

// ListFilters returns the list of Gmail filters in the settings.
func (g *GmailAPI) ListFilters() (filter.Filters, error) {
	lmap, err := g.getLabelMap()
	if err != nil {
		return nil, err
	}

	apires, err := g.service.Users.Settings.Filters.List(gmailUser).Do(g.opts...)
	if err != nil {
		return nil, annotateError(err)
	}
	return api.Import(apires.Filter, lmap)
}

// DeleteFilters deletes all the given filter IDs.
func (g *GmailAPI) DeleteFilters(ids []string) error {
	for _, id := range ids {
		err := g.service.Users.Settings.Filters.Delete(gmailUser, id).Do(g.opts...)
		if err != nil {
			return fmt.Errorf("deleting filter %q: %w", id, annotateError(err))
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

	gfilters, err := api.Export(fs, lmap)
	if err != nil {
		return err
	}

	for i, gfilter := range gfilters {
		_, err = g.service.Users.Settings.Filters.Create(gmailUser, gfilter).Do(g.opts...)
		if err != nil {
			return fmt.Errorf("creating filter %d: %w", i, annotateError(err))
		}
	}

	return nil
}

// ListLabels lists the user labels.
func (g *GmailAPI) ListLabels() (label.Labels, error) {
	apires, err := g.service.Users.Labels.List(gmailUser).Do(g.opts...)
	if err != nil {
		return nil, annotateError(err)
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
		err := g.service.Users.Labels.Delete(gmailUser, id).Do(g.opts...)
		if err != nil {
			return fmt.Errorf("deleting label %q: %w", id, annotateError(err))
		}
	}
	return nil

}

// AddLabels creates the given labels.
func (g *GmailAPI) AddLabels(lbs label.Labels) error {
	for _, lb := range lbs {
		_, err := g.service.Users.Labels.Create(gmailUser, labelToGmailAPI(lb)).Do(g.opts...)
		if err != nil {
			return annotateError(fmt.Errorf("creating label %q: %w", lb.Name, err))
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
		_, err := g.service.Users.Labels.Patch(gmailUser, lb.ID, labelToGmailAPI(lb)).Do(g.opts...)
		if err != nil {
			return annotateError(fmt.Errorf("patching label %q: %w", lb.Name, err))
		}
	}
	return nil
}

func (g *GmailAPI) getLabelMap() (api.LabelMap, error) {
	labels, err := g.ListLabels()
	if err != nil {
		return api.LabelMap{}, err
	}
	return api.NewLabelMap(labels), nil
}

func labelToGmailAPI(lb label.Label) *gmail.Label {
	var color *gmail.LabelColor
	if lb.Color != nil {
		color = &gmail.LabelColor{
			BackgroundColor: lb.Color.Background,
			TextColor:       lb.Color.Text,
		}
	}
	return &gmail.Label{
		Name:  lb.Name,
		Color: color,
	}
}

func annotateError(err error) error {
	var oerr *oauth2.RetrieveError
	if errors.As(err, &oerr) {
		if strings.Contains(string(oerr.Body), "invalid_grant") {
			return errors.WithDetails(err,
				"Possible expired token. Try refreshing with `gmailctl init --refresh-expired`",
				"More help at "+authExpiredURL,
			)
		}
	}

	var gerr *googleapi.Error
	if !errors.As(err, &gerr) {
		return err
	}
	if gerr.Code != http.StatusBadRequest {
		return err
	}
	for _, e := range gerr.Errors {
		if e.Reason == "invalidArgument" && strings.Contains(e.Message, "color palette") {
			return errors.WithDetails(err,
				fmt.Sprintf("See the allowed set of color values here: %s", labelDocsURL))
		}
	}
	return err
}

type keyOption string

func (k keyOption) Get() (string, string) {
	return "key", string(k)
}
