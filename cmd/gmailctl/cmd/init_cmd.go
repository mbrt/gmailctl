package cmd

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/mbrt/gmailctl/internal/data"
)

var (
	initReset          bool
	initRefreshExpired bool
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
		} else if initRefreshExpired {
			err = refreshToken()
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
	initCmd.Flags().BoolVar(&initRefreshExpired, "refresh-expired", false, "Refresh auth token if expired.")
}

func resetConfig() error {
	if err := APIProvider.ResetConfig(cfgDir); err != nil {
		return err
	}
	fmt.Println("Configuration reset.")
	return nil
}

func continueConfig() error {
	if err := handleCfgDir(); err != nil {
		return fmt.Errorf("configuring the main config directory: %w", err)
	}
	if err := APIProvider.InitConfig(cfgDir); err != nil {
		return err
	}
	fmt.Println("\nYou have correctly configured gmailctl to use Gmail APIs.")
	return nil
}

func refreshToken() error {
	if rt, ok := APIProvider.(TokenRefresher); ok {
		return rt.RefreshToken(context.Background(), cfgDir)
	}
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

	// Create default config files
	cfgFile := path.Join(cfgDir, "config.jsonnet")
	if err := createDefault(cfgFile, data.DefaultConfig()); err != nil {
		return err
	}
	libFile := path.Join(cfgDir, "gmailctl.libsonnet")
	return createDefault(libFile, data.GmailctlLib())
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
