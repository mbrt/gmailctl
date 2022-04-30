package localcred

import (
	"fmt"
	"net"
	"net/http"
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
			Addr: ":0",
			Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
				if err := req.ParseForm(); err != nil {
					resp.Write([]byte(err.Error()))
					resp.WriteHeader(400)
					return
				}
				state := req.Form.Get("state")
				if state != expectedState {
					resp.Write([]byte("Invalid state"))
					resp.WriteHeader(400)
					return
				}
				code := req.Form.Get("code")
				if code == "" {
					resp.Write([]byte("Missing code in request"))
					resp.WriteHeader(400)
					return
				}
				ch <- code
				close(ch)

				resp.Header().Add("Content-Type", "text/plain")
				resp.WriteHeader(200)
				resp.Write([]byte(successMsg))
			}),
		},
	}
}

func (s *oauth2Server) Start() (string, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", err
	}
	go s.srv.Serve(l)
	return fmt.Sprintf("localhost:%d", l.Addr().(*net.TCPAddr).Port), nil
}

func (s *oauth2Server) Close() {
	_ = s.srv.Close()
}

func (s *oauth2Server) WaitForCode() string {
	return <-s.code
}
