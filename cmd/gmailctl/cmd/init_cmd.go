package cmd

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/mbrt/gmailctl/pkg/api"
)

var (
	initReset bool
)

const (
	credentialsMissingMsg = `The credentials are not initialized.

To do so, head to https://console.developers.google.com

1. Create a new project if you don't have one
1. Go to 'Enable API and services' and select Gmail
2. Go to credentials and create a new one, by selecting 'Help me choose'
   2a. Select the Gmail API
   2b. Select 'Other UI'
   2c. Access 'User data'.
3. Download the credentials file into '%s'
   and execute the 'init' command again.

Documentation about Gmail API authorization can be found
at: https://developers.google.com/gmail/api/auth/about-auth
`

	defaultConfig = `# NOTE: This is a simple example.
# Please refer to https://github.com/mbrt/gmailctl#configuration for docs about
# the config format. Don't forget to change the configuration before to apply it
# to your own inbox!

version: v1alpha2
author:
  # This is optional and used only if you want to export your filters
  # with gmailctl export:
  #
  # name: YOUR NAME HERE
  # email: YOUR.MAIL@gmail.com

filters:
  - name: toMe
    query:
      to: myself+foo@gmail.com

rules:
  - filter:
      from: bar@gmail.com
    actions:
      markImportant: true
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
	initCmd.Flags().BoolVar(&initReset, "reset", false, "Reset the configuration.")
}

func resetConfig() error {
	if err := deleteFile(credentialsPath); err != nil {
		return err
	}
	if err := deleteFile(tokenPath); err != nil {
		return err
	}
	fmt.Println("Configuration reset.")
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

func handleCfgDir() (err error) {
	// Create the config dir
	if _, err := os.Stat(cfgDir); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err = os.MkdirAll(cfgDir, 0700); err != nil {
			return err
		}
	}

	// Create a default config file
	cfgFile := path.Join(cfgDir, "config.yaml")
	if _, err := os.Stat(cfgFile); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		f, err := os.Create(cfgFile)
		if err != nil {
			return err
		}
		defer func() {
			e := f.Close()
			if err == nil {
				err = e
			}
		}()
		_, err = f.WriteString(defaultConfig)
		return err
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

func deleteFile(path string) error {
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}
