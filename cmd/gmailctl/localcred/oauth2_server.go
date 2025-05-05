package localcred

import (
	"fmt"
	"html"
	"net"
	"net/http"
	"time"
)

const successMsg = `Successfully authenticated with Google OAuth2.
You may now close this page and return to the terminal.
`

type oauth2Server struct {
	srv  *http.Server
	code chan string
}

func newOauth2Server(expectedState string) *oauth2Server {
	ch := make(chan string)

	return &oauth2Server{
		code: ch,
		srv: &http.Server{
			Addr:              ":0",
			ReadHeaderTimeout: 10 * time.Second,
			Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
				if err := req.ParseForm(); err != nil {
					writeError(resp, err, http.StatusBadRequest)
					return
				}
				state := req.Form.Get("state")
				if state != expectedState {
					writeError(resp, fmt.Errorf("invalid state: %q", state), http.StatusBadRequest)
					return
				}
				code := req.Form.Get("code")
				if code == "" {
					writeError(resp, fmt.Errorf("missing code"), http.StatusBadRequest)
					return
				}
				ch <- code
				close(ch)

				resp.Header().Add("Content-Type", "text/plain")
				resp.WriteHeader(200)
				_, _ = resp.Write([]byte(successMsg))
			}),
		},
	}
}

// Start the oauth2Server asynchronously.
func (s *oauth2Server) Start(port int) (string, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port)) //nolint:gosec
	if err != nil {
		return "", err
	}
	go func() {
		// This always returns an error when the server is closed.
		_ = s.srv.Serve(l)
	}()
	return fmt.Sprintf("localhost:%d", l.Addr().(*net.TCPAddr).Port), nil
}

// Close shuts down the oauth2Server.
func (s *oauth2Server) Close() {
	_ = s.srv.Close()
}

// WaitForCode waits for the oauth2Server to receive a code.
// This is a blocking call. Errors are ignored.
func (s *oauth2Server) WaitForCode() string {
	return <-s.code
}

func writeError(resp http.ResponseWriter, err error, code int) {
	resp.WriteHeader(code)
	_, _ = resp.Write([]byte(html.EscapeString(fmt.Sprintf("Error: %v", err))))
}
