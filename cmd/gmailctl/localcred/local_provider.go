package localcred

import (
	"context"
	"fmt"
	"os"
	"path"

	"google.golang.org/api/gmail/v1"

	"github.com/mbrt/gmailctl/cmd/gmailctl/cmd"
	"github.com/mbrt/gmailctl/internal/engine/api"
	"github.com/mbrt/gmailctl/internal/errors"
)

const (
	// Keep in sync with https://rclone.org/drive/#making-your-own-client-id.
	// They hit the exact same issues with Google Drive.
	credentialsMissingMsg = `The credentials are not initialized.

To do so, head to https://console.developers.google.com

0. Create a new project if you don't have one.
1. Go to 'Enable API and services', search for Gmail and enable it.
2. Go to 'OAuth consent screen'.
    2a. If your account is managed by an organization, you have to
        select 'Internal' as 'User Type'. For individual accounts
        select 'External'.
    2b. Set an application name (e.g. 'gmailctl').
    2c. Use your email for 'User support email' and 'Developer
        contact information'. Save and continue.
    2d. Select 'Add or remove scopes' and add:
        * https://www.googleapis.com/auth/gmail.labels
        * https://www.googleapis.com/auth/gmail.settings.basic
    2e. Save and continue until you're back to the dashboard.
3. You now have a choice. You can either:
    * Click on 'Publish App' and avoid 'Submitting for
      verification'. This will result in scary confirmation
      screens or error messages when you authorize gmailctl with
      your account (but for some users it works), OR
    * You could add your email as 'Test user' and keep the app in
      'Testing' mode. In this case everything will work, but
      you'll have to login and confirm the access every week (token
      expiration).
4.  Go to Credentials on the left.
    4a. Click 'Create credentials'.
    4b. Select 'OAuth client ID'.
    4c. Select 'Desktop app' as 'Application type' and give it a name.
    4d. Create.
5. Download the credentials file into %q and execute the 'init'
   command again.

Documentation about Gmail API authorization can be found
at: https://developers.google.com/gmail/api/auth/about-auth
`
	authMessage = `Go to the following link in your browser and authorize gmailctl:

%v

NOTE that gmailctl runs a webserver on your local machine to
collect the token as returned from Google. This only runs until
the token is saved. If your browser is on another machine
without access to the local network, this will not work.
`
)

// Provider is a GMail credential provider that uses the local filesystem.
type Provider struct{}

func (Provider) Service(ctx context.Context, cfgDir string) (*gmail.Service, error) {
	auth, err := openCredentials(credentialsPath(cfgDir))
	if err != nil {
		return nil, err
	}
	return openToken(ctx, auth, tokenPath(cfgDir))
}

func (Provider) InitConfig(cfgDir string) error {
	cpath := credentialsPath(cfgDir)
	tpath := tokenPath(cfgDir)

	auth, err := openCredentials(cpath)
	if err != nil {
		fmt.Printf(credentialsMissingMsg, cpath)
		return err
	}
	_, err = openToken(context.Background(), auth, tpath)
	if err != nil {
		stderrPrintf("%v\n\n", err)
		err = setupToken(auth, tpath)
	}
	return err
}

func (Provider) ResetConfig(cfgDir string) error {
	if err := deleteFile(credentialsPath(cfgDir)); err != nil {
		return err
	}
	if err := deleteFile(tokenPath(cfgDir)); err != nil {
		return err
	}
	return nil
}

func (Provider) RefreshToken(ctx context.Context, cfgDir string) error {
	auth, err := openCredentials(credentialsPath(cfgDir))
	if err != nil {
		return errors.WithDetails(fmt.Errorf("invalid credentials: %w", err),
			"Please run 'gmailctl init' to initialize the credentials.")
	}
	svc, err := openToken(ctx, auth, tokenPath(cfgDir))
	if err != nil {
		return setupToken(auth, tokenPath(cfgDir))
	}
	// Check whether the token works by getting a label.
	if _, err := svc.Users.Labels.Get("me", "INBOX").Context(ctx).Do(); err != nil {
		return setupToken(auth, tokenPath(cfgDir))
	}
	return nil
}

func openCredentials(path string) (*api.Authenticator, error) {
	cred, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening credentials: %w", err)
	}
	return api.NewAuthenticator(cred)
}

func openToken(ctx context.Context, auth *api.Authenticator, path string) (*gmail.Service, error) {
	token, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("missing or invalid cached token: %w", err)
	}
	return auth.Service(ctx, token)
}

func setupToken(auth *api.Authenticator, path string) error {
	localSrv := newOauth2Server(auth.State)
	addr, err := localSrv.Start()
	if err != nil {
		return errors.WithDetails(fmt.Errorf("starting local server: %w", err),
			"gmailctl requires a temporary local HTTP server for the authentication flow.")
	}
	defer localSrv.Close()

	fmt.Printf(authMessage, auth.AuthURL("http://"+addr))
	authCode := localSrv.WaitForCode()
	if err := saveToken(path, authCode, auth); err != nil {
		return fmt.Errorf("caching token: %w", err)
	}
	return nil
}

func saveToken(path, authCode string, auth *api.Authenticator) (err error) {
	fmt.Printf("Saving credential file to %s\n", path)
	f, e := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if e != nil {
		return fmt.Errorf("creating token file: %w", e)
	}
	defer func() {
		e = f.Close()
		// Do not hide more important errors.
		if err == nil {
			err = e
		}
	}()

	return auth.CacheToken(context.Background(), authCode, f)
}

func credentialsPath(cfgDir string) string {
	return path.Join(cfgDir, "credentials.json")
}

func tokenPath(cfgDir string) string {
	return path.Join(cfgDir, "token.json")
}

func deleteFile(path string) error {
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}

func stderrPrintf(format string, a ...interface{}) {
	/* #nosec */
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}

// Verify that the interface is implemented.
var _ cmd.GmailAPIProvider = Provider{}
