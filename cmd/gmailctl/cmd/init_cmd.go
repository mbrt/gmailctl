package cmd

import (
	"context"
	"fmt"
	"os"
	"path"

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
3. Go to 'OAuth consent screen'
   3a. If your account is managed by an organization, you have to
       select 'Internal' as 'User Type' and Create (otherwise ignore)
   3b. Set an application name (e.g. 'gmailctl')
   3c. Update 'Scopes for Google API', by adding:
       * https://www.googleapis.com/auth/gmail.labels
       * https://www.googleapis.com/auth/gmail.metadata
       * https://www.googleapis.com/auth/gmail.settings.basic
5. IMPORTANT: you don't need to submit your changes for verification, as
   you're not creating a public App.
6. Save and go back to Credentials
   6a. Click 'Create credentials'
   6b. Select 'OAuth client ID'
   6c. Select 'Other' as 'Application type' and give it a name.
   6d. Create.
7. Download the credentials file into '%s' and execute the 'init'
   command again.

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
		return fmt.Errorf("configuring the main config directory: %w", err)
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

	// Handle legacy yaml configuration
	if hasYamlConfig(cfgDir) {
		return nil
	}

	// Create default config files
	cfgFile := path.Join(cfgDir, "config.jsonnet")
	if err := createDefault(cfgFile, defaultConfig()); err != nil {
		return err
	}
	libFile := path.Join(cfgDir, "gmailctl.libsonnet")
	return createDefault(libFile, gmailctlLib())
}

func hasYamlConfig(cfgDir string) bool {
	path := path.Join(cfgDir, "config.yaml")
	_, err := os.Stat(path)
	return err == nil
}

func createDefault(path, contents string) (err error) {
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer func() {
			e := f.Close()
			if err == nil {
				err = e
			}
		}()
		_, err = f.WriteString(contents)
		return err
	}
	return nil
}

func setupToken(auth *api.Authenticator) error {
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\nAuthorization code: ", auth.AuthURL())

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return fmt.Errorf("unable to retrieve token from web: %w", err)
	}

	if err := saveToken(tokenPath, authCode, auth); err != nil {
		return fmt.Errorf("caching token: %w", err)
	}
	return nil
}

func saveToken(path, authCode string, auth *api.Authenticator) (err error) {
	fmt.Printf("Saving credential file to %s\n", path)
	f, e := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if e != nil {
		return fmt.Errorf("creating token file: %w", err)
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
