package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/mbrt/gmailfilter/pkg/api"
	"github.com/mbrt/gmailfilter/pkg/config"
	"github.com/mbrt/gmailfilter/pkg/filter"
)

func readConfig(path string) (config.Config, error) {
	/* #nosec */
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return config.Config{}, errors.Wrap(err, fmt.Sprintf("cannot read %s", path))
	}

	var res config.Config
	err = yaml.Unmarshal(b, &res)
	return res, err

}

func errorf(format string, a ...interface{}) {
	/* #nosec */
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}

func fatal(format string, a ...interface{}) {
	errorf(format, a...)
	if !strings.HasSuffix(format, "\n") {
		errorf("\n") // Add newline
	}
	os.Exit(1)
}

func openAPI() (api.GmailAPI, error) {
	cred, err := os.Open("credentials.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open credentials")
	}
	auth, err := api.NewAuthenticator(cred)
	if err != nil {
		return nil, errors.Wrap(err, "invalid credentials")
	}

	token, err := os.Open("token.json")
	if err != nil {
		getTokenFromWeb(auth, "token.json")
		token, err = os.Open("token.json")
		if err != nil {
			return nil, errors.Wrap(err, "invalid cached token")
		}
	}

	return auth.API(context.Background(), token)
}

func getTokenFromWeb(auth api.Authenticator, tokenPath string) {
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\nAuthorization code: ", auth.AuthURL())

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		fatal("unable to retrieve token from web: %v", err)
	}

	if err := saveToken(tokenPath, authCode, auth); err != nil {
		fatal("unable to cache token: %v", err)
	}
}

func saveToken(path, authCode string, auth api.Authenticator) error {
	fmt.Printf("Saving credential file to %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrap(err, "unable create token file")
	}
	defer func() { _ = f.Close() }()

	return auth.CacheToken(context.Background(), authCode, f)
}

func main() {
	cfg, err := readConfig("config.yaml")
	if err != nil {
		fatal("error in config parse: %s", err)
	}
	newFilters, err := filter.FromConfig(cfg)
	if err != nil {
		fatal("errors exporting local filters: %s", err)
	}

	gmailapi, err := openAPI()
	if err != nil {
		fatal("cannot connect to Gmail: %s", err)
	}
	upstreamFilters, err := gmailapi.ListFilters()
	if err != nil {
		fatal("cannot get filters from Gmail: %s", err)
	}

	diff, err := filter.Diff(upstreamFilters, newFilters)
	if err != nil {
		fatal("cannot compare upstream with local filters")
	}

	if diff.Empty() {
		fmt.Println("No difference found from upstream.")
		return
	}
	fmt.Printf("You are going to apply the following changes to your settings:\n\n%s", diff)
}
