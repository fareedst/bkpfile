// Package main implements the bkpfile command-line interface for creating and managing file backups.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"bkpfile/internal/bkpfile"

	"github.com/spf13/cobra"
)

var (
	// Version information
	version     = "1.1.0"
	compileDate string
	platform    string

	// Global flags
	dryRun bool
	list   bool
	config bool

	// Root command
	rootCmd = &cobra.Command{
		Use:     "bkpfile [FILE_PATH] [NOTE]",
		Short:   "Single file backup CLI application",
		Long:    `A command-line application for creating and managing file backups.`,
		Version: fmt.Sprintf("%s (compiled %s) [%s]", version, compileDate, platform),
		Args:    cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Handle config flag first (exclusive operation)
			if config {
				return bkpfile.DisplayConfig()
			}

			// Require at least one argument for other operations
			if len(args) < 1 {
				return fmt.Errorf("file path is required")
			}

			filePath := args[0]
			note := ""
			if len(args) > 1 {
				note = args[1]
			}

			// Load configuration
			cfg, err := bkpfile.LoadConfig(".")
			if err != nil {
				// Configuration error should use the config error status code
				// Use default config to get the status code since loading failed
				defaultCfg := bkpfile.DefaultConfig()
				formatter := bkpfile.NewOutputFormatter(defaultCfg)
				formatter.PrintError(fmt.Sprintf("failed to load config: %v", err))
				os.Exit(defaultCfg.StatusConfigError)
			}

			// Create formatter with loaded configuration
			formatter := bkpfile.NewOutputFormatter(cfg)

			if list {
				// List backups
				backups, err := bkpfile.ListBackups(cfg.BackupDirPath, filePath)
				if err != nil {
					formatter.PrintError(fmt.Sprintf("failed to list backups: %v", err))
					os.Exit(cfg.StatusConfigError)
				}

				// Display backups
				for _, backup := range backups {
					// Get relative path for display
					relPath, err := filepath.Rel(".", backup.Path)
					if err != nil {
						relPath = backup.Path // Fallback to absolute path if relative path fails
					}
					formatter.PrintListBackup(relPath, backup.CreationTime.Format("2006-01-02 15:04:05"))
				}
				return nil
			}

			// Create backup
			err = bkpfile.CreateBackup(cfg, filePath, note, dryRun)
			if err != nil {
				// Check if this is a success status
				if backupErr, ok := err.(*bkpfile.BackupError); ok {
					isSuccess := backupErr.StatusCode == cfg.StatusCreatedBackup ||
						backupErr.StatusCode == cfg.StatusFileIsIdenticalToExistingBackup
					if isSuccess {
						os.Exit(backupErr.StatusCode)
					} else {
						formatter.PrintError(backupErr.Message)
						os.Exit(backupErr.StatusCode)
					}
				} else {
					formatter.PrintError(err.Error())
					os.Exit(1)
				}
			}
			return nil
		},
	}
)

func init() {
	// Add global flags
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without creating backups")
	rootCmd.PersistentFlags().BoolVar(&list, "list", false, "List all backups for the specified file")
	rootCmd.PersistentFlags().BoolVar(&config, "config", false, "Display computed configuration values and exit")

	// Customize help template to include version
	rootCmd.SetHelpTemplate(`{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}
Version: {{.Version}}
`)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		// Use default config to get status codes since cfg is not in scope here
		defaultCfg := bkpfile.DefaultConfig()
		formatter := bkpfile.NewOutputFormatter(defaultCfg)

		// Check if this is a success status
		if backupErr, ok := err.(*bkpfile.BackupError); ok {
			isSuccess := backupErr.StatusCode == defaultCfg.StatusCreatedBackup ||
				backupErr.StatusCode == defaultCfg.StatusFileIsIdenticalToExistingBackup
			if isSuccess {
				os.Exit(backupErr.StatusCode)
			} else {
				formatter.PrintError(backupErr.Message)
				os.Exit(backupErr.StatusCode)
			}
		} else {
			formatter.PrintError(err.Error())
			os.Exit(1)
		}
	}
}
