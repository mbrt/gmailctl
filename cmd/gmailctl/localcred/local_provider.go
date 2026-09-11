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
	// Keep in sync with Google's OAuth setup documentation:
	// https://developers.google.com/workspace/guides/configure-oauth-consent
	// https://developers.google.com/workspace/guides/create-credentials#desktop-app
	credentialsMissingMsg = `The credentials are not initialized.

To set them up, open https://console.cloud.google.com in your browser.

1. Create or select a project. Use this project for all the steps below.
2. Go to 'APIs & Services' > 'Library', find 'Gmail API' and enable it.
3. Go to 'Google Auth Platform' > 'Branding'. If it is not configured
   yet, click 'Get started':
    3a. Enter an 'App name' (e.g. 'gmailctl') and your 'User support email'.
    3b. Under 'Audience', choose 'External' for a personal Google account.
        'Internal' is available for projects owned by an organization
        and limits access to accounts in that organization. Otherwise,
        use 'External'.
    3c. Enter your email under 'Contact Information', review the terms
        under 'Finish', then click 'Continue' and 'Create' if you agree.
   If already configured, review these settings in 'Branding' and 'Audience'.
4. Go to 'Data Access' > 'Add or remove scopes' and select:
        * https://www.googleapis.com/auth/gmail.labels
        * https://www.googleapis.com/auth/gmail.settings.basic
   Click 'Update', then 'Save'.
5. For an 'External' audience, go to 'Audience' and choose either:
    * Keep 'Testing' status. Under 'Test users', click 'Add users', add
      the Google account you will use with gmailctl, and save. Access
      expires after seven days; run 'gmailctl init --refresh-expired'
      to authorize again.
    * Click 'Publish app' to switch to 'In production' and avoid the
      seven-day testing expiry. For your own personal use, verification
      is not required, but Google may show an 'unverified app' warning
      when you authorize access.
   Skip this step for an 'Internal' audience.
6. Go to 'Clients' > 'Create client'. Select 'Desktop app' as the
   'Application type', give it a name, and click 'Create'. Download the
   client JSON file before closing the dialog and save it as:
   %q
   Then rerun the same 'gmailctl init' command.

Setup documentation:
https://developers.google.com/workspace/guides/configure-oauth-consent
Personal-use verification exceptions:
https://support.google.com/cloud/answer/13464323
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

func (Provider) InitConfig(cfgDir string, port int) error {
	cpath := credentialsPath(cfgDir)
	tpath := tokenPath(cfgDir)

	auth, err := openCredentials(cpath)
	if err != nil {
		return errors.WithDetails(err,
			fmt.Sprintf(credentialsMissingMsg, cpath))
	}
	_, err = openToken(context.Background(), auth, tpath)
	if err != nil {
		stderrPrintf("%v\n\n", err)
		err = setupToken(auth, tpath, port)
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

func (Provider) RefreshToken(ctx context.Context, cfgDir string, port int) error {
	auth, err := openCredentials(credentialsPath(cfgDir))
	if err != nil {
		return errors.WithDetails(fmt.Errorf("invalid credentials: %w", err),
			"Please run 'gmailctl init' to initialize the credentials.")
	}
	svc, err := openToken(ctx, auth, tokenPath(cfgDir))
	if err != nil {
		return setupToken(auth, tokenPath(cfgDir), port)
	}
	// Check whether the token works by getting a label.
	if _, err := svc.Users.Labels.Get("me", "INBOX").Context(ctx).Do(); err != nil {
		return setupToken(auth, tokenPath(cfgDir), port)
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

func setupToken(auth *api.Authenticator, path string, port int) error {
	localSrv := newOauth2Server(auth.State)
	addr, err := localSrv.Start(port)
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
