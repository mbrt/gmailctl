package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/mbrt/gmailfilter/pkg/api"
)

var (
	initReset bool
)

const (
	credentialsMissingMsg = `The credentials are not initialized.

To do so, head to https://console.developers.google.com
1. Go to 'Enable API and services' and select Gmail;
2. Go to credentials and create a new one, by selecting 'Help me choose'
   2a. Select the Gmail API
   2b. Select 'Other UI'
   2c. Access 'User data'.
3. Download the credentials file into '%s'
   and execute the 'init' command again.

Documentation about Gmail API authorization can be found
at: https://developers.google.com/gmail/api/auth/about-auth
`
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the Gmail configuration",
	Long: `The init command initialize the Gmail configuration, asking
you for details and guiding you through the process of
setting up the API authorizations and initial settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if initReset {
			err = resetConfig()
		} else {
			err = continueConfig()
		}
		if err != nil {
			fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Flags and configuration settings
	rootCmd.Flags().BoolVar(&initReset, "reset", false, "Reset the configuration.")
}

func resetConfig() error {
	return nil
}

func continueConfig() error {
	if err := handleCfgDir(); err != nil {
		return errors.Wrap(err, "error configuring the main config directory")
	}
	auth, err := openCredentials()
	if err != nil {
		stderrPrintf("%v\n\n", err)
		fmt.Printf(credentialsMissingMsg, credentialsPath)
		return nil
	}
	_, err = openToken(auth)
	if err != nil {
		stderrPrintf("%v\n\n", err)
		err = setupToken(auth)
	}
	if err != nil {
		return err
	}

	fmt.Println("\nYou have correctly configured gmailctl to use Gmail APIs.")
	return nil
}

func handleCfgDir() error {
	if _, err := os.Stat(cfgDir); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err = os.MkdirAll(cfgDir, 0700); err != nil {
			return err
		}
	}
	return nil
}

func setupToken(auth api.Authenticator) error {
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\nAuthorization code: ", auth.AuthURL())

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return errors.Wrap(err, "unable to retrieve token from web")
	}

	if err := saveToken(tokenPath, authCode, auth); err != nil {
		return errors.Wrap(err, "unable to cache token")
	}
	return nil
}

func saveToken(path, authCode string, auth api.Authenticator) (err error) {
	fmt.Printf("Saving credential file to %s\n", path)
	f, e := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if e != nil {
		return errors.Wrap(err, "unable create token file")
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
