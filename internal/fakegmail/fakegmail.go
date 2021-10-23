// package fakegmail provides a fake implementation of Gmail APIs.
//
// Inspired by https://github.com/googleapis/google-cloud-go
// internal/examples/fake/fake_test.go.
// See https://github.com/googleapis/google-api-go-client/blob/fff2ed2561e512e6e4356cc522ad8bb6c5883296/internal/examples/fake/fake_test.go
package fakegmail

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gorilla/mux"
	gmailv1 "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// NewService returns a fake server that implements GMail APIs.
func NewService(ctx context.Context, t *testing.T) *gmailv1.Service {
	t.Helper()

	srv := &gmailServer{
		gmail{
			labels:     make(map[string]*gmailv1.Label),
			labelNames: make(map[string]bool),
			m:          &sync.Mutex{},
		},
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
	res, err := g.CreateLabel(req)
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
	res, err := g.UpdateLabel(req)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeResponse(w, res)
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
	w.Write(b)
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
	labelNames  map[string]bool
	labelNextID int
	m           *sync.Mutex
}

func (g *gmail) Labels() []*gmailv1.Label {
	g.m.Lock()
	defer g.m.Unlock()

	var res []*gmailv1.Label
	for _, l := range g.labels {
		res = append(res, l)
	}
	return res
}

func (g *gmail) CreateLabel(l gmailv1.Label) (*gmailv1.Label, error) {
	g.m.Lock()
	defer g.m.Unlock()

	if _, ok := g.labels[l.Id]; ok {
		return nil, fmt.Errorf("label with ID %q is already present", l.Id)
	}
	if _, ok := g.labelNames[l.Name]; ok {
		return nil, fmt.Errorf("label with name %q is already present", l.Name)
	}
	l.Id = fmt.Sprintf("ID%d", g.labelNextID)
	g.labelNextID++
	g.labels[l.Id] = &l
	g.labelNames[l.Name] = true

	return &l, nil
}

func (g *gmail) DeleteLabel(id string) error {
	g.m.Lock()
	defer g.m.Unlock()

	l, ok := g.labels[id]
	if !ok {
		return statusError{404, fmt.Errorf("id %q not found", id)}
	}
	delete(g.labelNames, l.Name)
	delete(g.labels, id)

	return nil
}

func (g *gmail) UpdateLabel(l gmailv1.Label) (*gmailv1.Label, error) {
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
		g.labelNames[l.Name] = true
		target.Name = l.Name
	}

	return &l, nil
}

type statusError struct {
	StatusCode int
	Err        error
}

func (s statusError) Error() string {
	return fmt.Sprintf("%v (status %d)", s.Err, s.StatusCode)
}
