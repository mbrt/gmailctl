package cmd

import (
	"fmt"
	"os"
	"path"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var (
	cfgDir          string
	credentialsPath string
	tokenPath       string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gmailctl",
	Short: "Declarative configuration for Gmail",
	Long: `Gmailctl is a command line utility that allows you to manage
your Gmail filters in a declarative way, making them easier
to maintain and understand.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgDir, "config", "", "config directory (default is $HOME/.gmailctl)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgDir != "" {
		// Use config file from the flag.
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		cfgDir = path.Join(home, ".gmailctl")
	}
	credentialsPath = path.Join(cfgDir, "credentials.json")
	tokenPath = path.Join(cfgDir, "token.json")
}
