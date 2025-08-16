package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var cfgDir string
var colorFlag string

// rootCmd is the command run when executing without subcommands.
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

// RemoveCommand removes a subcommand.
func RemoveCommand(name string) bool {
	for _, c := range rootCmd.Commands() {
		if c.Name() == name {
			rootCmd.RemoveCommand(c)
			return true
		}
	}
	return false
}

// AddCommand adds a subcommand.
func AddCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgDir, "config", "", "config directory (default is $HOME/.gmailctl)")
	rootCmd.PersistentFlags().StringVar(&colorFlag, "color", "auto",
		"whether to enable color output ('always', 'auto' or 'never')")
	rootCmd.PersistentFlags().Lookup("color").NoOptDefVal = "always"
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgDir != "" {
		// Use config file from the flag.
		return
	}
	// Find home directory.
	usr, err := user.Current()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cfgDir = path.Join(usr.HomeDir, ".gmailctl")
}

// shouldUseColorDiff decides, based on the value of the color flag and other
// factors, whether gmailctl should use color output.
func shouldUseColorDiff() bool {
	switch colorFlag {
	case "never":
		return false
	case "auto":
		return os.Getenv("TERM") != "dumb" &&
			(isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()))
	case "always":
		return true
	default:
		fatal(fmt.Errorf("--color must be 'always', 'auto' or 'never', not '%v'", colorFlag))
		return false
	}
}
