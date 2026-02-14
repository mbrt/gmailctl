package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	papply "github.com/mbrt/gmailctl/internal/engine/apply"
	"github.com/mbrt/gmailctl/internal/engine/config"
	"github.com/mbrt/gmailctl/internal/engine/rimport"
	"github.com/mbrt/gmailctl/internal/errors"
)

const pullHeader = `// Pulled from Gmail by 'gmailctl pull'.
//
// This file was generated to match your current Gmail filter state.
// You can edit this file and use 'gmailctl apply' to sync changes back.

// Uncomment if you want to use the standard library.
// local lib = import 'gmailctl.libsonnet';
`

var (
	pullFilename      string
	pullYes           bool
	pullForce         bool
	pullMerge         bool
	pullMergeStrategy string
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull Gmail filters to local config file",
	Long: `The pull command pulls the current Gmail filter state and writes it
to the local config file. This is the reverse of the 'apply' command.

By default, pull performs conflict detection to prevent overwriting
local changes. If your local config differs from Gmail, pull will
abort unless you provide the --force flag.

Use the --merge flag to interactively merge Gmail changes into your
local config. Filters that exist only in Gmail or only locally are
automatically merged, while conflicts require user resolution.

By default pull uses the configuration file inside the config
directory [config.jsonnet].`,
	Run: func(*cobra.Command, []string) {
		f := pullFilename
		if f == "" {
			f = configFilenameFromDir(cfgDir)
		}
		if err := pull(f, !pullYes, !pullForce); err != nil {
			fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)

	// Flags and configuration settings
	pullCmd.PersistentFlags().StringVarP(&pullFilename, "filename", "f", "", "config file to write")
	pullCmd.Flags().BoolVarP(&pullYes, "yes", "y", false, "don't ask for confirmation, just pull")
	pullCmd.Flags().BoolVar(&pullForce, "force", false, "overwrite local config even if conflicts detected")
	pullCmd.Flags().BoolVar(&pullMerge, "merge", false, "interactively merge Gmail changes into local config")
	pullCmd.Flags().StringVar(&pullMergeStrategy, "merge-strategy", "",
		"conflict resolution strategy when using --merge: 'local' (keep local), 'gmail' (take Gmail)")
}

func pull(path string, interactive, detectConflict bool) error {
	// Validate merge-strategy flag
	if pullMergeStrategy != "" && !pullMerge {
		return errors.New("--merge-strategy requires --merge flag")
	}
	strategy := MergeStrategy(pullMergeStrategy)
	if strategy != "" && strategy != StrategyLocal && strategy != StrategyGmail {
		return fmt.Errorf("invalid --merge-strategy: %q (valid: local, gmail)", pullMergeStrategy)
	}

	// Phase 1: Fetch upstream state
	gmailapi, err := openAPI()
	if err != nil {
		return configurationError(fmt.Errorf("connecting to Gmail: %w", err))
	}

	upstream, err := upstreamConfig(gmailapi)
	if err != nil {
		return err
	}

	// If merge mode and local file exists, do text-based merge (preserves original structure)
	if pullMerge {
		if _, err := os.Stat(path); err == nil {
			// File exists - do text-based merge
			parseRes, err := parseConfig(path, "", false)
			if err != nil {
				return fmt.Errorf("parsing local config for merge: %w", err)
			}

			// Get local filters and labels from parsed config
			localFilters := parseRes.Res.GmailConfig.Filters
			localLabels := parseRes.Res.GmailConfig.Labels

			// Analyze what needs to be merged (doesn't modify anything)
			mergeResult, err := analyzeForMerge(
				localFilters,
				localLabels,
				upstream.Filters,
				upstream.Labels,
				strategy,
				interactive,
			)
			if err != nil {
				return fmt.Errorf("analyzing merge: %w", err)
			}

			// Nothing to add
			if !mergeResult.HasNewFilters && len(mergeResult.NewLabels) == 0 {
				return nil
			}

			// Read original file as text
			originalContent, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("reading local config: %w", err)
			}

			// Do text-based append (preserves original structure)
			modifiedContent, err := appendRulesToFile(string(originalContent), mergeResult.NewRules, mergeResult.NewLabels)
			if err != nil {
				return fmt.Errorf("appending to file: %w", err)
			}

			// Show what will be added
			if len(mergeResult.NewLabels) > 0 {
				labelsText, _ := labelsToJsonnet(mergeResult.NewLabels)
				fmt.Println("\n--- New labels to append ---")
				fmt.Println(labelsText)
				fmt.Println("----------------------------")
			}

			if len(mergeResult.NewRules) > 0 {
				rulesText, _ := rulesToJsonnet(mergeResult.NewRules)
				fmt.Println("\n--- New rules to append ---")
				fmt.Println(rulesText)
				fmt.Println("----------------------------")
			}

			// Confirm and write
			if interactive {
				fmt.Printf("\nAppending %d new label(s) and %d new rule(s) to: %s\n",
					len(mergeResult.NewLabels), len(mergeResult.NewRules), path)
				if !askYN("Write changes to file?") {
					return nil
				}
			}

			return writeTextAtomic(path, modifiedContent)
		}
		// File doesn't exist - fall through to normal pull
		fmt.Println("No local config to merge into, creating new config...")
	}

	// Phase 2: Conflict detection (unless --force or --merge)
	if detectConflict {
		// Check if local config file exists
		if _, err := os.Stat(path); err == nil {
			// File exists - load local config for comparison
			parseRes, err := parseConfig(path, "", false)
			if err != nil {
				// If parse fails with ErrNotFound (shouldn't happen since Stat succeeded),
				// treat as no conflict. If other parse error, return it.
				if !errors.Is(err, config.ErrNotFound) {
					// Parse error in existing config - allow pull to potentially fix it
					stderrPrintf("Warning: Local config has parse errors: %v\n", err)
					stderrPrintf("  Proceeding with pull to overwrite invalid config.\n")
				}
			} else {
				// Config parsed successfully - check for semantic differences
				diff, err := papply.Diff(parseRes.Res.GmailConfig, upstream, false, 0, false)
				if err != nil {
					return fmt.Errorf("checking for conflicts: %w", err)
				}

				if !diff.Empty() {
					return errors.WithDetails(
						errors.New("conflict detected: local config differs from Gmail"),
						"Your local config has changes that differ from Gmail.\n"+
							"Use 'gmailctl diff' to see the differences.\n"+
							"Use 'gmailctl pull --force' to overwrite local changes.")
				}
			}
		}
		// If file doesn't exist, no conflict - proceed
	}

	// Phase 3: Export Gmail to config format
	cfg, err := rimport.Import(upstream.Filters, upstream.Labels)
	if err != nil {
		return err
	}

	// Phase 4: Show diff and confirm (unless --yes)
	if interactive {
		// Check if file exists to show appropriate message
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("Creating new config file: %s\n", path)
		} else {
			fmt.Printf("Updating config file: %s\n", path)
		}

		if !askYN("Write this configuration to file?") {
			return nil
		}
	}

	// Phase 5: Atomic file write
	err = writeConfigAtomic(path, cfg)
	if err != nil {
		return err
	}

	fmt.Printf("Configuration written to %s\n", path)
	return nil
}

func writeConfigAtomic(path string, cfg interface{}) error {
	// Create temp file in same directory (ensures same filesystem for atomic rename)
	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, ".gmailctl-pull-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Ensure cleanup on error
	defer func() {
		if tmpFile != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	// Write to temp file
	err = rimport.MarshalJsonnet(cfg, tmpFile, pullHeader)
	if err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	// Sync to disk before rename (ensures durability)
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("syncing temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}
	tmpFile = nil // Prevent deferred cleanup

	// Atomic rename (on both Unix and Windows with Go 1.24+)
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}

func writeTextAtomic(path string, content string) error {
	// Create temp file in same directory (ensures same filesystem for atomic rename)
	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, ".gmailctl-pull-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Ensure cleanup on error
	defer func() {
		if tmpFile != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	// Write content to temp file
	if _, err := tmpFile.WriteString(content); err != nil {
		return fmt.Errorf("writing content: %w", err)
	}

	// Sync to disk before rename (ensures durability)
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("syncing temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}
	tmpFile = nil // Prevent deferred cleanup

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("renaming temp file: %w", err)
	}

	fmt.Printf("Configuration updated: %s\n", path)
	return nil
}
