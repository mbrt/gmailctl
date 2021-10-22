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

	gmailv1 "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// NewService returns a fake server that implements GMail APIs.
func NewService(ctx context.Context, t *testing.T) *gmailv1.Service {
	t.Helper()

	srv := &gmailServer{
		g: gmail{
			labels: make(map[string]*gmailv1.Label),
			m:      &sync.Mutex{},
		},
	}
	mux := http.NewServeMux()
	mux.Handle("/gmail/v1/users/me/labels", http.HandlerFunc(srv.handleLabels))

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	svc, err := gmailv1.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(ts.URL))
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

type gmailServer struct {
	g gmail
}

func (g *gmailServer) handleLabels(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case http.MethodGet:
		resp := &gmailv1.ListLabelsResponse{
			Labels: g.g.Labels(),
		}
		writeResponse(w, resp)

	case http.MethodPost:
		var req gmailv1.Label
		if err := readRequest(&req, r.Body); err != nil {
			writeErr(w, err)
			return
		}
		if err := g.g.CreateLabel(req); err != nil {
			writeErr(w, err)
		}

	default:
		writeErr(w, fmt.Errorf("unsupported method %q", r.Method))
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
	w.Write(b)
}

func writeErr(w http.ResponseWriter, err error) {
	var se statusErr
	if errors.As(err, &se) {
		http.Error(w, se.Err.Error(), se.StatusCode)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

type gmail struct {
	labels      map[string]*gmailv1.Label
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

func (g *gmail) CreateLabel(l gmailv1.Label) error {
	g.m.Lock()
	defer g.m.Unlock()

	if _, ok := g.labels[l.Id]; ok {
		return fmt.Errorf("label with ID %q is already present", l.Id)
	}
	// TODO: Check that the name doesn't exist.
	l.Id = fmt.Sprintf("ID%d", g.labelNextID)
	g.labelNextID++
	g.labels[l.Id] = &l
	return nil
}

type statusErr struct {
	StatusCode int
	Err        error
}

func (s statusErr) Error() string {
	return fmt.Sprintf("%v (status %d)", s.Err, s.StatusCode)
}
