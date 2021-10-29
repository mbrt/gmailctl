// package fakegmail provides a fake implementation of Gmail APIs.
//
// Inspired by https://github.com/googleapis/google-cloud-go
// internal/examples/fake/fake_test.go.
// See https://github.com/googleapis/google-api-go-client/blob/fff2ed2561e512e6e4356cc522ad8bb6c5883296/internal/examples/fake/fake_test.go
package fakegmail

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"testing"

	"github.com/gorilla/mux"
	gmailv1 "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"github.com/mbrt/gmailctl/internal/stringset"
)

var defaultLabels = stringset.New(
	"INBOX",
	"TRASH",
	"IMPORTANT",
	"UNREAD",
	"SPAM",
	"STARRED",
	"CATEGORY_PERSONAL",
	"CATEGORY_SOCIAL",
	"CATEGORY_UPDATES",
	"CATEGORY_FORUMS",
	"CATEGORY_PROMOTIONS",
)

// NewService returns a fake server that implements GMail APIs.
func NewService(ctx context.Context, t *testing.T) *gmailv1.Service {
	t.Helper()

	srv := &gmailServer{
		gmail{
			labels:     make(map[string]*gmailv1.Label),
			labelNames: stringset.New(),
			filters:    make(map[string]*gmailv1.Filter),
			m:          &sync.Mutex{},
		},
	}
	for _, dl := range defaultLabels.ToSlice() {
		srv.labels[dl] = nil
	}

	mux := mux.NewRouter().StrictSlash(true)
	mux.Handle("/gmail/v1/users/me/labels",
		http.HandlerFunc(srv.HandleLabelsGet)).Methods(http.MethodGet)
	mux.Handle("/gmail/v1/users/me/labels",
		http.HandlerFunc(srv.HandleLabelsPost)).Methods(http.MethodPost)
	mux.Handle("/gmail/v1/users/me/labels/{id}",
		http.HandlerFunc(srv.HandleLabelDelete)).Methods(http.MethodDelete)
	mux.Handle("/gmail/v1/users/me/labels/{id}",
		http.HandlerFunc(srv.HandleLabelPatch)).Methods(http.MethodPatch)
	mux.Handle("/gmail/v1/users/me/settings/filters",
		http.HandlerFunc(srv.HandleFiltersGet)).Methods(http.MethodGet)
	mux.Handle("/gmail/v1/users/me/settings/filters",
		http.HandlerFunc(srv.HandleFiltersPost)).Methods(http.MethodPost)
	mux.Handle("/gmail/v1/users/me/settings/filters/{id}",
		http.HandlerFunc(srv.HandleFilterDelete)).Methods(http.MethodDelete)

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	svc, err := gmailv1.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(ts.URL))
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

type gmailServer struct {
	gmail
}

func (g *gmailServer) HandleLabelsGet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	resp := &gmailv1.ListLabelsResponse{
		Labels: g.Labels(),
	}
	writeResponse(w, resp)
}

func (g *gmailServer) HandleLabelsPost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req gmailv1.Label
	if err := readRequest(&req, r.Body); err != nil {
		writeErr(w, err)
		return
	}
	res, err := g.CreateLabel(&req)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeResponse(w, res)
}

func (g *gmailServer) HandleLabelDelete(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id := mux.Vars(r)["id"]
	if err := g.DeleteLabel(id); err != nil {
		writeErr(w, err)
	}
}

func (g *gmailServer) HandleLabelPatch(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	id := mux.Vars(r)["id"]
	var req gmailv1.Label
	if err := readRequest(&req, r.Body); err != nil {
		writeErr(w, err)
		return
	}
	req.Id = id
	res, err := g.UpdateLabel(&req)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeResponse(w, res)
}

func (g *gmailServer) HandleFiltersGet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	resp := &gmailv1.ListFiltersResponse{
		Filter: g.Filters(),
	}
	writeResponse(w, resp)
}

func (g *gmailServer) HandleFiltersPost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req gmailv1.Filter
	if err := readRequest(&req, r.Body); err != nil {
		writeErr(w, err)
		return
	}
	res, err := g.CreateFilter(&req)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeResponse(w, res)
}

func (g *gmailServer) HandleFilterDelete(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id := mux.Vars(r)["id"]
	if err := g.DeleteFilter(id); err != nil {
		writeErr(w, err)
	}
}

func readRequest(res interface{}, r io.Reader) error {
	return json.NewDecoder(r).Decode(res)
}

func writeResponse(w http.ResponseWriter, r interface{}) {
	b, err := json.Marshal(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to marshal response: %v", err), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(b); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func writeErr(w http.ResponseWriter, err error) {
	var se statusError
	if errors.As(err, &se) {
		http.Error(w, se.Err.Error(), se.StatusCode)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

type gmail struct {
	labels      map[string]*gmailv1.Label
	labelNames  stringset.Set
	labelNextID int
	filters     map[string]*gmailv1.Filter
	m           *sync.Mutex
}

func (g *gmail) Labels() []*gmailv1.Label {
	g.m.Lock()
	defer g.m.Unlock()

	var res []*gmailv1.Label
	for id, l := range g.labels {
		if defaultLabels.Has(id) {
			// Skip default labels.
			continue
		}
		res = append(res, l)
	}
	// To make things more deterministic.
	sort.Slice(res, func(i, j int) bool {
		return res[i].Id < res[j].Id
	})

	return res
}

func (g *gmail) CreateLabel(l *gmailv1.Label) (*gmailv1.Label, error) {
	g.m.Lock()
	defer g.m.Unlock()

	if l.Id != "" {
		return nil, statusError{http.StatusBadRequest, fmt.Errorf("cannot create label with non empty ID. Got: %q", l.Id)}
	}
	if g.labelNames.Has(l.Name) {
		return nil, statusError{http.StatusBadRequest, fmt.Errorf("label with name %q is already present", l.Name)}
	}
	l.Id = fmt.Sprintf("ID%d", g.labelNextID)
	g.labelNextID++
	g.labels[l.Id] = l
	g.labelNames.Add(l.Name)

	return l, nil
}

func (g *gmail) DeleteLabel(id string) error {
	g.m.Lock()
	defer g.m.Unlock()

	l, ok := g.labels[id]
	if !ok {
		return statusError{404, fmt.Errorf("id %q not found", id)}
	}
	g.labelNames.Remove(l.Name)
	delete(g.labels, id)

	return nil
}

func (g *gmail) UpdateLabel(l *gmailv1.Label) (*gmailv1.Label, error) {
	g.m.Lock()
	defer g.m.Unlock()

	if _, ok := g.labels[l.Id]; !ok {
		return nil, statusError{404, fmt.Errorf("id %q not found", l.Id)}
	}
	target := g.labels[l.Id]
	if l.Color != nil {
		// Only update the color if it was passed in.
		target.Color = l.Color
	}
	if target.Name != l.Name {
		delete(g.labelNames, target.Name)
		g.labelNames.Add(l.Name)
		target.Name = l.Name
	}

	return target, nil
}

func (g *gmail) Filters() []*gmailv1.Filter {
	g.m.Lock()
	defer g.m.Unlock()

	var res []*gmailv1.Filter
	for _, f := range g.filters {
		res = append(res, f)
	}
	// To make things more deterministic.
	sort.Slice(res, func(i, j int) bool {
		return res[i].Id < res[j].Id
	})

	return res
}

func (g *gmail) CreateFilter(f *gmailv1.Filter) (*gmailv1.Filter, error) {
	g.m.Lock()
	defer g.m.Unlock()

	h := hashFilter(f)
	if _, ok := g.filters[h]; ok {
		return nil, statusError{http.StatusBadRequest, fmt.Errorf("filter with hash %q already exists", h)}
	}
	// Check whether referenced labels exist.
	for _, id := range f.Action.AddLabelIds {
		if _, ok := g.labels[id]; !ok {
			return nil, statusError{http.StatusBadRequest, fmt.Errorf("invalid label %q", id)}
		}
	}
	f.Id = h
	g.filters[h] = f

	return f, nil
}

func (g *gmail) DeleteFilter(id string) error {
	g.m.Lock()
	defer g.m.Unlock()

	if _, ok := g.filters[id]; !ok {
		return statusError{404, fmt.Errorf("id %q not found", id)}
	}
	delete(g.filters, id)

	return nil
}

type statusError struct {
	StatusCode int
	Err        error
}

func (s statusError) Error() string {
	return fmt.Sprintf("%v (status %d)", s.Err, s.StatusCode)
}

func hashFilter(gf *gmailv1.Filter) string {
	// We want to hash only criteria and action and not the rest,
	// especially not the ID.
	f := struct {
		c gmailv1.FilterCriteria
		a gmailv1.FilterAction
	}{
		c: *gf.Criteria,
		a: *gf.Action,
	}
	return hashStruct(f)
}

func hashStruct(a interface{}) string {
	h := sha256.New()
	if _, err := h.Write([]byte(fmt.Sprintf("%#v", a))); err != nil {
		// This should be unreachable.
		panic(err)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
