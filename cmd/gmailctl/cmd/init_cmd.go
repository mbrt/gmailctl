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
	initUpdateLib      bool
	port               int
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the Gmail configuration",
	Long: `The init command initialize the Gmail configuration, asking
you for details and guiding you through the process of
setting up the API authorizations and initial settings.`,
	Run: func(*cobra.Command, []string) {
		var err error
		if initReset {
			err = resetConfig()
		} else if initRefreshExpired {
			err = refreshToken()
		} else if initUpdateLib {
			err = updateLib()
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
	initCmd.Flags().BoolVar(&initUpdateLib, "update-lib", false, "Update the library file.")
	initCmd.Flags().IntVar(&port, "port", 0, "OAuth server bind port (default is 0 meaning random port)")
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
	if err := APIProvider.InitConfig(cfgDir, port); err != nil {
		return err
	}
	fmt.Println("\nYou have correctly configured gmailctl to use Gmail APIs.")
	return nil
}

func refreshToken() error {
	if rt, ok := APIProvider.(TokenRefresher); ok {
		return rt.RefreshToken(context.Background(), cfgDir, port)
	}
	return nil
}

func handleCfgDir() error {
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

func createDefault(path, contents string) error {
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return createFile(path, contents)
	}
	return nil
}

func updateLib() error {
	if err := handleCfgDir(); err != nil {
		return err
	}
	libFile := path.Join(cfgDir, "gmailctl.libsonnet")
	return createFile(libFile, data.GmailctlLib())
}

func createFile(path, contents string) (err error) {
	var f *os.File
	f, err = os.Create(path)
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
