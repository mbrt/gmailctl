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
	credentialsMissingMsg = `The credentials are not initialized.

To do so, head to https://console.developers.google.com

1. Create a new project if you don't have one.
1. Go to 'Enable API and services' and select Gmail.
2. Go to credentials and create a new one, by selecting 'Help me choose'.
   2a. Select the Gmail API.
   2b. Select 'Other UI'.
   2c. Access 'User data'.
3. Go to 'OAuth consent screen'.
   3a. If your account is managed by an organization, you have to
       select 'Internal' as 'User Type' and Create (otherwise ignore).
   3b. Set an application name (e.g. 'gmailctl').
   3c. Update 'Scopes for Google API', by adding:
       * https://www.googleapis.com/auth/gmail.labels
       * https://www.googleapis.com/auth/gmail.settings.basic
5. IMPORTANT: you don't need to submit your changes for verification, as
   you're only going to access your own data. Save and 'Go back to
   Dashboard'.
   5a. Make sure that the 'Publishing status' is set to 'In production'.
       If it's set to 'Testing', Publish the app and ignore the
	   verification. Using the testing mode will make your tokens
	   expire every 7 days and require re-authentication.
6. Go back to Credentials.
   6a. Click 'Create credentials'.
   6b. Select 'OAuth client ID'.
   6c. Select 'Desktop app' as 'Application type' and give it a name.
   6d. Create.
7. Download the credentials file into '%s' and execute the 'init'
   command again.

Documentation about Gmail API authorization can be found
at: https://developers.google.com/gmail/api/auth/about-auth
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

	fmt.Printf("Go to the following link in your browser and authorize gmailctl: \n"+
		"%v\n\n"+
		"NOTE: If you are running on a remote machine this will not work.\n"+
		"Please execute this command on your desktop and copy the\n"+
		"credentials to the remote machine.\n", auth.AuthURL("http://"+addr))

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
		// do not hide more important error
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
